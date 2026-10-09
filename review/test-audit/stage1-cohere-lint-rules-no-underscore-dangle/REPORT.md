Unit u121, starting origin/main ce1c5a2fd91e40b05160e587ea5d88ed5bdfedb2.
One top-level Test exists; no family grouping or skipped rows.
Clean baseline passed; all four production mutants survived.
TestCompileProfiles: untrue under this four-mutant experiment; empty-answer probe passed.
All mutants compiled natively; every survivor has a changed-output witness.

```json
[
  {
    "test": "TestCompileProfiles",
    "package": "stage1/cohere/lint/rules/no-underscore-dangle",
    "file": "stage1/cohere/lint/rules/no-underscore-dangle/build_test.go:16",
    "seconds": 31.141,
    "oracle": "Self: successful load.Load, lower.Lower and native.Build in sanitized and release modes. Emitted JavaScript is only written to a file. The row runs no executable and compares no lint findings, messages, spans or fixes. Source-Node executions in this audit are survivor witnesses, not this row's oracle.",
    "oracle_kind": "self",
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": null,
    "verdict": "untrue",
    "subsumed_by": [],
    "mutants_in_matrix": 4,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": true,
    "bounded": false,
    "matrix_rows": [
      "TestCompileProfiles"
    ],
    "evidence": "timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/lint/rules/no-underscore-dangle/ -run . > P1.log 2>&1; --- PASS: TestCompileProfiles (21.53s)"
  }
]
```

| ID | Origin file:line | Change | Rows failed |
|---|---|---|---|
| M1 (production) | stage1/cohere/lint/rules/no-underscore-dangle/rule.a:8 | `name.startsWith('_')` → `name.startsWith('$')` |  |
| M2 (production) | stage1/cohere/lint/rules/no-underscore-dangle/rule.a:7 | `this.context.settings.read(name, fallback) === 'true'` → `this.context.settings.read(name, fallback) !== 'true'` |  |
| M3 (production) | stage1/cohere/lint/rules/no-underscore-dangle/profile.a:32 | `    visit(context, rule, root);` → `` |  |
| M4 (production) | stage1/cohere/lint/rules/no-underscore-dangle/rule.a:27 | `children[offset + 1] ?? -1` → `children[offset + 0] ?? -1` |  |
| P1 (probe) | stage1/cohere/lint/rules/no-underscore-dangle/profile.a:17 | `function run(row: string, countOnly: boolean): number {` → `function run(row: string, countOnly: boolean): number {
    if (row.length >= 0) return 0;` |  |

Survivors:
M1: unguarded runtime behavior in this package. The same source-Node profile command changes findings 6 → 1, with exit 0 and fixed source unchanged. See witness-M1.log, witness-baseline.log and the full witness_diff in results.jsonl.
M2: unguarded runtime behavior in this package. The same source-Node profile command changes findings 6 → 3, with exit 0 and fixed source unchanged. See witness-M2.log, witness-baseline.log and the full witness_diff in results.jsonl.
M3: unguarded runtime behavior in this package. The same source-Node profile command changes findings 6 → 0, with exit 0 and fixed source unchanged. See witness-M3.log, witness-baseline.log and the full witness_diff in results.jsonl.
M4: unguarded runtime behavior in this package. The same source-Node profile command changes findings 6 → 5, with exit 0 and fixed source unchanged. See witness-M4.log, witness-baseline.log and the full witness_diff in results.jsonl.

Brief ambiguities, findings and time costs:
The package is a standalone compilation gate, not the semantic agreement suite described by the rule README. The fresh go test -list output contains only TestCompileProfiles. The README describes historical validate.py runs and outside-package checks; none is treated as a current-session test observation. No rows moved or vanished relative to the fresh list.
The row loads profile.a, lowers it, compiles two native modes, writes emitted JavaScript, and logs that all three profiles compiled. Its emitted JavaScript is only written by os.WriteFile: it is not parsed by Node, executed, or compared. Successful writing is weaker than the log wording suggests. Neither native executable is invoked.
The code-under-test entry for the profile is run(row,countOnly). P1 returns zero immediately for every possible string input using if (row.length >= 0) return 0. The guard preserves TypeScript checking of the unreachable-at-runtime remainder. The row still passes, so vacuous=true refers to this entry, not to preparation helpers such as Load. The external witness confirms there are no findings or fixed-source output from that entry.
The untrue verdict is defined by the four selected validly compiling behavioral mutants. It does not claim the test can never fail: loading, lowering, file creation or native compilation errors can fail it. No intentionally invalid syntax, type error, broken import or clang failure was used as a semantic kill. This compile gate cannot distinguish these altered runtime answers.
There is no executable external oracle inside this Go row. The live Node hook runs the .a source only to prove before/after behavior changed for each survivor. Node is not counted as external-run for the row. No Go cohere oracle, ESLint, prior validation output or historical expected value was substituted for a current-session row assertion.
The fixed menu spans four distinct operations in three rule methods and the profile driver. It was recorded before any mutant was executed. Since there is only one row and a stage1 port needs native recompilation, four production mutants meet the fallback limit; no selector switch or harness edit was necessary. Each mutant receives /tmp/u121/cache/<id>. The standalone diffs apply to the actual starting origin/main and have no switch.
Both native modes are rebuilt by the existing row on every variant. Timing in this report combines loading, lowering, sanitized build, release build and JavaScript writing; the unchanged test does not emit separate native phase durations. Measuring individual phases would require additional instrumentation or separate builds, so these phase costs are not invented. Runtime output witnesses are separately timed in commands.jsonl.
The shared witness input and settings were chosen before mutation results. It includes declarations, a trailing underscore, a property, function parameters and an object-binding rename. Each survivor has an actual observed output change, not a hypothetical behavior claim or equivalent-candidate label. The witness is saved as review evidence, not added to the audited test package.
No build or run exceeded 90 seconds. The full package ran for every mutant. No package narrowing, unknown matrix cells, witness/setup harness weakening or skipped tests were needed. Package-wide survival is demonstrated here; other repository packages may catch these diffs during central replay.
The initial instruction names no historical test rows or commit for this unit. Fresh origin/main and the test list are therefore the complete scope. The environment was warm; cloud setup was skipped. npm ci in stage3/api preceded the baseline. The Node hook uses built-in stripping, so no additional node_modules directory was required.
A first regex inventory also recognized if/for control statements. Before mutant execution it was replaced with a TypeScript AST inventory of the exact owned function declarations, methods, constructor and sorting callback. Imported context/parser/scanner/compiler internals were not mutated and are outside the owned rule package.

Timing and coverage:
Warm toolchain: Go 1.27.1, nproc 5. Setup and npm installation were not separately timed. npm.log preserves installation output. Clean baseline binary seconds: 30.337.
Three isolated -count=1 binary samples: [31.141, 30.765, 32.633]; median 31.141 seconds.
| Variant | Test binary rebuild/check seconds | Process wall seconds | Witness wall seconds |
|---|---:|---:|---:|
| M1 | 30.47 | 32.430 | 0.150 |
| M2 | 30.75 | 32.382 | 0.207 |
| M3 | 31.25 | 32.947 | 0.179 |
| M4 | 31.3 | 36.479 | 0.190 |
| P1 | 21.53 | 23.870 | 0.142 |

Not covered: semantic parity against Go cohere or ESLint, execution of emitted JavaScript or either native witness binary, other packages, invalid-option behavior, full corpus, and compiler/runtime mutation coverage. These are intentionally not claimed.
Completed 2026-10-09T12:39:14.144605+00:00. Every source mutation was reversed; only evidence is committed.
