package demo

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type rpcReply struct {
	ID     json.RawMessage `json:"id"`
	Result json.RawMessage `json:"result"`
	Error  *struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

func converse(t *testing.T, server MCPServer, requests ...string) []rpcReply {
	t.Helper()
	var out bytes.Buffer
	require.NoError(t, ServeMCP(server, strings.NewReader(strings.Join(requests, "\n")+"\n"), &out))
	var replies []rpcReply
	for _, line := range strings.Split(strings.TrimSpace(out.String()), "\n") {
		if line == "" {
			continue
		}
		var r rpcReply
		require.NoError(t, json.Unmarshal([]byte(line), &r))
		replies = append(replies, r)
	}
	return replies
}

func text(t *testing.T, reply rpcReply) string {
	t.Helper()
	var result struct {
		Content []struct {
			Text string `json:"text"`
		} `json:"content"`
	}
	require.NoError(t, json.Unmarshal(reply.Result, &result))
	require.NotEmpty(t, result.Content)
	return result.Content[0].Text
}

func call(id int, tool, args string) string {
	return `{"jsonrpc":"2.0","id":` + string(rune('0'+id)) + `,"method":"tools/call","params":{"name":"` + tool + `","arguments":` + args + `}}`
}

func TestServeMCP(t *testing.T) {
	data := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(data, "tidepool.json"),
		[]byte(`[{"id":"TIDE-12","title":"Leap day","body":"Breaks on Feb 29.","labels":["bug"],"state":"open"}]`), 0o600))
	tracker := Tracker(data)

	t.Run("when a client initializes", func(t *testing.T) {
		replies := converse(t, tracker,
			`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-06-18","capabilities":{},"clientInfo":{"name":"t","version":"0"}}}`,
			`{"jsonrpc":"2.0","method":"notifications/initialized"}`)

		t.Run("it should answer the request and not the notification", func(t *testing.T) {
			assert.Len(t, replies, 1)
		})

		t.Run("it should agree to the client's protocol version", func(t *testing.T) {
			assert.Contains(t, string(replies[0].Result), `"protocolVersion":"2025-06-18"`)
		})
	})

	t.Run("when the client lists the tracker's tools", func(t *testing.T) {
		replies := converse(t, tracker, `{"jsonrpc":"2.0","id":2,"method":"tools/list"}`)
		var result struct {
			Tools []struct {
				Name string `json:"name"`
			} `json:"tools"`
		}
		require.NoError(t, json.Unmarshal(replies[0].Result, &result))
		var names []string
		for _, tool := range result.Tools {
			names = append(names, tool.Name)
		}

		t.Run("it should offer the four issue tools", func(t *testing.T) {
			assert.ElementsMatch(t, []string{"list_issues", "get_issue", "add_comment", "close_issue"}, names)
		})
	})

	t.Run("when the client lists a project's issues", func(t *testing.T) {
		replies := converse(t, tracker, call(3, "list_issues", `{"project":"tidepool"}`))

		t.Run("it should return them", func(t *testing.T) {
			assert.Contains(t, text(t, replies[0]), "TIDE-12")
		})
	})

	t.Run("when the client closes an issue with a comment", func(t *testing.T) {
		closed := Tracker(data)
		converse(t, closed,
			call(4, "add_comment", `{"id":"TIDE-12","body":"Fixed in parse.go."}`),
			call(5, "close_issue", `{"id":"TIDE-12"}`))
		replies := converse(t, Tracker(data), call(6, "get_issue", `{"id":"TIDE-12"}`))

		t.Run("it should still be closed for the next session", func(t *testing.T) {
			assert.Contains(t, text(t, replies[0]), `"state": "closed"`)
		})
	})

	t.Run("when the client calls a tool that does not exist", func(t *testing.T) {
		replies := converse(t, tracker, call(7, "delete_everything", `{}`))

		t.Run("it should answer with an error", func(t *testing.T) {
			assert.NotNil(t, replies[0].Error)
		})
	})

	t.Run("when the client asks the weather for the same place twice", func(t *testing.T) {
		first := converse(t, Weather(), call(8, "forecast", `{"place":"Gull Point"}`))
		second := converse(t, Weather(), call(8, "forecast", `{"place":"Gull Point"}`))

		t.Run("it should give the same forecast, so a re-recording tells the same story", func(t *testing.T) {
			assert.Equal(t, text(t, first[0]), text(t, second[0]))
		})
	})
}
