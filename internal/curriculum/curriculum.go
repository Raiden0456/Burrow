package curriculum

import (
	_ "embed"
	"fmt"
	"net/http"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/raiden0456/gonteractive/internal/httpx"
)

//go:embed topics.yaml
var topicsYAML []byte

type Topic struct {
	Slug         string   `yaml:"slug" json:"slug"`
	Title        string   `yaml:"title" json:"title"`
	TitleRu      string   `yaml:"title_ru" json:"titleRu"`
	Objectives   []string `yaml:"objectives" json:"objectives"`
	ObjectivesRu []string `yaml:"objectives_ru" json:"objectivesRu"`
	Module       string   `yaml:"-" json:"module"`
	ModuleRu     string   `yaml:"-" json:"moduleRu"`
}

type Module struct {
	Slug    string  `yaml:"slug" json:"slug"`
	Title   string  `yaml:"title" json:"title"`
	TitleRu string  `yaml:"title_ru" json:"titleRu"`
	Topics  []Topic `yaml:"topics" json:"topics"`
}

type Curriculum struct {
	Modules []Module
	bySlug  map[string]Topic
	order   []string
}

func MustLoad() *Curriculum {
	var doc struct {
		Modules []Module `yaml:"modules"`
	}
	if err := yaml.Unmarshal(topicsYAML, &doc); err != nil {
		panic(fmt.Sprintf("curriculum: %v", err))
	}
	c := &Curriculum{Modules: doc.Modules, bySlug: map[string]Topic{}}
	for mi := range c.Modules {
		m := &c.Modules[mi]
		if m.TitleRu == "" {
			panic("curriculum: missing title_ru for module " + m.Slug)
		}
		for ti := range m.Topics {
			m.Topics[ti].Module = m.Title
			m.Topics[ti].ModuleRu = m.TitleRu
			t := m.Topics[ti]
			if _, dup := c.bySlug[t.Slug]; dup {
				panic("curriculum: duplicate topic slug " + t.Slug)
			}
			if t.TitleRu == "" || len(t.ObjectivesRu) != len(t.Objectives) {
				panic("curriculum: incomplete Russian translation for topic " + t.Slug)
			}
			c.bySlug[t.Slug] = t
			c.order = append(c.order, t.Slug)
		}
	}
	return c
}

func (c *Curriculum) Topic(slug string) (Topic, bool) {
	t, ok := c.bySlug[slug]
	return t, ok
}

func (c *Curriculum) Lookup(slug string) (Topic, error) {
	t, ok := c.bySlug[slug]
	if !ok {
		return t, httpx.Coded(http.StatusNotFound, "unknown_topic", fmt.Sprintf("unknown topic %q", slug))
	}
	return t, nil
}

func (c *Curriculum) Ordered() []string {
	return c.order
}

func (t Topic) Brief() string {
	var b strings.Builder
	fmt.Fprintf(&b, "Topic: %s (module: %s)\nLearning objectives:\n", t.Title, t.Module)
	for _, o := range t.Objectives {
		fmt.Fprintf(&b, "- %s\n", o)
	}
	return b.String()
}

func (c *Curriculum) Routes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/curriculum", func(w http.ResponseWriter, r *http.Request) {
		httpx.JSON(w, http.StatusOK, c.Modules)
	})
}
