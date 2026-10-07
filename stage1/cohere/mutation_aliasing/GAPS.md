# What stage 0 couldn't lower in cohere's mutation_aliasing module

The port beside this file is cohere's `mutation_aliasing` module (mutation_aliasing.go and ranges.go), written as 0.1 Adamic, for #dnv6f2c: the effect vocabulary and React's mutable ranges, which cohere's high-level IR and Adamic's flow graph share, so the React rules have one source in Adamic. The gaps were found on `codex/cohere-bump-ca939db` (1ef4b34), with cohere at its main, febfb41.

Stage 0 refused the port three times, each for a gap it already knows and another slice already holds; none is new, so none went to @system_adamic. Each is below as the smallest program that shows it, stage 0's own message, and the way the port went around it. Each workaround is marked in the port with its gap number (`grep -n "gap N" *.ts`). The programs are in `gaps/`, and `gaps_test.go` holds them to this file, so the test says when one closes and its workaround can go.

## How it's held

`mutation_aliasing_test.go` runs `testdata/cohere_side_test.go` inside cohere's module by overlay. That side writes the cases the port reads: the module's own tests' functions, built with those tests' own helpers, and their helper tests as probes; one function rebuilding cohere's own focused test of created-from; the 1,010 functions React Compiler's own fixtures lower to in cohere's high-level IR, with their effects (`testdata/react_ranges.txt.gz`, README.md says how it's exported); and 2,000 functions generated from a seed. It checks each exported function gets the ranges cohere's own pass gave it, runs the module's passes on every case, unchanged, and prints what they made. The port (`main.ts`, on an IR of its own, the module tests' IR with what cohere's IR answers beside it) prints the same, natively under the sanitizers, on Node and through the JavaScript backend, and all three must match Go byte for byte, with the native port leaking nothing. Twenty-four mutants, at least one per algorithm (the vocabulary, a range's two predicates, the four graph builders' rules, identity and freezing, the phi kinds, the mutation gate, the deferred phi operand, the worklist four times, the widening's order guard, the definition half six times, and the block's first order), must each make the port disagree with Go, natively and on Node, without failing.

```sh
go test -v -count=1 -timeout 30m ./stage1/cohere/mutation_aliasing > "$TMPDIR/mutation_aliasing.log" 2>&1; echo "exit=$?"
```

## 1. An empty array as the fallback of ??

```ts
for(const each of lists.get(2) ?? []) {
```

```
gaps/1_empty_array_fallback.ts:4:35: stage 0 can't lower an array of never yet        (Node prints 9)
```

tsc types the literal `never[]`. The css, json and values slices hold the same gap.

**Around it:** ranges.ts falls back to typed empty constants (`noIdentifiers`, `noMembers`, `noPendingPhiOperands`), and the deferred phi operands are added with an `if`.

## 2. A method called through ?.

```ts
nodes.get(1)?.edges.push(7);
```

```
gaps/2_call_through_optional_chain.ts:6:1: stage 0 can't lower a call through ?. (an optional call) yet        (Node prints 1)
```

`internal/lower/lower_test.go` pins it, and the single assignment slice met it too.

**Around it:** `AliasingState.createFrom` reads the source node first and pushes onto its edges only when there is one.

## 3. panic as an arm of a conditional expression

```ts
console.log(gap === 'LoopCarriedInversion' ? 'loop-carried-inversion' : panic(`an unnamed gap: ${gap}`));
```

```
gaps/3_panic_in_a_conditional.ts:7:77: stage 0 can't lower reading panic yet        (Node prints loop-carried-inversion)
```

The gitignore, markdownblocks and selector slices hold the same gap.

**Around it:** the driver names a range gap with a switch, `rangeGapName` (main.ts).

## The cycle rule's cost

Not gaps: stage 0 is right by its rule (docs/memory.md, "Cycles") each time, and what the rule costs this port's shape is worth writing down, since an IR adopting the module will meet it.

- **A one-parameter function capturing the graph.** As in the single assignment slice (its `GAPS.md`), a function of one parameter that captures `graph` can be seen as `GraphInterface.identifierOf`, so the graph could hold the function that holds it. `buildAliasingGraph` gathers a phi's operands and a closure's captures with loops rather than `map`.
- **A place type a block can be seen as.** The driver's place was first `{ readonly id: IdentifierIdType }`. Its block has an `id` that is a number too, so a block could be seen as a place, stored in an instruction's place, and reach back to the instruction: stage 0 refused the place field as cycle-capable. The driver's place names its value `identifier`, which no block has.
- **Module-level code.** The driver's reader first ran as the module's own code, pushing each record into a module-level array. The cycle rule takes what module-level code writes as written into something no function made, and refused the array; as a function, `readCases`, the same code writes only into what it made.

## Where cohere's own lint and Adamic 0.1 disagree

`cohere --fix` formats the port to house style, and its `logical-assignment-operators` fix rewrote `hasFrozen = hasFrozen || ...` as `hasFrozen ||= ...`, which Adamic 0.1 refuses by design ("Adamic 0.1 refuses ||=; write the if"). The port writes the ifs (`derivePhiImmutable` in ranges.ts, and `closureOf` and `describe` in main.ts), which both accept. A port that runs the fixer will meet it again until the rule is off, or its fix writes the if, for files Adamic compiles.

## Where the port differs from the Go, and not because of stage 0

Each file says so at its top. In short: the kinds are unions of their names, with `aliasingEffectKindName` and `effectValueKindName` for Go's String methods; an effect's `from` is the place or `undefined`, so `hasFrom` is `from !== undefined`; a seam that may have nothing to say (`instructionOrder`, `returnValue`, `storedContextValue`, `closure`) answers `undefined` rather than a second result; `MutableRange` is a class whose fields never change, so a widening sets a new one; the alias node's back edges are Maps, which keep their insertion order without Go's key slices; and Go's definition walk struct is a class made once per call.
