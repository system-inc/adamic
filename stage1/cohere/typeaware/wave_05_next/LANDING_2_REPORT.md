Built: wave 05 rebased onto newly advanced origin/main, with all six implemented rules re-green; no new claims.
Commits: tested rebased tip fb23ac03801683596f0396d0fb6f859d6092a922; main base e8ba3d5d81de4d3773c723914fccd4c76248b965; previous published tip 6a1efc9f.
Commands and outputs: both frozen corpora, release and ASan/UBSan/LSan, production Go byte parity, bridge tests, filtered Node oracle and vet PASS.
Mutants: six rule verdicts and CFG corruption caught only by Go bytes; released registry caught by exact refusal assertion; skipped-parser probe also caught, not a React rule mutant.
Not covered: three React ports, their parity/mutants/timings, full repository gate, full upstream option matrix and shared emitted-JavaScript comparison.

# Second landing revalidation

Main advanced to the devirtualization merge. The requested rebase applied all 16 commits without conflicts. No wave-owned implementation changed, and no protected compiler files differ from the new main. This worker publishes only `codex/typeaware-wave-05`; it does not push main or an area branch. The user's explicit rebase instruction authorizes rewriting this owned branch, superseding the general CLAUDE.md prohibition. Publication uses an exact lease on the old remote tip.

The commands match [LANDING_REPORT.md](LANDING_REPORT.md), with first-wave artifact directory `/workspace/wave-05-landing2-first-artifacts` and logs under [landing2_evidence](landing2_evidence/). Each command sourced `/workspace/adamic-tools/env.sh`; subprocess output went directly to files. First-wave corpus environment variables were explicitly set to both frozen manifests and `/workspace/wave-05-typescript`.

- First-wave `TestWave05AgreementAndMutants`: PASS 189.494s. Controls: 20 findings; repository: 13 findings, 24,608 identical bytes; compiler: 70 findings, 23,501 identical bytes. Release and sanitized native both match production Go. All three verdict mutants compile and exit 0, empty stderr, then fail Go comparison. Released handle exits 70; registry-deletion mutant exits 0 and fails that assertion.
- `WAVE05_OUTPUT_SKIP_BENCH=1 python3 stage1/cohere/typeaware/wave_05_next/validate_output.py`: PASS output controls, 30 module programs, DOM/Node timer regression, both corpora, sanitizers, three comparison mutants and three released-handle protocols. Corpus counts remain zero for these continuation rules, with 18,485 repository and 5,318 compiler canonical bytes.
- Timer `validate.py`, with both corpus manifests: PASS positive controls, verdict mutant, exact released-handle refusal and both sanitized corpora. Its retained historical constructor-gap probe still refuses; the completed output rules avoid that constructor pattern.
- `go test ./bridge/tsgo/... -count=1 -timeout=15m`: PASS bridge 103.025s and checker 0.246s.
- Filtered `TestTheOracleCatchesOneByte` and Node agreement command from the first landing report: PASS 1.843s; log records all selected observations and cache accounting.
- React `validate_gap.py`: PASS reproducing all three production Go positives versus native parser exit 70, non-JSX clean control and no-op parser probe mutant.
- `go vet ./bridge/tsgo/...` and `git diff --check`: exit 0, empty output.

Checks overlapped, so timer-script timing observations are retained but not presented as uncontended performance. Previously measured native/Go medians remain in the original evidence: first three compiler 14.029s/0.588s, repository 0.532s/0.184s; continuation compiler 2.380s/0.348s, repository 0.276s/0.125s. Setup remains the prior successful 82s run, nproc 5.

# Remaining blocked claims

`react-hooks/globals`, `react-hooks/immutability`, and `react-hooks/no-deriving-state-in-effects` are still unported. Local and main parser remain blob `bc0ee72ab6fa5cdf6b2dcca1e096d9c7f50fae8d`, without JSX support. The separate `codex/stage1-jsx-lint` dependency requires shared parser/lookahead/scanner integration outside this worker's rule directories. Ahra's explicit instruction to stop on other blockers applies. No further rules were claimed. Re-green status covers the six implemented rules; it does not assert the three reservations are completed.
