Deleted exactly TestExplicitExportResolvesStarCollision and TestOneSeedOneProgram for #ajth108 and test audit #xphstyt.
Base: fc7252a6e30d5c36e27a774188928ec34887dd10, advanced from the initial b8bcadb main tip before validation.
Full load and fuzz packages and TestCallTargetReaders passed under the 89-second wall limit.
All seven relevant load mutants and fuzz S1 were caught by remaining tests.
No full repository gate, new fixtures, new tests or compiler behavior changes; counts.md needs no update.

TestGeneratedProgramsCheckAndLower, TestRegexProgramsPassTheChecker and TestStarExportCollisionNamesBothModules remain unchanged. The deleted load test uses writeProgram, also used by the retained collision test and other load tests. The deleted fuzz test uses Generate and Source, also used by the retained generator tests. shared-uses.log records the rg proof; neither test owned a helper or disk fixture, so nothing else was removed.

Evidence was read from test-defend/deletion-set/internal-load and test-defend/deletion-set/internal-fuzz before editing. Their matrix and inventory are retained here. Assumption: replay mutants whose candidates_failed includes one of the two authorized deletion targets. Thus the fuzz replay is S1; M10/G1/G2/G3/R1/R2/R3 concern the two explicitly retained tests and were not rerun. Patches are copied from the evidence branches, applied unchanged one at a time, and reversed in finally. No production edits remain.

Commands (each shell sources /workspace/adamic-tools/env.sh):

- export GOPROXY='https://proxy.golang.org|direct'; timeout 600 bash cloud/setup.sh: exit 0. Go ready 0.078s, Node ready 0.088s, clang ready 0.489s, markdown ready 1.144s, submodules ready 21.503s, shared cache ready 28.073s, go build ready 350.816s, build cache warm 350.915s, done 350.945s. nproc=5, CPU quota=4.
- timeout 120 npm ci --ignore-scripts --no-audit --no-fund in stage3/api: exit 0, added 3 packages in 2s. Setup did not install these loader-required locked dependencies. The initial warm load run failed with: load: node:* imports require @types/node 25.3.3 installed in stage3/api/node_modules/@types/node. After installation it passed.
- timeout 89 go test ./internal/load -count=1 -timeout 85s: exit 0, ok 1.103s (load.log).
- timeout 89 go test ./internal/fuzz -count=1 -timeout 85s: exit 0, ok 35.891s (fuzz.log).
- timeout 89 go test ./internal/ir -run '^TestCallTargetReaders$' -count=1 -timeout 85s: exit 0, ok 31.860s.
- timeout 89 go vet ./internal/load ./internal/fuzz: exit 0, no output.
- timeout 800 python3 review/compiler/test-audit-deletions-a/replay.py: exit 0; each individual go test has timeout 89 and -timeout 85s, ADAMIC_GATE_UNCACHED=1 and a distinct ADAMIC_BUILD_CACHE_DIR. Load mutants run the full remaining load package with -json -count=1. S1 selects ^(TestOctoberFeaturesAppear|TestUndefinedNumbersShapes)$ to avoid the known fixed-seed shrinker stall.
- gofmt on both changed Go files and git diff --check: clean.

Early cold test commands (load, fuzz, CallTargetReaders, S1) hit the 89-second compile wall before producing results while setup was still warming. They are not claimed as passing checks or mutant kills. cold-replay-results.json and cold-replay.log retain the S1 timeout. A separate timeout 180 go test -c -o /tmp/audit-del-a-load.test ./internal/load succeeded. All reported passing commands above were rerun after warming, with no mutant present. Restored full-package logs are also retained.

Observed mutant catches (every run exited 1 from test assertions, no timeout or build failure):

| Mutant | Remaining catcher | Wall seconds |
|---|---|---:|
| fuzz-S1 | TestOctoberFeaturesAppear; TestUndefinedNumbersShapes/vocabulary | 9.298 |
| load-D04 | TestStarExportCollisionNamesBothModules | 7.824 |
| load-D07 | TestStarExportCollisionNamesBothModules | 7.182 |
| load-D08 | TestStarExportCollisionNamesBothModules | 7.059 |
| load-M03 | TestStarExportCollisionNamesBothModules | 6.874 |
| load-M05 | TestStarExportCollisionNamesBothModules | 10.631 |
| load-M12 | TestStarExportCollisionNamesBothModules | 7.107 |
| load-M19 | TestStarExportCollisionNamesBothModules | 9.11 |

Every failing row and original Go JSON output is preserved in replay-results.json and the corresponding mutant log. The inference supported by these observations is redundancy against this recorded mutant set, not all possible future defects. No numbered roadmap step was specified for this unit; it delivers the named test-audit deletion task. Integration lane output is recorded separately after the required committed-branch check.

Change commit: 0618a123. Restored full package runs: load 1.752s and fuzz 13.094s, each exit 0 under timeout 89. Production-source git diff was empty.

Committed-branch lane command: git fetch -q origin main devtools/fast-gate cloud/merge-tree && git show origin/cloud/merge-tree:cloud/integration/lane-checks.py | python3 -. Exit 0: lane checks 4.2 s: gofmt and tools on 2 Go files, t.Parallel on 2 test packages; vet 2 packages.

Automatic approval review initially rejected an end-of-file deletion because it said the retained collision test could also be removed. The full file showed that test preceding the target. An exact-text edit with explicit preservation assertions was subsequently approved; no action remains blocked.
