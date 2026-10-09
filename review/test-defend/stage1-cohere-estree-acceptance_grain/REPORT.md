TestAcceptanceDiagnosticControl is defended in the bounded eighteen-test matrix.
D1 fails only this row; four semantic comparison rows and thirteen new product rows pass.
The defense establishes a positive acceptance control, not the refusal-check witness claimed in its log.

Started clean from origin/main 955e3eb92b9cd04aca420974d1006048b4615d51 on test-defend/stage1-cohere-estree-acceptance_grain. The audit branch was fetched with the full requested refspec. Its REPORT, rows, issues, reachability, inventory/menu and matrix were read. Current listing has 245 top-level tests versus 233 at audit: thirteen product tests were added and TestProduct_SyntaxMutantsSetup vanished. Every added test is included in D1-new-products.log. The target still exists at acceptance_test.go:53.

Before mutation the code under test and oracle were named. Code under test is Adamic's ESTree TypeScript port Parser/Converter path, with the mutation in Converter.convert. Oracle is live Go cohere's refusal of ++await 42;, together with the self-written expectation that the validation-disabled port succeeds and contains 0 Program. Node and sanitized native run the port. Go cohere, tests, builders, existing built-in mutation and comparison harness were unchanged.

The prior audit called this row untrue as a witness: weakening refusedBeforeDeadline did not affect it, because the row never calls that checker. Its positive acceptance assertion is separate and worth attempting to defend. The control's built-in mutant changes only pipeline.ts's syntax refusal condition to false. syntaxError still computes a refusal, but the port continues through conversion. No other current test disables this syntax guard; reached-callers.txt records the search. TestAcceptanceDiagnostics uses the exact same input with ordinary validation and refuses before conversion. TestGeneratedAgreement tests valid awaits and valid updates; TestRecoveredExpressions tests permissive await parsing with valid operands. TestAcceptanceGrammar guards successful ordinary conversion.

Clean coverage: control.cover was collected for ^TestAcceptanceDiagnosticControl$ with -coverpkg=github.com/system-inc/adamic/internal/lower,github.com/system-inc/adamic/internal/native. comparison.cover covers TestAcceptanceDiagnostics and TestGeneratedAgreement, and recovered.cover covers TestRecoveredExpressions. These instrument Go compiler/product builders, not the TypeScript port. No control-exclusive Go blocks were found against the combined comparison set. NODE_V8_COVERAGE also observed actual Converter source execution. No converter source lines were exclusive after mapping both source forms correctly. control-convert-transformed.js reproduces the existing temporary-copy import rewrites; comparison-convert-transformed.js uses original imports. Raw V8 records and coverage-difference.json preserve the result. The defense rests on a semantic input difference at a shared dispatch line: AwaitExpression underneath PrefixUnaryExpression with PlusPlusToken reaches conversion only in the validation-disabled control among the observed rows.

D1, planned before its failures were examined, flips an existing unary dispatch condition for precisely that node/parent combination. It adds no statement, panic, input literal, or test-specific environment switch. The converter falls through to its existing unrepresented-node panic. This is an aimed production condition mutant. The standalone D1.diff applies to origin/main at convert.ts:1552. It preserves normal valid-await and valid-update conversion and the ordinary invalid-input refusal path. Mutation cache is /tmp/estree-defense/cache/D1, separate from clean. The control's ordinary build helper lowers and builds sanitized native before invoking Node, so native compilation of this diff completed. Node fails first in evaluation of the result map; this failing run does not execute the native control binary. Native builds and native comparisons in the other semantic rows also completed successfully.

Observed failure: acceptance_test.go:61: exit status 70; adamic: panic: ESTree conversion not yet represented: AwaitExpression at 2.
Command: ADAMIC_NATIVE_SPLIT=1 ADAMIC_BUILD_CACHE_DIR=/tmp/estree-defense/cache/D1 timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/estree/ -run '^TestAcceptanceDiagnosticControl$' > D1-TestAcceptanceDiagnosticControl.log 2>&1

Observed passes:
- TestAcceptanceDiagnostics
- TestAcceptanceGrammar
- TestGeneratedAgreement
- TestProduct_DeepMutantsIR0
- TestProduct_DeepMutantsIR1
- TestProduct_DeepMutantsIR2
- TestProduct_SyntaxMutantIR_000
- TestProduct_SyntaxMutantIR_001
- TestProduct_SyntaxMutantIR_002
- TestProduct_SyntaxMutantNative_000
- TestProduct_SyntaxMutantsSetup_000
- TestProduct_SyntaxMutantsSetup_001
- TestProduct_SyntaxMutantsSetup_002
- TestProduct_UnattachedDecoratorIR
- TestProduct_UnattachedDecoratorLowered
- TestProduct_UnattachedDecoratorNative
- TestRecoveredExpressions

The target is the only failed row in this bounded matrix. Every selected invocation completed under 90 seconds. Full-package and repository-wide uniqueness remain unknown outside these rows. This is a bounded defense, not a package-wide uniqueness claim. The semantic rows cover the specific invalid input, valid updates and valid awaits; all newly added product rows are also observed. Other current package rows were not replayed after the whole baseline exceeded its budget. One mutant was enough for this bounded defense, so no second or third aimed attempt was spent.

Source was restored and the unchanged control passed again in 26.409 test-binary seconds. git apply --check D1.diff passed. Only evidence is committed, with no production or test edits.

Unclear, wrong or costly parts of the brief and audit:

- The requested 15 GB free threshold is impossible on /tmp, which has only 8.8 GB total capacity. Initial /tmp free was 2.1 GB. Named earlier-unit scratch/cache directories were removed, increasing it to 5.3 GB; final free is 5.1 GB. /workspace has about 15 GB free. Repository and installed tools were untouched. No baseline or mutation failed from disk exhaustion.
- The audit's untrue verdict is specifically about an absent refusal-check invocation. It does not establish that the row's positive assertion cannot fail. This session demonstrates that the positive control can fail under a real production converter mutant. The log text "acceptance check catches it" still promises an invocation the body does not perform; that owner finding remains even though the positive-control row is defended.
- The row asserts successful execution and a Program substring, not an exact AST. This is sufficient for its positive acceptance role but does not validate conversion details. The Go audit status check likewise requires an error status rather than a specific diagnostic.
- Go coverage cannot measure a TS subprocess port. V8 coverage required separate offset mapping because the control helper rewrites imports in its copy. Comparing those offsets against the original source would create false exclusive lines. Correct mapping finds none; shared-line semantic differences remain valid defense leads.
- Full clean baseline cooked at 90.027 seconds in TestDeepGrammar, after the target and preceding acceptance rows passed without ordinary failures. A combined comparison coverage run also cooked at 90.088 seconds. It was narrowed to a two-row run plus a separate recovered-expression run, both green. No over-budget mutant run was used as a kill.
- The verdict schema has no bounded field despite allowing large-package narrowing. rows.json adds bounded and the explicit eighteen-test matrix to avoid overstating uniqueness.
- Product tests added since the audit mainly check construction rather than runtime semantics. Their actual passes are recorded; no behavioral assertion is inferred from a build-only row.

Timing and limits: warm env.sh worked, so setup was skipped; nproc=5; npm ci in stage3/api reported 946 ms. Clean control coverage: 27.173 binary seconds. Narrowed comparison coverage: 60.130 seconds. Recovered-expression coverage: 31.760 seconds. D1 control: 50.617 binary seconds, including lowering/native rebuild, with command wall 54.217 seconds. Its build helper does not expose separate clang timing, so this is a rebuild-inclusive upper bound. D1 diagnostics: 32.622 binary seconds; all exact command walls and binary times are in run-results.json and matrix.json. D1's thirteen new products: 64.517 binary seconds. Restored control: 26.409 seconds. Total work approximately twenty minutes. No absent corpus/library opt-ins, other unselected package rows, full suite, or repository replay was covered. Every selected row ran without skipping. No performance verdict was attempted, so a cost mutant was unnecessary.
