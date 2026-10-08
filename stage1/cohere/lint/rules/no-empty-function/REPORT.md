# no-empty-function

Ported the core rule in `rule.a`, with exact messages and comment suggestions in `messages.a`. Registration and the typed Go options adapter are local to this directory. No compiler, shared helper, dispatch, oracle driver or corpus source was changed.

## Branch and oracle

The branch is `lint-rules/no-empty-function`, based on `origin/area/stage1-lint` at `9156bf5c579a44d687c9955d13e44f9ad8bbb6f8`. The initial remote scan found no descriptor for this rule across 2,021 fetched remote refs. `evidence/fetch.log.gz` and `evidence/duplicate-check.log.gz` retain that evidence.

The unchanged oracle is `rules.NoEmptyFunction` from cohere `7945d102a6c18dd36adf9114a758ce646e8b2359`. The adapter uses its own `DecodeNoEmptyFunctionOptions`. The captured upstream suite contains 2,323 unique source/rule/options combinations (the large arrays contain 2,313 rows, with additional test cases).

Shared helpers used: `propertyName` with `Static`, `consistentReturnIsGenerator`, and comments `forFile`. The small arena projections supply these helpers with parser fields; they do not duplicate their algorithms. Comment spans are UTF-8 byte offsets, so body offsets are converted before comparison. Findings and edits stay in the harness's source offsets.

## Behavioral evidence

Selected certification compares complete serialized findings, message text, ranges, automatic fixes and suggestions against Go on source Node, Adamic-emitted JavaScript, and native under ASan/UBSan. It includes all 2,323 upstream cases, all four owned witnesses, and every distinct source in the inherited generated corpus. The initial passing run reported 755,880 identical output bytes; its four-witness default-options mutant baseline reported 6,857 identical bytes. Final-source certification is recorded separately below.

Witnesses:

- `testdata/kinds.ts.txt`: function kinds, generator and async names, accessors, constructors and property-assigned anonymous functions.
- `testdata/options.ts.txt`, with `testdata/options.options.json`: typed allow options, including the distinction between an override method and constructor.
- `testdata/comments-and-parameters.ts.txt`: interior comments, exterior comments and constructor parameter properties.
- `testdata/unicode-and-keys.ts.txt`: UTF-8 prefixes, escaped and computed keys, numeric names, async generators and concise arrows.

Go behavior retained deliberately: async arrows use the `arrowFunctions` option; async generators use the generator category/name without an async prefix; property-assigned anonymous function expressions are named as methods but use function allow options. `overrideMethods` does not exempt an override constructor. The dynamically constructed unexpected message is used instead of the unused generic catalog description. No upstream disagreement was observed.

`mutant.json` widens only the suggestion edit start from `start + 1` to `start`. The mutant compiles and runs. Node, emitted JavaScript and sanitized native all disagree with Go on the first witness's suggestion range (31:32 instead of 32:32), while the finding span and message remain unchanged. The exact log is retained, not inferred from a compiler rejection.

## Reproduction

Run from the repository root. These commands produce logs rather than piping test output:

```bash
export GOPROXY='https://proxy.golang.org|direct'
bash cloud/setup.sh --wasi-sdk > /tmp/s13-setup.log 2>&1
source stage1/cohere/lint/rules/no-empty-function/evidence/inputs.sh
go run ./cmd/lint-registry > /tmp/s13-registry.log 2>&1
go run ./cmd/adamic types stage1/cohere/lint/rules/no-empty-function/rule.a > /tmp/s13-types.log 2>&1
```

`ADAMIC_TYPESCRIPT_SOURCE` points to a clean checkout of the required pin `050880ce59e30b356b686bd3144efe24f875ebc8`. For a new full-package run, choose a fresh profile directory and set both profile variables to that same directory. `ADAMIC_LINT_BENCH=1` and `ADAMIC_TEST_WASI=1` are also set; the setup environment supplies `WASI_SYSROOT`.

The selected test is an owned test overlay, not a shared harness edit. Write the following JSON with absolute paths appropriate to your checkout, then run:

```json
{"Replace":{"/workspace/adamic/stage1/cohere/lint/s13_no_empty_function_test.go":"/workspace/adamic/stage1/cohere/lint/rules/no-empty-function/testdata/selected_test.go.txt"}}
```

```bash
go test -overlay /tmp/s13-no-empty-function-overlay.json ./stage1/cohere/lint -run '^TestNoEmptyFunctionSelected$' -count=1 -v -timeout=30m > /tmp/s13-selected.log 2>&1
go test ./stage1/cohere/lint -count=1 -json -timeout=30m > /tmp/s13-whole.jsonl 2>&1
```

The overlay's top-level test calls `t.Parallel`. The ordinary package tests discover this descriptor, typed options, witnesses and mutant without the overlay.

## Environment and limits

The successful recovered setup reported node 0.144s, Go 0.146s, markdown dependencies ready 0.311s, submodules 0.423s, clang 0.920s, WASI SDK 11.999s, Go build 351.701s, tests deferred 352.381s, cache warm 352.386s, total 352.593s. `nproc` is 5; the container CPU quota is four CPUs. Setup and interrupted cold-build attempts were separated: only completed runs are claimed.

Cohere formats and checks both owned Adamic files successfully. Final style edits replace ambiguous local aliases with direct `this.context` accesses and change no rule logic. The whole-package run started before those spelling edits; selected certification is repeated on the final files.

No missing shared helper or unexpressible Adamic language construct was encountered. The full repository compiler/oracle gate was not run for this rule-only unit. The environment restarted during both final-source selected certification and the first whole-package attempt. Their test parents were killed, leaving incomplete logs, so neither is claimed as a passing run. `evidence/whole-interrupted.jsonl.gz` and `evidence/selected-interrupted.log.gz` retain these attempts. Certification is retried with detached logging and a fresh profile directory. Whole-package results, named skips and final selected results are recorded below when completed.

## Final selected result

`evidence/selected.log.gz` records the final files: 2,323 upstream cases, four witnesses and 18 distinct inherited generated sources, with 750,617 identical serialized bytes across Go, source Node, emitted JavaScript and sanitized native. The default-options witness baseline produced 6,837 identical bytes. The three compiled mutant catches are extracted with original log line numbers in `evidence/mutant.log`. Different temporary path lengths account for the byte totals differing from the initial run.

The test passed in 458.86s; the Go package reported 459.704s; the wrapper wall time including linking was 515.510s. `evidence/selected-summary.json` records the command, inputs, loads and zero exit status.

## Whole lint package result

The recovered whole-package attempt exited 1: `panic: test timed out after 30m0s` during `TestProfileArtifacts`, at `stage1/cohere/lint/profile_test.go:126`, in the monolithic `clang -O2 -g` profiling build. This is a capacity limit, not an observed output mismatch. The test had run for 1m55s after earlier serial throughput and mutant tests consumed most of the package timeout. Shared files were not changed to bypass this gate.

Observed terminal results: 18 test/subtest passes (10 top-level passes), zero individual test failure events, one package failure due to timeout, zero skips (none). The log announces 25 unfinished tests; later tests were also not started. These names, the exact command, all inputs and the package failure are in `evidence/whole-summary.json`; the complete JSON log is `evidence/whole.jsonl.gz`. A zero skip count here does not establish that every test executed. In particular, full registry agreement, the package-wide mutant suite and profile snapshot agreement remain uncertified by this attempt; the selected test does establish this rule's complete four-way agreement and mutant catches.

Package elapsed time was 1800.459s; total wall time including pre-test linking was 1854.622s. `nproc`: 5. Load before: 1.27, 2.25, 2.96; after: 1.04, 1.21, 1.70. WASI, the clean pinned TypeScript checkout and both profile variables were set. Both profile variables used the same fresh directory.

Stopped on the whole-package timeout. No rule algorithm, shared-helper or Adamic language gap remains. No complete repository gate or completed whole lint package pass is claimed.
