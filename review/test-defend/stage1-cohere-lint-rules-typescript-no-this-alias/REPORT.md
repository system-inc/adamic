TestCompileProfiles defended as a production compiler acceptance smoke test.
D1 uniquely fails the entire one-row package; no test or oracle was changed.
Starting main c4c5991448f93d9be6e17d9256617cea9a2a087a; source restored and standalone diff validated.

[
  {
    "test": "TestCompileProfiles",
    "package": "stage1/cohere/lint/rules/typescript-no-this-alias",
    "prior_verdict": "untrue",
    "subsumed_by": [],
    "defense": "defended",
    "unique_mutant": "D1 internal/lower/lower.go:22",
    "attempts": [
      {
        "mutant": "D1",
        "file_line": "internal/lower/lower.go:22",
        "change": "if len(files) != 1 { -> if len(files) == 1 {",
        "rows_failed": [
          "TestCompileProfiles"
        ]
      }
    ],
    "evidence": "ADAMIC_BUILD_CACHE_DIR=/tmp/defend-no-this-alias/cache/D1 timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/lint/rules/typescript-no-this-alias/ -run .; build_test.go:28: lower: stage 0 compiles a program from one entry file, got 1",
    "bounded": false,
    "rows_passed": []
  }
]

Code under test and oracle

The compiled program is profile.a, importing rule.a, Parser, Scanner, RuleContext, Settings and their dependencies. The executed code under test is Adamic load.Load, lower.Lower, native.C/Build and javascript.JavaScript. The test checks compilation success only. Its oracle is its self-written expected absence of errors and the clang native build result; no rule-answer oracle is invoked.

Coverage and matrix

The full package contains exactly TestCompileProfiles. Coverage command: timeout 120 go test -json -count=1 -timeout 90s -coverpkg=./internal/load,./internal/lower,./internal/native,./internal/javascript -coverprofile=coverage.out ./stage1/cohere/lint/rules/typescript-no-this-alias/ -run ^TestCompileProfiles$. Go projected covered lines: 7800. Reached-function list and individual blocks are retained. No .a runtime function executes in this test.

D1 internal/lower/lower.go:22: if len(files) != 1 { -> if len(files) == 1 {. Full package fails only TestCompileProfiles. Passing rows: [] because there are no other rows. ADAMIC_BUILD_CACHE_DIR=/tmp/defend-no-this-alias/cache/D1 timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/lint/rules/typescript-no-this-alias/ -run .; build_test.go:28: lower: stage 0 compiles a program from one entry file, got 1.

Brief ambiguities, costs and limits

1. The audit classified three runtime rule mutations as survivors. Those observations do not establish that a compilation smoke test cannot fail. Its runtime-only .a code-under-test framing omitted the compiler acceptance behavior its assertions actually exercise. This defense targets that production compiler behavior.
2. The supplied prior failing fragment is a passing log message, not a failure. Full-refspec audit evidence resolves that ambiguity; its report explicitly says compilation regressions can fail despite the untrue runtime verdict.
3. There is only one top-level test, unchanged from the audit. No subsumer or rest-of-package test exists to cover it. A second rest-of-package coverage run would execute no tests. Actual per-test coverage of load/lower/native/javascript and reached functions is retained; all projected covered Go lines are exclusive only within this one-row package.
4. The oracle is self-written expected success for load/lower and two clang native builds, plus successful writing of the JavaScript emitter output. No Go cohere, Node or ESLint runtime comparison runs. The JavaScript file is written but never syntax-checked or executed, despite the final log saying it compiled.
5. Compiler mutation D1 flips the valid single-entry guard in Lower. It is a small genuine production input-acceptance regression. The test catches it before C generation, rather than through a clang warning, runtime failure, oracle edit or harness mutation. go vet validates the changed Go package and git apply --check validates the standalone diff.
6. The mutation has its own ADAMIC_BUILD_CACHE_DIR. No mutated native product is created because Lower rejects first; the Go test binary recompiles the changed compiler. Its 10.530s command wall time and 0.102s binary duration are recorded separately.
7. One uniquely caught first attempt is enough; no second or third attempt is needed. No Node/native twin pair exists in this package. Profiles here means compiled owned profiles; neither the name nor assertions specify a speed or budget threshold, so it is not a cost row.
8. The test guards compilation acceptance, not rule findings, emitted runtime correctness, sanitizer effectiveness, JavaScript syntax, or performance. Prior runtime survivors and empty-entry findings are not disproven by this defense. No runtime semantic-quality claim is made.
9. Warm env.sh worked and setup was skipped; npm ci still ran. nproc=5. The clean whole package passed at 31.815s and its coverage run passed at 32.051s with 29.5% combined Go statement coverage. No timeout, narrowing, panic, skip, red baseline, test edit, oracle edit or PR occurred.
10. No unresolved row remains. The compile-oriented name matches the native-build assertions; the emitted-JavaScript compiled log overstates its write-only check. Wider package or repo-wide redundancy is outside this matrix.

Evidence includes original audit reports, clean logs, coverage, standalone D1.diff, matrix, Go vet output, apply-check output and exact wall/binary timings. No further native rebuild is claimed for the rejecting mutant.
