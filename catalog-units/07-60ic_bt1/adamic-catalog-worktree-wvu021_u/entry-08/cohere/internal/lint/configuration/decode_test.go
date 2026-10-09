package configuration

import (
	"encoding/json"
	"strings"
	"testing"
)

type guardOptions struct {
	LibraryDirectory string   `json:"libraryDirectory"`
	Allowed          []string `json:"allowed"`
}

type tuningOptions struct {
	MaximumLineCount int `json:"maximumLineCount"`
}

// TestRequiredOptionsFailLoudlyWhenAbsent is the whole point of this file.
//
// boundary-no-project-import was enabled and inert for months under the gate cohere replaces: it
// declines every file when its LibraryDirectory is empty, which is correct behavior for a
// misconfigured guard and completely indistinguishable from a rule with nothing to report. A
// liveness harness reporting `fixtures=54 live=53 dead=1` was the only thing that ever noticed.
func TestRequiredOptionsFailLoudlyWhenAbsent(t *testing.T) {
	t.Parallel()
	registry := OptionsRegistry{
		"guard": {Decode: DecodeInto[guardOptions](), Required: true},
	}

	if _, err := registry.Decode("guard", nil); err == nil {
		t.Fatal("a required option was absent and no error was raised, so the rule would decline every file in silence")
	}

	decoded, err := registry.Decode("guard", []json.RawMessage{json.RawMessage(`{"libraryDirectory":"/libraries/structure/"}`)})
	if err != nil {
		t.Fatalf("valid options failed to decode: %v", err)
	}
	typed, isTyped := decoded.(guardOptions)
	if !isTyped {
		t.Fatalf("options decoded to %T rather than the rule's own type, which fails its type assertion and makes it decline every file", decoded)
	}
	if typed.LibraryDirectory != "/libraries/structure/" {
		t.Fatalf("the decoded value is wrong: %+v", typed)
	}
}

// TestTuningOptionsFallBackToDefaults covers the other half. A rule whose options only adjust it
// must run on its own defaults rather than fail, or every unconfigured tunable becomes a hard error.
func TestTuningOptionsFallBackToDefaults(t *testing.T) {
	t.Parallel()
	registry := OptionsRegistry{
		"tunable": {Decode: DecodeInto[tuningOptions]()},
	}

	decoded, err := registry.Decode("tunable", nil)
	if err != nil {
		t.Fatalf("an absent optional option was treated as fatal: %v", err)
	}
	if decoded != nil {
		t.Fatalf("expected nil so the rule uses its own defaults, got %+v", decoded)
	}
}

// TestARuleWithNoDecoderGetsNil keeps the common case cheap: most rules take no options at all.
func TestARuleWithNoDecoderGetsNil(t *testing.T) {
	t.Parallel()
	decoded, err := OptionsRegistry{}.Decode("plain", nil)
	if err != nil {
		t.Fatalf("a rule with no decoder errored on a bare severity: %v", err)
	}
	if decoded != nil {
		t.Fatalf("expected nil for a rule that takes no options, got %+v", decoded)
	}
}

// TestARuleWithNoDecoderRefusesAnOption is the other half. This used to answer nil for
// `{"ignored":true}`, and the key name was the defect stated as a fixture: an option written for a
// rule that reads none was accepted and had no effect.
func TestARuleWithNoDecoderRefusesAnOption(t *testing.T) {
	t.Parallel()
	_, err := OptionsRegistry{}.Decode("plain", []json.RawMessage{json.RawMessage(`{"ignored":true}`)})
	if err == nil {
		t.Fatal("an option given to a rule that takes none was accepted and would never be read")
	}
	for _, want := range []string{"plain", `element 1 {"ignored":true}`} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("the refusal does not name %q, so the author cannot tell what to remove: %v", want, err)
		}
	}
}

/*
 * A second element given to a rule that takes one is refused, naming the rule and the element.
 *
 * This is the guard for every rule with a single-element decoder, which is most of them. Before it,
 * `["error", {...}, {...}]` handed the decoder the first object, dropped the second, loaded clean and
 * ran. The guard lives in the config layer rather than in each decoder because a decoder never sees
 * the list it was sliced from, which is the boundary PortingARule.md section 7c is about.
 */
func TestASecondElementOnASingleElementRuleIsRefused(t *testing.T) {
	t.Parallel()
	registry := OptionsRegistry{"tunable": {Decode: DecodeInto[tuningOptions]()}}

	_, err := registry.Decode("tunable", []json.RawMessage{
		json.RawMessage(`{"maximumLineCount": 3}`),
		json.RawMessage(`{"enforceForRenamedProperties": true}`),
	})
	if err == nil {
		t.Fatal("a second option element was accepted by a rule that reads one, so it was dropped in silence")
	}
	for _, want := range []string{"tunable", `element 2 {"enforceForRenamedProperties": true}`} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("the refusal does not name %q: %v", want, err)
		}
	}

	// The control: the same rule with only its first element decodes, so the refusal above is about
	// the count and not about the first element.
	decoded, err := registry.Decode("tunable", []json.RawMessage{json.RawMessage(`{"maximumLineCount": 3}`)})
	if err != nil {
		t.Fatalf("one element on a one-element rule was refused: %v", err)
	}
	if decoded.(tuningOptions).MaximumLineCount != 3 {
		t.Fatalf("the single element did not reach the decoder: %+v", decoded)
	}
}

// TestAListRuleReceivesEveryElement pins what DecodeList is handed: every element, as one JSON
// array, in the order written. And nothing at all for a bare severity, so the rule's own defaults
// apply.
func TestAListRuleReceivesEveryElement(t *testing.T) {
	t.Parallel()
	var received []string
	registry := OptionsRegistry{"listed": {DecodeList: func(raw json.RawMessage) (any, error) {
		received = append(received, string(raw))
		return string(raw), nil
	}}}

	if _, err := registry.Decode("listed", []json.RawMessage{
		json.RawMessage(`"always"`),
		json.RawMessage(`{"null": "ignore"}`),
		json.RawMessage(`"third"`),
	}); err != nil {
		t.Fatalf("a list rule refused its elements: %v", err)
	}
	if _, err := registry.Decode("listed", nil); err != nil {
		t.Fatalf("a list rule with a bare severity errored: %v", err)
	}

	want := []string{`["always",{"null":"ignore"},"third"]`, ``}
	if len(received) != len(want) {
		t.Fatalf("decoder called %d times, want %d", len(received), len(want))
	}
	for index := range want {
		if received[index] != want[index] {
			t.Errorf("call %d received %s, want %s", index, received[index], want[index])
		}
	}
}

// TestANullIsRefusedBeforeAnyDecoderSeesIt pins the seam: the refusal happens here, for both
// arities, so no decoder has to remember that Go reads a null as "leave the field alone". The
// decoders below accept anything, which is what proves the refusal is the config layer's.
func TestANullIsRefusedBeforeAnyDecoderSeesIt(t *testing.T) {
	t.Parallel()
	acceptAll := func(raw json.RawMessage) (any, error) { return string(raw), nil }
	registry := OptionsRegistry{"single": {Decode: acceptAll}, "listed": {DecodeList: acceptAll}}

	for _, testCase := range []struct {
		rule     string
		elements []string
		want     string
	}{
		{"single", []string{`null`}, "element 1 is null"},
		{"single", []string{`{"b": 1, "a": {"c": [0, null]}}`}, "element 1 at a.c.1 is null"},
		{"listed", []string{`"always"`, `{"null": null}`}, "element 2 at null is null"},
	} {
		elements := make([]json.RawMessage, 0, len(testCase.elements))
		for _, element := range testCase.elements {
			elements = append(elements, json.RawMessage(element))
		}
		_, err := registry.Decode(testCase.rule, elements)
		if err == nil {
			t.Errorf("%s %v was accepted", testCase.rule, testCase.elements)
			continue
		}
		for _, want := range []string{testCase.rule, testCase.want} {
			if !strings.Contains(err.Error(), want) {
				t.Errorf("%s %v: the refusal does not name %q: %v", testCase.rule, testCase.elements, want, err)
			}
		}
	}

	// The control: the string "null" and a key named null are values, not nulls, and pass through.
	if _, err := registry.Decode("listed", []json.RawMessage{json.RawMessage(`"null"`), json.RawMessage(`{"null": "ignore"}`)}); err != nil {
		t.Errorf("a string and a key spelled null were refused: %v", err)
	}
}

// TestMalformedOptionsAreAnError guards against a rule silently receiving a zero struct because its
// JSON did not parse.
func TestMalformedOptionsAreAnError(t *testing.T) {
	t.Parallel()
	registry := OptionsRegistry{"guard": {Decode: DecodeInto[guardOptions](), Required: true}}

	if _, err := registry.Decode("guard", []json.RawMessage{json.RawMessage(`{"libraryDirectory": 42}`)}); err == nil {
		t.Fatal("options of the wrong JSON type decoded successfully")
	}
}

// TestUnconfiguredIsNotScopedOff pins the distinction the coverage line depends on.
//
// "Someone turned this rule off" and "nobody has said whether this rule should run" are different
// facts. Reporting them as one describes a brand-new rule as though it had been deliberately
// excluded, which is a lie about who decided what.
func TestUnconfiguredIsNotScopedOff(t *testing.T) {
	t.Parallel()
	configuration := &Config{
		Rules: map[string]RuleSetting{
			"configured-on":  {Severity: SeverityError},
			"configured-off": {Severity: SeverityOff},
		},
	}
	resolved := configuration.Resolve("File.ts")

	if status, _ := resolved.StatusOf("configured-on"); status != StatusEnabled {
		t.Fatalf("an enabled rule reported status %d", status)
	}
	if status, _ := resolved.StatusOf("configured-off"); status != StatusScopedOff {
		t.Fatalf("a rule the config turned off reported status %d rather than scoped-off", status)
	}
	if status, _ := resolved.StatusOf("never-mentioned"); status != StatusUnconfigured {
		t.Fatalf("a rule absent from the config reported status %d rather than unconfigured", status)
	}

	ignoring := &Config{IgnorePatterns: []string{"**"}}
	if status, _ := ignoring.Resolve("File.ts").StatusOf("anything"); status != StatusFileIgnored {
		t.Fatal("a rule on an ignored file did not report the file as the reason")
	}
}

// TestTheLiveGuardRuleGetsItsOptions is the specific rule that was inert, against the real configuration.
func TestTheLiveGuardRuleGetsItsOptions(t *testing.T) {
	t.Parallel()
	loaded, err := Load(liveConfigPath(t))
	if err != nil {
		t.Fatalf("loading the live config: %v", err)
	}

	elements := loaded.Resolve("libraries/structure/source/api/Fetch.ts").RawOptionsFor("boundary-no-project-import")
	if len(elements) == 0 {
		t.Fatal("boundary-no-project-import received no options from the live config, which is the inert-rule defect")
	}
	if len(elements) != 1 {
		t.Fatalf("boundary-no-project-import takes one option element and the live config gives it %d", len(elements))
	}

	var typed guardOptions
	if err := json.Unmarshal(elements[0], &typed); err != nil {
		t.Fatalf("the live options did not decode: %v", err)
	}
	if typed.LibraryDirectory != "/libraries/structure/" {
		t.Fatalf("libraryDirectory did not survive the round trip: %q", typed.LibraryDirectory)
	}
}
