Built: result assertions in all seven enum acceptance tests, plus a Node comparison of member values and the last reverse alias, task #5w441z8.
Commits: tests bbf1476dfacddb5e3b0750e04ada0425378d12ce; evidence f3ad4fbbcc73e777773eee22b959cc83f9149921; current main merged in 60fa8bb3672956d097be66f3c81e4cab6eb43b54.
Commands and outputs: eight focused lowering leaves and the existing enums.a oracle pass after restoration; lane checks, including go vet, pass.
Mutants: M01, M02, M04, M14, M15, M16 and each of the seven individual empty-answer probes fail named assertions.
Not covered: other audit mutants, stage1 ports, whole packages, or the full gate.

This is test-only work toward the named enum-lowering guard task. The seven formerly acceptance-only tests now require exact forward member constants and reverse names in the emitted enum object. Alias cases require one reverse slot containing the last declared alias, preserving the field's original position. The never-default test requires its runtime panic guard. Enum name iteration requires ObjectKeys IR and Node's numeric-key-first order; inline iteration requires the ordered numeric array and its bitwise-OR operands. Both iteration tests compare source Node, generated JavaScript and native output.

Observed: numeric IR has no flag-domain annotation, and open numeric enums can still lower successfully when flag recognition changes. The tests therefore also query the actual checker-backed flag proof. Literal spellings and the domain test require recognition of bit 30, whose emitted constant is exactly 1073741824. Domain checks require AND to accept a proven operand on either side, accept two proven operands, and reject two ordinary numbers; OR with an ordinary number is rejected, while XOR of two proven flags is accepted. The numeric-enum test also directly requires the whole enum's open-type proof.

Observed: all exact audit diffs apply to main. They come from `test-audit/internal-lower-enum_flags` at b074cb8226d6c9594b638616c83265e78364ceed, whose reported audit base was 8171b3173bdb. Replay applies one diff at a time and reverses it in a finally block. No production mutation remains. Every mutant test run exits 1 through an assertion; no replay log has a build failure, panic, or timeout.

| Mutant | Strengthened original test that catches it | Observed mismatch |
| --- | --- | --- |
| M01 | All seven | Numeric member constants increase by one |
| M02 | All seven | Reverse names become empty strings |
| M04 | TestFlagEnumMemberAliases; TestEnumNameEnumeration | A duplicate reverse field is appended instead of updating the existing slot |
| M14 | TestFlagEnumsDomain; TestFlagEnumLiteralSpellings; TestFlagEnumMemberAliases | Flag recognition becomes false |
| M15 | TestFlagEnumsDomain; TestFlagEnumLiteralSpellings | The bit-30 flag declaration is no longer recognized |
| M16 | TestFlagEnumsDomain | Both left-proven and right-proven AND expressions lose their domain proof |

The exact audit `P_LOWER.diff` returns nil, nil from Lower. Each of the seven tests was run separately under that probe, and each failed with `lowering returned empty IR`. Their seven logs and exact commands are recorded in `replay-results.json`. Restoration was followed by a passing run of all eight focused leaves.

The new independent witness is `enum E { A = 1, B = 2, Alias = A } console.log(E.A + ' ' + E.B + ' ' + E.Alias + ' ' + E[1]);`. Source Node prints exactly `1 2 1 Alias\n`. M01 makes both generated backends print `2 3 2 Alias\n`; M02 makes both print `1 2 1 \n`. These mutated executions exit successfully without sanitizer findings, so only the output comparison catches them. Native comparisons use ASan/UBSan and normal Linux leak checking. The fixture is written as a temporary .a file; no registered oracle fixture was added and allocation counts remain unchanged, so counts regeneration was not needed.

The existing registered fixture `internal/oracle/testdata/enums.a` also catches M01. It passes normally, then reports `stdout differs` for native and generated JavaScript with M01, with clean exits and empty stderr. For example, Node starts with `0: zero First`, while both mutated backends start with `1: zero Second`. After restoring production code, the same uncached oracle passes again; its restored leaf took 0.66 s. It also exercises native release and leak checking.

Inference: checking only successful admission permitted empty lowering answers and concealed changed enum representations. Assertions on the emitted object catch representation changes at their source; direct proof assertions cover behavior that is not stored in the IR; independent Node observations check values and enumeration order without deriving expectations from Adamic's own output.

Restored leaf durations on the four-CPU cgroup instance:

| Leaf | Seconds |
| --- | ---: |
| TestFlagEnumsDomain | 0.81 |
| TestEnumNeverDefault | 0.08 |
| TestFlagEnumLiteralSpellings | 0.17 |
| TestFlagEnumMemberAliases | 0.17 |
| TestEnumNameEnumeration | 0.54 |
| TestFlagEnumInlineIteration | 0.53 |
| TestNumericEnumsAreOpen | 0.70 |
| TestEnumMemberValuesAndReverseNameMatchNode | 0.57 |

The touched domain and numeric-enum tests now run their short probe tables inside top-level parallel leaves instead of adding subtests. Each external program has a 10 s process deadline. Each Go test has `-timeout 90s`; the shell commands have 120 s outer deadlines, and the replay driver bounds each Go invocation to 120 s and each git apply to 10 s.

Exact validation commands, from the repository root after sourcing `/workspace/adamic-tools/env.sh` and setting `GOPROXY='https://proxy.golang.org|direct'`:

```sh
timeout 120s go test ./internal/lower -run '^Test(FlagEnumsDomain|EnumNeverDefault|FlagEnumLiteralSpellings|FlagEnumMemberAliases|EnumNameEnumeration|FlagEnumInlineIteration|NumericEnumsAreOpen|EnumMemberValuesAndReverseNameMatchNode)$' -count=1 -v -timeout 90s
ADAMIC_GATE_UNCACHED=1 timeout 120s go test ./internal/oracle -run '^TestNativeAgreesWithNode$/^internal$/^oracle$/^testdata$/^enums.a$' -count=1 -v -timeout 90s
timeout 600s python3 review/compiler/enum-guards/replay.py
```

The replay driver runs only the seven named original tests for each semantic mutant, adding the independent Node witness for M01 and M02. It runs each original test separately for P_LOWER. The existing oracle's M01 run was applied and restored separately. Test output was redirected directly to files: `guards-baseline.log`, `guards-restored.log`, `oracle-baseline.log`, `oracle-M01.log`, `oracle-restored.log`, and the per-replay logs. Exact selected tests, exit codes and failed leaves are in `replay-results.json`.

Setup succeeded with `timeout 240s bash cloud/setup.sh`. Verbatim output is in `setup.log`: Go ready 0.022 s; Node ready 0.023 s; submodules ready 0.061 s; markdown validation step 0.008 s and ready 0.072 s; clang ready 0.173 s; Go build ready 36.205 s; test binaries deferred 36.365 s; build cache warm 36.366 s; setup done 36.392 s. `nproc` returned 5; `cpu.max` was `400000 100000`, a four-CPU quota. Versions: Go 1.27.1, Node 24.19.0, clang 20.1.8.

The branch started at main 83f3940ec77b8ba779da05ded113e9d9e7e710b6 and was fast-forwarded to current main 0942c5169d0ea736d9dfaa19881af1ad8adad162 before final validation. No unlanded worker branch was merged. Required lane checks after the test commit:

```sh
timeout 60s git fetch -q origin main devtools/fast-gate cloud/merge-tree
timeout 150s bash -o pipefail -c 'git show origin/cloud/merge-tree:cloud/integration/lane-checks.py | python3 -'
```

Output: `lane checks 1.3 s: gofmt and tools on 3 Go files, t.Parallel on 1 test packages; vet 1 packages`. The same required checks are repeated after the evidence commit before pushing. Delivery contains only test files and review evidence; mutant files are .diff, and no Go source lives under review.

Current main advanced to 68db8ddd145281a62655452496bdc32ef848bdf3 during final checks. It was merged normally in 60fa8bb3672956d097be66f3c81e4cab6eb43b54. The same eight focused tests and the same uncached oracle were rerun and passed, in 0.907 s and 0.694 s total respectively; the largest lowering leaf was 0.75 s. Lane output was `lane checks 1.0 s: gofmt and tools on 3 Go files, t.Parallel on 1 test packages; vet 1 packages`. The final-main evidence is in `current-main-guards.log`, `current-main-oracle.log`, and `current-main-lane.log`. Main's additional changes were confined to meter tests and their testdata.
