package registry

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/rule"
)

// ourTailwindSettings is the block our three Tailwind repositories write, verbatim.
const ourTailwindSettings = `{
	"entryPoint": "./app/_theme/styles/theme.css",
	"attributes": ["class", "className"],
	"callees": ["mergeClassNames", "createVariantClassNames"],
	"variables": [".*[Cc]lassName$", ".*[Cc]lassNames$"]
}`

func settingsOf(namespaces map[string]string) rule.OptionsBase {
	settings := map[string]json.RawMessage{}
	for namespace, block := range namespaces {
		settings[namespace] = json.RawMessage(block)
	}
	return rule.OptionsBase{ConfigDirectory: "/repo", Settings: settings}
}

func TestCheckSettingsAcceptsWhatTheRunReads(t *testing.T) {
	t.Parallel()
	accepted := map[string]rule.OptionsBase{
		"no settings":                    {ConfigDirectory: "/repo"},
		"our repositories' block":        settingsOf(map[string]string{"better-tailwindcss": ourTailwindSettings}),
		"upstream's other spelling":      settingsOf(map[string]string{"eslint-plugin-better-tailwindcss": `{"callees": ["cn"]}`}),
		"a key only one rule reads":      settingsOf(map[string]string{"better-tailwindcss": `{"order": "desc"}`}),
		"every location key":             settingsOf(map[string]string{"better-tailwindcss": `{"entryPoint": "a.css", "tailwindConfig": "b.css", "cwd": "web"}`}),
		"an empty block":                 settingsOf(map[string]string{"better-tailwindcss": `{}`}),
		"an empty list replacing a kind": settingsOf(map[string]string{"better-tailwindcss": `{"callees": []}`}),
	}
	for name, base := range accepted {
		if err := CheckSettings(base); err != nil {
			t.Errorf("%s was refused: %v", name, err)
		}
	}
}

// Each refusal names what it refuses, so the fix is one edit.
func TestCheckSettingsRefusesWhatWouldBeIgnored(t *testing.T) {
	t.Parallel()
	refused := []struct {
		name     string
		base     rule.OptionsBase
		mentions []string
	}{
		{"a plugin cohere reads no settings for", settingsOf(map[string]string{"react": `{"version": "19"}`}),
			[]string{`"react"`, `"better-tailwindcss"`}},
		{"both spellings at once", settingsOf(map[string]string{
			"better-tailwindcss":               `{}`,
			"eslint-plugin-better-tailwindcss": `{}`,
		}), []string{"both", "keep one"}},
		{"an upstream option not ported yet", settingsOf(map[string]string{"better-tailwindcss": `{"selectors": []}`}),
			[]string{`"selectors"`, "ignored"}},
		{"a key no rule reads", settingsOf(map[string]string{"better-tailwindcss": `{"calees": ["cn"]}`}),
			[]string{`"calees"`}},
		{"a value of the wrong type", settingsOf(map[string]string{"better-tailwindcss": `{"callees": "cn"}`}),
			[]string{"better-tailwindcss/", "callees"}},
		{"a value upstream's schema refuses", settingsOf(map[string]string{"better-tailwindcss": `{"order": "sideways"}`}),
			[]string{"better-tailwindcss/enforce-consistent-class-order", "order"}},
		{"null", settingsOf(map[string]string{"better-tailwindcss": `{"collapse": null}`}),
			[]string{"collapse"}},
	}
	for _, testCase := range refused {
		err := CheckSettings(testCase.base)
		if err == nil {
			t.Errorf("%s loaded", testCase.name)
			continue
		}
		for _, mention := range testCase.mentions {
			if !strings.Contains(err.Error(), mention) {
				t.Errorf("%s: the refusal does not mention %s: %v", testCase.name, mention, err)
			}
		}
	}
}

// The settings reach a rule through the options registry the run builds, merged under the rule's own
// element: the path a real config takes, not the decoder alone.
func TestSettingsReachARuleThroughTheRunsOptions(t *testing.T) {
	t.Parallel()
	base := settingsOf(map[string]string{"better-tailwindcss": `{"order": "desc", "callees": ["mergeClassNames"]}`})
	options := OptionsAt(base)

	bare, err := options.Decode("better-tailwindcss/enforce-consistent-class-order", nil)
	if err != nil {
		t.Fatal(err)
	}
	if encoded, _ := json.Marshal(bare); !strings.Contains(string(encoded), `"order":"desc"`) ||
		!strings.Contains(string(encoded), `"callees":["mergeClassNames"]`) {
		t.Fatalf("a bare severity decoded %s, want settings' order and callees", encoded)
	}

	overridden, err := options.Decode("better-tailwindcss/enforce-consistent-class-order",
		[]json.RawMessage{json.RawMessage(`{"order": "asc"}`)})
	if err != nil {
		t.Fatal(err)
	}
	if encoded, _ := json.Marshal(overridden); !strings.Contains(string(encoded), `"order":"asc"`) ||
		!strings.Contains(string(encoded), `"callees":["mergeClassNames"]`) {
		t.Fatalf("an options element decoded %s, want its own order over settings' and settings' callees kept", encoded)
	}
}
