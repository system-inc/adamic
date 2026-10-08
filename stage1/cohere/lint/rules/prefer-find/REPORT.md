# prefer-find on the shared checker harness

Base: origin/area/stage1-lint at 9b7547976. Branch: codex/typeaware-wave-15-checker.

The descriptor subscribes to CallExpression and ElementAccessExpression and receives the parser node directly. The rule uses RuleContext.checker for raw-shape, type-shape and name answers. It does not create a checker or copy a shared helper. Messages and ordered suggestion edits follow the pinned Go rule. Parser optional-chain propagation is distinguished from a direct QuestionDotToken using the shared scanner.

TestPreferFind captures all four upstream test functions: 19 pass cases, 28 failing cases, the syntax-only declaration test, and five additional quiet cases. All 53 unique captures agree byte for byte with live Go, sanitized native, source Node and emitted JavaScript replay, including findings, automatic-fix output and suggestions. The local witness includes literal constants, computed filter names, sequence/conditional receivers, comment trivia, strconv.ParseFloat versus JS prefix parsing, and optional calls/accesses. Its oracle emits findings and the witness comparison passes.

The mutant `prefer-find array verdict suppressed` preserves compilation but removes the receiver array verdict. It executes and disagrees with Go on native, Node and emitted JavaScript. See the caught lines in [logs/wave15-prefer-find-witness-mutant-fixed.log](logs/wave15-prefer-find-witness-mutant-fixed.log). The full gate reruns this mutant on the final implementation.

Across the 53 typed upstream projects, Go program creation and lint totaled 2703.213 ms; sanitized native program creation, lint and transcript recording totaled 3539.941 ms (1.310 times Go). This includes checker program startup per case and recording overhead; it is not isolated listener throughput. See [timing](logs/wave15-prefer-find-timing.json).

Focused TestRulesAgree passed in 155.27 s. Registry generation, gofmt, vet and whitespace checks passed. Setup timings and nproc are in [setup](logs/wave15-checker-port-setup.log).

Full-package validation passed: 114 tests/subtests passed, 0 failed, 1 skipped; 76 compiled and executed mutants caught. Package elapsed 1865.733 s; measured wall 1867.685 s. nproc 5, GOMAXPROCS 4. Load start 1.536/1.870/2.023, end 2.104/2.097/2.239; median one-minute load 1.289. Every input-dependent check ran, including 875 compiler/stage1 files, release throughput, 1/2/5 shard agreement and profile snapshot comparison. The only skip is TestCheckerBridgeRefusalPending, which explicitly awaits codex/tsgo-errors-as-values and returning TSGoError from the C error buffer; setting an input does not enable it. No shared test was changed. See [counts and exact command](logs/full-package-summary.json) and [full JSON log](logs/full-package.jsonl).

The final-source mutant catches are at [native line 6](logs/final-mutant.log#L6), [Node line 14](logs/final-mutant.log#L14) and [emitted JavaScript line 22](logs/final-mutant.log#L22). TestOwnedWitnesses, TestMutants and TestRulesAgree passed in this same package invocation.

Coverage limit: the shared 875-file corpus gate uses syntax-only manifests. Typed prefer-find parity is established on all 53 upstream projects and its owned witness, rather than claimed for a checker-enabled 875-file corpus run.

The remaining 17 claims stop at the exact missing shared question/helper named in [CLAIM_STATUS.md](CLAIM_STATUS.md); each linked note includes the Go symbol, file:line and reproducer. They have no descriptors, matched cases or mutants on this branch.

The historical branch was merged with area and pushed at 5fed7283f. Its full-input lint command fails to build because old private bridge questions still use pre-pin string source paths and SourceFile.Path. The failures are bridge/tsgo/checker/declaration_facts.go:17 and :22, and bridge/tsgo/checker/facts.go:303. Those shared bridge files are outside the authorized directory scope; they were not changed. See [build failure](logs/wave15-merged-lint.jsonl) and [metrics](logs/wave15-merged-lint-summary.json). The checker branch starts directly from area and contains only the owned rule directory and blocker evidence.
