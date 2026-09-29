## What the data can tell us

| Source | Gives | Notes |
| --- | --- | --- |
| `~/.claude/sessions/<pid>.json` | pid, session id, cwd, `kind` (interactive/background), `status`, `name`, start time | probe the pid; stale files outlive it |
| `claude agents --json [--all]` | the same for daemon-managed sessions, including finished ones | the supported surface for background sessions |
| `~/.claude/projects/<cwd>/<sid>.jsonl` | title, cwd, branch, model, effort, per-message `usage`, tool calls, MCP/skill attribution, compactions, PR links, file-history paths, whose turn it is | tail for state, head for metadata, full scan for totals |
| `~/.claude/projects/<cwd>/<sid>/subagents/**` | subagent and workflow transcripts with their own `usage` | ~25% of spend on this machine |
| `~/.claude/stats-cache.json` | per-day activity and per-model token totals | cheap daily rollups |
| `~/.claude.json` | MCP servers configured per project | the towers on the map |
| `~/.claude/teams/` | team rosters, to route `SendMessage` traffic into roads | done |
| `~/.claude/tasks/`, `~/.claude/plans/` | task lists, plans | later: on this machine every `tasks/session-*` dir is empty and plans are slug-named with no session link, so there is nothing to draw yet |
| Hooks | `SessionStart`, `SessionEnd`, `Stop`, `PreToolUse`, `PostToolUse`, `SubagentStart`, `SubagentStop`, `Notification` | real-time push; no polling |
