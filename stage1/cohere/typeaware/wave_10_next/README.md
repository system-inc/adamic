# Wave 10 continuation: timeout validated, two rules incomplete

Built the native timeout rule and registered its raw compiler-fact question; two control-flow rules remain partial.
Claims: `0ac7b51b`; earlier partial components: `72718a9d`; completed timeout work is in the commit adding this report.
Final native/Go gate passed in 64.698s: 37 controls, compiler77, repository287, sanitizers and released-handle checks.
Timeout, library-fact and released-registry mutants were caught; state/index component mutants were also caught by Go bytes.
The combined component build stalls in internal/fresh; CFG/root/callee and load-time/call work for the other two rules is unported.

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

## Exact remaining blocker and unfinished work

The combined owned component probe `gaps/native_atoms.a` still fails to finish
compilation. A bounded final-source run using the current stage-0 binary and
registered checker archive exited 124 at 45 seconds with empty compiler output.
The earlier three-minute run's SIGQUIT stack, preserved in
`evidence/fresh-stack.txt`, showed `internal/fresh.(*analysis).argument`, `made`
and `run` during lowering's freshness/cycle proof. Distinct literal class tags
did not resolve it. This is a compiler stall, not a shared-harness `.a` loading
failure. No compiler source was edited to bypass it.

Separate `gaps/state_only.a`, `gaps/index_only.a` and `timeout_probe.a` component
probes compile and run. An owned overlay test calls unchanged private production
Go helpers and matches their five output lines exactly:

```text
,1,
,
reaches streams
imported
lost
```

Those comparisons and the two component mutants validate only the prepared
helpers. They do not establish complete process-output or blocking-streams rule
behavior. Process-output still needs native CFG construction, root analysis,
callee following and its complete runner. Blocking-streams still needs native
CFG traversal, ordered call/load-time analysis and its complete runner. Their
full findings/fixes/suggestions comparisons, complete rule mutants and timings
remain unverified. The raw signature/program modes are prepared for this work;
origin and program-root facts have direct tests, but full signature/import-edge
coverage is not claimed.

Following Ahra's instruction to report other blockers and stop rather than edit
shared files, this continuation stops with the combined compiler stall recorded.
The two incomplete claims remain reserved. No further batch is claimed.

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
