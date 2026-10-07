# mutation_aliasing

cohere's `mutation_aliasing` module in Adamic: React's mutation and aliasing model, the effect vocabulary and the mutable ranges (React's `InferMutationAliasingRanges`: the alias graph the effects build, the mutate worklist over it, and the definition half), generic over a `GraphInterface<F, B, P>` adapter (Go's `Graph`) that extends the single assignment port's. It's the second of #dnv6f2c's two slices, the one source the React rules build on, and it follows the Go module file for file:

| Go | Here |
| --- | --- |
| `mutation_aliasing.go` | `mutation_aliasing.ts`: `AliasingEffectKindType`, `AliasingEffectInterface` and its three constructors, `EffectValueKindType`, `GraphInterface` and `OptionsInterface` |
| `ranges.go` | `ranges.ts`: `MutableRange`, `MutableRanges`, `inferMutableRanges`, `buildAliasingGraph` and `AliasingGraph`, the aliasing state and its `mutate` worklist, `validateMutableRanges`, `blockFirstOrder`, `rangeGaps` |

An IR uses it the way cohere's high-level IR uses the Go: implement `GraphInterface` once, over a function already in single assignment form with its evaluation order (`../static_single_assignment`), and call `inferMutableRanges` per function, nested ones each on their own. Import each file directly; there is no index. Where the port differs from the Go, each file's header says so and why: the kinds are unions of their names, a seam that may have nothing to say answers `undefined` rather than a second result, an effect's `from` is the place or `undefined`, and a range never changes once made. `GAPS.md` says what stage 0 couldn't lower and what the cycle rule cost.

## Running it

```sh
go test -v -count=1 -timeout 30m ./stage1/cohere/mutation_aliasing > "$TMPDIR/mutation_aliasing.log" 2>&1; echo "exit=$?"
```

It needs the cohere submodule at a commit that holds the two modules with the single assignment module's dominance fix (cohere 0cba6cd or later) and `go`, `node` and clang on the path, as the rest of the gate does. The test runs `testdata/cohere_side_test.go` inside cohere's `mutation_aliasing` module by overlay, in the package of the module's own tests and with `GOWORK=off`, which writes the cases and Go's answers. Then it runs the port on the same cases natively under ASan and UBSan, on Node and through the JavaScript backend, compares all three with Go byte for byte, checks the native port for leaks, and runs twenty-four mutants, each of which must be caught by its answers. `gaps_test.go` holds `GAPS.md`'s programs where `GAPS.md` says they stand.

| Variable | What it does |
| --- | --- |
| `COHERE_MUTATION_ALIASING_SEED` | the seed for the generated functions (20261007) |
| `COHERE_MUTATION_ALIASING_GENERATED` | how many to generate (2000) |
| `COHERE_MUTATION_ALIASING_GRAPHS` | a directory of `*.txt` cases files to run as well, such as a fresh export |

By hand, from the repository's root:

```sh
node oracle/node.mjs stage1/cohere/mutation_aliasing/main.ts stage1/cohere/mutation_aliasing/sample-cases.txt
go build -o /tmp/adamic ./cmd/adamic
/tmp/adamic build stage1/cohere/mutation_aliasing/main.ts -o /tmp/mutation_aliasing
/tmp/mutation_aliasing stage1/cohere/mutation_aliasing/sample-cases.txt
```

`main.ts` says what a cases file holds and what the port prints for each function: the block orders, every node of the alias graph with its edges and their indices, each mutation collected, and per value its range, its range after the widening alone, its abstract kind and where it contains.

## The corpus

1. **The module's own tests.** `cohere_side_test.go` builds the functions `mutation_aliasing_test.go` builds, with that file's own helpers, line for line, so the Go tests are ported by construction. The two tests that call the aliasing state's methods directly (the phi kinds and freezing) become one function per case whose effects make the same state, and the helper tests (`phiOpensBefore`, `phiOpenedRange`, the vocabulary, a range's predicates) become `probe` and `vocabulary` records. One more function rebuilds cohere's `TestCreateFromMutationPropagatesTransitively` (its IR's `ranges_test.go`), the only shape that shows a created-from keeping the mutation's own kind.
2. **React Compiler's fixtures, with their effects.** `testdata/react_ranges.txt.gz` is 1,010 functions: every function cohere's high-level IR lowers React Compiler's vendored fixtures to (355 fixtures, Flow left out, 695 functions with the nested ones) as `ForFunction` constructs them, and 315 more where `ForFunctionWithoutManualMemoization` prepares a fixture that names `useMemo` or `useCallback` (named `~memo`), each read through the IR's own range adapter, effects and all. Two of the adapter's answers aren't facts of the function: what `closure` says depends on the kinds the pass holds when it asks, so the export records the answer cohere gave at each closure and the case replays it; and each function carries the ranges cohere's own pass gave it, which the Go side checks it reproduces before it writes an answer, so the cases are shown to be cohere's. It was exported by cohere's `TestExportRangesForAdamic`, which runs only when its test flag names a directory:

   ```sh
   cd cohere && go run ./command/cohere-dev test -count=1 -run '^TestExportRangesForAdamic$' \
     ./internal/lint/ecmascript/high_level_intermediate_representation -args -export-ranges /tmp/react-ranges
   cd /tmp/react-ranges && ls *.txt | LC_ALL=C sort | xargs cat | gzip -9 -n > <this directory>/testdata/react_ranges.txt.gz
   ```

   Go's answers are computed fresh from it on every run, so it's a corpus of shapes, not an expectation; re-export it when the lowering or the effect inference changes enough to matter.
3. **Generated functions.** 2,000 from a fixed seed, finalized by the single assignment module: one to eight blocks with loops, joins and exceptional edges; parameters, captured values and a returns place; phis at joins, with an operand from a block that never precedes them; every effect kind over values made before, after or never; Creates of every value kind; closures under each rule (a recorded answer, frozen when every capture is immutable, or frozen when a named earlier function's alias graph, built with the captures' kinds as its context's, collects no mutation); stores into captured bindings; returned values; an instruction the function doesn't hold; a block finalize never numbered; and both options. One in six runs as `finalize-ranges`, so the driver finalizes it again.

## What the corpus can't show

The mutation kind a walk carries, conditional or definite, has no mutant. Within this pass it decides only whether `mutate` visits a node again, and a node is widened and its edges followed the same way at either kind, so no range depends on it; upstream reads it for its diagnostics, which this module doesn't make. Each collected mutation's kind is in the answers.
