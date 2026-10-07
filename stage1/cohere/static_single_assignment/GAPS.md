# What stage 0 couldn't lower in cohere's static_single_assignment module

The port beside this file is cohere's `static_single_assignment` module (static_single_assignment.go, graph.go, phi.go, construct.go, eliminate.go and verify.go), written as 0.1 Adamic, for #dnv6f2c: the graph orderings, Braun's construction, redundant phi elimination, the verifier and dominance that cohere's high-level IR and Adamic's flow graph share, so stage 2's flow graph and the React rules have one source in Adamic. The gaps were found on main f8013f0 with cohere 7945d10 (`codex/cohere-bump-ca939db`, 1ef4b34), the first pin that holds the module.

Every place stage 0 refused the port is below, as the smallest program that shows it, stage 0's own message, and the way the port went around it. Each workaround is marked in the port with its gap number (`grep -n "gap N" *.ts`). The programs are in `gaps/`, and `gaps_test.go` holds them to this file.

## How it's held

`static_single_assignment_test.go` runs `testdata/cohere_side_test.go` inside cohere's module by overlay. That side writes the cases the port reads: the module's own ten tests' functions, built with those tests' own helpers; the 695 functions React Compiler's own fixtures lower to in cohere's high-level IR (`testdata/react_graphs.txt.gz`, README.md says how it's exported); and 2,000 functions generated from a seed. It runs the module's passes on each, unchanged, and prints what they made. The port (`main.ts`, on an IR of its own, the module tests' IR with a terminal's places added) prints the same, natively under the sanitizers, on Node and through the JavaScript backend, and all three must match Go byte for byte, with the native port leaking nothing. Fourteen mutants, at least one per algorithm (reverse postorder twice, predecessors, evaluation order, construction three times, elimination, the operand order, the verifier three times, and dominance twice: over the block array, and one round rather than its fixed point), must each make the port disagree with Go, natively and on Node, without failing.

```sh
go test -v -count=1 -timeout 30m ./stage1/cohere/static_single_assignment > "$TMPDIR/ssa.log" 2>&1; echo "exit=$?"
```

README.md names the variables that ask for other generated functions or run a fresh export as well.

## 1. An arrow with a block body, returning a class

A function declaration returning a class instance lowers, and so does an arrow whose body is an expression. An arrow whose body is a block, returning the same instance from a `return`, is refused as a nominal-class view:

```ts
class Box {
	readonly id: number;
	constructor(id: number) {
		this.id = id;
	}
}
const make = (box: Box): Box => {
	return new Box(box.id + 1);
};
console.log(`${make(new Box(1)).id}`);
```

```
gaps/1_block_arrow_returning_class.ts:7:33: Adamic 0.1 refuses a value without nominal ancestry seen as Box; construct that class or a subclass; use an interface for structural values (adamic/nominal-class)        (Node prints 2)
```

The place it names is the body's `{`. `viewSite` (internal/lower/invariance.go) takes an arrow's body as a view of its return type whatever the body is, which is right for an expression and wrong for a block: `classViewRefusal` then asks what type the block itself has, and a block is no `Box`. The `return` inside is a view site of its own and is judged there. It's an inference from reading the code that the case wants `node.Kind != ast.KindBlock`; the program is the evidence.

**Around it:** the driver's adapter makes its placeholder block in a named function, `placeholderOf` (main.ts), which the adapter's `placeholder` arrow calls as an expression.

## 2. A call through an optional chain

```ts
const found = states.get(0)?.definitions.get(1);
```

```
gaps/2_call_through_optional_chain.ts:8:15: stage 0 can't lower a call through ?. (an optional call) yet        (Node prints 7 undefined)
```

Stage 0 knows it: `internal/lower/lower_test.go` pins the same message for a method called through `?.`. It's here because the port met it.

**Around it:** `renameReturns` (construct.ts) reads the block's state first and asks it only when there is one.

## The cycle rule's cost

Not a gap: stage 0 is right by its rule (docs/memory.md, "Cycles"), and what the rule costs this port's shape is worth writing down, since every IR that adopts these passes will meet it.

The Go hands its Graph callbacks that capture the graph: a visitor that renames a place through `graph.WithIdentifier`. A variable a function value captures is a slot, and the finder follows a function type to every function value in the program that can be seen as it. A visitor written `(place) => graph.withIdentifier(...)` takes one parameter, and a function of one parameter can be seen as one of two, so it can be seen as Graph's own `withIdentifier: (place: P, identifier: IdentifierId) => P`. The graph could then hold the visitor that holds the graph, and the finder refuses `graph`:

```
eliminate.ts:96:33: Adamic 0.1 refuses 'graph', a variable a function value captures and can be reached from what it holds, ... (adamic/cycle-capable)
```

Taking the role it ignores, `(place, _role) => ...`, makes the visitor a function of a `Role` second parameter, which no `IdentifierId` can be passed as, and the finder agrees nothing in the graph can hold it. Every visitor in the port is written with both parameters for that reason. `reversePostorder`'s `inRange` met the same rule more plainly (a `(id) => boolean` captured by the `retain` callback of the same type) and is a function declaration.

## Where the port differs from the Go, and not because of stage 0

Each file says so at its top. In short: the ids are numbers, Edge and Role are unions of their names, a visitor renames by returning the place (0.1 has no pointers), `returns` gives the place or undefined and `setReturns` writes it back, a phi's operands are an array the functions in phi.ts return anew, and the Go's pooled scratch memory (Go's costs, measured on Go) is working memory made per call.

## A precondition the Go didn't state (stated and checked since cohere 0cba6cd)

Construction's lookup assumes no edge enters the entry block, which both real IRs hold. A function whose entry has predecessors can put the entry on a cycle of blocks with one predecessor each, and a lookup around it after sealing recursed until the stack ran out, in Go (a fatal stack overflow, found by the first generated corpus) as in the port. Since cohere 0cba6cd the rule is in `GraphInterface.entry`'s contract: `verifySingleAssignment` reports a break as `EntryHasPredecessors`, and `construct` refuses one with a panic naming the entry and its predecessors. The generator now gives the finalizer-only functions edges into their entry, so the corpus shows the violation; `TestConstructRefusesAnEntryWithPredecessors` holds the refusal natively, on Node and through the JavaScript backend.

## Where the Go's dominance was wrong, and the port with it (fixed in cohere 0cba6cd)

Not a stage 0 gap, and not the port's: the port answered as Go cohere's module did, which is what it's held to. The port's corpus is where it showed.

`ComputeDominance` is Cooper-Harvey-Kennedy, which needs the blocks in a reverse postorder of the real edges: every block after some predecessor, so each immediate dominator comes before its block. It read the block slice for that order. `ReversePostorder` visits a structural fallthrough first, as upstream does, and keeps a block where the fallthrough first reached it, so a block whose real predecessors all come later (a loop reached through a fallthrough before its back edge) sat ahead of every one of them, and the fixed point settled on dominators that were too few.

Measured against dominance computed the plain way (each block's dominators are itself and the intersection of its predecessors', iterated to a fixed point) over the same predecessors: 8 of the 2,000 generated functions got a wrong answer from Go, the port with it, and none of the 695 React functions did. `generated-631` at seed 20261007 was the smallest to read: blocks in the order 2 6 4 7 5, with 4's predecessors 4 and 5, 5's 7, 7's 6 and 4, and 6's 2 and 7. Every path to 4 runs 2, 6, 7, 5, 4, and Go said only 2, 6 and 4 dominate it.

cohere 0cba6cd (#6v4a54x) gave dominance its own reverse postorder of the real edges, leaving the block slice's order alone, and the port follows it (`realReversePostorder` in verify.ts). The same change states and checks the entry precondition below. The slice's mutant `dominance over the block array` puts the old order back and must be caught. A mutant that changes how the pass combines predecessors can make a block its own immediate dominator over the old order, and `dominates` never leaves it, which is why the other dominance mutant stops the fixed point after one round instead.
