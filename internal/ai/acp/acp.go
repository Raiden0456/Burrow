package acp

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	sdk "github.com/coder/acp-go-sdk"

	"github.com/raiden0456/gonteractive/internal/ai"
)

const idleTimeout = 5 * time.Minute

type Provider struct {
	command string
	args    []string
	env     []string

	mu     sync.Mutex
	proc   *process
	active int
	idle   *time.Timer
}

func New(command string, args []string, env map[string]string) (*Provider, error) {
	if command == "" {
		return nil, errors.New("acp: command is required")
	}
	p := &Provider{command: command, args: args}
	for k, v := range env {
		p.env = append(p.env, k+"="+v)
	}
	return p, nil
}

func (p *Provider) Stream(ctx context.Context, req ai.Request, onDelta func(string) error) (string, error) {
	return p.prompt(ctx, render(req), onDelta)
}

func (p *Provider) Generate(ctx context.Context, req ai.Request, schema *ai.Schema) (string, error) {
	text := render(req) + "\n\nRespond with a single JSON object that conforms to the JSON Schema below. " +
		"Output only the JSON object: no prose, no markdown, no code fences.\n<schema>\n" + schema.String() + "\n</schema>"
	return p.prompt(ctx, text, func(string) error { return nil })
}

func (p *Provider) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.idle != nil {
		p.idle.Stop()
	}
	if p.proc != nil {
		p.proc.kill()
		p.proc = nil
	}
	return nil
}

func (p *Provider) prompt(ctx context.Context, text string, onDelta func(string) error) (string, error) {
	proc, err := p.acquire(ctx)
	if err != nil {
		return "", err
	}
	defer p.release()

	sess, err := proc.conn.NewSession(ctx, sdk.NewSessionRequest{Cwd: proc.workdir, McpServers: []sdk.McpServer{}})
	if err != nil {
		return "", fmt.Errorf("acp: new session: %w", proc.describe(err))
	}
	defer proc.closeSession(sess.SessionId)

	var (
		out     strings.Builder
		sinkErr error
	)
	proc.client.register(sess.SessionId, func(chunk string) {
		out.WriteString(chunk)
		if sinkErr == nil {
			sinkErr = onDelta(chunk)
		}
	})
	resp, err := proc.conn.Prompt(ctx, sdk.PromptRequest{
		SessionId: sess.SessionId,
		Prompt:    []sdk.ContentBlock{sdk.TextBlock(text)},
	})
	proc.client.unregister(sess.SessionId)

	if ctx.Err() != nil {
		proc.conn.Cancel(context.Background(), sdk.CancelNotification{SessionId: sess.SessionId})
		return "", ctx.Err()
	}
	if err != nil {
		return "", fmt.Errorf("acp: prompt: %w", proc.describe(err))
	}
	if sinkErr != nil {
		return "", sinkErr
	}
	switch resp.StopReason {
	case sdk.StopReasonRefusal:
		return "", ai.ErrRefused
	case sdk.StopReasonCancelled:
		return "", errors.New("acp: prompt was cancelled by the agent")
	}
	return out.String(), nil
}

func (p *Provider) acquire(ctx context.Context) (*process, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.idle != nil {
		p.idle.Stop()
	}
	if p.proc != nil && !p.proc.alive() {
		p.proc.kill()
		p.proc = nil
	}
	if p.proc == nil {
		proc, err := start(ctx, p.command, p.args, p.env)
		if err != nil {
			return nil, err
		}
		p.proc = proc
	}
	p.active++
	return p.proc, nil
}

func (p *Provider) release() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.active--
	if p.active > 0 {
		return
	}
	p.idle = time.AfterFunc(idleTimeout, func() {
		p.mu.Lock()
		defer p.mu.Unlock()
		if p.active == 0 && p.proc != nil {
			p.proc.kill()
			p.proc = nil
		}
	})
}

type process struct {
	cmd      *exec.Cmd
	conn     *sdk.ClientSideConnection
	client   *client
	workdir  string
	stderr   *tailBuffer
	canClose bool
}

func start(ctx context.Context, command string, args, env []string) (*process, error) {
	workdir, err := os.MkdirTemp("", "gonteractive-acp-")
	if err != nil {
		return nil, err
	}
	cmd := exec.Command(command, args...)
	cmd.Dir = workdir
	cmd.Env = append(os.Environ(), env...)
	stderr := &tailBuffer{limit: 4096}
	cmd.Stderr = stderr

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	if err := cmd.Start(); err != nil {
		os.RemoveAll(workdir)
		return nil, fmt.Errorf("acp: start %s: %w", command, err)
	}

	c := newClient()
	proc := &process{cmd: cmd, conn: sdk.NewClientSideConnection(c, stdin, stdout), client: c, workdir: workdir, stderr: stderr}

	initCtx, cancel := context.WithTimeout(ctx, time.Minute)
	defer cancel()
	resp, err := proc.conn.Initialize(initCtx, sdk.InitializeRequest{
		ProtocolVersion:    sdk.ProtocolVersionNumber,
		ClientCapabilities: sdk.ClientCapabilities{},
	})
	if err != nil {
		err = proc.describe(err)
		proc.kill()
		return nil, fmt.Errorf("acp: initialize %s: %w", command, err)
	}
	proc.canClose = resp.AgentCapabilities.SessionCapabilities.Close != nil
	return proc, nil
}

func (p *process) alive() bool {
	select {
	case <-p.conn.Done():
		return false
	default:
		return true
	}
}

func (p *process) closeSession(id sdk.SessionId) {
	if !p.canClose || !p.alive() {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	p.conn.CloseSession(ctx, sdk.CloseSessionRequest{SessionId: id})
}

func (p *process) describe(err error) error {
	var reqErr *sdk.RequestError
	if errors.As(err, &reqErr) {
		if strings.Contains(strings.ToLower(reqErr.Message), "auth") {
			return fmt.Errorf("the agent is not logged in (%s); log in inside the container first", reqErr.Message)
		}
		return fmt.Errorf("%s", reqErr.Message)
	}
	if tail := strings.TrimSpace(p.stderr.String()); tail != "" {
		return fmt.Errorf("%w (agent stderr: %s)", err, tail)
	}
	return err
}

func (p *process) kill() {
	p.cmd.Process.Kill()
	p.cmd.Wait()
	os.RemoveAll(p.workdir)
}

type tailBuffer struct {
	mu    sync.Mutex
	buf   []byte
	limit int
}

func (t *tailBuffer) Write(b []byte) (int, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.buf = append(t.buf, b...)
	if over := len(t.buf) - t.limit; over > 0 {
		t.buf = t.buf[over:]
	}
	return len(b), nil
}

func (t *tailBuffer) String() string {
	t.mu.Lock()
	defer t.mu.Unlock()
	return string(t.buf)
}

func render(req ai.Request) string {
	var b strings.Builder
	b.WriteString("<instructions>\n")
	b.WriteString("You are being used as a plain text-generation backend, not as a coding agent. ")
	b.WriteString("Do not use any tools, do not read or write files and do not run commands. Answer directly in your reply.\n")
	if req.System != "" {
		b.WriteString(req.System)
		b.WriteString("\n")
	}
	b.WriteString("</instructions>\n\n<conversation>\n")
	for _, m := range req.Messages {
		fmt.Fprintf(&b, "<%s>\n%s\n</%s>\n", m.Role, m.Content, m.Role)
	}
	b.WriteString("</conversation>\n\nWrite the next assistant reply to the last user message.")
	return b.String()
}
