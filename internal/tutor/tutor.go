package tutor

import (
	"fmt"
	"strings"

	"github.com/raiden0456/gonteractive/internal/settings"
)

func System(prefs settings.Preferences, role string) string {
	var b strings.Builder
	b.WriteString("You are an expert Go (golang) engineer and a patient, precise teacher. ")
	b.WriteString(role)
	fmt.Fprintf(&b, "\n\nWrite all prose in %s. Keep Go code, identifiers, package names and established technical terms in English. ", prefs.Language)
	b.WriteString("Use Markdown; put Go code in fenced ```go blocks. Target modern Go (1.24+) and idiomatic style.")
	if p := strings.TrimSpace(prefs.Profile); p != "" {
		fmt.Fprintf(&b, "\n\nAbout the learner: %s", p)
	}
	return b.String()
}
