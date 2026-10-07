# Slot 04 landing on current main and lint area

The final fetch during wave 21 detected main moving from 39638d9e to c7991b900362796aefd111474e65eb5398e91953 and lint area moving from d65a8f93 to b84a9d9314b65d3d0261ee017e233287b4f071da. The ancestry check failed and no implementation was pushed on the stale base. The tested wave 21 implementation was committed locally as 47459c43, then all 69 worker commits rebased without conflict onto current origin/area/stage1-lint, which contains current origin/main.

Rebased implementation tip is 696ce65bda727da57f9acda77a5a5cbc0dcc3e1f. The final handoff names the subsequent evidence commit. All 62 retained helpers are implemented; no additional helpers are claimed during landing. The previously pushed own branch tip is claim 3f7ed758c0330f129f0ffb5572d850725c44ab81; the final push uses an exact force-with-lease for only codex/lint-helpers-04, as the user explicitly requested rebasing the worker branch. No main or area ref is pushed.

Incoming main changes include proven predicates/relations lowering, native records/runtime declarations and their Go/Node checks, plus stage3 measurements and documentation. These changes are preserved exactly from the lint area base. No compiler, native runtime, oracle, shared harness, shared registration, CLAUDE.md or shared test file was edited to land this unit. A path-restricted diff against the area base is empty for those files.

Validation commands write directly to this directory's logs after sourcing /workspace/adamic-tools/env.sh:

- go test -p 2 -count=1 -v -timeout=20m ./stage1/cohere/lint/helpers/...: fresh run of all 22 helper packages, including every retained compiling semantic mutant and consumer omission. PASS, all 22 packages, no selected skips. New wave 21 PASS 71.916s with thirteen compiling semantic and seven omission mutants.
- go run ./cmd/lint-registry: PASS, regenerated 15 validated descriptors in the ignored registry before shared harness checks.
- go test -count=1 -v -timeout=15m ./stage1/cohere/lint -run '^(TestDotARename|TestCompleteSuggestionSerialization|TestEmittedJavaScriptMismatch)$': verify .a loading, full suggestion serialization and independent emitted-JavaScript comparison, including their semantic mutants. PASS 195.273s: emitted mismatch caught (53.66s), .a rename parity (72.30s), full suggestions and second-edit mutation (69.29s). The indented child failure in harness.log is the expected planted emitted mismatch, not a gate failure.
- go vet ./stage1/cohere/lint/helpers/...: PASS, all helper packages, empty log.
- ADAMIC_GATE_UNCACHED=1 go test -count=1 -v -timeout=10m ./internal/oracle -run '^TestRuntimeLastIndexOfMatchesNode$': external Node parity against release/sanitized native and emitted JavaScript, PASS 15.279s, 758 identical bytes; three native misses, two Node misses, zero cache hits.
- go test -count=1 -v -timeout=10m ./internal/native -run '^(TestRuntimeReleasePaths|TestRuntimeStringEquality)$': PASS 10.810s; string equality 0.70s, release paths 10.80s on the incoming runtime.

Wave 21's thirteen semantic mutations cover offsets, byte grouping, separator flags, source bounds/equality, type/inference/ratio/number/percentage checks; seven omission mutations cover all seven consumers. The initial inference inversion failed the TypeScript checker and is not credited. The final compiling inference-guard omission is compared against actual Go in all three execution modes. Full details and pre-rebase logs are in ../slot04_wave21/REPORT.md and evidence/.

The full repository gate and the 17 broader stage 1 TypeScript/postcss/graphql/parser correctness comparisons are outside this bounded worker gate. No selected check is skipped or relaxed, and no skip counts as correctness evidence. Callback dependencies, full native rule findings, arbitrary configurations and exact callback visit traces remain outside these helper ports. Toolchain setup succeeded before the base moved (94s, 5 processors), with unchanged Go 1.27.1, clang 20.1.8 and Node 24.19.0.

## Fresh package results

```
ok  	github.com/system-inc/adamic/stage1/cohere/lint/helpers	158.049s
ok  	github.com/system-inc/adamic/stage1/cohere/lint/helpers/comments	218.427s
ok  	github.com/system-inc/adamic/stage1/cohere/lint/helpers/slot04_wave10	21.642s
ok  	github.com/system-inc/adamic/stage1/cohere/lint/helpers/slot04_wave11	23.355s
ok  	github.com/system-inc/adamic/stage1/cohere/lint/helpers/slot04_wave12	17.493s
ok  	github.com/system-inc/adamic/stage1/cohere/lint/helpers/slot04_wave13	97.089s
ok  	github.com/system-inc/adamic/stage1/cohere/lint/helpers/slot04_wave14	19.075s
ok  	github.com/system-inc/adamic/stage1/cohere/lint/helpers/slot04_wave15	15.917s
ok  	github.com/system-inc/adamic/stage1/cohere/lint/helpers/slot04_wave16	43.245s
ok  	github.com/system-inc/adamic/stage1/cohere/lint/helpers/slot04_wave17	38.531s
ok  	github.com/system-inc/adamic/stage1/cohere/lint/helpers/slot04_wave18	54.740s
ok  	github.com/system-inc/adamic/stage1/cohere/lint/helpers/slot04_wave19	164.532s
ok  	github.com/system-inc/adamic/stage1/cohere/lint/helpers/slot04_wave2	19.805s
ok  	github.com/system-inc/adamic/stage1/cohere/lint/helpers/slot04_wave20	27.547s
ok  	github.com/system-inc/adamic/stage1/cohere/lint/helpers/slot04_wave21	71.916s
ok  	github.com/system-inc/adamic/stage1/cohere/lint/helpers/slot04_wave3_space	19.951s
ok  	github.com/system-inc/adamic/stage1/cohere/lint/helpers/slot04_wave4	21.986s
ok  	github.com/system-inc/adamic/stage1/cohere/lint/helpers/slot04_wave5	24.897s
ok  	github.com/system-inc/adamic/stage1/cohere/lint/helpers/slot04_wave6	27.898s
ok  	github.com/system-inc/adamic/stage1/cohere/lint/helpers/slot04_wave7	30.826s
ok  	github.com/system-inc/adamic/stage1/cohere/lint/helpers/slot04_wave8	48.633s
ok  	github.com/system-inc/adamic/stage1/cohere/lint/helpers/slot04_wave9	21.777s
```

Protected remote refs and the exact prior worker tip were checked again after every test completed and remained unchanged. The final evidence commit only changes slot 04 reports, logs and claim completion metadata.
