package lower

import (
	"errors"
	"slices"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

// dispatchVisit lowers source as main.ts and gives the program with its visit function, where each pinned program
// keeps its dispatch.
func dispatchVisit(t *testing.T, source string) (*ir.Program, ir.Function) {
	t.Helper()
	lowered, err := lowerSourceAs(t, source, ".ts")
	if err != nil {
		t.Fatalf("Lower: %v", err)
	}
	for _, function := range lowered.Functions {
		if function.Name == "visit" {
			return lowered, function
		}
	}
	t.Fatal("no function named visit")
	return nil, ir.Function{}
}

// dispatchHeldKinds is how many times function reads a field named kind, and each local a declaration in it
// initializes from such a read.
func dispatchHeldKinds(function ir.Function) (int, []int) {
	reads, held := 0, []int{}
	walk(function.Body, func(node any) bool {
		switch node := node.(type) {
		case ir.Property:
			if node.Name == "kind" {
				reads++
			}
		case ir.Declare:
			found := false
			walk(node.Value, func(inner any) bool {
				if property, isProperty := inner.(ir.Property); isProperty && property.Name == "kind" {
					found = true
				}
				return true
			})
			if found {
				held = append(held, node.Local)
			}
		}
		return true
	})
	return reads, held
}

// dispatchArmSeen is one arm of a lowered dispatch call: the statement before its call, and the call.
type dispatchArmSeen struct {
	check ir.If
	call  ir.CallClosure
}

// dispatchArms is every arm of the dispatch calls in function, by the field of the table its call reads: an Effects
// whose result calls a field's function value, with the If its body holds.
func dispatchArms(function ir.Function) map[string]dispatchArmSeen {
	arms := map[string]dispatchArmSeen{}
	walk(function.Body, func(node any) bool {
		effects, isEffects := node.(ir.Effects)
		if !isEffects {
			return true
		}
		call, isCall := effects.Result.(ir.CallClosure)
		if !isCall {
			return true
		}
		field, isField := call.Closure.(ir.Property)
		if !isField {
			return true
		}
		seen := dispatchArmSeen{call: call}
		if len(effects.Body) == 1 {
			seen.check, _ = effects.Body[0].(ir.If)
		}
		arms[field.Name] = seen
		return true
	})
	return arms
}

// dispatchOneComparison reports whether condition is exactly one comparison, the held kind local isn't key.
func dispatchOneComparison(condition ir.Expression, local int, key float64) bool {
	comparison, isBinary := condition.(ir.Binary)
	if !isBinary || comparison.Operator != ir.NotEqual {
		return false
	}
	read, isRead := comparison.Left.(ir.Read)
	constant, isConstant := comparison.Right.(ir.NumberConstant)
	return isRead && isConstant && read.Local == local && constant.Value == key
}

// dispatchCheckMessage is the words check panics with, "" where its body is no one panic with a constant message.
func dispatchCheckMessage(program *ir.Program, check ir.If) string {
	if len(check.Then) != 1 {
		return ""
	}
	stop, isPanic := check.Then[0].(ir.Panic)
	if !isPanic {
		return ""
	}
	message, isConstant := stop.Message.(ir.StringConstant)
	if !isConstant {
		return ""
	}
	return program.Strings[message.Index]
}

// Language's condition 3 on the IR both backends emit from: t00.ts's visit reads node's kind exactly once, into one
// local, and every comparison the dispatch makes, the presence test's two, the arms' two and the checks' two, reads
// that local. fn has no local of its own, and the lookup is one line of --explain-checks with its two keys.
// Mutant, red: wrong-kind-twice, the check reading x.kind again instead of the held local (three kind reads, and
// two comparisons whose left isn't the local).
func TestDispatchReadsKindOnce(t *testing.T) {
	t.Parallel()
	lowered, visit := dispatchVisit(t, dispatchT00)
	reads, held := dispatchHeldKinds(visit)
	if reads != 1 || len(held) != 1 {
		t.Fatalf("visit reads kind %d times into %d locals, want once into one", reads, len(held))
	}
	comparisons := 0
	walk(visit.Body, func(node any) bool {
		comparison, isBinary := node.(ir.Binary)
		if !isBinary || (comparison.Operator != ir.Equal && comparison.Operator != ir.NotEqual) {
			return true
		}
		comparisons++
		if read, isRead := comparison.Left.(ir.Read); !isRead || read.Local != held[0] {
			t.Errorf("a comparison reads %#v, not the held kind", comparison.Left)
		}
		return true
	})
	if comparisons != 6 {
		t.Fatalf("%d comparisons, want 6: the presence test's two, the arms' two and the checks' two", comparisons)
	}
	for _, local := range lowered.Locals {
		if local.Name == "fn" {
			t.Fatal("fn has a local: it holds no value")
		}
	}
	if len(lowered.DispatchChecks) != 1 || lowered.DispatchChecks[0].Keys != 2 {
		t.Fatalf("dispatch checks %+v, want one site with 2 keys", lowered.DispatchChecks)
	}
}

// The wrong-kind pin (Adamic's, from ruling 17): no admitted route calls an entry with a node of another kind, so
// the check is pinned on the IR. Each of t00.ts's arms holds it before its call: one comparison of the held kind
// with the arm's own key, panicking with row 9's words for x's type and the key's member.
// Mutant, red: wrong-no-check, the call lowered with no tag check (no If before either call).
func TestDispatchArmsHoldTheCheck(t *testing.T) {
	t.Parallel()
	lowered, visit := dispatchVisit(t, dispatchT00)
	_, held := dispatchHeldKinds(visit)
	if len(held) != 1 {
		t.Fatalf("%d held kinds, want 1", len(held))
	}
	arms := dispatchArms(visit)
	for _, want := range []struct {
		key     string
		value   float64
		message string
	}{{"0", 0, "cast failed: this Node is not a NodeA"}, {"1", 1, "cast failed: this Node is not a NodeB"}} {
		arm, found := arms[want.key]
		if !found {
			t.Fatalf("no arm calls key %s's entry (arms %v)", want.key, arms)
		}
		if !dispatchOneComparison(arm.check.Condition, held[0], want.value) {
			t.Fatalf("key %s's check is %#v, want the held kind compared once with %v", want.key, arm.check.Condition, want.value)
		}
		if got := dispatchCheckMessage(lowered, arm.check); got != want.message {
			t.Fatalf("key %s's check says %q, want %q", want.key, got, want.message)
		}
	}
}

// Her 20:46:47Z rule on t22.ts's IR: key 1's arm hands node, read as the parameter it is, with no CheckedCast or
// second test, to the entry declared for ImportAttributes, as the key's members ImportAttributes | AssertClause,
// which its check names; and the check is one comparison of the held kind with 1.
// Mutants, red: wrong-one-member, x passed as one member the checker lists (the message names one type); and
// wrong-per-member-check, the tag compared once per member under the key (the condition is no single comparison).
func TestDispatchTwoMembersUnderOneKey(t *testing.T) {
	t.Parallel()
	lowered, visit := dispatchVisit(t, dispatchT22)
	_, held := dispatchHeldKinds(visit)
	if len(held) != 1 {
		t.Fatalf("%d held kinds, want 1", len(held))
	}
	arm, found := dispatchArms(visit)["1"]
	if !found {
		t.Fatal("no arm calls key 1's entry")
	}
	if !dispatchOneComparison(arm.check.Condition, held[0], 1) {
		t.Fatalf("key 1's check is %#v, want one comparison of the held kind with 1", arm.check.Condition)
	}
	message := dispatchCheckMessage(lowered, arm.check)
	members, named := strings.CutPrefix(message, "cast failed: this Node is not a ")
	names := strings.Split(members, " | ")
	slices.Sort(names)
	if !named || strings.Join(names, ",") != "AssertClause,ImportAttributes" {
		t.Fatalf("key 1's check says %q, want node named as ImportAttributes | AssertClause", message)
	}
	if len(arm.call.Arguments) != 1 || len(visit.Parameters) != 1 {
		t.Fatalf("key 1's call passes %d arguments, visit takes %d", len(arm.call.Arguments), len(visit.Parameters))
	}
	if read, isRead := arm.call.Arguments[0].(ir.Read); !isRead || read.Local != visit.Parameters[0] {
		t.Fatalf("key 1's call hands %#v, want node's own read", arm.call.Arguments[0])
	}
	walk(visit.Body, func(node any) bool {
		if _, isCast := node.(ir.CheckedCast); isCast {
			t.Fatal("a CheckedCast in visit: the arm proved the tag, and the held kind is its only test")
		}
		return true
	})
}

// t24.ts lowers with its trailing argument passed: each arm hands node and then tag, read as visit's second
// parameter, to its entry, as the call condition admitted it.
// Mutant, red: the arms built from x alone, dropping the call's other arguments.
func TestDispatchPassesTheTrailingArgument(t *testing.T) {
	t.Parallel()
	_, visit := dispatchVisit(t, dispatchT24)
	arms := dispatchArms(visit)
	if len(arms) != 2 || len(visit.Parameters) != 2 {
		t.Fatalf("%d arms, visit takes %d parameters", len(arms), len(visit.Parameters))
	}
	for key, arm := range arms {
		if len(arm.call.Arguments) != 2 {
			t.Fatalf("key %s's call passes %d arguments, want 2", key, len(arm.call.Arguments))
		}
		if read, isRead := arm.call.Arguments[1].(ir.Read); !isRead || read.Local != visit.Parameters[1] {
			t.Fatalf("key %s's call passes %#v second, want tag's own read", key, arm.call.Arguments[1])
		}
	}
}

// A plain call no presence test guards declines by the lowering's guard, so the lookup keeps today's stop at its
// assertion, in its words: with no guard, no key matching would reach the arms' unreachable tail.
// Mutant, red: dispatchCallGuarded answering "" for every call (the scan admits it and Lower succeeds).
func TestDispatchUnguardedCallKeepsTheStop(t *testing.T) {
	t.Parallel()
	source := dispatchTwin(t, dispatchT00, 13, "\treturn fn(node);", false)
	_, roads := dispatchScan(t, source)
	if reasons := dispatchDeclineReasons(roads); len(roads.sites) != 0 || len(reasons) != 1 || reasons[0] != "escapes: a call of fn no presence test guards" {
		t.Fatalf("%d admitted, declines %q", len(roads.sites), reasons)
	}
	_, err := lowerSourceAs(t, source, ".ts")
	var pending *NotYet
	if !errors.As(err, &pending) || !strings.Contains(pending.Where, "main.ts:12:14") || !strings.Contains(pending.What, "a member typed any (0(0)) related to NodeA at the assertion site") {
		t.Fatalf("want today's stop at the assertion, main.ts:12:14, got %v", err)
	}
}
