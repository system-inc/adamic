Built: wave 05 rebased onto current main f8013f0b; all six implemented rules re-green; no new claims.
Commits: tested rebased tip e93c60774bb0c295e1b13c2043ed26af21728210; prior published tip 985dcb24; enclosing commit adds landing evidence.
Commands and outputs: both frozen corpora, production Go byte parity, sanitizers, released handles, bridge tests, filtered Node oracle, numeric listener comparison and vet PASS.
Mutants: six rule verdicts and CFG corruption caught only by Go bytes; registry mutant caught by refusal; six numeric-kind metadata mutations caught by enum comparison.
Not covered: three React ports, numeric handed-node dispatch/string-kind removal, full repository gate, full upstream option matrix and shared finding-model migration.

# Landing onto f8013f0b

Origin main advanced to `f8013f0baac41ddc340d76f83bddde38536a8f07`. The explicitly requested rebase applied all 18 commits without conflicts. No implementation or shared files were edited for landing. This is the only branch this worker publishes; push targets only `codex/typeaware-wave-05`, using an exact lease on the previous tip. No main or area branch push was attempted.

Commands are the full revalidation commands recorded in LANDING_2_REPORT.md, now with artifact directory `/workspace/wave-05-landing3-first-artifacts` and log prefix `/workspace/wave-05-landing3-`. Both frozen corpus manifest variables and the TypeScript source variable were explicitly set. Every toolchain shell sourced `/workspace/adamic-tools/env.sh`; all test output went directly to files. Logs are retained under [landing3_evidence](landing3_evidence/).

- `git fetch origin '+refs/heads/*:refs/remotes/origin/*' --no-recurse-submodules`, then `git rebase origin/main`: exit 0.
- First-wave `TestWave05AgreementAndMutants`: PASS 153.834s, including both corpora in normal and ASan/UBSan/LSan builds. Controls retain 20 findings; repository 13 findings and 24,608 identical bytes; compiler 70 findings and 23,501 identical bytes. Three verdict mutants exit 0 with empty stderr and disagree with Go. Released handle exits 70 exactly; registry-deletion mutant exits 0 and fails that assertion.
- `WAVE05_OUTPUT_SKIP_BENCH=1 python3 stage1/cohere/typeaware/wave_05_next/validate_output.py`: PASS direct controls, 30 module programs, DOM/Node timer regressions, corpora, sanitizers, process/blocking/CFG mutants and three released-handle protocols. Corpus counts remain zero for continuation rules: repository 18,485 bytes and compiler 5,318 bytes, identical to Go in both builds.
- Timer `validate.py`, with both corpus manifests: PASS timer controls, verdict mutant, released handle and both sanitized corpora. Its retained historical constructor-gap probe still refuses as expected; completed output ports avoid that constructor pattern.
- `go test ./bridge/tsgo/... -count=1 -timeout=15m`: PASS bridge 95.425s, checker 0.153s.
- Filtered Node agreement and `TestTheOracleCatchesOneByte`: PASS 10.599s; log records selected cases and cache accounting.
- React `validate_gap.py`: PASS reproducing Go-positive controls versus native parser exit 70, the non-JSX clean control and skipped-parser probe mutant. This is not React rule parity.
- `python3 stage1/cohere/typeaware/wave_05_next/check_listener_kinds.py`: PASS six declarations matching production Go listeners and pinned numeric enum values, plus six metadata mutants.
- `go vet ./bridge/tsgo/...` and `git diff --check`: exit 0, empty output.

Checks overlapped, so retained timer-run timing samples are not uncontended benchmarks. Previously measured native/Go medians remain: first three compiler 14.029s/0.588s, repository 0.532s/0.184s; continuation compiler 2.380s/0.348s, repository 0.276s/0.125s. No new performance claim is made. Setup remains the earlier successful 82s run, nproc 5.

# Outstanding dependencies

The existing claims `react-hooks/globals`, `react-hooks/immutability`, and `react-hooks/no-deriving-state-in-effects` remain unported. Main's parser blob is still `bc0ee72ab6fa5cdf6b2dcca1e096d9c7f50fae8d`, without JSX support. Integrating the parser/lookahead/scanner changes from `codex/stage1-jsx-lint` exceeds Ahra's restriction to own rule directories and explicit instruction to stop on other blockers.

The six completed rules already export numeric `listenerKinds`, verified against Go. They still use the existing string-kind parser and run-based drivers; the complete handed-node speed requirement awaits shared numeric parser/driver integration. No new rules were claimed or written, so no new rule.json was introduced. No shared finding-model migration SHA was supplied in this instruction. Full gate, full upstream option matrix and shared emitted-JavaScript comparison remain outside the reported checks.
