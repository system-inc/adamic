package registry

import (
	"fmt"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/internal/lint/testing"
)

// Every registration's decoder hands its rule the type the rule reads.
//
// Run takes `options any`, so nothing at compile time compares the type a Decode returns with the
// type its rule passes to rule.OptionsAs. `unified-signatures` asserted a pointer over the value its
// decoder returned, every configured option was ignored, and every fixture stayed green because the
// fixtures handed Run options directly instead of decoding them (#qmvkf83). OptionsAs now panics on
// such a disagreement, but only when a run reaches it with decoded options, so this test makes every
// rule reach it: each decoder is fed the smallest option that means "configured" (an empty object,
// or an empty option list), and whatever it decodes is handed to that rule's own Run.
//
// A decoder that refuses the empty input (a required key, a strict schema) is skipped and counted,
// and the count has a floor, so a harness that silently decoded nothing cannot pass.
//
// Running a rule is not reaching its options read: a rule that calls OptionsAs only inside a listener
// for a node kind the probe source lacks would run, pass, and check nothing (#zwd43jn). So each rule
// must also move rule.OptionsAsReads, unless it is named in optionsReadUnreachableOnProbe. Measured
// when this landed: 172 of 172 reach it. The counter is global, which is why this test is not
// parallel: a parallel test calling OptionsAs at the same moment would make a rule look reached.
// Not parallel: it reads the process-wide counter rule.OptionsAsReads before and after each rule's run to
// tell whether the run reached its options, so a parallel test calling OptionsAs at the same moment would
// make an unreached rule look reached.
func TestEveryDecoderHandsItsRuleTheTypeTheRuleReads(t *testing.T) {
	exercised := 0
	refused := 0
	var unreached []string
	for _, registration := range rule.Registered() {
		decoded, decodeError, decodes := decodeEmptyOption(registration)
		if !decodes {
			continue
		}
		if decodeError != nil || decoded == nil {
			refused++
			continue
		}
		exercised++
		subject := registration.Rule
		readsBefore := rule.OptionsAsReads.Load()
		// Not parallel: its parent reads rule.OptionsAsReads right after this subtest returns, so the run
		// must have finished by then.
		t.Run(subject.Name, func(t *testing.T) {
			defer func() {
				recovered := recover()
				if recovered == nil {
					return
				}
				message := fmt.Sprint(recovered)
				if strings.Contains(message, "registration's Decode") {
					t.Fatalf("%s: %s", subject.Name, message)
				}
				// Any other panic is the crash corpus's business, not this test's.
			}()
			if subject.NeedsTypeChecker {
				rule_testing.RunTypedWithOptions(t, subject, "/repository/source/Probe.tsx", "export const value = 1;\n", decoded)
				return
			}
			rule_testing.RunWithOptions(t, subject, "/repository/source/Probe.tsx", "export const value = 1;\n", decoded)
		})
		if rule.OptionsAsReads.Load() == readsBefore && !optionsReadUnreachableOnProbe[subject.Name] {
			unreached = append(unreached, subject.Name)
		}
	}
	if len(unreached) > 0 {
		t.Errorf("%d rules ran with decoded options without reaching rule.OptionsAs on the probe source, "+
			"so the agreement check never looked at them: %v. Read options at the top of Run, or name the "+
			"rule in optionsReadUnreachableOnProbe with why", len(unreached), unreached)
	}
	registered := map[string]bool{}
	for _, registration := range rule.Registered() {
		registered[registration.Rule.Name] = true
	}
	for name := range optionsReadUnreachableOnProbe {
		if !registered[name] {
			t.Errorf("optionsReadUnreachableOnProbe names %s, which is not registered", name)
		}
	}
	// About 110 registrations carry a decoder today. Under 80 exercised means the empty input stopped
	// decoding for most of them, or the walk is reading the wrong list, rather than that the tree
	// lost its options.
	if exercised < 80 {
		t.Fatalf("only %d decoders produced options to hand their rule (%d refused the empty input), "+
			"so this check proved little", exercised, refused)
	}
}

// decodeEmptyOption runs a registration's decoder on the smallest configured input it takes. decodes
// is false for a rule that takes no options, or whose decoder needs to know where the config lives.
func decodeEmptyOption(registration rule.Registration) (decoded any, decodeError error, decodes bool) {
	switch {
	case registration.Decode != nil:
		decoded, decodeError = registration.Decode([]byte(`{}`))
		return decoded, decodeError, true
	case registration.DecodeOptionList != nil:
		decoded, decodeError = registration.DecodeOptionList([]byte(`[{}]`))
		if decodeError != nil {
			decoded, decodeError = registration.DecodeOptionList([]byte(`[]`))
		}
		return decoded, decodeError, true
	}
	return nil, nil, false
}

// optionsReadUnreachableOnProbe names the rules whose options read cannot be reached on the probe
// source, each with why. Empty when it landed: all 172 exercised rules read their options at the top
// of Run. A rule added here is one the agreement check does not cover, so it needs its own decoder
// test that drives a real fixture through its registration's Decode.
//
// The fifteen React Compiler rules decode only `{}`, which ESLint writes into resolved configs and
// which carries nothing, so their Run has no option to read (#d21war2). A non-empty object is
// refused at decode; TestCompilerRulesAcceptOnlyTheEmptyObject drives each one's registered Decode.
//
// react-hooks/rules-of-hooks accepts `additionalHooks` and reads it nowhere, because upstream's rule
// reads it nowhere either; TestRulesOfHooksAdditionalHooksChangesNothing pins that.
var optionsReadUnreachableOnProbe = map[string]bool{
	"react-hooks/rules-of-hooks":               true,
	"react-hooks/config":                       true,
	"react-hooks/error-boundaries":             true,
	"react-hooks/globals":                      true,
	"react-hooks/immutability":                 true,
	"react-hooks/incompatible-library":         true,
	"react-hooks/no-deriving-state-in-effects": true,
	"react-hooks/preserve-manual-memoization":  true,
	"react-hooks/purity":                       true,
	"react-hooks/refs":                         true,
	"react-hooks/set-state-in-effect":          true,
	"react-hooks/set-state-in-render":          true,
	"react-hooks/static-components":            true,
	"react-hooks/unsupported-syntax":           true,
	"react-hooks/use-memo":                     true,
	"react-hooks/void-use-memo":                true,
}

type optionsReachControlOptions struct {
	Enabled bool
}

// optionsReachControl reads its options only inside a class listener, the shape the reach check
// exists to catch.
var optionsReachControl = rule.Rule{
	Name: "control/options-read-inside-a-listener",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		return rule.Listeners{ast.KindClassDeclaration: func(node *ast.Node) {
			rule.OptionsAs[optionsReachControlOptions](options)
		}}
	},
}

// The reach counter can fail: a rule reading its options only in a class listener does not move it
// on the probe source, and does on a class. Without this the reach assertion above could be passing
// because the counter moves on every run.
// Not parallel: it asserts the process-wide counter rule.OptionsAsReads does not move across a run, which
// a parallel test calling OptionsAs would break.
func TestTheOptionsReachCounterSeesARuleThatNeverReadsItsOptions(t *testing.T) {
	before := rule.OptionsAsReads.Load()
	rule_testing.RunWithOptions(t, optionsReachControl, "/repository/source/Probe.tsx", "export const value = 1;\n", optionsReachControlOptions{})
	if rule.OptionsAsReads.Load() != before {
		t.Fatal("the control moved the counter on a source with no class, so the counter cannot tell reached from run")
	}
	rule_testing.RunWithOptions(t, optionsReachControl, "/repository/source/Probe.tsx", "export class Probe {}\n", optionsReachControlOptions{})
	if rule.OptionsAsReads.Load() == before {
		t.Fatal("the control did not move the counter on a class, so the counter is not counting")
	}
}
