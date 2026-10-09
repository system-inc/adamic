u123: one row discovered and audited at ce1c5a2fd91e40b05160e587ea5d88ed5bdfedb2.
The clean whole-package baseline passed in 30.621 binary seconds.
TestCompileProfiles has a three-run median of 30.809 seconds; nproc=5.
Three semantic mutants survived, each with changed native output.
Verdict untrue is limited to semantic port behavior; the Lower nil probe fails.

```json
[
  {
    "test": "TestCompileProfiles",
    "package": "stage1/cohere/lint/rules/no-unsafe-optional-chaining",
    "file": "stage1/cohere/lint/rules/no-unsafe-optional-chaining/build_test.go:16",
    "seconds": 30.809,
    "oracle": "Handwritten successful-build expectation: lower.Lower and sanitized/release native.Build must succeed; emitted JavaScript is only written, not compiled or executed. No lint findings or runtime behavior are compared. Three semantic native-output changes pass.",
    "oracle_kind": "self",
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": null,
    "verdict": "untrue",
    "subsumed_by": [],
    "mutants_in_matrix": 3,
    "probe_kills": [
      "P1"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": false,
    "matrix_rows": [
      "TestCompileProfiles"
    ],
    "evidence": "M1/M2/M3: timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/lint/rules/no-unsafe-optional-chaining/ -run .; all --- PASS: TestCompileProfiles. P1 alone -run ^TestCompileProfiles$: --- FAIL: TestCompileProfiles; panic: runtime error: invalid memory address or nil pointer dereference; build_test.go:30. P1 is a probe, not a production mutant kill."
  }
]
```

Code under test and oracle, declared before mutation

The owned profile.a and rule.a are compiled by lower.Lower, native.C, native.Build and javascript.JavaScript. The oracle is self: the handwritten expectation that compilation/build and the file write succeed. The row executes no produced program. There are no default skips, other top-level rows, families, helpers or witnesses in this package. The lintoracle-tagged oracle.go contains no Test and is an external Go cohere adapter, left untouched. No Node or Go cohere differential comparison occurs in this row.

Function inventory and reach limitation

The directory has no default production Go functions; its sole Go Test calls the compiler APIs above after load.Load prepares the checked input. Port source functions compiled from this directory are Rule.constructor, Rule.pattern, Rule.reaching, Rule.visit and create in rule.a, plus ancestry, visit and run in profile.a and the profile top-level entry. The test reaches these as compiler input, not by executing their runtime bodies. Native witness runs separately reach them through the manifest input. Compiler internals span other packages; an exhaustive compiler call graph was not collected. Production mutants were confined to the listed port functions, so no claim of exhaustive compiler mutation coverage is made.

Fixed mutant menu, chosen before results

| ID | Original file:line | Change | Failed rows |
|---|---|---|---|
| M1 | rule.a:22 | flip `if(!node.optional)` to `if(node.optional)` | none |
| M2 | rule.a:32 | change arithmetic default from `'false'` to `'true'` | none |
| M3 | rule.a:28 | drop the entire report statement | none |
| P1, probe only | internal/lower/lower.go:20 | return nil, nil at Lower entry | TestCompileProfiles |

All rule.a locations use the directory named in the package field. M3 leaves no unused local. Each M diff has no switch and applies to the starting origin/main. Each was validated through the row's sanitized and release builds and a separate CLI native build. P1.diff is explicitly an empty-answer probe, guarded by constant true to leave ordinary imports referenced; go vet ./internal/lower/ succeeded. The audited package's go vet also succeeded. No test, comparison, oracle, dispatch, corpus or copied-file list was mutated.

The brief recommends a shared switch but also permits at most four separate native mutants when rebuilding. This audit chose three separate mutations, restored between them, rather than adding a runtime selector dependency. ADAMIC_BUILD_CACHE_DIR was distinct for M1, M2 and M3. No compiler production mutation occurred. P1 uses its own cache and fails before any native product build.

Survivors and native witnesses

All native commands use `/tmp/u123/manifest.txt`, naming `/tmp/u123/input.ts`, whose contents are retained as witness-input.ts.txt:
`(obj?.foo).bar;` followed by `obj?.foo + 1;`.

Clean: native output includes `finding 1 9 unsafeOptionalChain` and no unsafeArithmetic finding.
M1: native output has no finding lines; the parenthesized unsafe access loses its clean finding.
M2: native output adds `finding 16 24 unsafeArithmetic` even though the manifest supplies no arithmetic setting.
M3: native output has no finding lines; reporting has been removed.

Commands: `/tmp/u123/adamic build <package>/profile.a -o /tmp/u123/<id>` then `/tmp/u123/<id> /tmp/u123/manifest.txt`. Every executable exited 0. Separate clean/M*-witness.log files retain byte output. These witnesses establish changed behavior, not correctness against an outside authority, and are not counted as test kills. None of the mutants is an equivalent candidate.

Empty-answer probe

P1 makes Lower return nil, nil. The row fails at its native.C call with a nil-pointer panic at unchanged build_test.go:30. The whole package contains only this row; it was also rerun alone after the panic, and failed again. There are no unobserved later rows. probe_kills is [P1], while kills remains empty. vacuous=false is scoped to the Lower entry the Go row calls. It does not imply that runtime findings are checked. Other compiler-stage APIs and the port's runtime entry were not separately empty-probed.

Ambiguities, scope and cost

The unit is a stage1 port directory, but its only enumerated row proves buildability, not lint agreement. Under the brief's definition, zero semantic-mutant kills gives untrue. That label should not be read as proof that the compilation assertion can never fail: the nil probe demonstrably fails, and broken source could fail compilation. Intentionally creating compiler-rejected port source would test the compilation precondition and would not prove sensitivity to lint semantics, so no such artificial kill was credited.

The row logs that emitted JavaScript was compiled, but the code only obtains a string and writes emitted.mjs; it never asks Node to parse or run it. Sanitized native code is built but never executed, so this row does not establish sanitizer cleanliness. These are directly observed code facts, not mutation-based findings.

The phrase one probe per entry is ambiguous for a compilation row that chains Lower, C, Build and JavaScript. This audit chose Lower, the root compiler entry after preparation, and reports other stages as unprobed rather than claiming full empty-output coverage. vacuous=false only proves sensitivity to a nil IR, not to an empty executable or absent diagnostics.

The function-inventory requirement spans an entire compiler for this compilation row. The listed inventory is complete for functions defined in this owned port directory; compiler transitive functions were not exhaustively enumerated. Package uniqueness would be meaningful for a kill here because the whole package ran, but there are no production kills. Repo-wide and parent lint harness coverage are unknown and were not run.

The first witness script launch happened before the clean native CLI build completed and failed with FileNotFoundError. No mutation was applied at that point. It was rerun after the build finished, and all resulting logs are from completed runs. This was an orchestration error, not a test failure or mutant kill.

Timing and uncovered work

Warm env.sh worked, setup skipped (0 seconds), nproc=5. npm ci stage3/api: 0.434 wall seconds. Required discovery: 4.787 wall seconds. Clean baseline: 32.371 wall seconds, 30.621 test binary seconds. Separate binary timing runs: 30.528, 30.809, 30.980; median 30.809. CLI clean build: 11.509 seconds. M1/M2/M3 matrix wall seconds: 32.335, 32.241, 32.562. Their additional native witness builds: 11.540, 11.424, 11.575. Matrix times include Go invocation overhead and both sanitized/release builds; the test has no separately reported per-product build timing. Probe vet: 0.478 seconds; P1 first invocation: 9.449 seconds, binary 0.068 seconds including the panic. No run exceeded the prescribed budget. Detailed commands and durations are in timings.json, row-timings.json, mutation-timings.json, probe-timings.json and restored-timing.json.

No source changes remain. No other test package, optional lintoracle overlay, full lint agreement harness or repository gate was run. Three port mutants provide limited semantic coverage, not an exhaustive claim. Evidence was stored outside the checkout until all test runs finished, then copied to review/test-audit/stage1-cohere-lint-rules-no-unsafe-optional-chaining/. No PR was opened.
