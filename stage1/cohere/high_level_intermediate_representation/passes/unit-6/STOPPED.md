# Unit 6 stopped at shared graph contracts

Base: stage1-hir/wip 9addb0e8a48b1cf60621ae24ef2196c860ff33ce.
Oracle: cohere 7945d102a6c18dd36adf9114a758ce646e8b2359.
Only passes/unit-6 is changed. This is a dependency reproducer and tagged Go
fixture adapter, not an Adamic implementation or a completed unit certificate.

## Observed stop and reproducer

`oracle_test.go:46` constructs the smallest graph used here: one array allocation,
one identifier and a return. It supplies Go ranges and a singleton disjoint set,
then calls the real Go assignment, method alignment, align/merge, validation and
scope-terminal passes. No earlier Adamic analysis is substituted for Go input.

`testdata/single-scope.before.checkpoint` is the input immediately before
`BuildReactiveScopeTerminals`; `testdata/single-scope.after.checkpoint` is its
actual output. The latter contains one Scope terminal, one split original block,
zero rekeyed phi operands, and the range after finalization. The adapter exports
scope identities, range/group/member bindings and the outcome through the
owner's framing, and requires deterministic bytes on repeat encoding. The
fixture is synthetic and is not an original census record.

`replay_contract.a` imports `../../replay/index.ts`. All three executions match
the Go input graph byte for byte. All three executions of the Go after-state
stop with `unknown terminal Scope`. See validation/node-after.txt,
validation/emitted-javascript-after.txt and validation/native-after.txt.
The native executable was built with `--sanitize`; the successful prepass
execution enabled ASan/UBSan/LSan. There is no sanitizer success claim for
execution of an unimplemented pass.

The stopped-state comparison fails when Scope decoding becomes available,
requiring this assertion to be removed and the real pass certificate run.
It never labels the stopped after-state a matching graph.

## Requests to hir-01

1. Add the Go Scope terminal (`cohere/internal/lint/ecmascript/high_level_intermediate_representation/terminal.go:324`)
   to the shared TerminalType (`core.ts:90`), carrying a checked ScopeIndex,
   body BlockIndex and fallthrough BlockIndex, with Go-ID translation defined
   at the shared arena boundary. ScopeIndex already exists; do not add a second
   index class or use a bare numeric identity in the terminal. Add its fresh
   decoder (`replay/decode.ts:134` currently refuses it), dump/encode, clone and
   edge-adapter support. Go's body is a real successor; its fallthrough is a
   structured fallthrough edge. The sidecar schema/ScopeIdentity stays unit 6's
   responsibility; the small fixture encoder here is not its full final schema.
2. Expose a non-SSA finalization API using the existing SSA adapter/imports:
   reverse postorder, predecessors, evaluation order, in that order, after a
   lane rewrites blocks and rekeys phis. Go calls `Finalize(function)` at
   `cohere/internal/lint/ecmascript/high_level_intermediate_representation/scope_terminals.go:640`.
   `graph.ts:106` exports only constructHIR, which also calls construct at line
   114 and recurses. Calling that after scope splitting would reconstruct SSA,
   contrary to Go. No private adapter or copied SSA algorithm is introduced here.
3. Confirm the landed import path/API for the existing mutation_aliasing port.
   `stage1/cohere/mutation_aliasing/` is absent at this base. Go unit 6 uses
   `mutation_aliasing.MutableRange` and `MutableRanges` by reference:
   `cohere/internal/lint/ecmascript/high_level_intermediate_representation/scopes.go:264`,
   `scope_terminals.go:132`, and `cohere/mutation_aliasing/ranges.go:183,266`.
   Unit 5's public range/disjoint replay schema is also absent in this checkout.
   Do not create a private range implementation or guess the module's exports.
   This is an import/schema request, not a requirement for unit 5's Adamic
   algorithm to run first. Go state can supply the earlier analysis answers.

No owner-owned file, census exporter, shared arena, analysis module or cohere
production source is edited. No confirmed Adamic language gap is reported.

## Verification and limits

Setup: `bash cloud/setup.sh`, then `source /workspace/adamic-tools/env.sh`.
Setup passed in 40.114 seconds; nproc 5, cgroup quota 4 CPUs. Complete timing
lines are in validation/setup.txt.

Run from the repository root:

```
go test -v -count=1 -timeout 15m ./stage1/cohere/high_level_intermediate_representation/passes/unit-6
gofmt -l stage1/cohere/high_level_intermediate_representation/passes/unit-6/*.go
go vet ./stage1/cohere/high_level_intermediate_representation/passes/unit-6
```

Final comparison: PASS, 31.891 seconds. Formatting and vet: exit 0, no output.
The tagged adapter is overlaid beside Go HIR and imports the owner's printer,
framing and input-fact files by path, without modifying them. Its top-level
fixture test and this lane's top-level comparison test both call t.Parallel.
Checked-in fixtures are compared with freshly generated Go bytes on every run.

| Unit 6 certificate | Matched originals | Required originals |
| --- | ---: | ---: |
| Node | 0 | 1,465 |
| Emitted JavaScript | 0 | 1,465 |
| Sanitized native | 0 | 1,465 |

One synthetic prepass graph decodes exactly on each runtime. One synthetic
postpass graph reproduces the missing shared Scope contract on each runtime.
Neither is counted as a successful pass. The full census, the 23 Flow inputs,
all scope/alignment algorithms, declarations/outputs, memoization graph/levels,
error outcomes and semantic pass mutant remain unexecuted/unported. No semantic
mutant certificate is claimed. No full HIR or repository gate was run.
