Built the three source refusals on main and registered all seven native agreement programs.
Commit: recorded by git after this report; base d79302e50c022c997c00f7c68ddb7475f9ee65fc.
Focused tests PASS 7.020s; traced product build PASS 6.230s, 14 settled and 0 refused; read-only rows PASS 2.123s.
Mutants: every, findLast, literal and registration each failed its intended assertion; all restored.
Not covered: full packages/full gate, whole counts remeasurement, and later units (handled sequentially).

Source: squash of 9bc23cf851d81b4772b7e5a7219d46a8aef7ff7d.
The registry now owns the six array-narrowing control sources and FX5 mixed-literal source.
Their rows call lowersAndAgreesWithNodeNative with literal registered names. That helper
already compares source Node, generated JavaScript and native, so the duplicate JavaScript
call is removed. Assumption: the existing named-program registry is the conservative source
of truth for these controls. No new oracle fixture is registered, so no counts row is added.
The original lower testdata and historical source evidence remain in the source squash.

All commands ran at the repository root, after sourcing /workspace/adamic-tools/env.sh,
with GOPROXY=https://proxy.golang.org|direct. Output was redirected directly to logs.

Commands and results:
- timeout 600 bash cloud/setup.sh: exit 0. Node ready 0.064s; Go ready 0.083s;
  clang ready 0.557s; Markdown dependencies ready 1.255s; submodules ready 177.400s;
  shared cache ready 184.275s; go build ready 586.961s; cache warm 587.095s;
  done 587.189s. nproc=5, cpu.max=400000 100000. clang 20.1.8, Go 1.27.1, Node 24.19.0.
- timeout 300 go test ./internal/lower -run '^(TestNativeAgreementRowsNameRegisteredPrograms|TestProduct_LowerAgreementNative|TestArrayNarrowing.*|TestFX5.*)$' -count=1 -v -timeout 90s: PASS 7.020s (focus-warm.log).
- timeout 600 python3 review/compiler/loom-fx3-refusals-3/run-mutants.py: exit 0.
  Revert every: TestArrayNarrowingEveryRefused sees acceptance (got <nil>).
  Revert findLast: TestArrayNarrowingFindLastRefused sees acceptance (got <nil>).
  Revert literal: TestFX5NumberLiteralRefused sees acceptance (got <nil>).
  Drop FX5MixedLiteral registry entry: TestNativeAgreementRowsNameRegisteredPrograms
  rejects the unregistered literal. These are assertion kills, not sanitizer claims.
  Exact patches and individual logs accompany the runner.
- timeout 180 go test ./internal/ir -run '^TestCallTargetReaders$' -count=1 -timeout 90s:
  PASS 64.004s; final restored-source run PASS 29.094s (readers-final.log).
- timeout 180 go vet ./internal/lower: exit 0, no output, both warm and final runs.
- timeout 240 go test ./internal/oracle -run '^(TestPredicateMiscompileRefusals|TestPredicateDirectionCountsAreRecorded|TestCountsAreRecordedTable)$' -count=1 -v -timeout 90s:
  PASS 34.106s. Existing predicate controls and the counts table/predicate counters pass.
- timeout 180 go run ./internal/buildcache/cmd/traced -- go test ./internal/lower -run '^TestProduct_LowerAgreementNative$' -count=1 -v -timeout 90s:
  exit 0, product test 6.21s, 14 products settled, 0 refused (traced-products.log).
- ADAMIC_BUILD_CACHE=read timeout 180 go test ./internal/lower -run '^(TestArrayNarrowing.*|TestFX5.*|TestNativeAgreementRowsNameRegisteredPrograms)$' -count=1 -v -timeout 90s:
  PASS 2.123s (read-products-traced.log), proving rows fetch the products.
- git diff --cached --check -- internal cmd: exit 0, no output.
  Whole-diff whitespace check flags literal patch context (space before tab),
  including patches inherited unchanged from the source squash; those remain exact evidence.

Incomplete attempts are retained: focus.log ran before the TypeScript submodule existed;
focus-ready.log, readers.log and vet.log reached their 180s shell limits during cold
compilation with no test output. No test leaf exceeded its budget in those attempts.
The first read-only attempt lacked traced products, as read-products.log explains;
the repository trace launcher resolved this and the read-only rerun passed. The worker
transport briefly disconnected before the final command and then recovered.

Changed/source-test leaf seconds from the initial green focused run:
- TestArrayNarrowingEarlyReturnRefused: 0.13s
- TestArrayNarrowingFindLastRefused: 0.14s
- TestFX5NumberValueRefused: 0.11s
- TestFX5NumberLiteralRefused: 0.12s
- TestArrayNarrowingFindRefused: 0.11s
- TestArrayNarrowingCatRefused: 0.08s
- TestArrayNarrowingForwardRefused: 0.08s
- TestArrayNarrowingPropertyRefused: 0.07s
- TestArrayNarrowingInferredRefused: 0.04s
- TestArrayNarrowingReadonlyRefused: 0.04s
- TestArrayNarrowingPreservedUnionAgrees: 5.66s
- TestArrayNarrowingBooleanCallbackAgrees: 5.45s
- TestFX5MixedLiteralAgrees: 5.69s
- TestArrayNarrowingSomeNegatedAgrees: 5.01s
- TestArrayNarrowingEveryRefused: 0.09s
- TestNativeAgreementRowsNameRegisteredPrograms: 0.03s
- TestArrayNarrowingOptionalNumberAgrees: 0.48s
- TestArrayNarrowingFindLastIndexAgrees: 0.60s
- TestArrayNarrowingOptionalStringAgrees: 0.56s
- TestProduct_LowerAgreementNative: 7.01s
