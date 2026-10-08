Built: Linux x86_64 verification of the integrated WASI base, with current main merged; no authored production fixes.
Commits: claim 91df3fd; tested merge ab6390221319bcaf8a0a77774d049e7072ec6064, main parent f8013f0baac41ddc340d76f83bddde38536a8f07.
Results: excluding-markdownblocks gate exit 0, 32 packages pass, 0 fail, 8 with no tests; WASI 323 pass, 0 fail, 0 skip; formatting, vet, runtime and request checks pass.
Mutants: native stack macro substituted for WASI macro; stack_forever.a and stack_tail_call.a both fail exact comparisons, restored; built-in byte and exit mutants caught.
Not covered: Markdown blocks excluded at the user's request, independently fails on clean main; deployment, other architectures and custom engine stack limits.

The first commit and first push added only claim.md. The branch started at origin/wasm/integrate 6f7dce3dc1eace606fe081c8f4ab12034ae11bb4. Initially main e8ba3d5 was already an ancestor through merge 24f45a792a740eac25c5c1915e4b2e9fee0f6c63, so git merge --no-ff origin/main returned Already up to date. Main advanced during the gate. After fetching, git merge --no-ff origin/main created ab639022 with parent f8013f0. No rebase was used. There were no conflicts: ort automatically combined internal/native/runtime/adamic.h and internal/oracle/counts.md. No conflict resolution or authored production edit was needed. Only codex/wasm-landing is pushed.

Observed environment: Linux x86_64, Go 1.27.1, Node 24.19.0, native clang 20.1.8, WASI SDK 27 clang 20.1.8-wasi-sdk. nproc=5; cgroup cpu.max=400000 100000. bash cloud/setup.sh --wasi-sdk passed. Its timings: Go ready 0s, clang ready 1s, Node ready 1s, WASI SDK ready 3s, submodules ready 4s, build cache warm 145s, done 146s. Every test shell sourced /workspace/adamic-tools/env.sh, which sets WASI_SYSROOT=/workspace/adamic-tools/wasi-sdk/share/wasi-sysroot. Native clang remained the default; only TestWASI and the isolated runtime mutants prepended the SDK bin directory.

Final commands, each redirected directly to its named separate log, with exit records beside them:

```sh
gofmt -l cmd internal > final-gofmt.log 2>&1
go vet ./... > final-vet.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test -count=1 -timeout 30m ./... > final-gate.log 2>&1
ADAMIC_ORACLE_WASI=1 ADAMIC_GATE_UNCACHED=1 go test -json -count=1 -timeout 30m ./internal/oracle -run '^TestWASI' > final-wasi-oracle.jsonl 2>&1
ADAMIC_TEST_WASI=1 go test -count=1 -timeout 15m ./internal/native -run '^TestWASI$' > final-runtime.log 2>&1
ADAMIC_ORACLE_WASI=1 go test -count=1 ./cmd/adamic -run 'TestWASIRequest|TestRequest' > final-request.log 2>&1
```

Formatting and vet exited 0 with empty logs. The full WASI invocation exited 0 in 461.544s: TestWASIAgreesWithNode 371.47s, TestWASIEmission 87.94s, TestWASIOracleCatchesMutants 1.11s, TestWASIRunnerCatchesMutants 1.01s. Fixture counts are terminal subtest events parsed from the complete JSON log, counting the comparison and emission suites separately. Comparisons: 323 fixtures, 323 pass, 0 fail, 0 skip. Emission: 323 fixtures, 323 pass, 0 fail, 0 skip. No failing fixture exists to name, supply differing lines for, or hand off. The harness reads Node before running WASI and uses the JavaScript backend witness for checked fixtures. Runtime TestWASI exited 0 in 82.617s, covering its 35 fixture comparisons and 100,000-request lifetime/region probe. CLI request checks exited 0 in 9.004s.

The first baseline oracle also passed: 308 comparisons and 308 emission checks, no failures or skips. First runtime passed in 132.816s and first request checks in 11.789s. The initial full gate was stopped with exit 143 after main advanced, before changing its source tree. Its log is partial, not a full pass. A complete rerun with the dependencies installed was stopped at the user's request, as documented below; the final accepted scope excludes Markdown blocks.

Stack mutant procedure: copy the complete final runtime into /tmp/adamic-w4-final-stack-mutant/runtime, replace only the ADAMIC_TARGET_WASI ADAMIC_CHECK_STACK macro body with the native body from the same header, and select the copy with ADAMIC_WASI_RUNTIME. The repository header is never changed. Run:

```sh
ADAMIC_TEST_WASI=1 ADAMIC_WASI_RUNTIME=/tmp/adamic-w4-final-stack-mutant/runtime go test -v -count=1 -timeout 15m ./internal/native -run '^TestWASI$/internal/oracle/testdata/stack_(forever|tail_call)\.a$' > final-stack-mutant.log 2>&1
```

Both selected subtests actually ran and failed. All 48 runtime translation units compiled under strict C11; neither witness was killed by a warning or compilation failure. Read Node first: stack_forever.a has stdout "start\n", stderr "adamic: panic: RangeError: Maximum call stack size exceeded\n", exit 70. stack_tail_call.a has stdout "0\n", the same stderr, exit 70. Mutant WASI has empty stdout and exit 1 for both, with raw engine RangeError traces. First stdout line differs: Node "start" versus WASI EOF for stack_forever; Node "0" versus WASI EOF for stack_tail_call. The complete traces are retained in final-stack-mutant.log. This demonstrates that the retained runtime macro's addressed 64-byte frame prevents engine stack exhaustion from bypassing the panic/output contract on this host.

The mutant command exited 1. Restore by copying the exact repository header back over the scratch header; cmp exited 0 (recorded in final-mutant-exit.log). The first baseline mutant independently produced the same two failures and restoration. Final control TestWASI and the full oracle both passed with the unchanged repository header. Built-in oracle mutants change dedication's emitted string and use exit 23, and runner mutants change probe bytes and exit 23; both tests require the intended stdout/exit disagreement and passed.

Logs reside in this report directory. No test output was piped. This is observation on Node 24's default engine stack, not proof for custom smaller engine stacks or deployment hosts. docs/wasm.md contains older limitations that the integrated tests supersede for these witnesses; this unit did not edit those runtime-owned docs.

Upstream Markdown gate setup failure and independent main reproduction:

The first complete merged-tree gate exited 1. Its only failure was TestMarkdownUnicodeWidths in stage1/cohere/markdownblocks, whose Node witness could not import /tmp/adamic-markdown-width/node_modules/emoji-regex/index.js. The package finished in 1693.076s; this run did not time out. cloud/setup.sh does not install the three external width packages. The widths test explicitly uses ADAMIC_MARKDOWNWIDTH_DEPS or defaults to /tmp/adamic-markdown-width. Install only in scratch: npm install --prefix /tmp/adamic-markdown-width --ignore-scripts --no-audit --no-fund emoji-regex@10.6.0 get-east-asian-width@1.6.0 narrow-emojis@0.0.3. Install exited 0. A separate uncached TestMarkdownUnicodeWidths run then passed in 416.027s. No repository code was changed for this repair.

A subsequent whole-gate run with dependencies present was stopped at the user's request, exit 143, to spare the machine and switch to excluding that upstream package. It was progressing through distinct Markdown tests and mutants, not a demonstrated 15-minute single-test stall. The status log saves requested process snapshots and exact 40-line gate/oracle tails. Observed live tests progressed from TestMarkdownTextSplitting to TestMarkdownUnicodeWidths, TestWholeDocumentOraclePreflight/fence_length, TestMarkdownTableLayout/table_center and TestMarkdownCodeBlockLayout/fence_longest. Type-aware progressed through SixRule and Volume mutants and finished. No stall-based exclusion or 20-minute dump run was triggered.

User-directed final gate:

```sh
packages=$(go list ./... | grep -v stage1/cohere/markdownblocks)
ADAMIC_GATE_UNCACHED=1 go test -json -count=1 -timeout 30m $packages > excluding-markdownblocks-gate.jsonl 2>&1
```

Independent reproduction uses clean detached /tmp/adamic-w4-plain-main at origin/main f8013f0baac41ddc340d76f83bddde38536a8f07, with pinned submodules initialized. git status --short is empty. Because the active workspace's default scratch dependencies had been repaired, ADAMIC_MARKDOWNWIDTH_DEPS points to an empty /tmp/adamic-w4-main-empty-width-deps, reproducing a fresh cloud dependency state without removing working dependencies. No WASI/compiler/runtime changes from this branch are in that checkout.

```sh
ADAMIC_MARKDOWNWIDTH_DEPS=/tmp/adamic-w4-main-empty-width-deps ADAMIC_GATE_UNCACHED=1 go test -v -count=1 -timeout 10m ./stage1/cohere/markdownblocks > plain-main-markdownblocks.log 2>&1
ADAMIC_MARKDOWNWIDTH_DEPS=/tmp/adamic-w4-main-empty-width-deps ADAMIC_GATE_UNCACHED=1 go test -v -count=1 -timeout 10m ./stage1/cohere/markdownblocks -run '^TestMarkdownUnicodeWidths$' > plain-main-width.log 2>&1
```

The focused main reproduction exited 1 in 9.707s (test 9.66s). First diagnostic: width_test.go:78: original dependencies exit 1 stderr node:internal/modules/esm/resolve:271. First dependency error: Error [ERR_MODULE_NOT_FOUND]: Cannot find module '/tmp/adamic-w4-main-empty-width-deps/node_modules/emoji-regex/index.js' imported from /tmp/adamic-w4-plain-main/stage1/cohere/markdownblocks/testdata/width_library.mjs. This directly reproduces the failure without the WASI changes. The plain-main whole-package command exited 1 in 600.083s. Its first error line is "panic: test timed out after 10m0s". The goroutine dump names TestMarkdownSourceDecoding, running for 30s when the cumulative package budget expired. It did not reach TestMarkdownUnicodeWidths in this run. This is evidence of an independent plain-main package timeout, not proof that a single test hung for 15 minutes or that the missing module caused the timeout. The separate focused plain-main run proves the missing dependency directly. Both complete logs are retained.

The final user-directed excluding gate exited 0. JSON terminal package events: 40 packages total, 32 pass, 0 fail, 8 skip events for packages with no test files. Leaf test results, excluding parent tests that have subtests: 2,927 pass, 0 fail, 35 skip. All terminal test/subtest events, including parents: 3,025 pass, 0 fail, 37 skip. These two counts intentionally differ because parent completion events are counted only in the latter. Type-aware, the last package, passed in 586.779s. No failing test exists in the accepted scope. All skipped leaf tests are preserved in the JSON log; opt-in WASI tests in the ordinary gate are covered separately by the explicit WASI legs above.

There is no complete repository green claim including Markdown blocks. The requested Linux WASI gate and the user-directed excluding gate are green on merge ab639022. Clean-main reproduction and the excluded upstream package remain reported separately. git diff --check passes. Only claim.md, report.md and evidence logs in this directory are authored by W4.
