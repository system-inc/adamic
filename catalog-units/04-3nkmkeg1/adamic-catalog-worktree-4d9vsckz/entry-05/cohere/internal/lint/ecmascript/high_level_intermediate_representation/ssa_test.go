// Tests for single static assignment construction.
//
// # Why these lower through a real program rather than a bare parse
//
// `lowerSource` in lower_test.go parses and nothing more, which is right for testing the SHAPE of a
// lowered graph. It is wrong for testing SSA, and the reason is the whole point of this file: with
// no checker, every variable reference lowers to `LoadGlobal`, there is nothing to rename, and
// construction places zero phis while every structural assertion still passes. A suite built on
// that harness would be green and would be measuring nothing.
//
// So these go through `rule_testing.RunTyped`, which builds a real program with a real checker. It
// costs roughly a second per case, which is why the corpus test caps its sample.
package high_level_intermediate_representation

import (
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/internal/lint/testing"
	"github.com/system-inc/cohere/mutation_aliasing"
	"github.com/system-inc/cohere/static_single_assignment"
)

// lowerTypedForSSA lowers the first function in code with a real type checker attached.
func lowerTypedForSSA(t *testing.T, code string) *Function {
	t.Helper()
	var lowered *Function
	probe := rule.Rule{
		Name:             "ssa-test-harness",
		NeedsTypeChecker: true,
		Run: func(ctx rule.Context, options any) rule.Listeners {
			return rule.Listeners{
				ast.KindSourceFile: func(node *ast.Node) {
					if ctx.TypeChecker == nil {
						t.Fatal("the harness produced no type checker, so this test would prove nothing")
					}
					forEachFunctionLike(node, func(function *ast.Node) {
						if lowered == nil {
							lowered = Lower(function, ctx.TypeChecker)
						}
					})
				},
			}
		},
	}
	rule_testing.RunTyped(t, probe, "probe.tsx", code)
	if lowered == nil {
		t.Fatal("nothing lowered")
	}
	return lowered
}

// cohereEverywhere checks the SSA invariants on a function and every nested function.
func cohereEverywhere(t *testing.T, function *Function) {
	t.Helper()
	var each func(*Function)
	each = func(f *Function) {
		for _, violation := range VerifySSA(f) {
			t.Errorf("SSA violation: %s", violation.Detail)
		}
		for _, nested := range f.Functions {
			each(nested)
		}
	}
	each(function)
}

func TestSSAMatchesReactSpecFixtures(t *testing.T) {
	t.Parallel()

	for _, src := range []string{
		"export function f(){ let y = 2; if (y > 1) { y = 1; } else { y = 2; } let x = y; return x; }\n",
		"export function f(){ let x = 1; while (x < 10) { x = x + 1; } return x; }\n",
		"export function f(){ let x = 1; try { x = 2; } finally { x = 3; } return x; }\n",
	} {
		fn := lowerTypedForSSA(t, src)
		Construct(fn)
		t.Logf("SRC %s\n%s", src, Print(fn))
		if v := VerifySSA(fn); len(v) > 0 {
			for _, one := range v {
				t.Errorf("VIOLATION: %s", one)
			}
		}
		t.Logf("stats %+v", CollectSSAStats(fn))
	}
}

// Deliberately corrupt a correct SSA function four different ways and confirm
// VerifySSA reports each one. A checker that has never been shown to fail is
// not evidence that the property holds.
func TestSSAVerifierDetectsEachViolationClass(t *testing.T) {
	t.Parallel()

	build := func() *Function {
		fn := lowerTypedForSSA(t, "export function f(){ let y = 2; if (y > 1) { y = 1; } else { y = 2; } let x = y; return x; }\n")
		Construct(fn)
		if v := VerifySSA(fn); len(v) != 0 {
			t.Fatalf("baseline should be clean, got %v", v)
		}
		return fn
	}

	// 1. Misplace a phi: move it from the join to a block that does not dominate its uses.
	t.Run("phi moved off the join", func(t *testing.T) {
		t.Parallel()
		fn := build()
		var moved *Phi
		var from, to *BasicBlock
		for _, b := range fn.Blocks {
			if len(b.Phis) > 0 {
				moved, from = b.Phis[0], b
				break
			}
		}
		for _, b := range fn.Blocks {
			if b.Id != from.Id && len(b.Predecessors) == 1 {
				to = b
				break
			}
		}
		if moved == nil || to == nil {
			t.Skip("no phi to move")
		}
		from.Phis = nil
		to.Phis = append(to.Phis, moved)
		v := VerifySSA(fn)
		if len(v) == 0 {
			t.Fatal("moving a phi off its join was NOT detected")
		}
		t.Logf("detected: %v", v)
	})

	// 2. Point a use at a value defined on a sibling branch.
	t.Run("use reads a sibling branch definition", func(t *testing.T) {
		t.Parallel()
		fn := build()
		var join *BasicBlock
		for _, b := range fn.Blocks {
			if len(b.Phis) > 0 {
				join = b
				break
			}
		}
		if join == nil || len(join.Instructions) == 0 {
			t.Skip("no join")
		}
		// Rewrite the first use in the join to one predecessor's definition.
		operand := join.Phis[0].Operands.At(PhiOperandsInOrder(join.Phis[0])[0])
		done := false
		for _, id := range join.Instructions {
			EachInstructionPlacePointer(fn.Instructions[id], func(p *Place, role PlaceRole) {
				if role != PlaceRoleDefine && !done {
					p.Identifier = operand.Identifier
					done = true
				}
			})
			if done {
				break
			}
		}
		v := VerifySSA(fn)
		if len(v) == 0 {
			t.Fatal("a use reading a non-dominating sibling definition was NOT detected")
		}
		t.Logf("detected: %v", v)
	})

	// 3. Define one value twice.
	t.Run("value defined twice", func(t *testing.T) {
		t.Parallel()
		fn := build()
		var first, second *Instruction
		for _, b := range fn.Blocks {
			for _, id := range b.Instructions {
				ins := fn.Instructions[id]
				if first == nil {
					first = ins
				} else if second == nil {
					second = ins
				}
			}
		}
		if first == nil || second == nil {
			t.Skip("too few instructions")
		}
		second.LValue.Identifier = first.LValue.Identifier
		v := VerifySSA(fn)
		if len(v) == 0 {
			t.Fatal("a doubly-defined value was NOT detected")
		}
		t.Logf("detected: %v", v)
	})

	// 4. Read a value before it is defined within one block. Rather than hunting for an existing
	// use, rewrite the FIRST instruction's value to a LoadLocal of the LAST instruction's result,
	// which is unambiguously a backwards read inside one block.
	t.Run("use before definition in the same block", func(t *testing.T) {
		t.Parallel()
		fn := build()
		entry, _ := fn.Block(fn.Entry)
		if len(entry.Instructions) < 3 {
			t.Skip("entry too short")
		}
		last := fn.Instructions[entry.Instructions[len(entry.Instructions)-1]]
		first := fn.Instructions[entry.Instructions[0]]
		first.Value = &LoadLocal{Place: Place{Identifier: last.LValue.Identifier}}
		v := VerifySSA(fn)
		if len(v) == 0 {
			t.Fatal("a use before its definition was NOT detected")
		}
		t.Logf("detected: %v", v)
	})
}

// Constructs the corpus sample did not contain, plus the loop shapes most likely
// to break a construction that only looks right on straight-line code.
func TestSSAAcrossEveryControlFlowConstruct(t *testing.T) {
	t.Parallel()

	cases := []struct{ name, code string }{
		{"for-in", "export function f(o: any){ let n = 0; for (const k in o) { n = n + 1; } return n; }\n"},
		{"labeled-break", "export function f(){ let n = 0; outer: for (let i = 0; i < 3; i++) { while (true) { n = n + 1; break outer; } } return n; }\n"},
		{"labeled-continue", "export function f(){ let n = 0; outer: for (let i = 0; i < 3; i++) { for (let j = 0; j < 3; j++) { n = n + 1; continue outer; } } return n; }\n"},
		{"nested-loop", "export function f(){ let s = 0; for (let i = 0; i < 3; i++) { for (let j = 0; j < 3; j++) { s = s + i * j; } } return s; }\n"},
		{"do-while", "export function f(){ let x = 0; do { x = x + 1; } while (x < 5); return x; }\n"},
		{"for-of", "export function f(a: number[]){ let t = 0; for (const v of a) { t = t + v; } return t; }\n"},
		{"switch-fallthrough", "export function f(n: number){ let r = 0; switch (n) { case 1: r = 1; case 2: r = r + 2; break; default: r = 9; } return r; }\n"},
		{"try-catch-finally", "export function f(){ let x = 0; try { x = 1; } catch (e) { x = 2; } finally { x = x + 1; } return x; }\n"},
		{"loop-with-break", "export function f(a: number[]){ let t = 0; for (const v of a) { if (v < 0) { break; } t = t + v; } return t; }\n"},
		{"loop-with-continue", "export function f(a: number[]){ let t = 0; for (const v of a) { if (v < 0) { continue; } t = t + v; } return t; }\n"},
		{"ternary", "export function f(b: boolean){ let x = b ? 1 : 2; return x; }\n"},
		{"logical", "export function f(b: boolean, c: boolean){ let x = b && c; let y = b || c; return x ? y : x; }\n"},
		{"nested-function", "export function f(){ let n = 1; const g = () => { return n; }; n = 2; return g(); }\n"},
		{"reassigned-param", "export function f(a: number){ if (a > 0) { a = a * 2; } return a; }\n"},
		{"shadowed", "export function f(){ let y = 1; { let y = 2; y = y + 1; } y = y + 5; return y; }\n"},
		{"update-operators", "export function f(){ let i = 0; i++; ++i; i--; return i; }\n"},
		{"destructuring", "export function f(o: {a: number, b: number}){ let { a, b } = o; if (a > b) { a = b; } return a + b; }\n"},
	}
	for _, one := range cases {
		t.Run(one.name, func(t *testing.T) {
			t.Parallel()
			fn := lowerTypedForSSA(t, one.code)
			if fn == nil {
				t.Fatal("did not lower")
			}
			Construct(fn)
			var each func(*Function)
			each = func(f *Function) {
				for _, v := range VerifySSA(f) {
					t.Errorf("violation: %s", v.Detail)
				}
				for _, n := range f.Functions {
					each(n)
				}
			}
			each(fn)
			st := CollectSSAStats(fn)
			t.Logf("phis=%d named=%d uses=%d", st.Phis, st.NamedValues, st.Uses)
		})
	}
}

// TestMutatingVisitorCoversEveryValue pins visitor_mutate.go against visitor.go.
//
// The two switches must yield the same places in the same order for every instruction value. They
// are separate functions, so nothing but this stops one gaining a variant the other lacks, and the
// symptom of a drift is that renaming silently skips an operand: the graph stays well formed and a
// use keeps pointing at a pre-SSA identifier.
//
// Rather than list the variants, which would itself drift, this reflects over a constructed
// instance of every value the printer knows about, which is the same closed set.
func TestMutatingVisitorCoversEveryValue(t *testing.T) {
	t.Parallel()

	place := func(id static_single_assignment.IdentifierId) Place { return Place{Identifier: id} }
	values := []InstructionValue{
		&LoadLocal{Place: place(1)},
		&LoadContext{Place: place(1)},
		&DeclareLocal{LValue: place(1)},
		&DeclareContext{LValue: place(1)},
		&StoreLocal{LValue: place(1), Value: place(2)},
		&StoreContext{LValue: place(1), Value: place(2)},
		&Destructure{LValue: &PlacePattern{Place: place(1)}, Value: place(2)},
		&LoadGlobal{Name: "g"},
		&StoreGlobal{Name: "g", Value: place(1)},
		&PropertyLoad{Object: place(1)},
		&PropertyStore{Object: place(1), Value: place(2)},
		&PropertyDelete{Object: place(1)},
		&ComputedLoad{Object: place(1), Property: place(2)},
		&ComputedStore{Object: place(1), Property: place(2), Value: place(3)},
		&ComputedDelete{Object: place(1), Property: place(2)},
		&CallExpression{Callee: place(1), Args: []Argument{{Place: place(2)}}},
		&MethodCall{Receiver: place(1), Property: place(2), Args: []Argument{{Place: place(3)}}},
		&NewExpression{Callee: place(1), Args: []Argument{{Place: place(2)}}},
		&BinaryExpression{Left: place(1), Right: place(2)},
		&UnaryExpression{Value: place(1)},
		&PrefixUpdate{LValue: place(1), Value: place(2)},
		&PostfixUpdate{LValue: place(1), Value: place(2)},
		&Primitive{Value: nil},
		&RegExpLiteral{},
		&TemplateLiteral{Subexprs: []Place{place(1)}},
		&TaggedTemplateExpression{Tag: place(1), Subexprs: []Place{place(2)}},
		&TypeCastExpression{Value: place(1)},
		&MetaProperty{},
		&ObjectExpression{Properties: []ObjectProperty{{Value: place(1)}}},
		&ObjectMethod{},
		&ArrayExpression{Elements: []ArrayElement{{Place: place(1)}}},
		&FunctionExpression{Captures: []Place{place(1)}},
		&Await{Value: place(1)},
		&GetIterator{Value: place(1)},
		&IteratorNext{Iterator: place(1), Collection: place(2)},
		&NextPropertyOf{Value: place(1)},
		&JsxExpression{Props: []JsxAttribute{{Value: place(1)}}, Children: []Place{place(2)}},
		&JsxFragment{Children: []Place{place(1)}},
		&JsxText{},
		// Both root kinds, because the walkers SKIP a global root (it carries a name, not a
		// place) and a case listing only local roots would let a walker that skipped both pass.
		&StartMemoize{Deps: []ManualMemoDependency{
			{Root: ManualMemoRoot{Place: place(1)}, Path: []DependencyPathEntry{{Property: "a"}}},
			{Root: ManualMemoRoot{IsGlobal: true, Name: "g"}},
		}},
		&FinishMemoize{Value: place(1)},
		&Debugger{},
		&UnsupportedNode{},
	}

	for _, value := range values {
		var byValue, byPointer []struct {
			id   static_single_assignment.IdentifierId
			role PlaceRole
		}
		EachPlace(value, func(p Place, role PlaceRole) {
			byValue = append(byValue, struct {
				id   static_single_assignment.IdentifierId
				role PlaceRole
			}{p.Identifier, role})
		})
		EachPlacePointer(value, func(p *Place, role PlaceRole) {
			byPointer = append(byPointer, struct {
				id   static_single_assignment.IdentifierId
				role PlaceRole
			}{p.Identifier, role})
		})
		if len(byValue) != len(byPointer) {
			t.Errorf("%T: EachPlace yielded %d places, EachPlacePointer yielded %d",
				value, len(byValue), len(byPointer))
			continue
		}
		for index := range byValue {
			if byValue[index] != byPointer[index] {
				t.Errorf("%T: place %d differs: value=%+v pointer=%+v",
					value, index, byValue[index], byPointer[index])
			}
		}
	}
}

// TestPhiOperandsAreDeterministic pins the operand order.
//
// Phi.Operands is kept sorted by predecessor block id whatever order operands are set in, and
// PhiOperandsInOrder reads the ids in that order; this fails if either stops holding.
func TestPhiOperandsAreDeterministic(t *testing.T) {
	t.Parallel()

	phi := &Phi{Place: Place{Identifier: 9}}
	phi.Operands.Set(7, Place{Identifier: 3})
	phi.Operands.Set(2, Place{Identifier: 1})
	phi.Operands.Set(5, Place{Identifier: 2})
	want := []static_single_assignment.BlockId{2, 5, 7}
	for attempt := 0; attempt < 50; attempt++ {
		got := PhiOperandsInOrder(phi)
		if len(got) != len(want) {
			t.Fatalf("got %d operands, want %d", len(got), len(want))
		}
		for index := range want {
			if got[index] != want[index] {
				t.Fatalf("attempt %d: got %v, want %v", attempt, got, want)
			}
		}
	}
}

// Construction reaches a fixpoint: running it twice produces what running it once produced.
//
// The pass APPENDS phis and cannot reconcile ones it did not mint, so a second run over a graph that
// already carries phis leaves the old ones in place beside new ones naming the same merge. Their
// operands are identifiers from before the renumbering, so each defines a value nothing produces,
// joins no class in the disjoint partition, and takes no scope.
//
// This is not hypothetical: inlining an immediately invoked function expression restructures the
// graph and runs construction again, which is the path `preserve-manual-memoization` takes. Measured
// before the reset landed, a second run took corpus phis 1,548 to 3,096 and phis naming an undefined
// operand 221 to 1,730.
//
// Asserted as a fixpoint rather than a count, because the property that matters is "running it again
// changes nothing" and that survives every legitimate change to phi placement.
func TestConstructionIsIdempotent(t *testing.T) {
	t.Parallel()

	firstPhis, secondPhis := 0, 0
	firstStale, secondStale := 0, 0

	// A phi whose operand names no instruction, phi, or parameter defines a value the graph does not
	// produce.
	countStale := func(function *Function) (phis int, stale int) {
		defined := map[static_single_assignment.IdentifierId]bool{}
		for _, param := range function.Params {
			defined[param.Identifier] = true
		}
		for _, block := range function.Blocks {
			if block == nil {
				continue
			}
			for _, instructionId := range block.Instructions {
				if instruction := function.Instructions[instructionId]; instruction != nil {
					defined[instruction.LValue.Identifier] = true
				}
			}
			for _, phi := range block.Phis {
				defined[phi.Place.Identifier] = true
			}
		}
		for _, block := range function.Blocks {
			if block == nil {
				continue
			}
			for _, phi := range block.Phis {
				phis++
				for _, entry := range phi.Operands {
					if !defined[entry.Place.Identifier] {
						stale++
						break
					}
				}
			}
		}
		return phis, stale
	}

	forEachCorpusFunction(t, 300, func(function *Function, ranges *mutation_aliasing.MutableRanges,
		scopes *ReactiveScopes) {
		phis, stale := countStale(function)
		firstPhis += phis
		firstStale += stale
		Construct(function)
		phis, stale = countStale(function)
		secondPhis += phis
		secondStale += stale
	})

	if firstPhis == 0 {
		t.Fatal("the corpus produced no phis, so running construction twice proves nothing")
	}
	if secondPhis != firstPhis {
		t.Errorf("a second construction changed the phi count from %d to %d; the pass appends "+
			"rather than reconciles, so phis from the previous run must be dropped first",
			firstPhis, secondPhis)
	}
	if secondStale != firstStale {
		t.Errorf("a second construction changed the count of phis naming an undefined operand "+
			"from %d to %d", firstStale, secondStale)
	}
	t.Logf("stable across two runs: %d phis, %d naming an undefined operand", firstPhis, firstStale)
}

// Construction is not re-run after the inline, because upstream runs it once and after both.
//
// `Pipeline.ts` orders these deliberately: `dropManualMemoization` at 169,
// `inlineImmediatelyInvokedFunctionExpressions` at 173, `mergeConsecutiveBlocks` at 180, and
// `enterSSA` at 188. Single static assignment sees the spliced graph and versions it once. Nothing
// upstream constructs twice.
//
// A second run is not merely redundant here, it is destructive. `CopyNestedBodyInto` carries the
// nested function's phis across with their operands and blocks remapped, and those are exactly the
// phis a rebuild cannot re-derive: both arms of a logical define one temporary, which is versioned
// per block rather than merged, so the join has nothing to merge. Measured on
// `useMemo(() => p.a.b ?? [], [p.a?.b])`, the splice produces a correct phi at the join and a second
// construction destroys it, leaving a join with no phi.
//
// A grep rather than a behavioural assertion, for the same reason as the memo-inclusive census: the
// hazard is a caller adding the re-run back, and no assertion inside this package can see a pipeline
// assembled elsewhere.
var constructAfterInline = regexp.MustCompile(
	`(?s)Inline[A-Za-z]*InvokedFunctionExpressions[A-Za-z]*\(function\)[^}]*?Construct\(function\)`)

func TestConstructionIsNotReRunAfterTheInline(t *testing.T) {
	t.Parallel()

	paths, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range paths {
		source, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		// This file names the pattern in prose.
		if filepath.Base(path) == "ssa_test.go" {
			continue
		}
		if constructAfterInline.Match(source) {
			t.Errorf("%s re-runs Construct after the inline; upstream runs single static "+
				"assignment once, AFTER the splice, and a second run destroys the phis the copy "+
				"carried", path)
		}
	}
}
