package drill

import (
	"errors"
	"strings"
	"testing"
	"time"
)

const testFile = `package drill

import "testing"

func TestDouble(t *testing.T) {
	if got := Double(21); got != 42 {
		t.Fatalf("Double(21) = %d, want 42", got)
	}
}
`

func newTestRunner(t *testing.T) *Runner {
	t.Helper()
	r, err := NewRunner(t.TempDir())
	if errors.Is(err, ErrNoToolchain) {
		t.Skip("go toolchain not available")
	}
	if err != nil {
		t.Fatal(err)
	}
	return r
}

func TestRunner(t *testing.T) {
	r := newTestRunner(t)
	r.timeout = 20 * time.Second

	tests := []struct {
		name  string
		code  string
		check func(t *testing.T, res RunResult)
	}{
		{"passing", "package drill\n\nfunc Double(n int) int { return n * 2 }\n", func(t *testing.T, res RunResult) {
			if !res.Passed {
				t.Fatalf("expected pass, output:\n%s", res.Output)
			}
		}},
		{"failing", "package drill\n\nfunc Double(n int) int { return n }\n", func(t *testing.T, res RunResult) {
			if res.Passed || res.BuildError || !strings.Contains(res.Output, "want 42") {
				t.Fatalf("expected a test failure, got %+v", res)
			}
		}},
		{"build error", "package drill\n\nfunc Double(n int) int { return \"x\" }\n", func(t *testing.T, res RunResult) {
			if res.Passed || !res.BuildError {
				t.Fatalf("expected a build error, got %+v", res)
			}
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := r.Run(t.Context(), tt.code, testFile)
			if err != nil {
				t.Fatal(err)
			}
			tt.check(t, res)
		})
	}
}

func TestRunnerTimeout(t *testing.T) {
	r := newTestRunner(t)
	r.timeout = 3 * time.Second
	code := "package drill\n\nfunc Double(n int) int { for {} }\n"
	res, err := r.Run(t.Context(), code, testFile)
	if err != nil {
		t.Fatal(err)
	}
	if !res.TimedOut || res.Passed {
		t.Fatalf("expected a timeout, got %+v", res)
	}
}
