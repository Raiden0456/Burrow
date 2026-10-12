package drill

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/raiden0456/gonteractive/internal/httpx"
)

const (
	runTimeout     = 30 * time.Second
	maxOutputBytes = 64 << 10
	maxConcurrent  = 2
)

var ErrNoToolchain = &httpx.Error{
	Status:  http.StatusServiceUnavailable,
	Code:    "no_toolchain",
	Message: "the Go toolchain is not available on the server, so drills cannot run",
}

type RunResult struct {
	Passed     bool   `json:"passed"`
	BuildError bool   `json:"buildError"`
	TimedOut   bool   `json:"timedOut"`
	Output     string `json:"output"`
	DurationMs int64  `json:"durationMs"`
}

type Runner struct {
	goBin     string
	goVersion string
	cacheDir  string
	timeout   time.Duration
	slots     chan struct{}
}

func NewRunner(cacheDir string) (*Runner, error) {
	goBin, err := exec.LookPath("go")
	if err != nil {
		return nil, ErrNoToolchain
	}
	out, err := exec.Command(goBin, "env", "GOVERSION").Output()
	if err != nil {
		return nil, fmt.Errorf("go env GOVERSION: %w", err)
	}
	version := strings.TrimPrefix(strings.TrimSpace(string(out)), "go")
	if parts := strings.SplitN(version, ".", 3); len(parts) >= 2 {
		version = parts[0] + "." + parts[1]
	}
	if err := os.MkdirAll(cacheDir, 0o755); err != nil {
		return nil, err
	}
	return &Runner{goBin: goBin, goVersion: version, cacheDir: cacheDir, timeout: runTimeout, slots: make(chan struct{}, maxConcurrent)}, nil
}

// Run executes `go test` for code against tests in a throwaway module (package drill, stdlib only,
// no network). A non-nil error means the runner itself failed; failing tests are reported in RunResult.
func (r *Runner) Run(ctx context.Context, code, tests string) (RunResult, error) {
	select {
	case r.slots <- struct{}{}:
		defer func() { <-r.slots }()
	case <-ctx.Done():
		return RunResult{}, ctx.Err()
	}

	dir, err := os.MkdirTemp("", "gonteractive-drill-")
	if err != nil {
		return RunResult{}, err
	}
	defer os.RemoveAll(dir)

	files := map[string]string{
		"go.mod":        fmt.Sprintf("module drill\n\ngo %s\n", r.goVersion),
		"drill.go":      code,
		"drill_test.go": tests,
	}
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
			return RunResult{}, err
		}
	}

	runCtx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	cmd := exec.CommandContext(runCtx, r.goBin, "test", "-count=1", "-v", ".")
	cmd.Dir = dir
	cmd.Env = []string{
		"PATH=" + os.Getenv("PATH"),
		"HOME=" + dir,
		"GOCACHE=" + r.cacheDir,
		"GOPATH=" + filepath.Join(dir, "gopath"),
		"GOPROXY=off",
		"GOFLAGS=-mod=mod",
		"GOTOOLCHAIN=local",
		"GOTELEMETRY=off",
		"CGO_ENABLED=0",
	}
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Cancel = func() error { return syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL) }
	cmd.WaitDelay = 2 * time.Second

	var out bytes.Buffer
	cmd.Stdout = &limitedWriter{buf: &out, limit: maxOutputBytes}
	cmd.Stderr = cmd.Stdout

	start := time.Now()
	runErr := cmd.Run()
	res := RunResult{
		Output:     strings.ReplaceAll(out.String(), dir+string(filepath.Separator), ""),
		DurationMs: time.Since(start).Milliseconds(),
	}

	if ctx.Err() != nil {
		return RunResult{}, ctx.Err()
	}
	var exitErr *exec.ExitError
	switch {
	case runErr == nil:
		res.Passed = true
	case errors.Is(runCtx.Err(), context.DeadlineExceeded):
		res.TimedOut = true
		res.Output += fmt.Sprintf("\n\ntimed out after %s", r.timeout)
	case errors.As(runErr, &exitErr):
		res.BuildError = strings.Contains(res.Output, "[build failed]") || strings.Contains(res.Output, "[setup failed]")
	default:
		return RunResult{}, fmt.Errorf("run go test: %w", runErr)
	}
	return res, nil
}

type limitedWriter struct {
	buf       *bytes.Buffer
	limit     int
	truncated bool
}

func (w *limitedWriter) Write(p []byte) (int, error) {
	if room := w.limit - w.buf.Len(); room > 0 {
		w.buf.Write(p[:min(len(p), room)])
	} else if !w.truncated {
		w.truncated = true
		w.buf.WriteString("\n… output truncated")
	}
	return len(p), nil
}
