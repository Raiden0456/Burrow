package acp

import (
	"context"
	"os"
	"strings"
	"sync"
	"testing"

	sdk "github.com/coder/acp-go-sdk"

	"github.com/raiden0456/gonteractive/internal/ai"
)

const fakeAgentEnv = "GONTERACTIVE_FAKE_ACP_AGENT"

func TestMain(m *testing.M) {
	if os.Getenv(fakeAgentEnv) == "1" {
		runFakeAgent()
		return
	}
	os.Exit(m.Run())
}

type fakeAgent struct {
	conn *sdk.AgentSideConnection
}

func runFakeAgent() {
	a := &fakeAgent{}
	a.conn = sdk.NewAgentSideConnection(a, os.Stdout, os.Stdin)
	<-a.conn.Done()
}

func (a *fakeAgent) Initialize(context.Context, sdk.InitializeRequest) (sdk.InitializeResponse, error) {
	return sdk.InitializeResponse{ProtocolVersion: sdk.ProtocolVersionNumber, AuthMethods: []sdk.AuthMethod{}}, nil
}

func (a *fakeAgent) NewSession(context.Context, sdk.NewSessionRequest) (sdk.NewSessionResponse, error) {
	return sdk.NewSessionResponse{SessionId: sdk.SessionId("s-" + strings.Repeat("x", 4))}, nil
}

func (a *fakeAgent) Prompt(ctx context.Context, req sdk.PromptRequest) (sdk.PromptResponse, error) {
	text := req.Prompt[0].Text.Text

	perm, err := a.conn.RequestPermission(ctx, sdk.RequestPermissionRequest{
		SessionId: req.SessionId,
		ToolCall:  sdk.ToolCallUpdate{ToolCallId: "call-1"},
		Options: []sdk.PermissionOption{
			{OptionId: "allow", Name: "Allow", Kind: sdk.PermissionOptionKindAllowOnce},
			{OptionId: "reject", Name: "Reject", Kind: sdk.PermissionOptionKindRejectOnce},
		},
	})
	if err != nil {
		return sdk.PromptResponse{}, err
	}
	if perm.Outcome.Selected == nil || perm.Outcome.Selected.OptionId != "reject" {
		a.say(ctx, req.SessionId, "PERMISSION_GRANTED")
		return sdk.PromptResponse{StopReason: sdk.StopReasonEndTurn}, nil
	}

	switch {
	case strings.Contains(text, "<schema>") && strings.Contains(text, "That response is invalid"):
		a.say(ctx, req.SessionId, `{"reply": "fixed"}`)
	case strings.Contains(text, "<schema>"):
		a.say(ctx, req.SessionId, "```json\n{\"wrong\": 1}\n```")
	default:
		for _, chunk := range []string{"Hello", ", ", "world"} {
			a.say(ctx, req.SessionId, chunk)
		}
	}
	return sdk.PromptResponse{StopReason: sdk.StopReasonEndTurn}, nil
}

func (a *fakeAgent) say(ctx context.Context, id sdk.SessionId, text string) {
	a.conn.SessionUpdate(ctx, sdk.SessionNotification{SessionId: id, Update: sdk.UpdateAgentMessageText(text)})
}

func (a *fakeAgent) Authenticate(context.Context, sdk.AuthenticateRequest) (sdk.AuthenticateResponse, error) {
	return sdk.AuthenticateResponse{}, nil
}

func (a *fakeAgent) Logout(context.Context, sdk.LogoutRequest) (sdk.LogoutResponse, error) {
	return sdk.LogoutResponse{}, nil
}

func (a *fakeAgent) Cancel(context.Context, sdk.CancelNotification) error { return nil }

func (a *fakeAgent) CloseSession(context.Context, sdk.CloseSessionRequest) (sdk.CloseSessionResponse, error) {
	return sdk.CloseSessionResponse{}, nil
}

func (a *fakeAgent) ListSessions(context.Context, sdk.ListSessionsRequest) (sdk.ListSessionsResponse, error) {
	return sdk.ListSessionsResponse{}, nil
}

func (a *fakeAgent) ResumeSession(context.Context, sdk.ResumeSessionRequest) (sdk.ResumeSessionResponse, error) {
	return sdk.ResumeSessionResponse{}, nil
}

func (a *fakeAgent) SetSessionConfigOption(context.Context, sdk.SetSessionConfigOptionRequest) (sdk.SetSessionConfigOptionResponse, error) {
	return sdk.SetSessionConfigOptionResponse{}, nil
}

func (a *fakeAgent) SetSessionMode(context.Context, sdk.SetSessionModeRequest) (sdk.SetSessionModeResponse, error) {
	return sdk.SetSessionModeResponse{}, nil
}

func newFakeProvider(t *testing.T) *Provider {
	t.Helper()
	p, err := New(os.Args[0], nil, map[string]string{fakeAgentEnv: "1"})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { p.Close() })
	return p
}

func TestStreamCollectsChunksAndRejectsPermissions(t *testing.T) {
	p := newFakeProvider(t)

	var mu sync.Mutex
	var deltas []string
	out, err := p.Stream(t.Context(), ai.Request{Messages: []ai.Message{{Role: ai.RoleUser, Content: "hi"}}}, func(d string) error {
		mu.Lock()
		defer mu.Unlock()
		deltas = append(deltas, d)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if out != "Hello, world" {
		t.Fatalf("got %q, want %q", out, "Hello, world")
	}
	if strings.Join(deltas, "|") != "Hello|, |world" {
		t.Fatalf("unexpected deltas %q", deltas)
	}
}

func TestGenerateJSONRetriesInvalidOutput(t *testing.T) {
	p := newFakeProvider(t)
	schema := ai.MustSchema("reply", []byte(`{"type":"object","properties":{"reply":{"type":"string"}},"required":["reply"],"additionalProperties":false}`))

	var out struct {
		Reply string `json:"reply"`
	}
	if err := ai.GenerateJSON(t.Context(), p, ai.Request{Messages: []ai.Message{{Role: ai.RoleUser, Content: "ping"}}}, schema, &out); err != nil {
		t.Fatal(err)
	}
	if out.Reply != "fixed" {
		t.Fatalf("got %q, want %q", out.Reply, "fixed")
	}
}

func TestProcessIsReusedAcrossCalls(t *testing.T) {
	p := newFakeProvider(t)
	req := ai.Request{Messages: []ai.Message{{Role: ai.RoleUser, Content: "hi"}}}
	if _, err := ai.Complete(t.Context(), p, req); err != nil {
		t.Fatal(err)
	}
	first := p.proc
	if _, err := ai.Complete(t.Context(), p, req); err != nil {
		t.Fatal(err)
	}
	if p.proc != first {
		t.Fatal("expected the agent process to be reused")
	}
}
