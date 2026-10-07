Rebased the existing branch onto main and repaired its private mutant module graph; nine completed rule ports are green again.
Commits: tested code `5fa46b0e2f5dde6a074edff8a3e700edc195aa27`, main `e8ba3d5d81de4d3773c723914fccd4c76248b965`; the report commit is the final pushed tip.
Commands/results: three complete rule gates PASS (252.976s, 242.551s, 241.610s); bridge PASS 111.446s; 15-fixture Node gate PASS 5.474s; vet and formatting clean.
Mutants: all nine rule mutants exit 0 with empty stderr and are caught only by Go bytes; ten inherited rule mutants, seven bridge mutants, released-registry and Node-byte mutants are also caught.
Not covered: the three React claims remain unported; full repository gate, non-default rule options and shared production registration/profile/JavaScript integration are not claimed.

## Landing state

Only `codex/typeaware-wave-28` has been pushed by this unit. It is rebased onto the fetched main tip `e8ba3d5d`; no pull request is opened. The old remote tip was `1e3a68ef8365ccebbf820213449f79814d5e150a`. Main advanced during the first validation pass from `e011f8f6` to `e8ba3d5d` with new method call-target emission, so the branch was rebased a second time and the nine rule gates were repeated. The first-main full inherited compiler gate is retained separately from the final-main gate.

The rebased port commits are `b1a488fc` (original three), `9e4b1982` (Nexus three), and `1584d875` (core three). `5fa46b0e` repairs the private Nexus mutant graph. Original historical SHAs in previous reports are retained as historical evidence, not current branch tips.

The production rules did not need behavior changes. The updated compiler refuses a mutant importing two separately declared nominal classes: the runner instantiated the copied NoProcessExitAfterOutput while RequireBlockingStandardStreams still expected the original class. The owned checks_test.go now copies that dependent module with its behavior unchanged, replacing only its import so both clients refer to one nominal class declaration. No interface/cast bypass was added. The initial refusal is retained in `landing_validation/nominal-mutant-refusal.log`; the final exit mutant compiles, exits 0 with empty stderr, and is caught solely by its finding bytes.

No shared registration generator, shared harness or protected compiler file was edited. `git diff origin/main...HEAD` for internal/native/emit.go, internal/lower/lower.go, internal/native/native.go and internal/oracle/oracle_test.go is empty. Rebase brought main's compiler changes through its parent, not as unit edits. The bridge dependency commit is rebased as `9b9427d9`.

## Final main comparisons

| Batch | Controls | Findings | Canonical bytes |
| --- | ---: | ---: | ---: |
| Original | 255 | 126 | 45,929 |
| Nexus expanded | 132 | 108 | 66,867 |
| Nexus isolated DOM controls | 18 | 14 | 8,600 |
| Core | 322 | 181 | 68,850 |

Normal and ASan/UBSan/LSan native output matches Go byte for byte on these controls, including fixes and suggestions. The core controls retain one legacy-octal input with one Go recovery diagnostic, as documented in the earlier core report; the other 321 are syntax-valid. No control was newly omitted. Each batch also matches Go on all 287 repository roots and 77 TypeScript compiler roots in normal and sanitized builds: zero findings, 18,485 and 5,010 canonical bytes respectively. Positive controls and mutants guard against vacuous zero-corpus agreement.

The same repository manifest is used, but seven input files changed when rebasing onto main: internal/oracle/testdata/class_layouts.a, internal/oracle/testdata/closures_throw.a, stage1/cohere/lint/finding.ts, stage1/cohere/lint/lint.ts, stage1/cohere/lint/main.ts, stage1/cohere/values/parser.ts and stage1/cohere/values/tokenize.ts. Final comparisons use those current-main versions. `source-changes.json` preserves both hashes and the final root manifests/hashes are retained. All 77 compiler input hashes remain unchanged. This is the original root set, not every new file introduced on main.

The inherited ten-rule dependency suite on the first main tip also passed its full compiler/repository gate: 53 control findings, 180 repository findings, 16,589 compiler findings and 7,120,613 identical compiler bytes under sanitizers, in 801.685s. On the final main tip it passed controls, all ten mutants, normal/sanitized repository comparisons and released handles in 335.767s. Its large compiler corpus was not repeated after the second rebase. The nine unit rules' full compiler corpus WAS repeated on the final tip. These scopes are distinct in the retained logs.

## Every mutant and its check

| Rule/check | Mutation | Final check |
| --- | --- | --- |
| Verify optional parity | Test only root flags rather than union members | Go finding bytes |
| Block scoped var | Accept every use inside the file | Go finding bytes |
| Getter return | Change both conditional branches from AND to OR | Go finding bytes |
| No process exit after output | Report every reachable exit | Go finding bytes |
| No uncleared race timeout | Reverse read/write target classification | Go finding bytes |
| Require blocking streams | Force initialization order true | Go finding bytes |
| No throw literal | Change conditional could-be-error from OR to AND | Go finding bytes |
| No useless backreference | Count the first group as another group | Go message bytes |
| Prefer arrow callback | Insert an extra trailing space after => | Go fix bytes |
| Inherited before/cast/methods/coercion/caller/parameter/invariant/optional/alias/unused | Existing ten production judgment/span mutations | Go finding/fix bytes, each exits 0 |
| Released registry | Keep the released checker handle live | Required exact panic 70 |
| Bridge C input length | Add one byte to input length | ASan heap-buffer-overflow |
| Bridge output length | Add one byte to returned length | ASan heap-buffer-overflow |
| Bridge stale handle | Keep released handle live | Stale-handle assertion |
| Bridge source position | Ask type from SourceFile position | Independent oracle bytes |
| Bridge link guard | Remove opt-in guard | Required refusal |
| Bridge output free | Omit C output free | LeakSanitizer |
| Bridge region allocation | Allocate region result on heap | LeakSanitizer |
| Node oracle | Change one output byte | Node/native/JavaScript byte mismatch |
| React blocker probe | Skip parser.file | Required JSX rejection, all three probes distinguish exit 0 |

Exact byte offsets and observations are in the individual logs and mutants.json. Original and sanitized released probes for stream-symbol and reference-symbol still produce exactly `adamic: panic: invalid or released checker handle` with exit 70 and no sanitizer findings. The bridge gate additionally holds 100 C ABI queries and 162 independent positions (3,261 bytes) with outputs surviving release and zero/stale handle rejection. The Node filter runs 15 fixtures through Node, emitted JavaScript and sanitized native, including the newly merged call-target and devirtualize fixtures, plus the one-byte mutant. Native oracle cache hits are explicitly recorded in node.log.

## Isolated native time against Go

Three fresh-process rounds per workload alternate execution order, after all concurrent gates completed. Every timed stdout is independently checked for equality and native stderr is empty. Timings include startup, checker load, parsing, visits and canonical serialization; compilation is excluded. The gate logs' contended timings are not used here.

| Batch | Repository native / Go median | Compiler native / Go median |
| --- | --- | --- |
| Original | 0.393s / 0.213s (1.84x) | 9.827s / 1.994s (4.93x) |
| Nexus | 0.433s / 0.218s (1.99x) | 2.966s / 0.452s (6.57x) |
| Core | 0.425s / 0.230s (1.84x) | 3.275s / 0.518s (6.32x) |

Native is slower on these samples. Individual rounds and output hashes are in timing.json; this makes no overall speed claim.

## Commands

All tests write to log files. Toolchain setup remains the same completed setup: Go 1.27.1 ready 0s, clang 20.1.8 ready 1s, Node 24.19.0 ready 1s, submodules 1s, cache/setup 114s; nproc 5, CPU quota 4. The env file is /workspace/adamic-tools/env.sh.

```sh
source /workspace/adamic-tools/env.sh
ADAMIC_WAVE28_ARTIFACTS=/workspace/wave-28-latest-first ADAMIC_WAVE28_REPOSITORY_MANIFEST=/workspace/wave-28-artifacts/repository.manifest ADAMIC_WAVE28_COMPILER_MANIFEST=/workspace/wave-28-artifacts/compiler.manifest ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave-28-corpus go test ./stage1/cohere/typeaware -run '^TestWave28AgreementAndMutants$' -count=1 -timeout=25m -v >/tmp/wave-28-latest-first.log 2>&1
ADAMIC_WAVE28_NEXT_ARTIFACTS=/workspace/wave-28-latest-next go test ./stage1/cohere/typeaware/wave28_next -count=1 -timeout=25m -v >/tmp/wave-28-latest-next.log 2>&1
ADAMIC_WAVE28_THIRD_ARTIFACTS=/workspace/wave-28-latest-third ADAMIC_WAVE28_THIRD_FULL=1 go test ./stage1/cohere/typeaware/wave28_third -count=1 -timeout=25m -v >/tmp/wave-28-latest-third.log 2>&1
ADAMIC_COVERAGE_ARTIFACTS=/workspace/wave-28-latest-coverage ADAMIC_COVERAGE_REPOSITORY_MANIFEST=/workspace/wave-28-artifacts/repository.manifest go test ./stage1/cohere/typeaware -run '^TestCoverageAgreementAndMutants$' -count=1 -timeout=25m -v >/tmp/wave-28-latest-coverage.log 2>&1
go test ./bridge/tsgo/... -count=1 -timeout=15m -v >/tmp/wave-28-latest-bridge.log 2>&1
go test ./internal/oracle -run '^TestTheOracleCatchesOneByte$|^TestNativeAgreesWithNode$/internal/oracle/testdata/(maps_and_text|sorting|string_index|lone_surrogates|functions|closures|call_targets.*|devirtualize)\.a$' -count=1 -timeout=15m -v >/tmp/wave-28-latest-node.log 2>&1
go vet ./bridge/tsgo/... ./stage1/cohere/typeaware/... >/tmp/wave-28-latest-vet.log 2>&1
gofmt -l cmd internal bridge/tsgo stage1/cohere/typeaware >/tmp/wave-28-latest-gofmt.log 2>&1
ADAMIC_WAVE28_STAGE0=/workspace/wave-28-latest-third/adamic ADAMIC_WAVE28_FOURTH_ARTIFACTS=/workspace/wave-28-latest-fourth go test ./stage1/cohere/typeaware/wave28_fourth -count=1 -timeout=10m -v >/tmp/wave-28-latest-fourth.log 2>&1
```

The earlier full inherited gate used ADAMIC_COVERAGE_COMPILER_MANIFEST and ADAMIC_TYPESCRIPT_SOURCE as well. The isolated timings run /tmp/wave-28-latest-timing.py after every gate finishes, with three alternating process rounds and full-output comparisons. Logs, paired canonical streams, stream hashes, mutants, root hashes and timings are retained in landing_validation. Streams are compressed losslessly with deterministic gzip to preserve intentional trailing spaces in fixes.

## Remaining claims

react-hooks/set-state-in-effect, react-hooks/set-state-in-render and react-hooks/static-components remain claimed, blocked and unported. The final blocker reproduction passes in 15.455s: Go reports one issue per control while the current native parser exits 70 at each JSX tag; the non-JSX control parses. JSX support is published on origin/codex/stage1-jsx-lint but has not been integrated into this main. Native SSA, captures, post-dominators and ref control-dominance support required by these production Go validators is still absent from the inspected stage1 sources. This is an IR/parser dependency gap, not the shared .a/suggestion harness gap. No placeholder clean rule was registered. The earlier fourth-batch report contains the dependency inventory. No new rules were claimed.

Full go test ./..., full repository-wide vet, non-default rule option/project matrices, and shared production registration/profile/emitted-JavaScript rule comparisons are not claimed. The unit's native findings/fixes/suggestions comparisons and filtered Node oracle are complete for the nine existing ports. Stopping on the remaining substrate blocker preserves the instruction to keep changes inside owned rule directories.
