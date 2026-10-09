Built: behavior comparisons for every row in the seven enum tests, replacing IR snapshots in task #5w441z8.
Commit: e80dcea028d9166e1ea6716935f2c340bc6a9aaa; current main c0a7667baadaf161d6bb9066e0838b32774caa6c merged in cc992c34761411c757fce275beaa36d954d6a66c.
Commands and outputs: ten focused tests pass after restoration; lane checks, including go vet, pass.
Mutants: M01 and M02 fail all seven behavior tests; M04 and M14-M16 fail documented targeted checks; all seven individual empty-answer probes fail.
Limit: M04 and M14-M16 leave the seven behavior tests green, so six failures from behavior alone cannot be reported.

Each row now runs its source through oracle/node.mjs and its lowered program through the JavaScript backend on Node, comparing stdout, stderr and exit code. Functions are called and stored results are printed: bitwise updates, aliases, masking, fields, arrays, maps, parameter passing, optional parameters, switches, numeric casts, mutable containers and optional objects all produce observations. Rows also print enum member values and reverse names. Source must finish successfully and print nonempty output, preventing a silent empty-answer probe.

There was no pushed compiler/lower-agree branch when this revision started: the initial remote-head query returned no matching branch. A minimal local enumLowersAndAgreesWithNode helper was added in internal/lower/enum_agree_test.go. It has no dependency on an unlanded helper branch and contains no enum representation snapshots. Source Node runs before lowering. Every Node process has a 10 s deadline. The seven tests retain top-level parallel leaves and short sequential row tables.

Observed from applying the exact original audit diffs:

| Mutant | Seven behavior tests | Narrowly targeted test |
| --- | --- | --- |
| M01: numeric value plus one | All seven fail output comparison | Independent witness also fails in JavaScript and native |
| M02: reverse name empty | All seven fail output comparison | Independent witness also fails in JavaScript and native |
| M04: duplicate-key update removed | All seven pass | TestEnumReverseMappingUsesSingleSlot fails: duplicate enum field slot "1" |
| M14: literal equality reversed | All seven pass | TestEnumFlagProofsWithoutObservableLoweringEffect fails flag recognition |
| M15: bit 30 excluded | All seven pass | Same targeted proof test fails bit-30 recognition |
| M16: AND requires both operands | All seven pass | Same targeted proof test fails both left-proven and right-proven domain checks |

The targeted tests have comments naming why behavior cannot expose their respective mutations. JavaScript object literals overwrite duplicate keys, preserving M04's final value and key enumeration even though native object fields need unique slots. The targeted IR check asserts only slot uniqueness, not values or an entire object snapshot. Open numeric enum lowering bypasses the classification/domain proofs changed by M14-M16; their booleans are queried directly with the real checker. There are no IR value or iteration-shape assertions in the seven behavior tests.

Inference: demanding six JavaScript output failures would require changing production behavior or claiming an output difference that the replay does not observe. The conservative choice is to complete the behavioral replacement and retain the narrowly explained exceptions permitted by the ruling. This delivers a guard for each mutant while keeping observable enum semantics held to Node.

The independent member/reverse-name witness still compares both JavaScript and ASan/UBSan native execution with source Node. M01 prints `2 3 2 Alias` instead of `1 2 1 Alias`; M02 prints a blank reverse name. Both failures are clean output disagreements. Existing evidence also names the registered internal/oracle/testdata/enums.a fixture, which catches M01 through both backends. That existing fixture was not rerun in this revision; its prior uncached proof remains in the parent directory.

Each original test was separately rerun with exact P_LOWER.diff, whose Lower returns nil, nil. All seven failed `lowering returned empty IR`; there were no panics or build failures. All mutations were reversed in a finally block before restored validation. Original diffs remain .diff under the parent review directory. No production changes remain and no Go files live under review.

Restored durations:

| Leaf | Seconds |
| --- | ---: |
| TestFlagEnumsDomain | 2.22 |
| TestEnumNeverDefault | 0.24 |
| TestFlagEnumLiteralSpellings | 0.29 |
| TestFlagEnumMemberAliases | 0.24 |
| TestEnumNameEnumeration | 0.24 |
| TestFlagEnumInlineIteration | 0.23 |
| TestNumericEnumsAreOpen | 1.31 |
| TestEnumMemberValuesAndReverseNameMatchNode | 0.41 |
| TestEnumFlagProofsWithoutObservableLoweringEffect | 0.15 |
| TestEnumReverseMappingUsesSingleSlot | 0.05 |

Exact commands after sourcing /workspace/adamic-tools/env.sh and setting GOPROXY='https://proxy.golang.org|direct':

```sh
timeout 120s go test ./internal/lower -run '^Test(FlagEnumsDomain|EnumNeverDefault|FlagEnumLiteralSpellings|FlagEnumMemberAliases|EnumNameEnumeration|FlagEnumInlineIteration|NumericEnumsAreOpen|EnumMemberValuesAndReverseNameMatchNode|EnumFlagProofsWithoutObservableLoweringEffect|EnumReverseMappingUsesSingleSlot)$' -count=1 -v -timeout 90s
timeout 600s python3 review/compiler/enum-guards/behavior/replay.py
timeout 60s git fetch -q origin main devtools/fast-gate cloud/merge-tree
timeout 150s bash -o pipefail -c 'git show origin/cloud/merge-tree:cloud/integration/lane-checks.py | python3 -'
```

Replay gives each Go command a 120 s deadline and each git apply a 10 s deadline. Exact selected test names, commands and exit codes are in results.json. Tests log directly to files. baseline.log records an initial optional reverse-lookup witness rejected by the checker; changing its console.log to use an explicit fallback made baseline-fixed.log pass. restored.log records all ten restored tests passing in 2.229 s total.

Lane output: `lane checks 1.2 s: gofmt and tools on 4 Go files, t.Parallel on 1 test packages; vet 1 packages`. These checks are repeated after committing this evidence before pushing. No additional tool was introduced. The new witness sources are temporary .a files, so no registered oracle fixture or recorded allocation count changed.

The existing setup evidence is ../setup.log: setup succeeded in 36.392 s, with nproc 5 and cpu.max 400000 100000 (four-CPU quota). Go ready 0.022 s; Node ready 0.023 s; submodules ready 0.061 s; markdown validation step 0.008 s and ready 0.072 s; clang ready 0.173 s; Go build ready 36.205 s; test binaries deferred 36.365 s; cache warm 36.366 s. Setup was reused for this same-unit revision. No full package, stage1 port or full gate was run.

Current main then advanced to 859dee825a4ef89f5c4ecc0f00e6620ac75c4994 and was merged in 4809214fc72d478731e3b6921299edea7620f0ed. It introduced an FS-specific lowersAndAgreesWithNode helper with a different signature. The initial merged test build caught the name collision before any push. The local enum helper was renamed enumLowersAndAgreesWithNode without changing its comparison logic or mutation assertions. All ten focused tests then passed on merged main in 2.325 s total, with a largest leaf of 2.32 s. These results are in current-main-restored.log; helper-collision.log records the initial build diagnostic. Final lane checks run on the committed merged tree before pushing.
