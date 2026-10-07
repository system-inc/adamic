Built: both claimed Nexus stream rules in .a, with native control flow; all three continuation algorithms are now implemented, with production registration pending.
Commits: claim 3fc985765422fb29b2438cd105ec821741703de2; timer f8ef4017bc3ac617b3bd42668b163bbb8aa72a37; stream implementation 29ab11a7738cb1d49966e5d1e9f67e07c31eae28; this report and its evidence are committed separately.
Commands and outputs: stream agreement PASS 173.316s; disposal expansion PASS 74.771s; release and pending-registration checks PASS 34.229s; checker PASS 0.218s; go vet ./... PASS; setup 137s; nproc 5.
Mutants: process suppression, blocking suppression and await-using suppression caught only by Go byte comparison; five raw-fact mutants caught by direct checker contracts; retained released handles caught by required-panic checks.
Uncovered: shared production registration, emitted-JavaScript comparison, the complete upstream multifile fixture suite and full repository test gate; native remains slower than Go, and no new rules were claimed.

## What is complete

The original three rules remain complete and pushed, as recorded in
WAVE_19_REPORT.md. The continuation claim was pushed before implementation.
The timer implementation and its independent comparisons remain recorded in
WAVE_19_TIMEOUT_REPORT.md. This checkpoint completes the two stream algorithms
without changing the shared harness, registration generator or checker dispatch.

- no_process_exit_after_output.a identifies the ambient global console and
  NodeJS.Process members by raw symbol declarations, following import aliases
  for process members. It follows one checker-resolved helper body, excluding
  generators, unawaited async calls, never-returning functions and helpers that
  themselves exit. Helper bodies in other modules are parsed natively.
- require_blocking_standard_streams.a computes runtime import edges, executable
  entries and the reverse blocking closure in Adamic. It follows local invoked
  functions and handed callbacks, handles cycles in may-block analysis, declines
  imported entries and modules blocked at import load, and reports the first
  qualifying exit with the exact count phrase.
- stream_graph.a builds syntax-only native control-flow blocks. Rule judgments
  remain in Adamic, including write chains, catch resets, argument exits,
  unreachable exits, loop back-edges and both finally copies. Await-using reads
  the parser's declaration flags, including comments between its keywords.

The existing isolated declaration-ancestry question gained an explicit alias
parameter. Name extraction is limited to Identifier names: calling Node.Text()
on an ancestor's binding pattern panicked on compiler/debug.ts during the first
corpus run. The corrected corpus runs below pass, and a direct checker regression
and mutant hold that guard. No Go lint judgment was added. Resolved-callee and
program-modules continue to supply raw checker and program metadata only.

New Adamic modules are .a. StandardStreams.ts inside the archived temporary
controls is generated TypeScript oracle input representing Nexus's pinned path;
it is not an Adamic implementation module. Existing helper filenames and the
frozen corpus populations are preserved.

## Independent agreement and memory evidence

Go oracles invoke the unmodified production rules from pinned cohere
715ba94f3608a6500086b1076ce5cb7e51b836db and TypeScript-go
8d550c837c90bd1805b047b7eeccc2baac2d5e7a. They import no bridge code.
Comparisons include byte positions, rule names, IDs, complete messages, fixes,
suggestions, file headers and totals. These rules have no fixes or suggestions;
their zero counts are serialized and compared.

| Rule and population | Inputs | Findings | Identical bytes | Normal and ASan/UBSan/LSan |
| --- | ---: | ---: | ---: | --- |
| Process controls, including 48 pinned production rows | 58 | 32 | 19,664 | PASS |
| Blocking controls, including commented await-using | 43 | 24 | 17,206 | PASS |
| Process, TypeScript compiler | 77 | 0 | 5,241 | PASS |
| Process, repository | 287 | 0 | 18,485 | PASS |
| Blocking, TypeScript compiler | 77 | 0 | 5,241 | PASS |
| Blocking, repository | 287 | 0 | 18,485 | PASS |

The compiler is TypeScript 6.0.3 at
050880ce59e30b356b686bd3144efe24f875ebc8. Corpus manifests are the base's frozen
77 compiler and 287 repository roots, not a newly expanded source glob.
Controls cover direct and imported process members, console and process stream
writes, shadows, loops, catches, finally blocks, helper bodies, async and never
returns, dead exits, exported unused functions, const-bound functions, callbacks,
generators, Nexus imports, blocks before and after guards, recursive blockers,
argument evaluation and shebang wording. The entire upstream blocking multifile
suite was not reconstructed, and the corpus findings are zero; positive controls
and their mutants are therefore necessary evidence.

## Mutants and the actual integration boundary

Every stream rule mutant compiled, exited 0 and emitted empty stderr. Only
comparison with unchanged Go killed it:

| Mutant | Check that killed it |
| --- | --- |
| Suppress process findings | Go bytes, first difference 48 |
| Suppress blocking findings | Go bytes, first difference 49 |
| Disable await-using recognition | Go bytes, first difference 14,262 |
| Do not follow alias declarations | Direct checker contract: wrong alias ancestry |
| Remove binding-pattern name guard | Direct checker regression: BindingPattern Node.Text panic |
| Erase global augmentation fact | Direct checker contract: wrong global-augmentation metadata |
| Erase resolved return flags | Direct checker contract: wrong resolved callee |
| Erase resolved module target | Direct checker contract: wrong resolved module edge |
| Retain released checker handles | All three required-panic checks; mutant exits 0 with empty stderr |

The raw-fact mutants are checker contract proofs, distinct from the byte-only
lint mutants. The prior timeout lost-handle mutant remains recorded in
WAVE_19_TIMEOUT_REPORT.md, including its byte-only kill at 45.

All three new questions reject released handles in normal and sanitized builds.
The released-registry mutant accepts them and is caught only by requiring the
invalid-or-released-handle panic. The unmodified production archive independently
refuses declaration-ancestry, resolved-callee and program-modules with panic 70
and the exact unsupported-question message.

Comparisons use a scratch Go build overlay containing the three pending dispatch
lines. The working-tree shared files remain untouched. The exact unapplied
wave_19_next_registration.patch passes git apply --check. Ahra's correction
forbids editing shared files, so production registration is still the concrete
integration blocker. No claim beyond these three was made while it is pending.
Native runners are owned entry points; shared emitted-JavaScript comparison is
left to codex/lint-harness-dot-a. Direct cohere CLI lint of .a remains the base's
reported filename-extension gap.

## Commands and timings

All test output was written to logs. The executed checks were:

```
source /workspace/adamic-tools/env.sh
ADAMIC_WAVE19_STREAM_ARTIFACTS=/tmp/wave19-process-verified ADAMIC_WAVE19_BLOCKING_ARTIFACTS=/tmp/wave19-blocking-verified ADAMIC_WAVE19_RELEASE_ARTIFACTS=/tmp/wave19-stream-release ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave19-typescript go test ./stage1/cohere/typeaware -run '^TestWave19(ProcessPendingRegistration|BlockingPendingRegistration|StreamReleasedHandles)$' -count=1 -timeout 15m -v > /tmp/wave19-stream-verified.log 2>&1
ADAMIC_WAVE19_BLOCKING_ARTIFACTS=/tmp/wave19-blocking-disposal ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave19-typescript go test ./stage1/cohere/typeaware -run '^TestWave19BlockingPendingRegistration$' -count=1 -timeout 10m -v > /tmp/wave19-blocking-disposal.log 2>&1
ADAMIC_WAVE19_RELEASE_ARTIFACTS=/tmp/wave19-stream-release go test ./stage1/cohere/typeaware -run '^TestWave19StreamReleasedHandles$' -count=1 -timeout 10m -v > /tmp/wave19-stream-release-final.log 2>&1
go test ./bridge/tsgo/checker -count=1 -v > /tmp/wave19-stream-checker-final.log 2>&1
python3 stage1/cohere/typeaware/testdata/prove_wave_19_next_facts.py /tmp/wave19-stream-fact-mutants > /tmp/wave19-stream-fact-mutants.log 2>&1
go vet ./... > /tmp/wave19-stream-vet-final.log 2>&1
```

The full repository test gate was not rerun. The original wave's bridge and
filtered Node-oracle gates remain documented in WAVE_19_REPORT.md. Setup already
completed in this environment: Go/clang/Node ready 0s, submodules 1s, cache 137s,
done 137s; the environment has nproc 5.

Timings use bridge/tsgo/profile/volume_bench.py, three alternating complete
count-only runs for each implementation and corpus. These are observed process
medians, including checker load and native parsing, not isolated rule time.
All timed runs report zero findings. The exact benchmark invocations were:

```
python3 bridge/tsgo/profile/volume_bench.py /tmp/wave19-process-verified/process /tmp/wave19-process-verified/process-oracle /tmp/wave19-process-bench --corpus compiler /workspace/wave19-typescript/src/compiler/tsconfig.json /tmp/wave19-process-verified/compiler.manifest --corpus repository /workspace/adamic/tsconfig.json /tmp/wave19-process-verified/repository.manifest --rounds 3 > /tmp/wave19-stream-bench.log 2>&1
python3 bridge/tsgo/profile/volume_bench.py /tmp/wave19-blocking-disposal/blocking /tmp/wave19-blocking-disposal/blocking-oracle /tmp/wave19-blocking-final-bench --corpus compiler /workspace/wave19-typescript/src/compiler/tsconfig.json /tmp/wave19-blocking-disposal/compiler.manifest --corpus repository /workspace/adamic/tsconfig.json /tmp/wave19-blocking-disposal/repository.manifest --rounds 3 > /tmp/wave19-blocking-final-bench.log 2>&1
```

| Rule | Corpus | Go seconds | Native seconds | Native / Go |
| --- | --- | ---: | ---: | ---: |
| Process | Compiler | 0.453433 | 1.885205 | 4.158 |
| Process | Repository | 0.168813 | 0.277867 | 1.646 |
| Blocking, final disposal binary | Compiler | 0.385728 | 1.801388 | 4.670 |
| Blocking, final disposal binary | Repository | 0.224027 | 0.311736 | 1.392 |

Native is slower on these populations. No speedup is claimed. Exact timing
records, passing logs, portable manifests, compressed input/output streams and
SHA-256 hashes are in validation-wave-19-streams. Earlier failed ancestry runs
were corrected before these accepted comparisons and are not presented as passes.
