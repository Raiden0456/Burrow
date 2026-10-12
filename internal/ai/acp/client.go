package acp

import (
	"context"
	"errors"
	"sync"

	sdk "github.com/coder/acp-go-sdk"
)

var errUnsupported = errors.New("not supported by gonteractive")

type client struct {
	mu    sync.Mutex
	sinks map[sdk.SessionId]func(string)
}

var _ sdk.Client = (*client)(nil)

func newClient() *client {
	return &client{sinks: map[sdk.SessionId]func(string){}}
}

func (c *client) register(id sdk.SessionId, sink func(string)) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.sinks[id] = sink
}

func (c *client) unregister(id sdk.SessionId) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.sinks, id)
}

func (c *client) SessionUpdate(_ context.Context, n sdk.SessionNotification) error {
	chunk := n.Update.AgentMessageChunk
	if chunk == nil || chunk.Content.Text == nil {
		return nil
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if sink, ok := c.sinks[n.SessionId]; ok {
		sink(chunk.Content.Text.Text)
	}
	return nil
}

func (c *client) RequestPermission(_ context.Context, req sdk.RequestPermissionRequest) (sdk.RequestPermissionResponse, error) {
	for _, opt := range req.Options {
		if opt.Kind == sdk.PermissionOptionKindRejectOnce || opt.Kind == sdk.PermissionOptionKindRejectAlways {
			return sdk.RequestPermissionResponse{Outcome: sdk.NewRequestPermissionOutcomeSelected(opt.OptionId)}, nil
		}
	}
	return sdk.RequestPermissionResponse{Outcome: sdk.NewRequestPermissionOutcomeCancelled()}, nil
}

func (c *client) ReadTextFile(context.Context, sdk.ReadTextFileRequest) (sdk.ReadTextFileResponse, error) {
	return sdk.ReadTextFileResponse{}, errUnsupported
}

func (c *client) WriteTextFile(context.Context, sdk.WriteTextFileRequest) (sdk.WriteTextFileResponse, error) {
	return sdk.WriteTextFileResponse{}, errUnsupported
}

func (c *client) CreateTerminal(context.Context, sdk.CreateTerminalRequest) (sdk.CreateTerminalResponse, error) {
	return sdk.CreateTerminalResponse{}, errUnsupported
}

func (c *client) KillTerminal(context.Context, sdk.KillTerminalRequest) (sdk.KillTerminalResponse, error) {
	return sdk.KillTerminalResponse{}, errUnsupported
}

func (c *client) TerminalOutput(context.Context, sdk.TerminalOutputRequest) (sdk.TerminalOutputResponse, error) {
	return sdk.TerminalOutputResponse{}, errUnsupported
}

func (c *client) ReleaseTerminal(context.Context, sdk.ReleaseTerminalRequest) (sdk.ReleaseTerminalResponse, error) {
	return sdk.ReleaseTerminalResponse{}, errUnsupported
}

func (c *client) WaitForTerminalExit(context.Context, sdk.WaitForTerminalExitRequest) (sdk.WaitForTerminalExitResponse, error) {
	return sdk.WaitForTerminalExitResponse{}, errUnsupported
}
