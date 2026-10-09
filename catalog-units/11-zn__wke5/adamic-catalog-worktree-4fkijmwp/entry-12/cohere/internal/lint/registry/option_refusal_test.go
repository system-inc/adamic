package registry

import (
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/rule"
)

/*
 * Every registered decoder refuses an option key it does not declare, and names the key.
 *
 * option_decoding_test.go reads the source and proves every decode is strict. This one runs every
 * decoder and proves it behaves strictly, because a strict call can still be defeated: a decoder that
 * swallows the error and falls back to its defaults turns "unknown key refused" into "whole option
 * silently dropped", which is worse than the lenient decode it replaced. ban-ts-comment and
 * no-restricted-types both had that shape and were fixed for it (#4a4yse4).
 *
 * Each row is a pair. The baseline is the smallest options value the decoder accepts, and it must
 * decode; without it, a probe refused for some other reason (a required key missing, a mode string
 * expected) would pass as a refusal. The probe is the baseline plus one key no schema declares, and
 * it must be refused with an error naming that key.
 */

// unknownOptionKey is a key no rule declares, spelled so that no case-insensitive match can rescue it.
const unknownOptionKey = "cohereUnknownOptionKey"

// lenientOptionRules are the rules whose upstream schema leaves the top-level options object open,
// so ESLint, and this, loads an unknown key there. Their probe must decode, which pins the leniency
// as a decision rather than an accident. The sites are in lenientOptionDecodes with the schema each
// one checked.
var lenientOptionRules = map[string]bool{
	"react/forbid-prop-types":       true,
	"react/jsx-no-useless-fragment": true,
	"react/no-danger":               true,
	"react/style-prop-object":       true,
}

// optionListModes are the first elements tried for a DecodeOptionList rule that refuses `[{}]`,
// because its first element is a mode string and its object comes second. The first that decodes
// with an empty object after it is the baseline.
var optionListModes = []string{`"always"`, `"as-needed"`, `"never"`, `"functions"`, `"anyOrder"`, `"^[a-z]+$"`}

// TestEveryDecoderRefusesAnUnknownTopLevelKey probes the object every decoder reads first.
func TestEveryDecoderRefusesAnUnknownTopLevelKey(t *testing.T) {
	t.Parallel()
	probed, lenient, skipped := 0, 0, []string{}
	for _, registration := range rule.Registered() {
		name := registration.Rule.Name
		baseline, probe, decode := topLevelProbe(registration)
		if decode == nil {
			continue
		}
		if _, err := decode(baseline); err != nil {
			skipped = append(skipped, name)
			continue
		}
		_, err := decode(probe)
		if lenientOptionRules[name] {
			lenient++
			if err != nil {
				t.Errorf("%s: upstream's schema leaves its options object open, and %s was refused: %v", name, probe, err)
			}
			continue
		}
		probed++
		if err == nil {
			t.Errorf("%s: %s decoded, and the unknown key was dropped without a word", name, probe)
			continue
		}
		if !strings.Contains(err.Error(), unknownOptionKey) {
			t.Errorf("%s: %s was refused, but not for the key: %v", name, probe, err)
		}
	}
	t.Logf("%d decoders refused the unknown key, %d accepted it as upstream does, %d have no baseline here: %v",
		probed, lenient, len(skipped), skipped)
	if lenient != len(lenientOptionRules) {
		t.Errorf("probed %d of the %d lenient rules, so a lenient rule was renamed or lost its decoder", lenient, len(lenientOptionRules))
	}
	// About 190 registrations carry a decoder. Under 150 probed means the baselines stopped decoding
	// for most of them, and the refusals above stopped meaning anything.
	if probed < 150 {
		t.Fatalf("only %d decoders were probed, so this check proved little", probed)
	}
}

// topLevelProbe builds a registration's baseline and probe, and the decoder to run them through, or a
// nil decoder for a rule that takes no options or reads paths and so needs a real config to decode.
func topLevelProbe(registration rule.Registration) (baseline []byte, probe []byte, decode func([]byte) (any, error)) {
	object := `{}`
	probeObject := `{"` + unknownOptionKey + `": true}`
	switch {
	case registration.Decode != nil:
		return []byte(object), []byte(probeObject), registration.Decode
	case registration.DecodeOptionList != nil:
		if _, err := registration.DecodeOptionList([]byte(`[` + object + `]`)); err == nil {
			return []byte(`[` + object + `]`), []byte(`[` + probeObject + `]`), registration.DecodeOptionList
		}
		for _, mode := range optionListModes {
			if _, err := registration.DecodeOptionList([]byte(`[` + mode + `, ` + object + `]`)); err == nil {
				return []byte(`[` + mode + `, ` + object + `]`), []byte(`[` + mode + `, ` + probeObject + `]`), registration.DecodeOptionList
			}
		}
		return []byte(`[` + object + `]`), []byte(`[` + probeObject + `]`), registration.DecodeOptionList
	}
	return nil, nil, nil
}

// nestedOptionCase is one object below the top level, whose schema closes it too.
type nestedOptionCase struct {
	rule     string
	baseline string
	probe    string
	// refusal is what the error must contain; the unknown key unless the case is about spelling.
	refusal string
}

// nestedOptionCases reach the objects a top-level probe cannot: list entries, a polymorphic value's
// object form, and an option element after a mode string. Each is a decode this task converted.
var nestedOptionCases = []nestedOptionCase{
	{rule: "no-restricted-globals", baseline: `[{"name": "event", "message": "m"}]`, probe: `[{"name": "event", "mesage": "m"}]`, refusal: "mesage"},
	{rule: "no-restricted-globals", baseline: `[{"globals": ["event"]}]`, probe: `[{"globals": ["event"], "checkGlobalObjects": true}]`, refusal: "checkGlobalObjects"},
	{rule: "no-restricted-imports", baseline: `[{"paths": [{"name": "lodash"}]}]`, probe: `[{"paths": [{"name": "lodash", "importNames": ["map"], "mesage": "m"}]}]`, refusal: "mesage"},
	{rule: "no-restricted-imports", baseline: `[{"patterns": [{"group": ["lodash/*"]}]}]`, probe: `[{"patterns": [{"group": ["lodash/*"], "mesage": "m"}]}]`, refusal: "mesage"},
	{rule: "no-restricted-properties", baseline: `[{"object": "Math", "property": "pow"}]`, probe: `[{"object": "Math", "propery": "pow"}]`, refusal: "propery"},
	{rule: "no-restricted-exports", baseline: `{"restrictDefaultExports": {"direct": true}}`, probe: `{"restrictDefaultExports": {"directly": true}}`, refusal: "directly"},
	{rule: "react/forbid-elements", baseline: `{"forbid": [{"element": "button"}]}`, probe: `{"forbid": [{"element": "button", "mesage": "m"}]}`, refusal: "mesage"},
	{rule: "react/forbid-dom-props", baseline: `{"forbid": ["id"]}`, probe: `{"forbid": ["id"], "forbidden": ["id"]}`, refusal: "forbidden"},
	{rule: "@typescript-eslint/no-restricted-types", baseline: `{"types": {"Foo": {"message": "m"}}}`, probe: `{"types": {"Foo": {"mesage": "m"}}}`, refusal: "mesage"},
	{rule: "@typescript-eslint/ban-ts-comment", baseline: `{"ts-ignore": {"descriptionFormat": "^: "}}`, probe: `{"ts-ignore": {"descriptionFormt": "^: "}}`, refusal: "descriptionFormt"},
	{rule: "@typescript-eslint/no-misused-promises", baseline: `{"checksVoidReturn": {"arguments": false}}`, probe: `{"checksVoidReturn": {"argument": false}}`, refusal: "argument"},
	{rule: "@typescript-eslint/no-misused-promises", baseline: `{"checksVoidReturn": false}`, probe: `{"checksVoidReturnOpts": {"arguments": false}}`, refusal: "checksVoidReturnOpts"},
	{rule: "@typescript-eslint/no-floating-promises", baseline: `{"allowForKnownSafePromises": [{"from": "file", "name": "Safe"}]}`, probe: `{"allowForKnownSafePromises": [{"from": "file", "name": "Safe", "paths": "x"}]}`, refusal: "paths"},
	{rule: "@typescript-eslint/only-throw-error", baseline: `{"allow": [{"from": "file", "name": "Thrown"}]}`, probe: `{"allow": [{"from": "file", "name": "Thrown", "paths": "x"}]}`, refusal: "paths"},
	{rule: "@typescript-eslint/prefer-promise-reject-errors", baseline: `{"allow": [{"from": "file", "name": "Rejected"}]}`, probe: `{"allow": [{"from": "file", "name": "Rejected", "paths": "x"}]}`, refusal: "paths"},
	{rule: "@typescript-eslint/prefer-nullish-coalescing", baseline: `{"ignorePrimitives": {"string": true}}`, probe: `{"ignorePrimitives": {"strings": true}}`, refusal: "strings"},
	{rule: "no-shadow-restricted-names", baseline: `{"reportGlobalThis": false}`, probe: `{"allowGlobalThis": true}`, refusal: "allowGlobalThis"},
	{rule: "base/security-require-context-access", baseline: `{"requirements": [{"contextKey": "user", "requiresAny": ["Authenticated"]}]}`, probe: `{"requirements": [{"contextKey": "user", "requiresAny": ["Authenticated"], "requireAll": true}]}`, refusal: "requireAll"},

	// Spelling, which a tag now declares: a key that matches its field only case-insensitively.
	{rule: "no-empty", baseline: `{"allowEmptyCatch": true}`, probe: `{"AllowEmptyCatch": true}`, refusal: "AllowEmptyCatch"},
	{rule: "@typescript-eslint/restrict-plus-operands", baseline: `{"allowAny": true}`, probe: `{"allowany": true}`, refusal: "allowany"},
	{rule: "@typescript-eslint/switch-exhaustiveness-check", baseline: `{"requireDefaultForNonUnion": true}`, probe: `{"RequireDefaultForNonUnion": true}`, refusal: "RequireDefaultForNonUnion"},
	{rule: "nexus/consistency-no-shouting", baseline: `{"allow": ["NASA"]}`, probe: `{"Allow": ["NASA"]}`, refusal: "Allow"},
}

func TestEveryNestedOptionObjectRefusesAnUnknownKey(t *testing.T) {
	t.Parallel()
	registrations := map[string]rule.Registration{}
	for _, registration := range rule.Registered() {
		registrations[registration.Rule.Name] = registration
	}
	for _, testCase := range nestedOptionCases {
		t.Run(testCase.rule+" "+testCase.refusal, func(t *testing.T) {
			t.Parallel()
			registration, registered := registrations[testCase.rule]
			if !registered {
				t.Fatalf("%s is not registered under that name", testCase.rule)
			}
			decode := registration.Decode
			if decode == nil {
				decode = registration.DecodeOptionList
			}
			if decode == nil {
				t.Fatalf("%s registers neither Decode nor DecodeOptionList", testCase.rule)
			}
			if _, err := decode([]byte(testCase.baseline)); err != nil {
				t.Fatalf("baseline %s was refused, so the probe could be refused for that instead: %v", testCase.baseline, err)
			}
			_, err := decode([]byte(testCase.probe))
			if err == nil {
				t.Fatalf("%s decoded, and %q did nothing", testCase.probe, testCase.refusal)
			}
			if !strings.Contains(err.Error(), testCase.refusal) {
				t.Fatalf("%s was refused, but the error does not name %q: %v", testCase.probe, testCase.refusal, err)
			}
		})
	}
}
