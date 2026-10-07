# Wave 10 continuation: source ports complete, native graph compilation blocked

Built both remaining .a rule implementations, their CFG and raw AST field bridge; timeout remains fully validated.
Claim commit: `0ac7b51b`; timeout commit: `93916207`; this update is the commit adding the source-port evidence.
Source/Go comparison: 92 controls, 77 process-output and 30 blocking-streams findings; native timeout rerun passed in 80.023s.
Both source-rule mutants exit 0 and fail byte comparison; timeout and registry mutants pass their checks; raw static-body mutant is caught.
Native graph runners time out in internal/fresh; their corpus parity, sanitizers and native/Go timings remain unverified, so no new rules are claimed.

## Scope and selection

The original three ports are completed and pushed in `8cfa5d02`:
no-unmodified-loop-condition, no-redundant-type-constituents and prefer-includes.
The continuation claim was pushed before implementation after fetching all
origin heads and reading their claim files. Its rules, all with combined volume
zero, are:

- nexus/correctness-no-process-exit-after-output
- nexus/correctness-no-uncleared-race-timeout
- nexus/correctness-require-blocking-standard-streams

Selection used VOLUME_REPORT.md's 197 checker-dependent names and the linked
compiler/repository count tables. Main/base source had 25 ranked ports; the
26th existing port was not checker-dependent. Distinct claim blobs mentioned
96 remaining ranked names, leaving 76 candidates. These were the first three.
No additional rules were claimed. Original setup succeeded in 88s; nproc was 5
with a four-CPU quota. The existing tools environment was reused.

## Completed timeout port

`no_uncleared_race_timeout.a` implements the production rule's library-member
qualification, const promise lookup, executor traversal, global timer declaration
classification and lost-handle checks. Binding reads use checker identity,
including shorthand properties, shadowing, assignment targets and reads inside
nested functions. The message and byte spans match Go. This rule supplies no
fixes or suggestions; the serializer compares both counts explicitly.

`runtime_context.go` supplies raw compiler observations, decoded and classified
in `runtime_context.a`. The only shared source change is one import and one
switch arm in bridge/tsgo/checker/facts.go. No registration generator, shared
harness or protected compiler file was edited. The adapter reuses the caller's
checker lease. [runtime_context.md](runtime_context.md) documents its wire format.

Two defects in the earlier prepared adapter were corrected: colon framing was
incompatible with the existing newline decoder, and unconditional alias following
changed the timeout rule's symbol identity. `origin` now preserves identity;
`alias-origin` follows an alias only when native code requests that operation.
The process-member component explicitly selects the latter.

The owned Go harness compiles `timeout_suite.a` and an independent oracle overlay
inside pinned cohere. That oracle calls its unchanged production rule and loader;
it imports no bridge code. All output is logged to files. Controls include the
literal source tables extracted from production tests, extra lost/read handle
cases, shadowing, duplicate promise references, spreads, Unicode/CRLF, DOM timers,
and imported Promise/timer aliases. The dynamically assembled PhiSocial fixture
is not extracted by this harness; its literal table cases are covered.

| Dataset | Files | Findings | Exact serialized bytes | Sanitized comparison |
| --- | ---: | ---: | ---: | --- |
| Controls | 37 | 15 | 10,260 | same bytes, empty stderr |
| Frozen repository corpus | 287 | 0 | 18,485 | same bytes, empty stderr |
| TypeScript v6.0.3 src/compiler | 77 | 0 | 5,318 | same bytes, empty stderr |

The compiler checkout is `050880ce59e30b356b686bd3144efe24f875ebc8`.
The manifests are the same frozen corpora used by the original wave. Sanitized
archives use clang address/undefined instrumentation, nonrecovering undefined
behavior checks, and the native runner's `--sanitize` link. Linux leak detection
uses ASan's default enabled setting. Go's own memory accesses are not instrumented.

| Mutant | Result | Check that caught it |
| --- | --- | --- |
| Native timeout reports kept handles instead of lost handles | exit 0, empty stderr | production Go byte oracle, byte 54 |
| Raw adapter clears the default-library flag | Go test exit 1 | encoded library-owner assertion |
| Archive retains a released program in its registry | native exit 0 | required exit 70 with exact released-handle panic |
| Partial native write-chain reduction negates its containment condition | native exit 0 | production helper byte comparison |
| Partial native import index reverses its importer reachability condition | native exit 0 | production helper byte comparison |
| Existing Node one-byte oracle mutant | filtered test passed | independent Node/native byte comparison |

The normal released runtime-context query exits 70 with exactly
`adamic: panic: invalid or released checker handle`. Checker package tests passed
in 0.163s. The filtered Node oracle passed in 24.785s, including eight selected
behavior fixtures and its one-byte mutant. Vet passed. The six owned production
and runner `.a` files have zero configured Go cohere lint findings and were
formatted through the existing direct formatter API gate.

## Native time against Go

Three alternating complete-process runs per dataset were measured; every run's
stdout was compared. These are wall times including loading, checker setup,
parsing, linting and serialization, not isolated rule CPU measurements. The
machine was shared with other checks; no speedup claim is made.

| Dataset | Native median seconds | Go median seconds | Native / Go |
| --- | ---: | ---: | ---: |
| Controls | 0.112841 | 0.037875 | 2.979x |
| Repository | 0.252722 | 0.140324 | 1.801x |
| Compiler | 1.598204 | 0.284353 | 5.620x |

Raw samples and outputs are in `evidence/validated-timeout/benchmarks.json` and
its neighboring compressed output files.

## Remaining source ports and exact native blocker

`control_flow.a` ports Go cohere's graph construction using numeric block links.
`syntax_fields.go` supplies raw AST field membership; `syntax_fields.a` maps those
byte spans into the native parse tree. Neither Go adapter computes lint decisions.
The process-output port now contains root selection, reachable-path write state,
catch filtering, never/generator/async restrictions and local/foreign callee
following. The blocking-streams port contains import reachability, entry detection,
ordered call/load-time analysis, callback and function reachability and exact
policy-message serialization. The import index is in its own `blocking_index.a`.
Separate `.a` runners and independent production Go oracle overlays are provided.

Source execution under Node matches Go exactly over 64 extracted/extra process
controls and 28 additional blocking controls. The latter exercise empty Nexus
blocker declarations, aliases/reexports, blocking before/after exits, branches,
awaits, direct/indirect callbacks, generator and unused function roots, recursion,
catch/finally and shebang entry detection. Both rules produce zero fixes and zero
suggestions; those serialized fields are compared, not inferred.

| Source comparison | Controls | Findings | Identical bytes |
| --- | ---: | ---: | ---: |
| Process exit after output | 92 | 77 | 45,901 |
| Require blocking standard streams | 92 | 30 | 24,200 |

`TestSourceAgreementAndMutants` passed in 95.251s. It executes the owned source
with Node's type stripping against independently recorded raw compiler facts.
The records contain no production lint verdicts. A separate production Go loader
and rule supply expected findings. This is **source execution**, not native
parity or the compiler's emitted-JavaScript comparison. The raw record stream is
large (about 277 MiB compressed); it is reproducible but not committed.

The process source mutant drops newly recorded writes: exit 0, empty stderr,
comparison catches byte 114. The blocking source mutant disables ordered blocking:
exit 0, empty stderr, comparison catches byte 13,904. New syntax regression tests
cover absent loop fields, static bodies and destructuring ancestry. Clearing the resolved never return flag causes its Go assertion to fail. Omitting the
static body causes the Go check to exit 1 with `static block body absent`.
Source comparison also caught omitted loop/jump fields and a missing static-body
edge during development. A raw declaration adapter panic on binding-pattern
ancestor names was corrected by reading text only from supported name kinds.

The final native process build is bounded at 45s and exits 124; its owned Go test
fails after 53.129s. The final blocking build likewise exits 124 at 45s. Both
SIGQUIT traces show `internal/fresh.(*analysis).argument`, `made`, `value` and
freshness analysis before emission. An earlier process run stalled over three
minutes. This establishes a compiler-pass stall; its precise root cause is not
proven. No protected compiler, shared registration generator or shared harness
was edited. A nested-constructor refusal encountered earlier was avoided by
constructing the process helper in the blocking runner and injecting it.

The isolated native `runtime_probe.a` confirms a decoded static `Block` and a
resolved never return flag of 262144. Its released syntax query exits 70 with the
exact released-handle panic. The same probe passes address/undefined
and leak sanitizer checks with empty stderr. Omitting the raw static body in the
bridge makes the native probe exit 0 with empty stderr; byte comparison catches
byte 0. This validates raw bridge/decoder behavior only, not either
full graph rule. Probe evidence and statuses are preserved alongside the build
traces; see `evidence/source-ports`.

The completed timeout rule was rerun after these bridge changes. Normal and
sanitized outputs match Go for 37 controls (15 findings, 10,371 bytes at the new
artifact path), compiler77 (0 findings, 5,318 bytes) and repository287 (0 findings,
18,485 bytes). Its native decision mutant exits 0 and differs at byte 57; released
handle and retaining-registry mutant checks pass. Touched Go packages and vet pass.
All 12 top-level owned `.a` files format and lint with zero configured findings.
Previous timeout wall-time samples above remain the native/Go timing evidence.
Go-only compiler77 and repository287 baselines for each new rule have zero findings;
that observation does not establish native or source corpus agreement.

The earlier component-only evidence under `evidence` and `validated-timeout`
remains historical evidence for the previous commits. The source implementations
supersede its statements that CFG/callee/load-time logic was unported. Full native
findings/fixes/suggestions comparisons, complete native rule mutants, sanitizer
checks and native timings for these two rules remain blocked. Dynamic production
fixtures not assembled by the extraction helper, the full repository gate,
shared-harness integration and emitted-JavaScript comparison are not covered.
Following Ahra's instruction to stop at other blockers instead of editing shared
files, both claims remain reserved and no additional batch is claimed.

## Reproduction and evidence

```sh
source /workspace/adamic-tools/env.sh
ADAMIC_WAVE10_NEXT_VALIDATE=1 \
ADAMIC_WAVE10_NEXT_ARTIFACTS=/workspace/wave-10-next-validation \
ADAMIC_WAVE10_REPOSITORY_MANIFEST=/workspace/wave-10-repository.manifest \
ADAMIC_WAVE10_COMPILER_MANIFEST=/workspace/wave-10-compiler.manifest \
ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave-10-typescript \
go test ./stage1/cohere/typeaware/wave_10_next -count=1 -timeout=10m -v \
  > /tmp/wave-10-next-timeout-validation.log 2>&1

go test ./bridge/tsgo/checker -count=1 -v \
  > /tmp/wave-10-next-checker-tests.log 2>&1
go test ./internal/oracle \
  -run '^TestTheOracleCatchesOneByte$|^TestNativeAgreesWithNode$/internal/oracle/testdata/(maps_and_text|sorting|string_index|lone_surrogates|functions|closures)\.a$' \
  -count=1 -timeout=5m -v > /tmp/wave-10-next-node-oracle.log 2>&1
go vet ./stage1/cohere/typeaware/wave_10_next ./bridge/tsgo/checker \
  > /tmp/wave-10-next-vet.log 2>&1

timeout 45 /workspace/wave-10-next-validation/adamic build \
  stage1/cohere/typeaware/wave_10_next/gaps/native_atoms.a \
  -o /workspace/wave-10-next-atoms \
  --tsgo /workspace/wave-10-next-validation/checker.a \
  > /tmp/wave-10-next-atoms-recheck.log 2>&1
```

The private-helper test uses a Go overlay mapping an added virtual nexus test
file to this directory's `testdata/cohere_atoms_test.go`. Set
`ADAMIC_WAVE10_NEXT_ATOMS` to the colon-separated state/index/timer probe binaries
and run only `TestWave10NextNativeAtomsAgainstProduction` in cohere. Substitute
either helper mutant binary to obtain the logged comparison failures.

`evidence/validated-timeout` contains final test logs, all harness subprocess
stdout/stderr compressed without altering bytes, control sources, corpus
manifests, timing samples, expected exit statuses and SHA-256 hashes. Earlier
partial evidence remains under `evidence`; its unsupported-question failures and
failed combined comparisons are superseded only for the completed timeout port
and the separate helper probes. No whole-repository gate, complete control-flow
rule equivalence, shared-harness integration or emitted-JavaScript comparison is
claimed.

## Reproduce the source-port checks

```sh
source /workspace/adamic-tools/env.sh
ADAMIC_WAVE10_PROCESS_VALIDATE=1 \
ADAMIC_WAVE10_PROCESS_ARTIFACTS=/workspace/wave-10-process-validation \
go test ./stage1/cohere/typeaware/wave_10_next \
  -run '^TestProcessNativeAgreement$' -count=1 -timeout=2m -v \
  > /tmp/wave-10-process-native-final.log 2>&1
# This prepares the 64 controls and independent oracle before its bounded build.
ADAMIC_WAVE10_BLOCKING_FIXTURES=/workspace/wave-10-process-validation \
go test ./stage1/cohere/typeaware/wave_10_next \
  -run '^TestBlockingFixturePreparation$' -count=1 -v \
  > /tmp/wave-10-blocking-fixtures.log 2>&1
go build -o /workspace/wave-10-context-recorder \
  ./stage1/cohere/typeaware/wave_10_next/testdata/record_context.go \
  > /tmp/wave-10-context-recorder-build.log 2>&1
/workspace/wave-10-context-recorder \
  /workspace/wave-10-process-validation/tsconfig.json \
  /workspace/wave-10-process-validation/all-controls.manifest \
  /workspace/wave-10-all-context.json.gz > /tmp/wave-10-all-context.log 2>&1
ADAMIC_WAVE10_SOURCE_ARTIFACTS=/workspace/wave-10-process-validation \
ADAMIC_WAVE10_SOURCE_RECORDS=/workspace/wave-10-all-context.json.gz \
go test ./stage1/cohere/typeaware/wave_10_next \
  -run '^TestSourceAgreementAndMutants$' -count=1 -timeout=5m -v \
  > /tmp/wave-10-next-source-comparison.log 2>&1
```
