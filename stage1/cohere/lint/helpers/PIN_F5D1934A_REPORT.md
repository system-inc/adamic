# Cohere f5d1934a helper recapture

Base: area/stage1-lint 9156bf5c5. Merged, without rebasing, cohere-pin/f5d1934a-land e6534c401 and stage1-format/tsprinter-method-shorthand 06c6e7cbf.

Capture provenance now reads the parent repository's HEAD gitlink through internal/coherepin and helpers/testdata/pin.py. Tests also reject a checked-out cohere commit that differs from that gitlink. Stored captures retain their actual capture commit; changing the gitlink requires recapturing them. Live Go captures write capture-pin.json alongside their answers.

## Recapture observations

| Helper package or capture | Old versus new Go |
| --- | --- |
| helpers imports | Byte-identical on the same captured inputs: BindingsOf 228,830 bytes; NormalizedFileName 23,086; ImportedNameOf 5,111; CallExpressionSource 128,072; HasPathSegment 10,700. Both pins run the package's existing Go oracle overlays. |
| helpers JSX | Byte-identical: 1,284 upstream inputs plus the oracle control, 6,231,005 answer bytes. |
| core | All 8,230 call records and 2,552 frames agree after removing capture provenance and comparing calls as a multiset. Serialized call order differs because upstream tests run concurrently. kind_number.a was regenerated from Go ast.Kind; its table body is byte-identical and its generated provenance header changed. |
| ecmascript/text | 103,386 to 103,672 calls. No previous call removed or answer changed. The 286 added calls come from the two BEFORE_send exceptionPatterns cases added to TestIdLengthStaysSilent by cohere c2e39b75. They include two GraphemeCount answers of 11 and their grapheme classifier calls. |
| ecmascript React | Byte-identical on 4,065 captured sources: 4,530,973 answer bytes. |
| rules-react | Leaf calls 2,054 to 2,063, adding nine isReactComponentBaseName(Component, false) inputs. AST calls 28,317 to 28,365; added skipParenthesesOptional 27, isReactComponentBase 9, isComponentClass 9 and semanticParentOf 3. Comparing arenas without their capture-order IDs proves all previous calls and answers remain present and unchanged. New observations come from c2e39b75's boolean-prop-naming and sort-comp option-pattern cases, including newly exercised getter/static-property cases. No helper implementation changes were needed. |
| slot05/batch32 | Byte-identical against old Go: value 31046 bytes, after 30904 bytes, returns 30896 bytes. |
| slot05/batch33 | Byte-identical against old Go: component 31425 bytes, type 3841 bytes, network 310 bytes. |

The Go TypeScript gitlink is d92d9bfee114c80be2c375d72edae966176e3a4f at both cohere commits; the TypeScript-shim tree is also identical. The old Go comparisons use a detached worktree and the same current captured inputs, rather than assuming unchanged implementations imply unchanged output.

## Validation

The new provenance guard passes its git-repository fixture. Replacing its capture-pin comparison with false is caught by TestCaptureAndCheckoutMustMatchGitlink with stale capture accepted. Checkout drift is tested independently. Logs: /tmp/pin-land2/pin-test.log and /tmp/pin-land2/pin-mutant.log.

All requested packages run with -count=1 -timeout 3h -p 1 -parallel 4. Test output goes directly to log files. The environment restart interrupted an earlier attempt; those partial results are excluded and retained under /tmp/pin-land2/interrupted/.

Input configuration: /tmp/pin-land2/gate-env.sh. External oracles: pinned TypeScript 6.0.3, Prettier 3.9.6, typescript-estree 8.65.0, and the setup-provided markdown width packages. ESTree's completed Go/Node audit covers every eligible tracked source under stage1 plus pinned TypeScript src/compiler: 1,114 files, 1,110 identical and four separately recorded port refusals. Its manifest, Go answers and port records are in /tmp/pin-land2/estree/. JSON's -O2 -g profile binary is freshly compiled from this merged tree.

Toolchain setup succeeded in 46.724 seconds. nproc: 5, CPU quota: 4. Commands and package logs: /tmp/pin-land2/gate/; recapture/comparison logs: /tmp/pin-land2/.

The first lint run rejected the uncommitted generated kind table through its immutable-source guard. Committing the recapture restored that guard; the entire lint package then passed. JSON initially lacked 18 locked stage3/api dependency inputs; npm ci supplied them and the entire JSON package passed on rerun. No guard was relaxed.

ESTree initially failed TestRepositoryAgreement because its capture included ignored lint/.generated/registry.ts, which lint regenerated during the gate. The reproducer and original records are retained in /tmp/pin-land2/volatile-estree-audit/. Recapturing git ls-files stage1 plus the pinned compiler corpus excludes this volatile build output. Both corpus-dependent tests passed on retry in 286.032 seconds: /tmp/pin-land2/estree-corpus-retry.jsonl. The other 62 passing checks, including mutants, were retained; the full ESTree package was not rerun after the corrected corpus.

The only remaining skip is lint TestCheckerBridgeRefusalPending (stage1/cohere/lint/checker_pending_test.go:51). It awaits tsgoInspect returning TSGoError from the C error buffer; internal/load/prelude.d.ts:17 currently declares a string return. All external inputs were supplied. This bridge language gap is outside the helper-pin unit and was left intact.

Gofmt and vet are clean: /tmp/pin-land2/gofmt.log and /tmp/pin-land2/vet.log. The generated kind table also reproduces byte-for-byte: /tmp/pin-land2/kinds-verify.log.

## Package gate results

Counts include named subtests. Wall times are seconds, including go test startup. Every invocation used go test -json -count=1 -timeout 3h -p 1 -parallel 4.

| Package | Pass | Fail | Skip | Wall seconds |
| --- | ---: | ---: | ---: | ---: |
| internal/coherepin | 1 | 0 | 0 | 0.199 |
| stage1/cohere/lint/helpers | 30 | 0 | 0 | 440.419 |
| stage1/cohere/lint/helpers/comments | 10 | 0 | 0 | 140.878 |
| stage1/cohere/lint/helpers/core | 19 | 0 | 0 | 161.491 |
| stage1/cohere/lint/helpers/ecmascript/text | 24 | 0 | 0 | 242.693 |
| stage1/cohere/lint/helpers/from_wave1_11 | 1 | 0 | 0 | 30.501 |
| stage1/cohere/lint/helpers/module | 11 | 0 | 0 | 12.703 |
| stage1/cohere/lint/helpers/react | 1 | 0 | 0 | 148.255 |
| stage1/cohere/lint/helpers/rules-react | 42 | 0 | 0 | 863.207 |
| stage1/cohere/lint/helpers/slot03/batch3 | 1 | 0 | 0 | 14.105 |
| stage1/cohere/lint/helpers/slot04_wave20 | 8 | 0 | 0 | 25.879 |
| stage1/cohere/lint/helpers/slot05/batch32 | 3 | 0 | 0 | 27.852 |
| stage1/cohere/lint/helpers/slot05/batch33 | 3 | 0 | 0 | 32.147 |
| stage1/cohere/lint | 154 | 0 | 1 | 873.361 |
| stage1/cohere/json | 44 | 0 | 0 | 516.200 |
| stage1/cohere/estree | 62 | 1 | 0 | 985.355 |
| stage1/cohere/markdownblocks | 118 | 0 | 0 | 234.476 |
| stage1/cohere/tsprinter | 51 | 0 | 0 | 811.060 |
| stage1/typescript/parser | 178 | 0 | 0 | 859.087 |
| stage1/cohere/estree corpus retry (two tests) | 2 | 0 | 0 | 286.032 |

ESTree's initial failure is the volatile-corpus input described above; its corrected corpus tests pass, and all other package tests passed in the original run. Lint's one skip is named above. No other package skipped a test. The interrupted parser run is excluded from these counts and retained in /tmp/pin-land2/interrupted-parser/.

Final machine observation: nproc 5; CPU quota 4; load averages 6.40, 6.74, 6.53. No Go helper answer changed on an existing input, so no port implementation needed changing. Nothing else was stopped on.

Package JSON logs and summary: /tmp/pin-land2/gate/. Individual comparison records include /tmp/pin-land2/imports-old-comparison.log, /tmp/pin-land2/react-old-comparison.log, /tmp/pin-land2/batch-comparison.json and /tmp/pin-land2/react-delta.json. The actual JSX and React captures, with capture-pin.json, are in /tmp/pin-land2/jsx/ and /tmp/pin-land2/react/.
