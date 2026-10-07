Rebased the existing branch onto main and repaired its private mutant module graph; nine completed rule ports are green again.
Commits: tested code `5fa46b0e2f5dde6a074edff8a3e700edc195aa27`, main `e8ba3d5d81de4d3773c723914fccd4c76248b965`; the report commit is the final pushed tip.
Commands/results: three complete rule gates PASS (252.976s, 242.551s, 241.610s); bridge PASS 111.446s; 15-fixture Node gate PASS 5.474s; vet and formatting clean.
Mutants: all nine rule mutants exit 0 with empty stderr and are caught only by Go bytes; ten inherited rule mutants, seven bridge mutants, released-registry and Node-byte mutants are also caught.
Not covered: the three React claims remain unported; full repository gate, non-default rule options and shared production registration/profile/JavaScript integration are not claimed.

## Landing state

Only `codex/typeaware-wave-28` has been pushed by this unit. It is rebased onto the fetched main tip `e8ba3d5d`; no pull request is opened. The old remote tip was `1e3a68ef8365ccebbf820213449f79814d5e150a`. Main advanced during the first validation pass from `e011f8f6` to `e8ba3d5d` with new method call-target emission, so the branch was rebased a second time and the nine rule gates were repeated. The first-main full inherited compiler gate is retained separately from the final-main gate.

The rebased port commits are `b1a488fc` (original three), `9e4b1982` (Nexus three), and `1584d875` (core three). `5fa46b0e` repairs the private Nexus mutant graph. Original historical SHAs in previous reports are retained as historical evidence, not current branch tips.

The production rules did not need behavior changes. The updated compiler refuses a mutant importing two separately declared nominal classes: the runner instantiated the copied NoProcessExitAfterOutput while RequireBlockingStandardStreams still expected the original class. The owned checks_test.go now copies that dependent module with its behavior unchanged, replacing only its import so both clients refer to one nominal class declaration. No interface/cast bypass was added. The initial refusal is retained in `landing_validation/nominal-mutant-refusal.log.gz`; the final exit mutant compiles, exits 0 with empty stderr, and is caught solely by its finding bytes.

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

The earlier full inherited gate used ADAMIC_COVERAGE_COMPILER_MANIFEST and ADAMIC_TYPESCRIPT_SOURCE as well. The isolated timings run /tmp/wave-28-latest-timing.py after every gate finishes, with three alternating process rounds and full-output comparisons. Logs, paired canonical streams, stream hashes, mutants, root hashes and timings are retained in landing_validation. Streams are compressed losslessly with deterministic gzip to preserve intentional trailing spaces in fixes. The initial refusal log is also compressed because its diagnostic output contains a whitespace-only line; its bytes and hash are preserved.

## Remaining claims

react-hooks/set-state-in-effect, react-hooks/set-state-in-render and react-hooks/static-components remain claimed, blocked and unported. The final blocker reproduction passes in 15.455s: Go reports one issue per control while the current native parser exits 70 at each JSX tag; the non-JSX control parses. JSX support is published on origin/codex/stage1-jsx-lint but has not been integrated into this main. Native SSA, captures, post-dominators and ref control-dominance support required by these production Go validators is still absent from the inspected stage1 sources. This is an IR/parser dependency gap, not the shared .a/suggestion harness gap. No placeholder clean rule was registered. The earlier fourth-batch report contains the dependency inventory. No new rules were claimed.

Full go test ./..., full repository-wide vet, non-default rule option/project matrices, and shared production registration/profile/emitted-JavaScript rule comparisons are not claimed. The unit's native findings/fixes/suggestions comparisons and filtered Node oracle are complete for the nine existing ports. Stopping on the remaining substrate blocker preserves the instruction to keep changes inside owned rule directories.

## Landing refresh onto f8013f0b

Main advanced to `f8013f0baac41ddc340d76f83bddde38536a8f07`. The own branch rebased cleanly; tested code tip is `ceedfcab77e65eeb81c7a82a5e7ed9d39fcf64d2`. No owned implementation changed. The three complete gates passed again in 212.445s, 212.889s and 211.621s. Original controls: 255 sources, 126 findings, 45,164 bytes. Nexus expanded controls: 132 sources, 108 findings, 66,471 bytes; isolated DOM controls: 14 findings, 8,546 bytes. Core controls: 322 sources, 181 findings, 67,884 bytes; the documented legacy-octal recovery remains included. Normal and sanitizer streams agree byte for byte including fixes and suggestions. All batches also agree on the same 287 repository roots (18,485 bytes) and 77 compiler roots (5,010 bytes), zero findings, normal and ASan/UBSan/LSan. Control stream byte sizes differ from prior runs because artifact path lengths differ.

Every nine-rule mutant in the table above was repeated, exited zero with empty stderr, and was caught solely by independent Go bytes: parity 36,449; scope 48; getter 21,523; exit 1,816; timeout 4,816; blocking 7,909; throw 5,333; backreference 941; arrow fix whitespace 3,711. Released handles panic 70; Nexus/core repeat this under sanitizers without findings. All three released-registry mutants exit zero and fail the required-panic assertion. The 15-fixture Node/native/emitted-JavaScript oracle plus its one-byte mutant passed in 28.220s. Vet for bridge/typeaware and diff whitespace checks pass.

Fresh React blocker reproduction passes in 15.757s on the rebuilt compiler: each Go rule reports one finding, while JSX parsing exits 70 at bytes 101/75/62. The parser-bypass mutant exits zero on all three and is distinguished by the required rejection. No native React rule mutant is claimed. Main still has the string-only ParseNode API, unchanged parser blobs and no integrated JSX/native React SSA substrate. The shared-SSA branch remains design work. These three claims remain blocked and unported; numeric rule.json listener migration is also blocked by the integrated parser API. No new claims, main pushes or area pushes.

Gate timing observations were concurrent and are not isolated performance benchmarks. Original compiler native/Go processes: 9.098s/3.080s; Nexus compiler round medians: 3.040s/0.591s; core compiler round medians: 2.957s/0.501s. Native remains slower. Earlier isolated timing evidence remains historical. Full repository gate, inherited base rule/bridge mutants, non-default options and shared production registration were not repeated in this refresh.

Commands are the earlier full gate commands with artifact directories `/workspace/wave28-f801-first`, `-next`, `-third`, `-fourth`, and logs `/tmp/wave28-f801-{first,next,third,node,vet,fourth}.log`. Original uses both manifests and ADAMIC_TYPESCRIPT_SOURCE; core uses ADAMIC_WAVE28_THIRD_FULL=1; React uses ADAMIC_WAVE28_STAGE0=/workspace/wave28-f801-third/adamic. Exact execution logs are retained in landing_validation/f801-*.log. Toolchain remains the completed 165-second setup, nproc 5.

## Landing refresh onto c01907a7

Rebased cleanly onto `c01907a7036a22c2ea7ee686ed5fe4c6cd4bbc06`; tested source tip `f19a5f788af1a950927175273e69976784129236`. Incoming main adds stage3 sources/evidence and internal/oracle/stage3_hook_test.go. No compiler, bridge, stage1 parser or rule implementation changed; no shared leak-helper diff arrived in this rebase. All incoming changes were preserved. Push destination remains only codex/typeaware-wave-28 with an exact lease on old own tip 55b516036dde0ba2ca1c14a69323780a0803ea55.

Complete original/Nexus/core gates pass in 183.877s / 193.365s / 187.133s. Controls retain 126 / 108 (+14 DOM) / 181 findings and 45,164 / 66,471 (+8,546 DOM) / 67,884 bytes, normal and ASan/UBSan/LSan. The same 287 repository roots and 77 compiler roots agree byte for byte under normal and sanitizer runs, zero findings, 18,485 / 5,010 bytes per batch. The documented one core legacy-octal recovery input remains included. The exact full-stream comparisons include fixes and suggestions. All nine rule mutants in the table above are caught again by independent Go bytes after normal exit with empty stderr: parity 36,449; scope 48; getter 21,523; exit 1,816; timeout 4,816; blocking 7,909; throw 5,333; backreference 941; arrow fix whitespace 3,711. All three released-registry mutants are caught by required panic; normal released handles exit 70, and Nexus/core sanitizer released probes have no findings.

The 15-fixture filtered Node/native/JavaScript oracle and one-byte mutant pass in 6.548s; cache observations remain visible in its log. Active fifth-batch numeric listener/blocker gate passes in 16.260s: all three numeric declarations agree with production Go listener maps; each wrong-kind mutant is caught, each Go-positive JSX control still hits native parser rejection, and skipping parser.file is caught on all three. Vet for bridge/typeaware and whitespace checks pass. The exact commands are the previous full gate commands with c019 replacing f801 in artifact/log directories. Fifth-batch gate uses ADAMIC_WAVE28_STAGE0=/workspace/wave28-c019-third/adamic and ADAMIC_WAVE28_FIFTH_ARTIFACTS=/workspace/wave28-c019-fifth. All logs are retained as landing_validation/c019-*.log.

Contended gate timing observations, not isolated benchmarks: original compiler process native 5.942s / Go 2.925s; Nexus compiler three-round medians native 3.154s / Go 0.487s; core native 4.497s / Go 0.614s. Native remains slower. Historical isolated timing remains separately identified above. Toolchain is the completed 165-second setup, nproc 5. No new Go regexp matcher exists in the three active rule sources; no hand-rolled matcher was introduced.

The three earlier React HIR/SSA/capture claims remain parked under the explicit instruction and #dnv6f2c. The three active fifth-batch claims remain unported and blocked by current main's shared JSX parser, not the parked analysis. Main and area/stage1-lint still have the old parser blob. No further claims were taken. Full repository/stage3 gate, base dependency/bridge mutants, non-default options, new native React rule parity and shared production registration were not repeated or claimed.

## Landing refresh onto b8fb957a

Rebased cleanly onto main b8fb957aa839a9e8cb0b54279dd9864fa317bd30; tested source tip dfcafd843abfb4048c83f99872d17c89594b3010. Incoming main changes inherited static field emission and its Node oracle fixture/counts; all incoming changes were preserved. No shared leak-helper change arrived. The branch is pushed only to codex/typeaware-wave-28, using an exact lease on old own tip e67fc456aeeeab00f95bb7fb247163492892bde9.

Full original, Nexus and core gates pass in 194.019s / 205.495s / 211.363s. All controls, findings, fixes and suggestions still agree byte for byte: original 255 sources / 126 findings / 45,164 bytes; Nexus expanded 132 / 108 / 66,471 bytes plus isolated DOM 14 findings / 8,546 bytes; core 322 / 181 / 67,884 bytes, with the documented single legacy-octal recovery retained. Normal and ASan/UBSan/LSan controls and both frozen corpora pass for each batch: 287 repository roots, 77 compiler roots, zero findings, 18,485 / 5,010 identical bytes. The new main fixture is included in the filtered Node oracle, not added to the frozen lint population.

Every nine-rule mutant is caught after normal exit with empty stderr by independent Go bytes: parity 36,449; scope 48; getter 21,523; exit 1,816; timeout 4,816; blocking 7,909; throw 5,333; backreference 941; arrow-fix whitespace 3,711. Released-registry mutants exit zero and are caught by required panic assertions for all three batches. Normal released handles panic 70; Nexus/core sanitizer release checks are clean. The 16-fixture Node/native/emitted-JavaScript gate includes inherited_static_field_read and the one-byte mutant and passes in 7.010s. The active named-listener/JSX blocker gate passes in 24.983s, including three Identifier-substitution metadata mutants compared with production Go listener names, parser bypass on all three Go-positive controls, and the ordinary non-JSX control. Vet and whitespace checks pass.

Rule.json listener kinds now use ast.Kind names, per the explicit corrected contract. Earlier numeric metadata and numeric parser-API blocker descriptions in historical sections are superseded; JSX parsing is the actual remaining blocker. Main and area/stage1-lint still lack JSX descent and the named harness ab70f38d4. The earlier rejected separate-history rebase was not retried. HIR/SSA/capture-dependent claims stay parked under #dnv6f2c; the three active JSX rules remain unported pending parser integration. No new claims or shared source edits.

Logs are retained as landing_validation/b8fb-*.log. Commands are the earlier complete gate commands with b8fb in artifact/log directories. Fifth uses ADAMIC_WAVE28_STAGE0=/workspace/wave28-b8fb-third/adamic and ADAMIC_WAVE28_FIFTH_ARTIFACTS=/workspace/wave28-b8fb-fifth. Node's fixture filter adds inherited_static_field_read. Toolchain remains the completed 165-second setup, nproc 5. Contended compiler gate timing observations (not isolated benchmarks): original native process 6.034s / Go 2.394s; Nexus three-round medians 3.140s / 0.507s; core 3.456s / 0.493s. Native remains slower. Full repository gate, base dependency/bridge mutants, non-default options, production shared registration and completed new JSX rule parity are not claimed.
