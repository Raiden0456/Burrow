package daily

import (
	"context"
	"log/slog"
)

type messages struct {
	dailyTitle string
	dailyBody  string
	testBody   string
}

var catalog = map[string]messages{
	"en": {
		dailyTitle: "Daily Go quiz is ready",
		dailyBody:  "A fresh set of questions is waiting for you.",
		testBody:   "Push notifications work.",
	},
	"ru": {
		dailyTitle: "Ежедневный квиз по Go готов",
		dailyBody:  "Свежая порция вопросов уже ждёт.",
		testBody:   "Пуш-уведомления работают.",
	},
}

func (s *Scheduler) messages(ctx context.Context) messages {
	ui, err := s.settings.Interface(ctx)
	if err != nil {
		slog.Warn("daily: load interface settings", "err", err)
	}
	if m, ok := catalog[ui.Language]; ok {
		return m
	}
	return catalog["en"]
}
