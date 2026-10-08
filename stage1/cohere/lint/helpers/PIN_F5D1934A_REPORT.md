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

All requested packages run with -count=1 -timeout 3h -p 1 -parallel 4. Test output goes directly to log files. Package results and wall times will be recorded below after the gate finishes. The environment restart interrupted an earlier attempt; those partial results are excluded and retained under /tmp/pin-land2/interrupted/.

Input configuration: /tmp/pin-land2/gate-env.sh. External oracles: pinned TypeScript 6.0.3, Prettier 3.9.6, typescript-estree 8.65.0, and the setup-provided markdown width packages. ESTree's completed Go/Node audit covers every eligible source under stage1 plus TypeScript src/compiler: 1,115 files, 1,111 identical and four separately recorded port refusals. Its manifest, Go answers and port records are in /tmp/pin-land2/estree/. JSON's -O2 -g profile binary is freshly compiled from this merged tree.

Toolchain setup succeeded in 46.724 seconds. nproc: 5, CPU quota: 4. Commands and package logs: /tmp/pin-land2/gate/; recapture/comparison logs: /tmp/pin-land2/.
