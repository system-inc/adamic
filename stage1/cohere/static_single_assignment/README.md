# static_single_assignment

cohere's `static_single_assignment` module in Adamic: the graph orderings (reverse postorder, predecessors, evaluation order), Braun's single static assignment construction, redundant phi elimination, the verifier and dominance, generic over a `GraphInterface<F, B, P>` adapter (Go's `Graph`) an IR implements on its own function, block and place types. It's the one source stage 2's flow graph and the React rules build on (#dnv6f2c), and it follows the Go module file for file:

| Go | Here |
| --- | --- |
| `static_single_assignment.go` | `static_single_assignment.ts`: the ids, `EdgeType`, `RoleType`, `PhiInterface` and `GraphInterface` |
| `graph.go` | `graph.ts`: `reversePostorder`, `markPredecessors`, `markEvaluationOrder` |
| `phi.go` | `phi.ts`: a phi's operands, sorted by predecessor |
| `construct.go` | `construct.ts`: `construct` |
| `eliminate.go` | `eliminate.ts`: `eliminateRedundantPhis` |
| `verify.go` | `verify.ts`: `verifySingleAssignment`, `collectSingleAssignmentStats`, `computeDominance` |

An IR uses it the way cohere's high-level IR uses the Go: implement `GraphInterface` once, then finalize (`reversePostorder`, `markPredecessors`, `markEvaluationOrder`, in that order) and `construct`, recursing into nested functions itself. Import each file directly; there is no index. The one seam that differs from Go's is how a pass renames a place: a visitor returns the place, and the adapter writes it back, since 0.1 has no pointers. The types carry the house suffixes, each comment naming the Go type it mirrors. Each file's header says where it differs from the Go and why, and GAPS.md what stage 0 couldn't lower.

## Running it

```sh
go test -v -count=1 -timeout 30m ./stage1/cohere/static_single_assignment > "$TMPDIR/ssa.log" 2>&1; echo "exit=$?"
```

It needs the cohere submodule at a commit that holds the module with its dominance fix (cohere 0cba6cd or later) and `go`, `node` and clang on the path, as the rest of the gate does. The test runs `testdata/cohere_side_test.go` inside cohere's module by overlay (with `GOWORK=off`, since the module is one of its own), which writes the cases and Go's answers. Then it runs the port on the same cases natively under ASan and UBSan, on Node and through the JavaScript backend, compares all three with Go byte for byte, checks the native port for leaks, and runs fourteen mutants, each of which must be caught. `TestConstructRefusesAnEntryWithPredecessors` holds construct's refusal of an entry some edge enters. `gaps_test.go` holds GAPS.md's programs where GAPS.md says they stand.

| Variable | What it does |
| --- | --- |
| `COHERE_STATIC_SINGLE_ASSIGNMENT_SEED` | the seed for the generated functions (20261007) |
| `COHERE_STATIC_SINGLE_ASSIGNMENT_GENERATED` | how many to generate (2000) |
| `COHERE_STATIC_SINGLE_ASSIGNMENT_GRAPHS` | a directory of `*.txt` cases files to run as well, such as a fresh export |

By hand, from the repository's root:

```sh
node oracle/node.mjs stage1/cohere/static_single_assignment/main.ts stage1/cohere/static_single_assignment/sample-cases.txt
go build -o /tmp/adamic ./cmd/adamic
/tmp/adamic build stage1/cohere/static_single_assignment/main.ts -o /tmp/ssa
/tmp/ssa stage1/cohere/static_single_assignment/sample-cases.txt
```

`main.ts` says what a cases file holds and what the port prints for each function.

## The corpus

1. **The module's own tests.** `cohere_side_test.go` builds the ten functions `static_single_assignment_test.go` builds, with that file's own helpers, line for line, so the Go tests are ported by construction.
2. **React Compiler's fixtures.** `testdata/react_graphs.txt.gz` is every function cohere's high-level IR lowers React Compiler's vendored fixtures to (355 fixtures, Flow left out, 695 functions with the nested ones), as the IR hands them to construction, read through the IR's own adapter. It was exported by cohere c18103d (`TestExportGraphsForAdamic`, on 7945d10), which runs only when its test flag names a directory:

   ```sh
   cd cohere && go run ./command/cohere-dev test -count=1 -run '^TestExportGraphsForAdamic$' \
     ./internal/lint/ecmascript/high_level_intermediate_representation -args -export-graphs /tmp/react-graphs
   cd /tmp/react-graphs && ls *.txt | LC_ALL=C sort | xargs cat | gzip -9 -n > <this directory>/testdata/react_graphs.txt.gz
   ```

   Go's answers are computed fresh from it on every run, so it's a corpus of shapes, not an expectation; re-export it when the lowering changes enough to matter.
3. **Generated functions.** 2,000 from a fixed seed: diamonds, loops, self loops, unreachable blocks, structural fallthroughs, exceptional edges, edges to blocks that aren't there, gaps in block ids, parameters, a returns place, captured bindings written by context stores, terminals with places, and temporaries read where their definition may not dominate or defined twice, for the verifier.
