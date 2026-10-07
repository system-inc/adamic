Built: the remaining claimed leaked-number-render judgment and traversal port, with raw numeric and constraint facts.
Commits: implementation 23fe1096; previous completed work is on origin through edf39f73.
Checks: differential suite PASS 40.953s; checker facts PASS 0.196s; production Go JSX matrix PASS 0.043s; vet and diff check empty.
Mutant: changing the decline return from false to true compiled and exited 0; independent Go bytes caught byte 6236.
Not covered: real native JSX source parsing and emitted-JavaScript parity remain shared integration gaps; two push attempts failed for missing HTTPS credentials.

The new request explicitly allowed porting everything about a rule whose shared
harness integration remains blocked. The native rule now implements the entire
production rendered/falsy traversal, child versus attribute/spread exclusions,
numeric literals including bigint, union decline, branded intersections, base
constraints and the production depth bound. No checker question returns a lint
verdict or a fix. New `numeric-literal` and `base-constraint-shape` questions
each have separate Go and .a implementations and one dispatcher case.
Shared parser/scanner files, registration generators, and existing test harness
files were not edited. The new standalone runner and test use existing helpers.

Native source parsing of real JSX remains unavailable on this branch. The
previous probe and its exit-70 refusal are recorded in
WAVE_15_CONTINUATION_REPORT.md and validation-wave-15-continuation.
`origin/codex/lint-harness-dot-a` is absent from the fetched origin heads.

For positive rule evidence, the control runner adds explicit raw JSX wrapper
nodes around parsed expression statements. The independent Go oracle adds the
same syntactic contexts and invokes the unchanged production rule callback.
The wrappers contain no type decisions, report decisions or edits. Leaves remain
the original checker-resolved syntax nodes, retaining contextual and narrowed
types. This checks native rule logic while leaving the real JSX parser gap
visible; it is not described as end-to-end JSX source agreement.

Fifteen .a control files cover numbers, bigint and literal unions, zero/nonzero
bigint literals, enums, branded numbers, nullish coalescing, logical chains,
ternaries, constraints, unconstrained generics, strings and mixed unions, any,
unknown, attributes, spread children, outside-JSX expressions, fragments,
Unicode and CRLF. Both normal and ASan/UBSan/LSan native runs match Go:

| Population | Findings | Identical bytes |
| --- | ---: | ---: |
| Wrapper controls | 22 | 8866 |
| Same 77 compiler roots | 0 | 5318 |
| Same frozen 287 repository roots | 0 | 18485 |

The zero-finding corpora use actual source parsing and no wrapper adaptation.
The production rule has no fixes or suggestions; their empty lists match too.
Canonical streams and matching SHA-256 hashes are stored beside this report.
Paths in headers are original absolute paths, so hashes describe this run.
The same controls also match under sanitizers. A query through numeric-literal
after program release exits 70 with invalid or released checker handle.

The final mutant only replaces the unique `return false;` in canLeak with
`return true;`. It compiles, finishes normally with exit 0 and empty stderr,
and fails solely through the independent production Go output at byte 6236.
A preceding comparison-inversion mutant did not compile due to TS2367; it
was discarded and is not counted as proof. Direct Go tests compare four
checker-held numeric literals including 0n/1n, compare constraint identities
with direct checker answers, and reject malformed/noncanonical type identities.
The unchanged upstream Go JSX fixture matrix also passes, including the real
pagination, checkbox-grid and WebSockets sites. This baseline is independent
of the native wrapper adapter, but does not establish native JSX parsing.

Final single process timing observations: repository native 0.194499s versus
Go 0.112722s (1.73x); compiler native 2.445546s versus Go 0.564929s (4.33x).
The compiler timing overlapped the Go upstream test command and is not an
isolated benchmark or a median. Corpora have no trigger shapes and zero bridge
queries. Wrapper control timing is recorded in suite.log. No speed parity is claimed.

Commands sent test output to logs:

```sh
source /workspace/adamic-tools/env.sh
ADAMIC_WAVE15_LEAKED_ARTIFACTS=/workspace/wave15-leaked-final-2 ADAMIC_WAVE15_COMPILER_MANIFEST=/workspace/wave-15-compiler.manifest ADAMIC_WAVE15_REPOSITORY_MANIFEST=/workspace/wave-15-repository.manifest ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave-15-typescript go test -v -count=1 -timeout=15m ./stage1/cohere/typeaware -run '^TestWave15LeakedAgreementAndMutants$' > /workspace/wave15-leaked-final-2.log 2>&1
go test -v -count=1 ./bridge/tsgo/checker > /workspace/wave15-leaked-facts.log 2>&1
(cd cohere && go test -v -count=1 ./internal/lint/rules/nexus -run '^TestCorrectnessNoLeakedNumberRender' > /workspace/wave15-leaked-upstream.log 2>&1)
go vet ./bridge/tsgo/... ./stage1/cohere/typeaware > /workspace/wave15-leaked-vet.log 2>&1
git diff --check > /workspace/wave15-leaked-diffcheck.log 2>&1
```

Toolchain remains the previously recorded Go 1.27.1, clang 20.1.8 and
Node 24.19.0; setup last completed in 30s with timing lines go/clang/node/
submodules ready 0s, cache warm 30s. nproc 5, CPU quota 4. All four new .a
files were formatted through production cohere stdin with virtual .ts aliases;
no Adamic .ts files were authored. Submodule pins are unchanged.

After implementation commit 23fe1096, both the requested push and its retry
failed with the exact message:

```text
fatal: could not read Username for 'https://github.com': No such device or address
```

A format-patch series is provided under /workspace/wave15-unpushed-patches.
No next rules were claimed: the user's prerequisite that already claimed work
be pushed has not been met. A read-only ranking audit after fetching 325
origin refs found 33 claim documents mentioning 96 ranked rules. The provisional
first remaining names were process-exit-after-output, uncleared-race-timeout
and require-blocking-standard-streams, but these were not reserved or edited.
A future selection must fetch and audit again. Full repository gates and
emitted-JavaScript comparison were not run.
