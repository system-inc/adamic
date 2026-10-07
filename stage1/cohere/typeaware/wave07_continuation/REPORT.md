Built: three continuation rules in native .a, validated through private checker registration overlays; production registration remains pending.
Commits: claim 59bdcd34e807795fa50933d1989b6a54b021510f; implementation 0532b3dfc814fbf844fa17f64b0dd3f1cd46d984; evidence follows in this report commit.
Commands and outputs: three byte-comparison gates, sanitizer gates, checker tests, vet and 18-file source lint passed; 11/16/16 positive control findings.
Mutants: three rule messages and three raw checker facts caught by independent Go bytes; registry deletion caught by the required released-handle panic.
Not covered: production registration, shared harness integration, full repository test gate, emitted-JavaScript comparison and exhaustive upstream option/JSX matrices.

# Continuation report

The claimed rules are `nexus/correctness-no-uncleared-race-timeout`,
`nexus/correctness-no-process-exit-after-output`, and
`nexus/correctness-require-blocking-standard-streams`. The original trio remains
complete and pushed, with its report in [WAVE_07_REPORT.md](../WAVE_07_REPORT.md).
No further rules were claimed. The selection and earlier blocker diagnosis in
[BLOCKED.md](BLOCKED.md) are historical; this report records the subsequent ports.

## Implementation and integration boundary

Each rule has its own .a implementation. Native helpers perform timer recognition,
callee following, write/exit classification, try-chain path state, entry selection,
import closure and blocking-call analysis. The new Go questions expose raw symbol,
resolved-signature declaration, structural control flow, and resolved module facts.
They never return a lint verdict. Structural flow code retains upstream attribution
and its MIT license. The independent oracle calls the pinned production Go rules
and imports no bridge implementation.

The shared dispatcher, registration generator and existing test harness were not
edited in this continuation. Private validation builds use Go overlays to register
three questions. The normal archive still refuses those unknown questions. The
exact three integration routes and schemas are in [QUESTIONS.md](QUESTIONS.md).
These results prove the isolated implementations; they do not establish that the
normal production archive can run them. Ahra's shared-file restriction prevents
finishing that registration here. All new Adamic implementation files are .a;
TypeScript fixture strings in evidence describe the independent oracle's inputs.

## Validation

Run from the repository with `source /workspace/adamic-tools/env.sh`:

```sh
python3 stage1/cohere/typeaware/wave07_continuation/validate_timer.py /workspace/wave-07-timer --compiler /workspace/wave-07-typescript
python3 stage1/cohere/typeaware/wave07_continuation/validate_process.py /workspace/wave-07-process --compiler /workspace/wave-07-typescript
python3 stage1/cohere/typeaware/wave07_continuation/validate_streams.py /workspace/wave-07-streams-final --compiler /workspace/wave-07-typescript
python3 stage1/cohere/typeaware/wave07_continuation/validate_questions.py /workspace/wave-07-questions
go test ./bridge/tsgo/checker ./bridge/tsgo/wave07_flow -count=1 -v
go vet ./bridge/tsgo/checker ./bridge/tsgo/wave07_flow
```

All output was saved to files. The scripts preserve complete command arguments,
exit codes and timing records in commands.json. Compressed stdout/stderr, input
fixture snapshots, manifests, command records and benchmark records are committed
under [evidence](evidence/median-seconds.json). Every comparison checks complete
canonical findings, including fixes and suggestions, against production Go bytes.
All three rules have no default fixes or suggestions; those empty fields are still
serialized and compared. This does not exercise a nonempty suggestion serializer.

| Rule | Control findings | Canonical bytes | Repository findings | Compiler findings |
| --- | ---: | ---: | ---: | ---: |
| Uncleared race timeout | 11 | 6942 | 0 | 0 |
| Exit after output | 16 | 9638 | 0 | 0 |
| Blocking standard streams | 16 | 11990 | 0 | 0 |

The frozen repository corpus has 287 sources; TypeScript src/compiler has 77.
Each zero-finding result still compares the entire canonical stream: 18485 bytes
for repository inputs and 5318 for compiler inputs. Controls test positive and
negative paths, shadowing, nested bodies, try/catch/finally, generator and async
calls, ordered blocking, imported load-time blocking, recursion and source offsets.
DOM comparisons are recorded separately: timer and process retain 11 and 16 findings; streams has zero findings with its DOM-only ambient declarations. Sanitized native builds use address,
undefined-behavior and leak sanitizers and agree on controls and both corpora with
empty sanitizer stderr. Go's heap itself is not instrumented by those C sanitizers.

Per-rule mutants replace the finding message ID. Every mutant builds, exits 0,
and has empty stderr; only the byte comparison with independent Go catches it.
Raw-fact mutants zero shorthand value identity (first differing byte 5390), remove
CFG successor edges (618), or remove resolved module targets (9645). These also
build and exit normally and are caught by Go output. All three live questions
succeed, then reject a released handle with exit 70 and
`adamic: panic: invalid or released checker handle`. A separate registry deletion
mutant exits 0 after release, and the required panic assertion catches it.
A preliminary unused-variable compile failure was corrected and is not counted as
a successful mutant check.

Direct checker tests pass, vet is clean, and the isolated .a-capable cohere CLI
reports `0.1s (276 rules • 18 checked)` with no findings. Its newer parser is used
only for implementation-source lint; finding oracles remain pinned at cohere
715ba94f3608a6500086b1076ce5cb7e51b836db and TypeScript-go
8d550c837c90bd1805b047b7eeccc2baac2d5e7a. Compiler corpus commit is
050880ce59e30b356b686bd3144efe24f875ebc8. Original setup took 81 seconds;
`nproc` reported 5 (four-core quota), using Go 1.27.1, clang 20.1.8 and Node 24.19.0.

## Native time against Go

Three quiet alternating rounds compare complete output using the original
validation-wave-07/full_bench.py. These are process-time medians, including loading
and traversal. The zero-hit corpora do not measure positive-control query cost.
Native is slower on these inputs.

| Rule | Repository native / Go seconds | Compiler native / Go seconds |
| --- | --- | --- |
| timer | 0.231783 / 0.111282 | 1.549636 / 0.285661 |
| process | 0.235780 / 0.121675 | 1.617633 / 0.292161 |
| streams | 0.289598 / 0.121115 | 1.673026 / 0.292310 |

The original bridge buffer/decoder mutants and filtered Node oracle gates are
recorded with the original wave, and were not repeated for this continuation.
There was no full repository gate, all-26-rule rerun, emitted-JavaScript comparison,
or exhaustive upstream options/JSX matrix. Production dispatch registration and
shared harness integration remain required before treating these ports as shipped.
