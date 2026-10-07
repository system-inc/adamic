Built: rebased the twelve completed ports onto current origin/main; the three reserved React rules remain blocked.
Commits: main e011f8f6; rebased tip 3d81aa65; published ancestry preserved at 07645922; this report committed separately.
Commands and outputs: four wave-22 oracles PASS (700.318s); bridge PASS (108.445s), checker PASS (1.164s).
Mutants: all twelve per-rule mutants plus scope-export and released-handle registry mutants caught again after rebase.
Not covered: full repository gate, three pending React ports, new benchmarks; the overly broad first test run was interrupted.

## Landing

The only branch pushed for this unit is codex/typeaware-wave-22. Its previous
published tip was 5b86729d. Fetching all origin heads succeeded. Rebase onto
origin/main e011f8f60899586d6373a5ccb07335ad82cfbf3c completed without conflicts.

CLAUDE.md prohibits force-pushing. A merge with the previous published tip as
an additional parent preserves its ancestry while retaining the exact rebased
tree. Commit 076459228aabd23c642906ee0400a8d914518c55 has the same tree as its
first parent. Both current main and the previous published branch tip are
ancestors, allowing a normal push. No shared parser, harness or compiler fixes
were made in this landing unit. The working tree was clean before rebase.

## Verification

```sh
source /workspace/adamic-tools/env.sh
ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave-22-typescript-pinned go test ./stage1/cohere/typeaware -run '^TestWave22(AgreementAndMutants|NextAgreement|ThirdAgreement|FourthAgreement)$' -count=1 -timeout=30m -v
ADAMIC_TSGO_CORPUS=/workspace/wave-22-typescript-pinned go test ./bridge/tsgo/... -count=1 -timeout=15m -v
```

Each command redirected output directly to a log. Saved logs are in
validation-wave-22-landing. Results:

| Oracle | Result | Seconds |
| --- | --- | ---: |
| Fourth preference batch | PASS | 177.88 |
| Continuation promises/spread/concurrency batch | PASS | 225.00 |
| Original three rules | PASS | 161.30 |
| Third global-binding batch | PASS | 136.13 |
| Entire focused package invocation | PASS | 700.318 |
| Bridge package | PASS | 108.445 |
| Checker package | PASS | 1.164 |

Every batch compared its positive controls and the frozen repository/compiler
corpora byte for byte against production Go cohere, including repairs and
suggestions, in normal and sanitizer builds. Each frozen corpus emitted zero
findings (repository 18485 bytes; compiler 5857 bytes). These corpus checks are
supported by positive control findings, not zero-findings alone.

All twelve per-rule mutants compiled and executed with exit zero and empty
stderr; only the byte oracle caught their changed findings, fixes or suggestions:
numeric-prefix, has-own-fix-span, spread-parens, promise-condition,
spread-await-edit, lost-write-span, nullish-suggestion, qualifier-fix,
private-read, global-provenance, global-declaration-span and timer-string.
The scope-export mutant was also caught. Released handles still panic with
exit 70; the retained-registry mutant exits zero and is caught by the refusal
requirement. The bridge independently matched 54982 bytes across 1600 positions
in four files under ASan, UBSan and LSan, and reran its bridge mutants.

The original rule suite recorded native/Go wall times of 0.608547/0.252550s
on repository and 4.638112/0.475953s on compiler. These are single observations,
not replacement benchmark medians. Native remains slower in these observations.
Other batches' previous benchmark medians remain in their original reports.

The initial command used the overly broad filter TestWave22|Test, which matches
all package tests. It ran other workers' coverage/volume suites and was stopped
with its child processes before the package timeout. Its incomplete log is
preserved as interrupted.log and is not reported as a pass. The exact-filter
run above completed; no test harness was edited to obtain the result.

The existing toolchain setup was reused: setup total 118s, nproc 5, Go 1.27.1,
clang 20.1.8, Node 24.19.0. The full repository gate was not run.

## Pending claims

react-hooks/set-state-in-effect, react-hooks/set-state-in-render and
react-hooks/static-components remain reserved and unimplemented. The rebased
parser.ts blob bc0ee72ab6fa5cdf6b2dcca1e096d9c7f50fae8d is identical to the
previous published parser that rejects the saved JSX reproduction. Native
React HIR/SSA prerequisites are still absent from this branch. The precise
blockers and independent production Go tests are in WAVE_22_FIFTH_REPORT.md.

Ahra's instruction to stop at other blockers without editing shared files
continues to apply. No new claims were made. The twelve completed ports are
re-green on main; no coverage is claimed for the three pending React rules.
