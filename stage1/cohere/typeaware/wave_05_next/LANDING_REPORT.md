Built: wave-05 rebased without conflicts onto current origin/main; six completed rules revalidated, three React claims still blocked.
Commits: tested rebased tip 51f195feb1ddf0feaff4f4feb8f93b03fea5861e; base e011f8f6; previous remote tip ad6a485b.
Commands and outputs: six-rule production Go comparisons, both frozen corpora, sanitizers, released handles, bridge tests, filtered Node oracle and vet PASS.
Mutants: three first-wave verdicts, timer verdict and two output verdicts plus CFG caught only by Go bytes; released registry and skipped-parser probe caught by refusal assertions.
Not covered: React ports and their parity/mutants/timings, complete repository gate, complete upstream option matrix and shared emitted-JavaScript harness.

# Landing revalidation

The user explicitly requested rebase and republication before further claims. This supersedes CLAUDE.md's general prohibition on rewriting history for this owned branch. The rebase applied all 15 branch commits without conflicts. No protected compiler files differ from main, and no rule source was changed for landing. This is the only branch published by this worker. No new rule was claimed.

All commands sourced `/workspace/adamic-tools/env.sh`. Output went directly to logs, retained in [landing_evidence](landing_evidence/).

- `git fetch origin '+refs/heads/*:refs/remotes/origin/*' --no-recurse-submodules` and `git rebase origin/main`: exit 0.
- `go test ./stage1/cohere/typeaware -run '^TestWave05AgreementAndMutants$' -count=1 -timeout=20m -v`: PASS 120.762s for controls. Repeated with `ADAMIC_WAVE05_ARTIFACTS=/workspace/wave-05-landing-first-artifacts`, `ADAMIC_WAVE05_REPOSITORY_MANIFEST=/workspace/wave-05-repository.manifest`, `ADAMIC_WAVE05_COMPILER_MANIFEST=/workspace/wave-05-compiler.manifest`, and `ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave-05-typescript`: PASS 148.845s including both corpora.
- `WAVE05_OUTPUT_SKIP_BENCH=1 python3 stage1/cohere/typeaware/wave_05_next/validate_output.py`: PASS direct controls, 30 module programs, DOM/Node timer regressions, corpora, ASan/UBSan/LSan, comparison mutants and three released-handle protocols.
- `python3 stage1/cohere/typeaware/wave_05_next/validate.py`, with the same three corpus/source environment variables: PASS timer controls, comparison mutant, released handle and both sanitized corpora. Its historical constructor-gap probe still refuses as expected; the completed output implementation already avoids that constructor pattern.
- `go test ./bridge/tsgo/... -count=1 -timeout=15m`: PASS bridge 106.777s, checker 0.191s.
- `go test ./internal/oracle -run '^TestTheOracleCatchesOneByte$|^TestNativeAgreesWithNode$/internal/oracle/testdata/(maps_and_text|sorting|string_index|lone_surrogates|functions|closures)\.a$' -count=1 -timeout=10m -v`: PASS 11.727s. Go's subtest matching selected additional names containing those expressions; the log records all 19 Node/native observations.
- `python3 stage1/cohere/typeaware/wave_05_react/validate_gap.py`: PASS exact blocker reproduction, not React implementation parity.
- `go vet ./bridge/tsgo/...` and `git diff --check`: exit 0, empty logs.

The first three ports retain 13 repository findings (24,608 canonical bytes) and 70 compiler findings (23,501 bytes), in release and sanitized native compared with production Go. Output/timer drivers retain zero findings on both corpora (18,485 repository bytes; 5,318 compiler bytes), plus positive direct controls. The template, type-parameter, dead-store, timer, process-exit, blocking-stream and CFG mutants all compiled and exited 0 with empty stderr, then differed from Go. Released checker handles exit 70 with the exact required diagnostic; disabling registry deletion exits 0 and fails that assertion.

Checks overlapped on this machine, so timer-run timing samples are retained as observations and are not used as uncontended benchmarks. Previously measured native/Go medians remain in the original reports: first three compiler 14.029s/0.588s and repository 0.532s/0.184s; continuation compiler 2.380s/0.348s and repository 0.276s/0.125s. No new performance claim is made. Setup was previously successful: 82s total, nproc 5.

# Remaining prerequisite

`react-hooks/globals`, `react-hooks/immutability`, and `react-hooks/no-deriving-state-in-effects` remain unported. The unchanged local/main parser blob is `bc0ee72ab6fa5cdf6b2dcca1e096d9c7f50fae8d`. All three positive production Go controls still encounter native JSX refusal 70 before rule evaluation. JSX support exists on `origin/codex/stage1-jsx-lint` at `a8a62d62ca49db7415e14c3887dd305022b17309`, but integrating its parser/lookahead/scanner changes is outside this worker's rule directories. Ahra's instruction to stop on other blockers still applies. Landing validation of the six implemented rules is green; the three unfinished claims remain explicitly blocked rather than advertised as complete.
