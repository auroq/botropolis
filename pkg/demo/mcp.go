package demo

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"hash/fnv"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
)

// MCPServer is a fictional service for recorded sessions to call, so
// the city has radio towers with real traffic on them. It speaks just
// enough of the Model Context Protocol over stdio for Claude Code:
// initialize, tools/list and tools/call.
type MCPServer struct {
	Name  string
	Tools []MCPTool
}

type MCPTool struct {
	Name        string
	Description string
	Params      map[string]string
	Call        func(args map[string]string) (string, error)
}

type rpcRequest struct {
	ID     json.RawMessage `json:"id"`
	Method string          `json:"method"`
	Params json.RawMessage `json:"params"`
}

// ServeMCP answers newline-delimited JSON-RPC requests until in ends.
// Notifications, which carry no id, are read and not answered.
func ServeMCP(server MCPServer, in io.Reader, out io.Writer) error {
	scan := bufio.NewScanner(in)
	scan.Buffer(make([]byte, 1<<20), 16<<20)
	enc := json.NewEncoder(out)
	for scan.Scan() {
		var req rpcRequest
		if err := json.Unmarshal(scan.Bytes(), &req); err != nil {
			continue
		}
		if len(req.ID) == 0 {
			continue
		}
		result, rpcErr := server.handle(req)
		reply := map[string]any{"jsonrpc": "2.0", "id": req.ID}
		if rpcErr != nil {
			reply["error"] = map[string]any{"code": -32601, "message": rpcErr.Error()}
		} else {
			reply["result"] = result
		}
		if err := enc.Encode(reply); err != nil {
			return err
		}
	}
	return scan.Err()
}

func (s MCPServer) handle(req rpcRequest) (any, error) {
	switch req.Method {
	case "initialize":
		var p struct {
			ProtocolVersion string `json:"protocolVersion"`
		}
		_ = json.Unmarshal(req.Params, &p)
		return map[string]any{
			"protocolVersion": p.ProtocolVersion,
			"capabilities":    map[string]any{"tools": map[string]any{}},
			"serverInfo":      map[string]any{"name": s.Name, "version": "1.0.0"},
		}, nil
	case "ping":
		return map[string]any{}, nil
	case "tools/list":
		tools := make([]any, 0, len(s.Tools))
		for _, t := range s.Tools {
			props := map[string]any{}
			required := []string{}
			for name, desc := range t.Params {
				props[name] = map[string]any{"type": "string", "description": desc}
				required = append(required, name)
			}
			sort.Strings(required)
			tools = append(tools, map[string]any{
				"name": t.Name, "description": t.Description,
				"inputSchema": map[string]any{"type": "object", "properties": props, "required": required},
			})
		}
		return map[string]any{"tools": tools}, nil
	case "tools/call":
		var p struct {
			Name      string         `json:"name"`
			Arguments map[string]any `json:"arguments"`
		}
		if err := json.Unmarshal(req.Params, &p); err != nil {
			return nil, err
		}
		for _, t := range s.Tools {
			if t.Name != p.Name {
				continue
			}
			args := map[string]string{}
			for k, v := range p.Arguments {
				args[k] = fmt.Sprint(v)
			}
			text, err := t.Call(args)
			if err != nil {
				return map[string]any{"content": []any{map[string]any{"type": "text", "text": err.Error()}}, "isError": true}, nil
			}
			return map[string]any{"content": []any{map[string]any{"type": "text", "text": text}}}, nil
		}
		return nil, fmt.Errorf("no tool %q", p.Name)
	}
	return nil, fmt.Errorf("no method %q", req.Method)
}

// Issue is one ticket in the fictional tracker.
type Issue struct {
	ID       string    `json:"id"`
	Title    string    `json:"title"`
	Body     string    `json:"body"`
	Labels   []string  `json:"labels"`
	State    string    `json:"state"`
	Comments []Comment `json:"comments,omitempty"`
}

type Comment struct {
	Body string `json:"body"`
}

// Tracker is an issue tracker over <dir>/<project>.json. Changes are
// written back, so an issue one recorded session closes is closed for
// the next.
func Tracker(dir string) MCPServer {
	var mu sync.Mutex
	load := func(project string) ([]Issue, error) {
		data, err := os.ReadFile(filepath.Join(dir, project+".json"))
		if err != nil {
			return nil, fmt.Errorf("no project %q in the tracker", project)
		}
		var issues []Issue
		return issues, json.Unmarshal(data, &issues)
	}
	save := func(project string, issues []Issue) error {
		data, err := json.MarshalIndent(issues, "", "  ")
		if err != nil {
			return err
		}
		return os.WriteFile(filepath.Join(dir, project+".json"), append(data, '\n'), 0o644)
	}
	find := func(id string) (string, []Issue, int, error) {
		paths, _ := filepath.Glob(filepath.Join(dir, "*.json"))
		for _, p := range paths {
			project := strings.TrimSuffix(filepath.Base(p), ".json")
			issues, err := load(project)
			if err != nil {
				continue
			}
			for i, issue := range issues {
				if strings.EqualFold(issue.ID, id) {
					return project, issues, i, nil
				}
			}
		}
		return "", nil, 0, fmt.Errorf("no issue %q", id)
	}
	pretty := func(v any) (string, error) {
		data, err := json.MarshalIndent(v, "", "  ")
		return string(data), err
	}
	change := func(id string, apply func(*Issue)) (string, error) {
		mu.Lock()
		defer mu.Unlock()
		project, issues, i, err := find(id)
		if err != nil {
			return "", err
		}
		apply(&issues[i])
		if err := save(project, issues); err != nil {
			return "", err
		}
		return pretty(issues[i])
	}
	return MCPServer{Name: "tracker", Tools: []MCPTool{
		{
			Name: "list_issues", Description: "List a project's issues with their state and labels.",
			Params: map[string]string{"project": "the project's name, e.g. tidepool"},
			Call: func(a map[string]string) (string, error) {
				mu.Lock()
				defer mu.Unlock()
				issues, err := load(a["project"])
				if err != nil {
					return "", err
				}
				return pretty(issues)
			},
		},
		{
			Name: "get_issue", Description: "Read one issue, with its comments.",
			Params: map[string]string{"id": "the issue id, e.g. TIDE-12"},
			Call: func(a map[string]string) (string, error) {
				mu.Lock()
				defer mu.Unlock()
				_, issues, i, err := find(a["id"])
				if err != nil {
					return "", err
				}
				return pretty(issues[i])
			},
		},
		{
			Name: "add_comment", Description: "Comment on an issue.",
			Params: map[string]string{"id": "the issue id", "body": "the comment"},
			Call: func(a map[string]string) (string, error) {
				return change(a["id"], func(i *Issue) { i.Comments = append(i.Comments, Comment{Body: a["body"]}) })
			},
		},
		{
			Name: "close_issue", Description: "Close an issue.",
			Params: map[string]string{"id": "the issue id"},
			Call: func(a map[string]string) (string, error) {
				return change(a["id"], func(i *Issue) { i.State = "closed" })
			},
		},
	}}
}

// Weather forecasts coastal conditions for a named place. The forecast
// is a hash of the name, so the same place always gets the same
// weather and a re-recording tells the same story.
func Weather() MCPServer {
	winds := []string{"N", "NE", "E", "SE", "S", "SW", "W", "NW"}
	seas := []string{"calm", "slight", "moderate", "rough", "very rough"}
	return MCPServer{Name: "weather", Tools: []MCPTool{{
		Name: "forecast", Description: "Today's wind, visibility and sea state for a coastal place.",
		Params: map[string]string{"place": "the place's name, e.g. Gull Point"},
		Call: func(a map[string]string) (string, error) {
			place := strings.TrimSpace(a["place"])
			if place == "" {
				return "", errors.New("no place given")
			}
			h := fnv.New32a()
			_, _ = h.Write([]byte(strings.ToLower(place)))
			n := h.Sum32()
			data, err := json.MarshalIndent(map[string]any{
				"place":         place,
				"wind":          fmt.Sprintf("%s %d kn", winds[n%8], 5+n/8%30),
				"visibility_nm": 1 + n/256%12,
				"sea_state":     seas[n/4096%5],
			}, "", "  ")
			return string(data), err
		},
	}}}
}
