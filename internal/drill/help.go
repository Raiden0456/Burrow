package drill

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/raiden0456/gonteractive/internal/ai"
	"github.com/raiden0456/gonteractive/internal/providers"
	"github.com/raiden0456/gonteractive/internal/store/db"
	"github.com/raiden0456/gonteractive/internal/tutor"
)

const helpRole = "You coach a learner through a coding exercise. You never write the full solution."

type Hint struct {
	ID        int64  `json:"id"`
	Content   string `json:"content"`
	CreatedAt int64  `json:"createdAt"`
}

func (s *Service) StreamHelp(ctx context.Context, id int64, code, output string, onDelta func(string) error) (Hint, error) {
	d, err := s.q.GetDrill(ctx, id)
	if err != nil {
		return Hint{}, err
	}
	previous, err := s.q.ListDrillHints(ctx, sql.NullInt64{Int64: id, Valid: true})
	if err != nil {
		return Hint{}, err
	}
	prefs, err := s.settings.Preferences(ctx)
	if err != nil {
		return Hint{}, err
	}
	provider, err := s.providers.For(ctx, providers.TaskChat)
	if err != nil {
		return Hint{}, err
	}

	var b strings.Builder
	fmt.Fprintf(&b, "<task>\n%s\n</task>\n\n<tests>\n%s\n</tests>\n\n<learner_code>\n%s\n</learner_code>\n", d.Description, d.TestCode, code)
	if Mode(d.Mode) == ModeBugHunt {
		b.WriteString("\nThis is a Bug Hunt: coach diagnosis of an existing bug. Start with a small observation or a question about the failing case, then make later hints progressively more specific. Do not reveal the full fix.\n")
	}
	if strings.TrimSpace(output) != "" {
		fmt.Fprintf(&b, "\n<latest_go_test_output>\n%s\n</latest_go_test_output>\n", output)
	}
	for i, h := range previous {
		fmt.Fprintf(&b, "\n<earlier_help n=\"%d\">\n%s\n</earlier_help>\n", i+1, h.Content)
	}
	b.WriteString("\nHelp the learner make the next step: point out what is wrong or missing in their code and what to do next. ")
	b.WriteString("You may show a tiny snippet for one specific line, but never the complete solution. Do not repeat earlier help. Keep it short.")

	content, err := provider.Stream(ctx, ai.Request{
		System:   tutor.System(prefs, helpRole),
		Messages: []ai.Message{{Role: ai.RoleUser, Content: b.String()}},
	}, onDelta)
	if err != nil {
		return Hint{}, err
	}
	row, err := s.q.CreateHint(ctx, db.CreateHintParams{Kind: "drill", DrillID: sql.NullInt64{Int64: id, Valid: true}, Content: content})
	if err != nil {
		return Hint{}, err
	}
	return Hint{ID: row.ID, Content: row.Content, CreatedAt: row.CreatedAt}, nil
}
