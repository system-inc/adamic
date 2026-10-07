Rebased: the existing helper branch onto origin/main e8ba3d5d81de4d3773c723914fccd4c76248b965; no new claims or production helper changes.
Commits: the previous pushed tip was 599028b3618df9082e4fb6d9ec785911e187ba04; the rebased implementation tip before this evidence commit is 75688ff313b35736f8e340a549e0481502972b19.
Checks: fresh owned and inherited oracle results are retained beside this report, with setup and the filtered comparison oracle.
Mutants: normalization suffix/newline/join, segment final-empty/unmatched-closer/escape, separator guard and floating-point counter approximation remain mandatory checks.
Uncovered: exact nextBuildCount still refuses bigint-return lowering; complete rule parity, missing consumer repository fixtures and the repository-wide gate are not claimed.

The landing cap takes precedence over selecting more helpers. Both previously pushed branches were audited and rebased. The sibling rule branch still has shared integration failures, so no fourth helper is claimed and the inventory is not described as exhausted.

Setup: ADAMIC_TOOLS=/workspace/adamic-tools bash cloud/setup.sh completed in 197 seconds: Go ready 0s, clang/Node/submodules 1s, cache warm 197s, done 197s. nproc is 5 with a four-core quota. Go 1.27.1, clang 20.1.8 and Node 24.19.0; cohere remains pinned at 715ba94f3608a6500086b1076ce5cb7e51b836db.

The delivered helpers normalizeValueFunctionArgument and segment remove two listed dependencies from each of better-tailwindcss/enforce-canonical-classes, enforce-consistent-class-order, enforce-consistent-variant-order, enforce-shorthand-classes, no-conflicting-classes and no-unknown-classes. Zero rules lose their final helper blocker from these helpers alone. Their 40,221 rows contain 1,164,929 exact observation bytes across actual Go, source Node, emitted JavaScript and ASan/UBSan native. Unsupported segment inputs refuse explicitly. The counter support probe is a recorded compiler boundary, not a delivered helper.

All historical evidence and reports remain historical. landing-*.log records this rebase verification; it does not reinterpret old commit hashes as current tips. No shared compiler, registration, harness or readiness implementation is changed.

Current-main commands: ADAMIC_GATE_UNCACHED=1 go test ./stage1/cohere/lint/helpers/wave12 -count=1 -v -timeout=15m PASS 89.466s. ADAMIC_GATE_UNCACHED=1 go test ./stage1/cohere/lint/helpers ./stage1/cohere/lint/helpers/comments -count=1 -v -timeout=20m PASS: helpers 118.342s, comments 177.409s. Every owned negative control and all inherited helper/comment mutants are caught. Existing comment consumer/JSX adapter boundaries remain explicit; inherited checks cover source Node and sanitized native and do not newly claim emitted-JavaScript coverage.

go vet ./stage1/cohere/lint/helpers/... exits 0 with no output. ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestTheOracleCatchesOneByte$' -count=1 -v -timeout=10m passes; landing-comparison-oracle.log retains the time and cache counters.
