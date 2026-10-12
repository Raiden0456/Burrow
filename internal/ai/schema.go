package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"strings"

	"github.com/santhosh-tekuri/jsonschema/v6"
)

type Schema struct {
	Name     string
	Doc      map[string]any
	raw      string
	compiled *jsonschema.Schema
}

func MustSchema(name string, raw []byte) *Schema {
	var doc map[string]any
	if err := json.Unmarshal(raw, &doc); err != nil {
		panic(fmt.Sprintf("schema %s: %v", name, err))
	}
	inst, err := jsonschema.UnmarshalJSON(bytes.NewReader(raw))
	if err != nil {
		panic(fmt.Sprintf("schema %s: %v", name, err))
	}
	c := jsonschema.NewCompiler()
	if err := c.AddResource(name+".json", inst); err != nil {
		panic(fmt.Sprintf("schema %s: %v", name, err))
	}
	compiled, err := c.Compile(name + ".json")
	if err != nil {
		panic(fmt.Sprintf("schema %s: %v", name, err))
	}
	return &Schema{Name: name, Doc: doc, raw: string(raw), compiled: compiled}
}

func (s *Schema) String() string { return s.raw }

func (s *Schema) Validate(data []byte) error {
	inst, err := jsonschema.UnmarshalJSON(bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("not valid JSON: %w", err)
	}
	return s.compiled.Validate(inst)
}

type Validator interface {
	Validate() error
}

// GenerateJSON asks the provider for a schema-shaped response, decodes it into out and, if out
// implements Validator, checks it. On any failure the model gets one retry with the error attached.
func GenerateJSON(ctx context.Context, p Provider, req Request, schema *Schema, out any) error {
	var lastErr error
	for range 2 {
		text, err := p.Generate(ctx, req, schema)
		if err != nil {
			return err
		}
		if lastErr = decode(text, schema, out); lastErr == nil {
			return nil
		}
		req.Messages = append(slices.Clone(req.Messages),
			Message{Role: RoleAssistant, Content: text},
			Message{Role: RoleUser, Content: fmt.Sprintf("That response is invalid: %v\nReturn only the corrected JSON object.", lastErr)},
		)
	}
	return fmt.Errorf("model returned invalid output: %w", lastErr)
}

func decode(text string, schema *Schema, out any) error {
	data := []byte(extractJSON(text))
	if err := schema.Validate(data); err != nil {
		return err
	}
	if err := json.Unmarshal(data, out); err != nil {
		return err
	}
	if v, ok := out.(Validator); ok {
		return v.Validate()
	}
	return nil
}

func extractJSON(text string) string {
	t := strings.TrimSpace(text)
	start := strings.Index(t, "{")
	end := strings.LastIndex(t, "}")
	if start >= 0 && end > start {
		return t[start : end+1]
	}
	return t
}
