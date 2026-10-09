TestCompileProfiles is defended for native profile buildability.
D1 uniquely fails the only current package row; sanitized succeeds and release linking fails.
Semantic lint correctness, sanitizer execution and JavaScript validity remain unchecked.

Starting origin/main: bc9edc5560379b63d4688a21006efb6032b60434.
Read CLAUDE.md, README.md, docs/0.1.md, docs/memory.md, audit REPORT.md, rows.json, plan.json and matrix.json. Audit origin/test-audit/stage1-cohere-lint-rules-no-unsafe-optional-chaining was fetched with a full remote refspec. Its historical starting commit ce1c5a2f differs from this session's main. No historical run is used as new proof.

CODE UNDER TEST: Adamic lower.Lower, native.C, native.Build and javascript.JavaScript applied to the owned profile.a and its imports. The row compiles the port rather than running Rule.visit. ORACLE: self, successful preparation/lowering, successful sanitized and release native builds, successful emitted JavaScript file write. These were declared before mutation. The native compiler's runtime linking is production code under test, not the oracle or harness. No test, oracle, port or harness edits were made.

Discovery: go test -list . ./stage1/cohere/lint/rules/no-unsafe-optional-chaining/ found TestCompileProfiles and no other top-level tests. No family, twin, helper or opt-in row exists in this package. No newly added row was omitted. TestCompileProfiles has no performance threshold or timing assertion, so no answer-preserving cost-growth attempt is required.

Baseline: npm ci --prefix stage3/api completed before the whole-package baseline. Warm env.sh worked, setup skipped, nproc=5. Clean baseline passed in 29.661 binary seconds, row 29.65 seconds. No skipped rows. Baseline and dependency logs are retained.

Coverage: row-coverage.log ran timeout 120 go test -json -count=1 -timeout 90s -run '^TestCompileProfiles$' -coverpkg=github.com/system-inc/adamic/internal/lower,github.com/system-inc/adamic/internal/native,github.com/system-inc/adamic/internal/javascript -coverprofile=/tmp/defend-unsafe-profile/row.cover ./stage1/cohere/lint/rules/no-unsafe-optional-chaining/. Rest used the same command with -run '^$' and rest.cover, because the current package has no other row. Own binary passed in 30.296 seconds; empty-rest binary passed in 0.027 seconds. 4,750 Go blocks were executed by the own run, 4,741 exclusively against the empty rest. Package initialization explains the nonexclusive blocks. In particular native.go:145.2,146.37 containing the libm argument is exclusive. These are Go compiler coverage profiles, not runtime coverage of .a bodies. All positive blocks and original profiles are retained. With no other rows, exclusivity is trivial, but the release link failure still establishes the promised compilation check can catch a real compiler regression.

Aimed attempt, recorded in plan.json before execution: D1, fixed menu drop a statement, internal/native/native.go:145 on starting origin/main. Remove arguments = append(arguments, "-lm"). The existing source comment explains that Linux release builds need this explicitly, while sanitizer runtime happens to supply it. This targets the difference between sanitized and release profiles, which prior semantic rule mutations did not affect. Go vet ./internal/native/ succeeds (0.273 wall seconds), so the production Go compiler mutant builds. The standalone unified diff applies cleanly to starting origin/main, has no switch, and does not insert any statement.

Whole-package matrix: timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/lint/rules/no-unsafe-optional-chaining/ -run ., ADAMIC_BUILD_CACHE_DIR=/tmp/defend-unsafe-profile/cache/D1. Exit 1, 35.571 wall seconds, 30.361 binary seconds. Failure: build_test.go:34: native: clang failed: exit status 1; main.c undefined reference to floor and fmod; --- FAIL: TestCompileProfiles (30.35s). Line 34 is the release Build assertion. The sanitized Build at line 30 succeeded, since the row reached line 34. This is an unresolved-symbol link error, not -Werror, an invalid source edit, an empty-answer probe, or a fabricated return error. Passing other rows: []; unknown other package rows: []; current package has exactly one row. Repo-wide uniqueness is unknown, and other packages were not run. The build assertion is the intended observation, so no native runtime witness is necessary.

Stop after one successful unique defense, per up to three attempts. The production source is restored and the standalone diff rechecked with git apply --check. Restored whole-package run is recorded separately in restored-baseline.log and restored-timing.json. No tests were removed, rewritten or weakened. Evidence was accumulated outside the checkout until all runs finished.

Brief findings and costs: The audit called this row untrue after changing semantic lint answers, but the row's assertions promise buildability, not agreement. A compiler linking mutant reaches that actual promise. Restricting defense to rule.a mutants would miss it. The per-test coverage comparison with the rest is an empty-set comparison here, since discovery finds one row. The test log says emitted JavaScript was compiled, but the assertion only writes a string; Node never parses or executes it. Native sanitizer instrumentation is built but never executed. Its name CompileProfiles fits native compilation, while the emitted-JavaScript success message overstates validation. Keep this row for buildability and do not read this defense as proof of lint findings, emitted JavaScript validity, native output or sanitizer cleanliness. No cost claim is asserted. No run exceeded 90 binary seconds, no narrowing was needed, and no setup installation was needed. npm ci duration was not measured separately. Timing facts above are observed; no estimated setup/build subdivision is reported.

Restored confirmation: {"Time": "2026-10-09T15:24:59.85291468Z", "Action": "pass", "Package": "github.com/system-inc/adamic/stage1/cohere/lint/rules/no-unsafe-optional-chaining", "Elapsed": 30.391}
{
  "command": [
    "timeout",
    "120",
    "go",
    "test",
    "-json",
    "-count=1",
    "-timeout",
    "90s",
    "./stage1/cohere/lint/rules/no-unsafe-optional-chaining/",
    "-run",
    "."
  ],
  "exit": 0,
  "wall_seconds": 32.11503979199915
}
