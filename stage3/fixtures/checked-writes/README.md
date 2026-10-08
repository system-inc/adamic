# Checked writes through wider views

Step 10 (#7hc82dv). Fixtures derived from TypeScript 6.0.3, original commit
`050880ce59e30b356b686bd3144efe24f875ebc8`, Copyright Microsoft Corporation,
Apache-2.0. Copied/reduced declarations and statements are identified in each
fixture; drivers and verifier are new. The survey is pinned to Adamic commit
`9b8ebd77bb4fd3994fc037d2a3e9f68980132797`, whose adapted source commit is
`12fef29628bfda6cf8bec4ba2a326fa4ec4a66d5`.

The ruled .ts contract is a check at the write against the field's actual shape
type: fitting values are stored, misfits exit 70 before the subsequent output,
naming the field and both types. These files remain authored as .a. Their first
line records the strict .a refusal measured on main `45487a80`; the verifier
copies their unchanged text to scratch .ts files for compatibility execution.
Node executes each original .a through oracle/node.mjs, independently of Adamic.
The different Node and Adamic outcomes on negative twins are intentional.

## Selection and provenance

The survey's coarse `family` has only three groups. This unit groups all **515
class-d records by source_family**, ranks by descending class-d count, and breaks
ties lexically. The largest 20 cover **381 sites**; 134 sites in smaller families
are outside this corpus. manifest.json contains every selected site ID, one full
representative record per family, original locations/reasons, reduction limits,
expected outputs, and the exact input mutant for each pair. With --survey, the
verifier checks the entire survey digest and recomputes the ranking and IDs.

Class d means unresolved, not a proved unsafe or reachable upstream write. The
fixtures preserve the relevant tsc type relation and use small boundary drivers.
They do not pretend these incompatible driver calls occur in tsc's test suite.
Most unrelated fields are omitted. SourceFile retains fileName; SymbolTable's
values are numbers; NodeFlags use numeric literal domains; generated emit data
retains autoGenerate; flow variants retain the node kind. The Expression and
NodeArray position drivers legally narrow readonly pos to 0. setTextRangePos's
body is copied unchanged. Compound flags writes, the emitNode clear, array
mutators, and other field stores are extracted or parameterized as stated in the
individual fixture comments. Diagnostic-file and tracker writers are new drivers
based on the real field declarations, not copied upstream functions.

There is no fitting element write to never[]. Its positive is explicitly a
**no-write empty-sentinel control**, and its negative is a push into the same
sentinel. An inhabited positive element write for that family is impossible.
The three array-element families and NodeArray's array-plus-range shape also
exercise boundaries beyond a plain named-field store; they are retained because
the task selected families by count, not by compiler support.

| Rank | Source family | Class-d sites | Pair | Input mutant caught by exact Node stdout |
|---:|---|---:|---|---|
| 1 | DiagnosticWithLocation | 144 | [in](01_located-diagnostic_in.a), [out](01_located-diagnostic_out.a) | `file` to `undefined` |
| 2 | never[] | 82 | [in](02_shared-empty_in.a), [out](02_shared-empty_out.a) | `false` to `true` |
| 3 | FlowNode | 26 | [in](03_flow-node_in.a), [out](03_flow-node_out.a) | `binary` to `binding` |
| 4 | ResolvedType | 17 | [in](04_resolved-members_in.a), [out](04_resolved-members_out.a) | `members` to `undefined` |
| 5 | T | 17 | [in](05_generic-range_in.a), [out](05_generic-range_out.a) | `0` to `1` |
| 6 | NodeBuilderContext | 11 | [in](06_builder-tracker_in.a), [out](06_builder-tracker_out.a) | `tracker` to `fallback` |
| 7 | DiagnosticWithLocation[] | 9 | [in](07_diagnostic-array_in.a), [out](07_diagnostic-array_out.a) | `located` to `detached` |
| 8 | GeneratedIdentifier | 8 | [in](08_generated-identifier_in.a), [out](08_generated-identifier_out.a) | `emitNode` to `undefined` |
| 9 | Identifier | 8 | [in](09_identifier-flags_in.a), [out](09_identifier-flags_out.a) | `0` to `1` |
| 10 | Expression | 7 | [in](10_expression-range_in.a), [out](10_expression-range_out.a) | `0` to `1` |
| 11 | Node | 7 | [in](11_node-flags_in.a), [out](11_node-flags_out.a) | `1` to `0` |
| 12 | CapturedThis | 6 | [in](12_captured-this_in.a), [out](12_captured-this_out.a) | `emitNode` to `undefined` |
| 13 | Declaration[] | 6 | [in](13_declaration-array_in.a), [out](13_declaration-array_out.a) | `declaration` to `expression` |
| 14 | NodeArray<Statement> | 6 | [in](14_statement-array-range_in.a), [out](14_statement-array-range_out.a) | `0` to `1` |
| 15 | DiagnosticWithDetachedLocation | 5 | [in](15_detached-diagnostic_in.a), [out](15_detached-diagnostic_out.a) | `undefined` to `file` |
| 16 | FlowArrayMutation \| FlowAssignment | 5 | [in](16_flow-assignment-union_in.a), [out](16_flow-assignment-union_out.a) | `binary` to `binding` |
| 17 | GeneratedIdentifier \| GeneratedPrivateIdentifier \| Node | 5 | [in](17_generated-node-union_in.a), [out](17_generated-node-union_out.a) | `emitNode` to `undefined` |
| 18 | SyntheticSuper | 5 | [in](18_synthetic-super_in.a), [out](18_synthetic-super_out.a) | `emitNode` to `undefined` |
| 19 | Mutable<GeneratedIdentifier> | 4 | [in](19_mutable-generated_in.a), [out](19_mutable-generated_out.a) | `emitNode` to `undefined` |
| 20 | Diagnostic | 3 | [in](20_diagnostic-union_in.a), [out](20_diagnostic-union_out.a) | `file` to `undefined` |

## Expectations and observations

manifest.json separates the Node truth from the required .ts safety outcome.
Node expectations are exact stdout/stderr/exit triples. All negatives store the
misfit silently and exit 0. Required Adamic negatives exit 70, preserve only
`before write\n` on stdout, and have an `adamic: panic:` diagnostic naming the
field and both types. Positive Adamic runs must match Node exactly. Diagnostic
wording has not been invented: the six supported negative diagnostics are pinned
byte-for-byte; unsupported cases specify required diagnostic content until they
can run. observations.json records only what actually ran.

Observed on main `45487a809f89885a3fc651cd590e7dabf31362dc`:

- 40 exact Node observations passed; 40 strict .a refusal headers matched.
- 20 real-input mutants ran: change each positive's fitting argument to the
  negative's misfit (or enable never[]'s write). All exited 0 with empty stderr;
  each was caught by the positive's exact stdout check, then independently
  matched its negative twin's Node expectation.
- 80 header-predicate mutants were rejected: remove each header or replace its
  reason. This exercises this verifier's exact predicate, not cohere's Gate.
- Node's negative silent-store outputs are also rejected by the .ts runtime
  acceptance predicate, modeling a missing check. No implementation mutant is
  claimed from this predicate test.
- All 40 .ts build attempts were refused by main's existing invariant-mutable
  check before execution. **No exit-70 native/JS behavior is verified yet.**
  The strict --require-checked-writes run exits 1 and names all 40 pending files;
  early refusal is never accepted as a successful checked write.

main-observations.json retains that main baseline. observations.json retains
the latest dependency run with complete diagnostics, binary hash, source hashes,
Node outputs, per-pair mutants, runtime results, native counts, and totals. counts.md is this corpus's refreshed
scoreboard. These programs are not registered in internal/oracle, so its shared
counts table and registry were not changed; native allocation counts are recorded locally for the 12 supported compatibility
inputs, and remain unavailable for the others.

## Reproduce

Run from Adamic's root, sourcing the setup script's printed environment:

```sh
source /workspace/adamic-tools/env.sh
git show 9b8ebd77:stage3/adapt/71-writable-views/language-decision/results.json > /tmp/step10-results.json
go build -o /tmp/step10-adamic ./cmd/adamic > /tmp/step10-build.log 2>&1
python3 stage3/fixtures/checked-writes/verify.py --compiler /tmp/step10-adamic --survey /tmp/step10-results.json --observe > /tmp/step10-verify.log 2>&1
python3 stage3/fixtures/checked-writes/verify.py --compiler /tmp/step10-adamic --survey /tmp/step10-results.json --require-checked-writes > /tmp/step10-runtime-gate.log 2>&1
```

The build and observation run exited 0. The final strict runtime gate exited 1
on this base, as described above. Pass a compiler built from #63x2441 to discharge
the runtime gate. Successful builds run native with ASan/UBSan, release native with allocation
counting, and the generated JavaScript through oracle/node.mjs on Node. Finished positives enable leak detection; panic negatives
skip leak detection but retain address/undefined-behavior checks, because panic
intentionally ends without cleanup. The expected stdout, exit, and panic content
must match in both backends. Every subprocess has a 120-second timeout and runs
sequentially. --observe refreshes evidence; --require-checked-writes is the strict
acceptance mode and does not refresh it. No whole Go package or full gate ran.

Setup passed after setting `GOPROXY=https://proxy.golang.org|direct`. Timings:
Go 0.095s, Node 0.106s, clang 0.529s, markdown dependencies 0.951s, submodules
4.401s, build cache 177.348s, total 177.376s; nproc 5, cpu.max 400000 100000.
Go 1.27.1, Node 24.19.0, clang 20.1.8. The initial recursive `git fetch origin`
was interrupted while fetching nested TypeScript revisions; fetching main with
fetch.recurseSubmodules=false succeeded, and setup installed the pinned
submodules normally. No compiler implementation, other adaptation, runtime,
reference baseline, or shared gate was edited.

## Implementation dependency proof

Found `origin/codex/checked-wider-writes` at
`6a093f039c0aa71105c857b2187c6fe680d5c607` (#63x2441), and built it in an isolated
worktree. It uses the same cohere gitlink as fixture main. No dependency commit
was merged into this fixture branch. Its README explicitly retains container,
callable and structural-reference refusals.

```sh
git worktree add --detach /workspace/step10-compiler origin/codex/checked-wider-writes
# Reuse the installed matching cohere via a symlink in this external worktree.
go -C /workspace/step10-compiler build -buildvcs=false -o /tmp/step10-checked-adamic ./cmd/adamic > /tmp/step10-checked-build.log 2>&1
python3 stage3/fixtures/checked-writes/verify.py --compiler /tmp/step10-checked-adamic --compiler-revision 6a093f039c0aa71105c857b2187c6fe680d5c607 --survey /tmp/step10-results.json --observe > /tmp/step10-checked-verify.log 2>&1
python3 stage3/fixtures/checked-writes/verify.py --compiler /tmp/step10-checked-adamic --compiler-revision 6a093f039c0aa71105c857b2187c6fe680d5c607 --survey /tmp/step10-results.json --require-checked-writes > /tmp/step10-checked-runtime-gate.log 2>&1
```

VCS stamping initially failed on the shared submodule symlink; -buildvcs=false
resolved it. An initial generated-JavaScript invocation bypassed the runtime
resolver and failed module lookup. The verifier now uses oracle/node.mjs for
both source and generated JavaScript; the corrected complete run passed.

Observed: **12 of 40 runtime contracts pass**, in these six families:
DiagnosticWithLocation, T, Identifier, Expression, Node, Diagnostic. Each fitting
fixture matches Node in sanitized native, counted release native, and JavaScript.
Each negative exits 70 with exactly the pinned panic in all three. The six
finished sanitized positives pass leak detection, and counted allocations equal
frees plus regions. counts.md includes all 12 native allocation rows.

**26 .ts inputs remain Refused; 2 remain NotYet** (NodeArray's array-plus-range
representation). These are 14 complete pairs, not failed in-type computations
or silent miscompiles. The strict gate exits 1 on exactly those 28 files. No
claim of all-20-family runtime coverage is made.

mutant.go.txt is a small tool compiled with the dependency. It clears actual
ir.SetProperty.WriteCheck values after lowering, verifies at least one check was
removed, and builds sanitized native and JavaScript. It never edits the compiler.

```sh
mkdir -p /workspace/step10-compiler/stage3/fixtures/checked-writes/mutant-tool
cp stage3/fixtures/checked-writes/mutant.go.txt /workspace/step10-compiler/stage3/fixtures/checked-writes/mutant-tool/main.go
go -C /workspace/step10-compiler build -buildvcs=false -o /tmp/step10-mutant ./stage3/fixtures/checked-writes/mutant-tool > /tmp/step10-mutant-build.log 2>&1
python3 stage3/fixtures/checked-writes/mutants.py --tool /tmp/step10-mutant > /tmp/step10-runtime-mutants.log 2>&1
```

Both commands exited 0. Each of the six supported pairs caught its real
check-removal mutant by the **exit-70 runtime comparison in both backends**.
The mutant instead exits 0, stores silently, and produces exactly the negative
twin's Node output. Sanitizers and leak detection pass, so an unrelated compiler
error or memory failure did not kill it. runtime-mutants.json records all six
catches and names the 14 blocked families. The separate 20 fitting-argument
mutants remain caught by exact Node stdout, one per selected pair.

## Remaining proof

The checked-write compiler dependency is not on this origin/main. Runtime
checks and true implementation mutants for the 14 larger-contract families
remain blocked, rather than claimed green. The original 134 smaller
family sites, full tsc computation, arbitrary heap aliases, callback escapes,
recursive shapes, and actual reachability of class-d sites are not covered.
