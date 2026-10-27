package oracle

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/lower"
)

// A checked .ts predicate used as a value (#vdt9f1q, ruled Oct 25) is the checked forwarder
// functionValue makes: it checks its own claim's true side, or an asserts on return, at every call
// through it, wherever the value went. visitor.ts is honest and agrees with Node in tsc's visitor
// shape, through an argument, a field, an element, a return, an alias and === identity. Four of
// Language's controls the checker narrows nothing by (#1761dah) were stopped at "an indirect call
// of a checked predicate" until the value carried its check; they lower now and agree with Node.
// The other three stay in predicate_false_flow_stopped: c33, isS.call, at its unbound-method
// refusal, and c24 and c45 at lowering's optional call, which no value reaches yet.
func init() {
	names := []string{
		"predicate_values/visitor.ts",
		"predicate_values/control_c34_satisfies.ts",
		"predicate_values/control_c44_inner_switch.ts",
		"predicate_values/control_c46_logical_assign.ts",
		"predicate_values/control_c53_inner_decl.ts",
	}
	for _, name := range names {
		fixtures = append(fixtures, struct {
			path    string
			lowers  bool
			checked bool
		}{"internal/oracle/testdata/" + name, true, false})
	}
}

// Each program lies on its predicate's true side (an asserts one returns for an Expression), and
// reaches the lie by one road: an argument in tsc's visitor shape, a field, a let alias, a returned
// function, an array element, an asserts value passed in, and a claim on the second parameter. Two
// roads stop short of a value: an optional call through one isn't lowered yet, and bind is refused
// as an unbound method. Node reads the missing field as undefined. As .ts, every road stops at the
// forwarder's check on sanitized and released native and the JavaScript backend, its message
// located at the predicate's declaration (12:1 in each), never at a call: one forwarder serves
// every call through the value (Optimizer's should). As .a, the unproven predicate stays refused.
// Mutant: the forwarder without its check leaves both backends' exit-70 contract, so every row goes
// red.
var predicateValueFixtures = []struct {
	name, function, nodeOut, checkedOut, branch, source string
}{
	{"argument", "isS", "undefined\n", "", "true", "Node"},
	{"field", "isS", "undefined\n", "", "true", "Node"},
	{"alias", "isS", "undefined\n", "", "true", "Node"},
	{"returned", "isS", "undefined\n", "", "true", "Node"},
	{"element", "isS", "undefined\n", "", "true", "Node"},
	{"asserts", "assertS", "undefined\n", "", "asserts", "Node"},
	{"second_parameter", "isS", "undefined\n", "", "true", "Node"},
}

func TestPredicateValuesCheckTheirClaim(t *testing.T) {
	t.Parallel()
	for _, probe := range predicateValueFixtures {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			witness, path := predicateValueInput(t, probe.name)
			node := onNode(t, witness)
			if difference := disagreement(run{stdout: []byte(probe.nodeOut)}, node); difference != "" {
				t.Fatal("source Node: " + difference)
			}
			var refused *lower.Refused
			if _, err := lowered(t, witness); !errors.As(err, &refused) {
				t.Fatalf(".a witness: want a refusal of the unproven predicate, got %v", err)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			message := fmt.Sprintf("adamic: panic: predicate %s used as a value, declared at %s:12:1, %s branch: source %s, target Statement\n", probe.function, path, probe.branch, probe.source)
			want := run{stdout: []byte(probe.checkedOut), stderr: []byte(message), exitCode: 70}
			sanitized, _ := nativelyUncached(t, program)
			for _, got := range []run{sanitized, releasedUncached(t, program), onJavaScriptBackend(t, program)} {
				if difference := disagreement(want, got); difference != "" {
					t.Fatal(difference)
				}
			}
			changes := 0
			for i := range program.Functions {
				if !strings.HasSuffix(program.Functions[i].Name, "_value") {
					continue
				}
				program.Functions[i].Body = mutateReadiness(program.Functions[i].Body, func(node any) any {
					if panic, ok := node.(ir.Panic); ok {
						constant, ok := panic.Message.(ir.StringConstant)
						if ok && strings.Contains(program.Strings[constant.Index], " used as a value, ") {
							changes++
							return ir.Evaluate{Value: ir.NumberConstant{Value: 0}}
						}
					}
					return node
				})
			}
			if changes != 1 {
				t.Fatalf("check mutant removed %d checks", changes)
			}
			mutant, _ := nativelyUncached(t, program)
			for _, got := range []run{mutant, onJavaScriptBackend(t, program)} {
				if disagreement(want, got) == "" || strings.Contains(string(got.stderr), " used as a value, ") {
					t.Fatalf("the forwarder without its check still stopped at it: %+v", got)
				}
			}
		})
	}
}

// tsc's Debug.assertNotNode (debug.ts:307) believes the test's false answer through its overload's
// claim, asserts node is Exclude<T, U>, with no call narrowing on false (Adamic's must on #vdt9f1q,
// unless main already refused it). isKeyword says no to a Keyword, and its value's check never sees a
// true answer to check, but R-T80 checks the overload's own claim at the call, so the call stops
// there before node.text is read. Mutant: the call's check removed, neither backend stops there.
func TestPredicateValueThroughAssertNotNode(t *testing.T) {
	t.Parallel()
	witness, path := predicateValueInput(t, "assert_not_node")
	node := onNode(t, witness)
	if difference := disagreement(run{stdout: []byte("undefined\n")}, node); difference != "" {
		t.Fatal("source Node: " + difference)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	var site *ir.PredicateCallCheck
	for i := range program.PredicateChecks.Sites {
		if program.PredicateChecks.Sites[i].Function == "assertNotNode" {
			site = &program.PredicateChecks.Sites[i]
		}
	}
	if site == nil {
		t.Fatalf("no check at assertNotNode's call: %+v", program.PredicateChecks.Sites)
	}
	want := run{stderr: []byte(fmt.Sprintf("adamic: panic: predicate assertNotNode at %s asserts branch: source Node, target Identifier\n", site.Where)), exitCode: 70}
	sanitized, _ := nativelyUncached(t, program)
	for _, got := range []run{sanitized, releasedUncached(t, program), onJavaScriptBackend(t, program)} {
		if difference := disagreement(want, got); difference != "" {
			t.Fatal(difference)
		}
	}
	changes := 0
	for i := range program.Functions {
		if !strings.HasSuffix(program.Functions[i].Name, "_predicate_check") {
			continue
		}
		program.Functions[i].Body = mutateReadiness(program.Functions[i].Body, func(node any) any {
			if _, ok := node.(ir.Panic); ok {
				changes++
				return ir.Evaluate{Value: ir.NumberConstant{Value: 0}}
			}
			return node
		})
	}
	if changes != 1 {
		t.Fatalf("check mutant removed %d checks", changes)
	}
	mutant, _ := nativelyUncached(t, program)
	for _, got := range []run{mutant, onJavaScriptBackend(t, program)} {
		if disagreement(want, got) == "" {
			t.Fatalf("the call's check removed, it still stopped there: %+v", got)
		}
	}
}

// tsc's visitNode(node, visitor, isIdentifier) over an open SyntaxKind (Optimizer, #vdt9f1q): a kind
// proves no layout, so the value's check is the tag test and the true branch reads the target
// through #k1t2wp9's checked view. The view is keyed by the target's type across the program, so a
// read narrowed by a call through the value reads through it too: the honest program agrees with
// Node, and a Keyword sharing Identifier's kind stops at its first field read with the view's own
// message, on sanitized and released native and the JavaScript backend. The forwarder without its
// check registers no view, so the lie would read a missing field unchecked.
func TestPredicateValuesOverAnOpenTag(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, nodeOut, stderr string }{
		{"open_tag", "x\nother\n", ""},
		{"open_tag_lie", "undefined\n", "adamic: panic: field read failed: node.text is not initialized; expected string, found missing\n"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			witness, path := predicateValueInput(t, probe.name)
			node := onNode(t, witness)
			if difference := disagreement(run{stdout: []byte(probe.nodeOut)}, node); difference != "" {
				t.Fatal("source Node: " + difference)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			want := node
			if probe.stderr != "" {
				want = run{stderr: []byte(probe.stderr), exitCode: 70}
			}
			sanitized, binary := nativelyUncached(t, program)
			for index, got := range []run{sanitized, releasedUncached(t, program), onJavaScriptBackend(t, program)} {
				if difference := disagreement(want, got); difference != "" {
					t.Fatalf("%s on %s: exit %d, stdout %q, stderr %q", difference, []string{"sanitized native", "release native", "the JavaScript backend"}[index], got.exitCode, got.stdout, got.stderr)
				}
			}
			if want.exitCode == 0 {
				if report := leaksUncached(t, program, binary); report != "" {
					t.Fatal(report)
				}
			}
		})
	}
}

// The witness is the stored .a, which Node runs and which stays refused; checked mode is its .ts copy.
func predicateValueInput(t *testing.T, name string) (string, string) {
	t.Helper()
	return predicateInputFrom(t, "internal/lower/testdata/predicate_values", name)
}
