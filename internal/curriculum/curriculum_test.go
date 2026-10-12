package curriculum

import "testing"

func TestCurriculumLoads(t *testing.T) {
	c := MustLoad()
	if len(c.Ordered()) == 0 {
		t.Fatal("no topics loaded")
	}
	for _, slug := range c.Ordered() {
		topic, ok := c.Topic(slug)
		if !ok || topic.ModuleRu == "" || len(topic.Objectives) == 0 {
			t.Fatalf("topic %q is incomplete: %+v", slug, topic)
		}
	}
}
