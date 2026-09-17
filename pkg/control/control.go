package control

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"syscall"
)

const claudeBinary = "claude"

var (
	jobIDPattern        = regexp.MustCompile(`^[0-9a-f]{8}$`)
	backgroundedPattern = regexp.MustCompile(`backgrounded\s*\x{B7}?\s*([0-9a-f]{8})`)
	ansiPattern         = regexp.MustCompile(`\x1b\[[0-9;]*[A-Za-z]`)
)

type Runner interface {
	Run(ctx context.Context, dir string, name string, args ...string) (string, error)
	Start(dir string, name string, args ...string) error
}

type Control struct {
	runner Runner
	getenv func(string) string
	onPath func(string) bool
}

func New(runner Runner, getenv func(string) string, onPath func(string) bool) *Control {
	return &Control{runner: runner, getenv: getenv, onPath: onPath}
}

func Default(terminal string) *Control {
	getenv := func(key string) string {
		if key == "BOTROPOLIS_TERMINAL" && terminal != "" {
			return terminal
		}
		return os.Getenv(key)
	}
	return New(execRunner{}, getenv, func(name string) bool {
		_, err := exec.LookPath(name)
		return err == nil
	})
}

func (c *Control) New(ctx context.Context, dir, prompt string) (string, error) {
	args := []string{"--bg"}
	if prompt != "" {
		args = append(args, prompt)
	}
	return c.background(ctx, dir, args)
}

func (c *Control) Resume(ctx context.Context, dir, sessionID string) (string, error) {
	return c.background(ctx, dir, []string{"--bg", "--resume", sessionID})
}

func (c *Control) background(ctx context.Context, dir string, args []string) (string, error) {
	out, err := c.runner.Run(ctx, dir, claudeBinary, args...)
	if err != nil {
		return "", err
	}
	if m := backgroundedPattern.FindStringSubmatch(ansiPattern.ReplaceAllString(out, "")); m != nil {
		return m[1], nil
	}
	return c.newestJob(ctx, dir)
}

type Agent struct {
	ID        string `json:"id"`
	SessionID string `json:"sessionId"`
	CWD       string `json:"cwd"`
	Kind      string `json:"kind"`
	Status    string `json:"status"`
	State     string `json:"state"`
	PID       int    `json:"pid"`
	StartedAt int64  `json:"startedAt"`
}

type agentJSON = Agent

func (c *Control) Agents(ctx context.Context, all bool) ([]Agent, error) {
	args := []string{"agents", "--json"}
	if all {
		args = []string{"agents", "--all", "--json"}
	}
	out, err := c.runner.Run(ctx, "", claudeBinary, args...)
	if err != nil {
		return nil, err
	}
	var agents []Agent
	if err := json.Unmarshal([]byte(out), &agents); err != nil {
		return nil, fmt.Errorf("claude agents --json is unreadable: %w", err)
	}
	return agents, nil
}

func (c *Control) newestJob(ctx context.Context, dir string) (string, error) {
	out, err := c.runner.Run(ctx, "", claudeBinary, "agents", "--json")
	if err != nil {
		return "", fmt.Errorf("claude started a session but printed no job id, and listing agents failed: %w", err)
	}
	var agents []agentJSON
	if err := json.Unmarshal([]byte(out), &agents); err != nil {
		return "", fmt.Errorf("claude started a session but printed no job id, and agents --json is unreadable: %w", err)
	}
	var newest agentJSON
	for _, a := range agents {
		if a.Kind == "background" && a.CWD == dir && a.ID != "" && a.StartedAt > newest.StartedAt {
			newest = a
		}
	}
	if newest.ID == "" {
		return "", fmt.Errorf("claude started a session in %s but no job id could be found for it", dir)
	}
	return newest.ID, nil
}

func (c *Control) Attach(id string) error {
	jobID, err := JobID(id)
	if err != nil {
		return err
	}
	argv, err := TerminalCommand(c.getenv, c.onPath, []string{claudeBinary, "attach", jobID})
	if err != nil {
		return err
	}
	return c.runner.Start("", argv[0], argv[1:]...)
}

func (c *Control) Stop(ctx context.Context, id string) error {
	return c.simple(ctx, "stop", id)
}

func (c *Control) Remove(ctx context.Context, id string) error {
	return c.simple(ctx, "rm", id)
}

func (c *Control) simple(ctx context.Context, verb, id string) error {
	jobID, err := JobID(id)
	if err != nil {
		return err
	}
	_, err = c.runner.Run(ctx, "", claudeBinary, verb, jobID)
	return err
}

func JobID(id string) (string, error) {
	if len(id) > 8 && id[8] == '-' {
		id = id[:8]
	}
	if !jobIDPattern.MatchString(id) {
		return "", fmt.Errorf("%q is not a job id (eight hex characters) or a session id", id)
	}
	return id, nil
}

type execRunner struct{}

func (execRunner) Run(ctx context.Context, dir string, name string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Dir = dir
	var stderr strings.Builder
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			msg = strings.TrimSpace(string(out))
		}
		if msg != "" {
			return string(out), fmt.Errorf("%s %s: %w: %s", name, strings.Join(args, " "), err, msg)
		}
		return string(out), fmt.Errorf("%s %s: %w", name, strings.Join(args, " "), err)
	}
	return string(out), nil
}

func (execRunner) Start(dir string, name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	cmd.Stdin, cmd.Stdout, cmd.Stderr = nil, nil, nil
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	if err := cmd.Start(); err != nil {
		return err
	}
	go func() { _ = cmd.Wait() }()
	return nil
}

var errNoTerminal = errors.New("no terminal found: set BOTROPOLIS_TERMINAL (a command prefix, e.g. \"wezterm start --\") or TERMINAL")
