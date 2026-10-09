package registry

import (
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/system-inc/cohere/internal/lint/rule"
)

// tailwindPortedOptions are the better-tailwindcss options cohere honors, by rule, keyed by the option
// key a sample sets. The options common to every rule are listed under "*". An option joins this list
// in the commit that ports it, and not before: an option accepted and ignored is worse than one
// refused, because the config loads and the setting silently does nothing (#gj5nm6e).
var tailwindPortedOptions = map[string][]string{
	"*": {"entryPoint", "tailwindConfig", "cwd"},
	"better-tailwindcss/enforce-canonical-classes":      {"collapse", "logical"},
	"better-tailwindcss/enforce-consistent-class-order": {"order", "unknownClassOrder", "unknownClassPosition", "componentClassOrder", "componentClassPosition"},
	"better-tailwindcss/no-unnecessary-whitespace":      {"allowMultiline"},
}

// tailwindAcceptedBefore are the keys cohere's decoders took before any of #gj5nm6e, which a sample
// may set beside a ported one.
var tailwindAcceptedBefore = map[string][]string{
	"*": {"attributes", "callees", "variables"},
	"better-tailwindcss/enforce-canonical-classes":             {"ignore"},
	"better-tailwindcss/no-unknown-classes":                    {"ignore"},
	"better-tailwindcss/enforce-consistent-important-position": {"position"},
	"better-tailwindcss/enforce-consistent-variable-syntax":    {"syntax"},
}

// TestTailwindForwardSamplesLoadOncePorted runs every better-tailwindcss option list that upstream
// 4.7.0's schema accepts (#pd2chkx's forward sweep, which found cohere refusing all of them) through
// cohere's real decoders. A list whose every key is ported (or was accepted before) must load; a list
// setting any option not yet ported must still be refused. The second half is the point as much as the first: it is what stops
// an option from being accepted before it does anything.
func TestTailwindForwardSamplesLoadOncePorted(t *testing.T) {
	t.Parallel()

	data, err := os.ReadFile(filepath.Join("testdata", "better-tailwindcss-forward-samples.json"))
	if err != nil {
		t.Fatal(err)
	}
	var file struct {
		Samples []struct {
			Rule     string            `json:"rule"`
			Elements []json.RawMessage `json:"elements"`
		} `json:"samples"`
	}
	if err := json.Unmarshal(data, &file); err != nil {
		t.Fatal(err)
	}
	if len(file.Samples) == 0 {
		t.Fatal("no samples were read, so this test would prove nothing")
	}

	options := OptionsAt(rule.OptionsBase{ConfigDirectory: t.TempDir(), ProjectRoot: t.TempDir()})
	loaded, refused := 0, 0
	for _, sample := range file.Samples {
		var element map[string]json.RawMessage
		if len(sample.Elements) != 1 || json.Unmarshal(sample.Elements[0], &element) != nil {
			t.Fatalf("%s %s: each sample is one object element", sample.Rule, sample.Elements)
		}
		described := sample.Rule + " " + string(compactJSON(t, sample.Elements[0]))

		ported := true
		for key := range element {
			ported = ported && isTailwindPorted(sample.Rule, key)
		}
		_, decodeError := options.Decode(sample.Rule, sample.Elements)
		switch {
		case ported && decodeError != nil:
			t.Errorf("%s: ported, and refused: %v", described, decodeError)
		case !ported && decodeError == nil:
			t.Errorf("%s: not ported, and loaded; it must stay refused until it does what it says", described)
		case ported:
			loaded++
		default:
			refused++
		}
	}
	t.Logf("%d samples load, %d stay refused until their options are ported", loaded, refused)
	if loaded == 0 {
		t.Fatal("no sample loaded, so the ported list is not being exercised")
	}
}

// isTailwindPorted reports whether one rule honors one option key, ported or accepted before.
func isTailwindPorted(ruleName string, key string) bool {
	for _, keys := range []map[string][]string{tailwindPortedOptions, tailwindAcceptedBefore} {
		if slices.Contains(keys["*"], key) || slices.Contains(keys[ruleName], key) {
			return true
		}
	}
	return false
}

// compactJSON is a sample's element without whitespace, so a refused value is named one way.
func compactJSON(t *testing.T, raw json.RawMessage) []byte {
	t.Helper()
	var value any
	if err := json.Unmarshal(raw, &value); err != nil {
		t.Fatal(err)
	}
	encoded, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	return encoded
}
