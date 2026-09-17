package claude

import (
	"encoding/json"
	"errors"
	"os"
	"sort"
)

type MCPServer struct {
	Name    string
	Type    string
	Command string
	Args    []string
	URL     string
}

type MCPConfig struct {
	Global   []MCPServer
	Projects map[string][]MCPServer
}

type mcpServerJSON struct {
	Type    string   `json:"type"`
	Command string   `json:"command"`
	Args    []string `json:"args"`
	URL     string   `json:"url"`
}

type mcpConfigJSON struct {
	Servers  map[string]mcpServerJSON `json:"mcpServers"`
	Projects map[string]struct {
		Servers map[string]mcpServerJSON `json:"mcpServers"`
	} `json:"projects"`
}

func ReadMCPConfig(path string) (MCPConfig, error) {
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return MCPConfig{}, nil
	}
	if err != nil {
		return MCPConfig{}, err
	}
	var raw mcpConfigJSON
	if err := json.Unmarshal(data, &raw); err != nil {
		return MCPConfig{}, err
	}
	config := MCPConfig{Global: mcpServers(raw.Servers)}
	for cwd, project := range raw.Projects {
		if len(project.Servers) == 0 {
			continue
		}
		if config.Projects == nil {
			config.Projects = map[string][]MCPServer{}
		}
		config.Projects[cwd] = mcpServers(project.Servers)
	}
	return config, nil
}

func (c MCPConfig) ForProject(cwd string) []MCPServer {
	byName := map[string]MCPServer{}
	for _, s := range c.Global {
		byName[s.Name] = s
	}
	for _, s := range c.Projects[cwd] {
		byName[s.Name] = s
	}
	return sortedServers(byName)
}

func mcpServers(raw map[string]mcpServerJSON) []MCPServer {
	byName := map[string]MCPServer{}
	for name, s := range raw {
		byName[name] = MCPServer{
			Name:    name,
			Type:    mcpType(s),
			Command: s.Command,
			Args:    s.Args,
			URL:     s.URL,
		}
	}
	return sortedServers(byName)
}

func mcpType(s mcpServerJSON) string {
	switch {
	case s.Type != "":
		return s.Type
	case s.Command != "":
		return "stdio"
	case s.URL != "":
		return "http"
	}
	return ""
}

func sortedServers(byName map[string]MCPServer) []MCPServer {
	servers := make([]MCPServer, 0, len(byName))
	for _, s := range byName {
		servers = append(servers, s)
	}
	sort.Slice(servers, func(i, j int) bool { return servers[i].Name < servers[j].Name })
	return servers
}
