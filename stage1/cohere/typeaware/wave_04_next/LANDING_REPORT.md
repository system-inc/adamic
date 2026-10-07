Rebased: wave 04 onto current main `e8ba3d5d`; no code changes or conflicts.
Commits: validated rebased tip `9854da3a` replaces remote `2f79da56`; this evidence commit follows.
Checks: six complete ports match Go over controls, 77 compiler roots and 287 repository roots, normal and sanitized; partial React checks pass.
Mutants: all six rule mutants, both handle mutants, six reporting/refusal mutants, four refs-kernel mutants and the private-question rejection mutant caught again.
Uncovered: three React source analyses still lack native HIR/SSA/reactive scopes; no new claims or full repository gate.

The only pushed branch owned by this unit is `codex/typeaware-wave-04`. All 16
existing commits rebase without conflicts; the preserved range-diff marks every
patch identical. Main advanced from `e011f8f6` to `e8ba3d5d` during the first
validation, so the branch was rebased a second time and every owned oracle was
rerun with a newly built stage 0. A final remote read confirms main still at
`e8ba3d5d` and the wave branch at the expected old remote SHA. The user explicitly
requested rebasing and pushing the previously pushed branch; a force-with-lease
against that exact old SHA protects concurrent remote updates.

## Final checks

All output goes to log files in `/workspace/typeaware-wave-04-landing-current`.
Sources use `/workspace/adamic-tools/env.sh`; no toolchain or submodule pins
changed. Original setup remains 77 seconds with nproc 5.

- `go build -o /workspace/typeaware-wave-04-landing-current/adamic ./cmd/adamic`: PASS.
- `TestWave04AgreementAndMutants`: PASS 219.392s; controls 39 findings / 17,295 bytes,
  repository 14 / 23,598, compiler 200 / 77,043; normal and ASan/UBSan/leaks identical.
- `wave_04_next/validate.py`: PASS; 26 timer controls, 59 process roots and 45 blocking
  programs, plus both requested corpora; complete normal/sanitized Go byte agreement.
- `wave_04_react/validate_partial.py`: PASS reporting only, 11 / 5,948 bytes across Go,
  native, sanitized native, source Node and emitted JavaScript. Source analyses refuse.
- `wave_04_react/validate_refs_kernel.py`: PASS kernel only, 592 / 10,151 bytes across
  the same backends; this is not a full refs rule verdict.
- All checker package tests, scoped Go vet and the filtered Node oracle: PASS.

The original corpus test command uses `ADAMIC_WAVE04_ARTIFACTS`,
`ADAMIC_WAVE04_COMPILER_MANIFEST`, `ADAMIC_WAVE04_REPOSITORY_MANIFEST` and
`ADAMIC_TYPESCRIPT_SOURCE`, then `go test ./stage1/cohere/typeaware -run
'^TestWave04AgreementAndMutants$' -count=1 -v -timeout=30m`. The continuation uses
its existing validator with `--adamic` pointing to the fresh build and the same
frozen manifests and extracted controls as [its report](REPORT.md). Both React
validators also receive that fresh `--adamic` path. The filtered Node command is
preserved in the earlier reproduction report; this round selects maps/text,
sorting, functions, closures and the one-byte comparator oracle.

## Mutants and ownership

All six completed-rule mutants compile, exit zero and have empty stderr; only Go
byte comparison catches them: casing end+1 at byte 5816, matching message at 126,
void end+1 at 1604, race timeout end+1 at 84, output/exit end+1 at 79 and blocking
streams end+1 at 9245. Both original and continuation released-handle probes panic
70. Their registry-retention mutants compile and exit zero, caught by the missing
required panic. Sanitizer corpus runs have empty native stderr.

Three React renderer span mutants and three removed-refusal mutants pass their
respective comparison/panic checks again. Four refs kernel mutants are caught
only by Go bytes. The unknown-private-question mutant compiles under an owned Go
overlay and fails exactly `unknown private question must reject: <nil>`; its log
records the expected test failure. These partial and checker mutants are distinct
from the six complete rule mutants.

## Timing and limits

Three quiet alternating count rounds for the completed correctness continuation:
compiler native 1.905s / Go 0.318s;
repository native 0.296s / Go 0.130s.
Native is slower. Counts agree in every round. This does not benchmark React or
remeasure the original trio's timing; its earlier timing remains historical.

Evidence, hashes, exact output streams, mutant logs, raw timing data and patch
mapping are in [landing-evidence](landing-evidence). Generated control headings
use longer scratch paths, so control byte lengths and mutant byte offsets differ
from prior runs while the Go comparisons still hold exact streams.

No protected compiler file, shared harness or generator was edited during this
landing task. The full repository gate was not run. React HIR lowering,
SSA/capture propagation and memoization reactive scopes remain unavailable, as
recorded in the [partial React report](../wave_04_react/REPORT.md). Those three
rules remain unfinished and claimed; no further rules were claimed.
