Built wave 3 of #41bkfdw: 76 acceptance rows use the shared Node agreement helper; one checked-failure acceptance row is explicitly exempted.
Implementation commit: 15ca2016; current-main merge: ca1ad1f1; delivery SHA is reported with the push.
Focused tests and TestCallTargetReaders pass with -timeout 90s; lane checks pass.
Two new one-line mutants pass old tests and fail converted tests; nine existing guard mutants remain caught, M12 by its original paired IR guard.
Not covered: full packages, full gate, native agreement beyond the preserved enum guard, or ordinary source agreement for the readiness soundness probe.

| File | Accept | Refuse | IR-only |
| --- | ---: | ---: | ---: |
| namespaces_test.go | 3 | 35 | 0 |
| nested_functions_test.go | 4 | 9 | 0 |
| enums_test.go | 22 | 23 | 0 |
| enum_flags_test.go | 43 | 4 | 0 |
| enum_guards_test.go | 2 | 0 | 2 |
| statics_guards_test.go | 3 | 0 | 0 |

The census has 150 entry rows: 77 accept, 71 refuse, two IR-only. The readiness-message acceptance row remains a checked-failure soundness probe: source Node reads undefined and prints NaN, while Adamic intentionally exits 70. It uses the shared bounded Node observation runner and preserves the exact panic text. No rows deleted; no golden IR snapshots.

Erasure, environment allocation count and layout, captured ownership, closed-frame ownership proof invalidation, and native enum field-slot uniqueness retain direct assertions with comments naming what JavaScript behavior cannot observe. Flag-domain proof rows remain IR-only because open numeric lowering bypasses those proofs.

The local runners in enum_guards_test.go and statics_guards_test.go are removed. enum_agree_test.go stays outside this unit because enums_open_test.go still calls it. requireLoweredOutput stays for existing callers outside this unit. M12 is invisible in the factory-output row; the original paired TestParserFactoryBindingHoisting still catches its lost NamespaceVar metadata. The surviving output-only attempt is recorded honestly in mutants.json.

Commands and observations:

- `export GOPROXY='https://proxy.golang.org|direct'; timeout 240 bash cloud/setup.sh`: tooling and submodules ready, then timeout during whole-repository cache warming. `source /opt/adamic-tools/env.sh` failed because that path does not exist. The installed environment is `/workspace/adamic-tools/env.sh`.
- Cached retry `timeout 120 bash cloud/setup.sh`: exit 0. Timing lines: Go 0.073s, Node 0.075s, Markdown dependencies 0.181s, submodules 0.227s, clang 0.511s, build ready 62.922s, cache warm 63.039s, done 63.071s. `nproc`: 5; cgroup quota: four CPUs.
- First broad selector exhausted its outer timeout while warming the initial compiler cache and also selected unrelated tests. It is superseded by the exact assigned-file selector in test-pattern.txt.
- `timeout 100 go test ./internal/lower -run "$(cat review/compiler/agree-modules/test-pattern.txt)" -timeout 90s -count=1 -v > review/compiler/agree-modules/tests-final.log 2>&1`: PASS, package 4.552s. Every selected leaf below 60s; timings.json records each passing test's seconds. No new top-level tests.
- `timeout 100 go test ./internal/ir -run '^TestCallTargetReaders$' -timeout 90s -count=1 -v > review/compiler/agree-modules/call-targets.log 2>&1`: PASS, 19.926s.
- `timeout 900 python3 /tmp/mutants-agree.py`, followed by `timeout 230 python3 /tmp/mutants-finish.py`: each mutation is restored in finally; commands, statuses and failures are recorded in mutants.json. Initial M12 output-only attempt survived, then the original paired replay caught it. The corrected reproducible script is saved as replay.py.
- `git fetch -q origin main devtools/fast-gate cloud/merge-tree && git show origin/cloud/merge-tree:cloud/integration/lane-checks.py | python3 -`: PASS. This checkout only tracked main, so explicit refspec fetches first populated origin/cloud/merge-tree and origin/devtools/fast-gate.

Mutants and what caught them:

| Mutant | Result |
| --- | --- |
| enum-value.diff: numeric enum constant + 1 | Old TestConstEnumErasesRuntimeObject passed; converted stdout 5 versus source 4 failed. |
| nested-rest.diff: RestElement = 0 | Old TestNestedRestIsSupported passed; converted stdout undefined versus source 1 failed. |
| enum M01 | Shared source stdout comparison in TestEnumMemberValuesAndReverseNameMatchNode. |
| enum M02 | Shared source stdout comparison in TestEnumMemberValuesAndReverseNameMatchNode. |
| enum M04 | Named duplicate native field-slot assertion in TestEnumReverseMappingUsesSingleSlot. |
| enum M14 | Direct flag-classification proof assertion. |
| enum M15 | Direct bit-30 flag-classification proof assertion. |
| enum M16 | Direct flag-domain proof assertion. |
| statics M06 | Shared source stdout comparison in TestPrivateAndPublicStaticsAgreeWithNode. |
| statics M12 | Original paired TestParserFactoryBindingHoisting metadata assertion; factory output alone passes. |
| statics M13 | Exact readiness panic message includes new Box().n. |

No new oracle fixtures were added, so counts.md requires no refresh. Assumption: source-row count treats an import dependency as part of its main entry row and the six functions in the domain-proof program as one source row. Refusal rows retain their original programs and contracts.
