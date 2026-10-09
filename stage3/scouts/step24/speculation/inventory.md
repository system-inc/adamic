# Speculative parsing inventory

TypeScript 6.0.3, commit 050880ce59e30b356b686bd3144efe24f875ebc8. All calls in the two complete ASTs; declarations and exported aliases are not calls.

“One” means one runtime representation family, not one literal value or one object shape. Boolean literals and numeric enums retain their exact values. “Checked” is a conservative specialization boundary, not an instruction to coerce the callback result. Generic forwarding must be instantiated at its callers. Known unions can remain tagged unions; a narrowed result projection must be checked when proof is missing. Assertions are not proof.

Every lookAhead/tryParse/tryScan result is also tested by the helper using JavaScript `!result`; lookAhead always rewinds, tryParse/tryScan commit only truthy results. scanRange always restores. Full expressions, contexts, assigned-local uses, body return expression types (bodyReturns) and assertion input types are in inventory.json.

## Counts

```json
{
  "counts": {
    "src/compiler/parser.ts": {
      "lookAhead": 60,
      "speculationHelper": 3,
      "resetTokenState": 4,
      "tryScan": 2,
      "tryParse": 20,
      "scanRange": 1
    },
    "src/compiler/scanner.ts": {
      "scanRange": 1,
      "speculationHelper": 2,
      "resetTokenState": 1
    }
  },
  "classifications": {
    "one proven runtime family": 67,
    "rewind, void result": 5,
    "checked result specialization": 22
  },
  "rescans": 21,
  "textResets": 5,
  "semanticDiagnostics": 0
}
```

## Calls

| Site | Call | Callback declaration / checker result (annotation honored) | Result consumer | Decision |
|---|---|---|---|---|
| src/compiler/parser.ts:1674 | lookAhead | (inferred, no return annotation) / boolean at src/compiler/parser.ts:1674 | truthiness (condition) | one proven runtime family |
| src/compiler/parser.ts:1683 | lookAhead | (inferred, no return annotation) / boolean at src/compiler/parser.ts:1683 | truthiness (condition) | one proven runtime family |
| src/compiler/parser.ts:1889 | speculationHelper | (inferred, no return annotation) / void at src/compiler/parser.ts:1889 | value stored/passed or discarded; see source context | one proven runtime family |
| src/compiler/parser.ts:1892 | resetTokenState | (no callback) / (no callback) | value stored/passed or discarded; see source context | rewind, void result |
| src/compiler/parser.ts:2273 | lookAhead | T / T at src/compiler/parser.ts:2256 | value stored/passed or discarded; see source context; truthiness (!result); truthiness (short circuit); truthiness (condition); returned to caller | checked result specialization |
| src/compiler/parser.ts:2274 | tryScan | T / T at src/compiler/parser.ts:2256 | value stored/passed or discarded; see source context; truthiness (!result); truthiness (short circuit); truthiness (condition); returned to caller | checked result specialization |
| src/compiler/parser.ts:2296 | speculationHelper | T / T at src/compiler/parser.ts:2295 | returned to caller | checked result specialization |
| src/compiler/parser.ts:2305 | speculationHelper | T / T at src/compiler/parser.ts:2304 | returned to caller | checked result specialization |
| src/compiler/parser.ts:2665 | tryScan | (inferred, no return annotation) / boolean at src/compiler/parser.ts:2665 | truthiness (condition) | one proven runtime family |
| src/compiler/parser.ts:2755 | tryParse | (inferred, no return annotation) / boolean at src/compiler/parser.ts:2766 | returned to caller | one proven runtime family |
| src/compiler/parser.ts:2774 | lookAhead | boolean / boolean at src/compiler/parser.ts:2824 | returned to caller | one proven runtime family |
| src/compiler/parser.ts:2777 | lookAhead | boolean / boolean at src/compiler/parser.ts:2802 | returned to caller | one proven runtime family |
| src/compiler/parser.ts:2808 | tryParse | (inferred, no return annotation) / boolean at src/compiler/parser.ts:2766 | returned to caller | one proven runtime family |
| src/compiler/parser.ts:2830 | lookAhead | (inferred, no return annotation) / boolean at src/compiler/parser.ts:7153 | truthiness (short circuit); returned to caller | one proven runtime family |
| src/compiler/parser.ts:2831 | lookAhead | (inferred, no return annotation) / boolean at src/compiler/parser.ts:7158 | returned to caller | one proven runtime family |
| src/compiler/parser.ts:2855 | lookAhead | boolean / boolean at src/compiler/parser.ts:4292 | returned to caller | one proven runtime family |
| src/compiler/parser.ts:2861 | lookAhead | boolean / boolean at src/compiler/parser.ts:7866 | truthiness (short circuit); returned to caller | one proven runtime family |
| src/compiler/parser.ts:2886 | lookAhead | (inferred, no return annotation) / boolean at src/compiler/parser.ts:2945 | returned to caller | one proven runtime family |
| src/compiler/parser.ts:2925 | lookAhead | (inferred, no return annotation) / boolean at src/compiler/parser.ts:7553 | truthiness (condition) | one proven runtime family |
| src/compiler/parser.ts:2983 | lookAhead | (inferred, no return annotation) / boolean at src/compiler/parser.ts:2989 | returned to caller | one proven runtime family |
| src/compiler/parser.ts:3046 | lookAhead | (inferred, no return annotation) / boolean at src/compiler/parser.ts:8369 | returned to caller | one proven runtime family |
| src/compiler/parser.ts:3181 | resetTokenState | (no callback) / (no callback) | value stored/passed or discarded; see source context | rewind, void result |
| src/compiler/parser.ts:3634 | lookAhead | (inferred, no return annotation) / boolean at src/compiler/parser.ts:7148 | value stored/passed or discarded; see source context; truthiness (condition) | one proven runtime family |
| src/compiler/parser.ts:3881 | tryParse | (inferred, no return annotation) / boolean at src/compiler/parser.ts:8361 | truthiness (condition) | one proven runtime family |
| src/compiler/parser.ts:4202 | lookAhead | (inferred, no return annotation) / boolean at src/compiler/parser.ts:4205 | returned to caller | one proven runtime family |
| src/compiler/parser.ts:4334 | lookAhead | (inferred, no return annotation) / boolean at src/compiler/parser.ts:4354 | truthiness (condition) | one proven runtime family |
| src/compiler/parser.ts:4465 | lookAhead | (inferred, no return annotation) / boolean at src/compiler/parser.ts:4457 | truthiness (condition) | one proven runtime family |
| src/compiler/parser.ts:4602 | tryParse | TypeNode \| undefined / TypeNode \| undefined at src/compiler/parser.ts:4523 | truthiness (short circuit); returned to caller | checked result specialization |
| src/compiler/parser.ts:4628 | lookAhead | (inferred, no return annotation) / boolean at src/compiler/parser.ts:4584 | truthiness (conditional); returned to caller | one proven runtime family |
| src/compiler/parser.ts:4641 | lookAhead | (inferred, no return annotation) / boolean at src/compiler/parser.ts:4542 | truthiness (conditional); returned to caller | one proven runtime family |
| src/compiler/parser.ts:4643 | lookAhead | (inferred, no return annotation) / boolean at src/compiler/parser.ts:4391 | truthiness (conditional); returned to caller | one proven runtime family |
| src/compiler/parser.ts:4651 | lookAhead | (inferred, no return annotation) / boolean at src/compiler/parser.ts:7148 | truthiness (conditional); returned to caller | one proven runtime family |
| src/compiler/parser.ts:4701 | lookAhead | (inferred, no return annotation) / boolean at src/compiler/parser.ts:4584 | returned to caller | one proven runtime family |
| src/compiler/parser.ts:4705 | lookAhead | (inferred, no return annotation) / boolean at src/compiler/parser.ts:4711 | returned to caller | one proven runtime family |
| src/compiler/parser.ts:4727 | lookAhead | (inferred, no return annotation) / boolean at src/compiler/parser.ts:2994 | truthiness (condition) | one proven runtime family |
| src/compiler/parser.ts:4770 | tryParse | (inferred, no return annotation) / TypeNode \| undefined at src/compiler/parser.ts:4758 | value stored/passed or discarded; see source context; value stored/passed or discarded; see source context | checked result specialization |
| src/compiler/parser.ts:4856 | lookAhead | (inferred, no return annotation) / boolean at src/compiler/parser.ts:4881 | truthiness (condition) | one proven runtime family |
| src/compiler/parser.ts:4860 | lookAhead | boolean / boolean at src/compiler/parser.ts:4847 | returned to caller | one proven runtime family |
| src/compiler/parser.ts:4914 | tryParse | (inferred, no return annotation) / Identifier \| undefined at src/compiler/parser.ts:4924 | value stored/passed or discarded; see source context; truthiness (condition); returned to caller | checked result specialization |
| src/compiler/parser.ts:4989 | lookAhead | (inferred, no return annotation) / boolean at src/compiler/parser.ts:4363 | returned to caller | one proven runtime family |
| src/compiler/parser.ts:5158 | lookAhead | (inferred, no return annotation) / boolean at src/compiler/parser.ts:7163 | returned to caller | one proven runtime family |
| src/compiler/parser.ts:5229 | tryParse | (inferred, no return annotation) / ArrowFunction \| undefined at src/compiler/parser.ts:5229 | returned to caller | checked result specialization |
| src/compiler/parser.ts:5238 | lookAhead | (inferred, no return annotation) / Tristate at src/compiler/parser.ts:5251 | returned to caller; parser.ts:5223 comparison triState === Tristate.False (0); parser.ts:5227 comparison triState === Tristate.True (1) | one proven runtime family |
| src/compiler/parser.ts:5301 | lookAhead | (inferred, no return annotation) / boolean at src/compiler/parser.ts:2963 | truthiness (condition) | one proven runtime family |
| src/compiler/parser.ts:5349 | lookAhead | (inferred, no return annotation) / boolean at src/compiler/parser.ts:5349 | value stored/passed or discarded; see source context; truthiness (condition) | one proven runtime family |
| src/compiler/parser.ts:5398 | lookAhead | Tristate / Tristate at src/compiler/parser.ts:5409 | comparison: lookAhead(isUnParenthesizedAsyncArrowFunctionWorker) === Tristate.True | one proven runtime family |
| src/compiler/parser.ts:5720 | lookAhead | (inferred, no return annotation) / boolean at src/compiler/parser.ts:7163 | returned to caller | one proven runtime family |
| src/compiler/parser.ts:5880 | lookAhead | (inferred, no return annotation) / boolean at src/compiler/parser.ts:2973 | truthiness (condition) | one proven runtime family |
| src/compiler/parser.ts:5932 | lookAhead | (inferred, no return annotation) / boolean at src/compiler/parser.ts:4354 | truthiness (condition) | one proven runtime family |
| src/compiler/parser.ts:5941 | lookAhead | (inferred, no return annotation) / boolean at src/compiler/parser.ts:4359 | truthiness (condition) | one proven runtime family |
| src/compiler/parser.ts:6028 | tryParse | (inferred, no return annotation) / NodeArray<TypeNode> \| undefined at src/compiler/parser.ts:6562 | value stored/passed or discarded; see source context; undefined comparison; value stored/passed or discarded; see source context | checked result specialization |
| src/compiler/parser.ts:6114 | tryParse | (inferred, no return annotation) / JsxElement \| JsxSelfClosingElement \| JsxFragment at src/compiler/parser.ts:6114 | value stored/passed or discarded; see source context; truthiness (condition); value stored/passed or discarded; see source context; value stored/passed or discarded; see source context; returned to caller | one proven runtime family |
| src/compiler/parser.ts:6390 | lookAhead | (inferred, no return annotation) / boolean at src/compiler/parser.ts:6381 | returned to caller | one proven runtime family |
| src/compiler/parser.ts:6490 | tryParse | (inferred, no return annotation) / NodeArray<TypeNode> \| undefined at src/compiler/parser.ts:6562 | value stored/passed or discarded; see source context; truthiness (condition); value stored/passed or discarded; see source context | checked result specialization |
| src/compiler/parser.ts:6526 | tryParse | (inferred, no return annotation) / NodeArray<TypeNode> \| undefined at src/compiler/parser.ts:6562 | value stored/passed or discarded; see source context; value stored/passed or discarded; see source context; truthiness (short circuit); truthiness (condition); value stored/passed or discarded; see source context; value stored/passed or discarded; see source context; value stored/passed or discarded; see source context | checked result specialization |
| src/compiler/parser.ts:6635 | lookAhead | (inferred, no return annotation) / boolean at src/compiler/parser.ts:7158 | truthiness (!result); truthiness (condition) | one proven runtime family |
| src/compiler/parser.ts:6939 | lookAhead | (inferred, no return annotation) / boolean at src/compiler/parser.ts:7342 | truthiness (short circuit); truthiness (condition) | one proven runtime family |
| src/compiler/parser.ts:6941 | lookAhead | (inferred, no return annotation) / boolean at src/compiler/parser.ts:7366 | truthiness (condition) | one proven runtime family |
| src/compiler/parser.ts:7243 | lookAhead | SyntaxKind / SyntaxKind at src/compiler/parser.ts:2207 | value stored/passed or discarded; see source context; comparison: currentToken === SyntaxKind.EqualsToken; truthiness (short circuit); comparison: currentToken === SyntaxKind.AsteriskToken; truthiness (short circuit); comparison: currentToken === SyntaxKind.OpenBraceToken; truthiness (short circuit); comparison: currentToken === SyntaxKind.DefaultKeyword; truthiness (short circuit); comparison: currentToken === SyntaxKind.AsKeyword; truthiness (short circuit); comparison: currentToken === SyntaxKind.AtToken | one proven runtime family |
| src/compiler/parser.ts:7265 | lookAhead | boolean / boolean at src/compiler/parser.ts:7168 | returned to caller | one proven runtime family |
| src/compiler/parser.ts:7299 | lookAhead | (inferred, no return annotation) / boolean at src/compiler/parser.ts:4363 | returned to caller | one proven runtime family |
| src/compiler/parser.ts:7324 | lookAhead | (inferred, no return annotation) / boolean at src/compiler/parser.ts:7148 | truthiness (!result); returned to caller | one proven runtime family |
| src/compiler/parser.ts:7339 | lookAhead | (inferred, no return annotation) / boolean at src/compiler/parser.ts:7331 | returned to caller | one proven runtime family |
| src/compiler/parser.ts:7354 | lookAhead | (inferred, no return annotation) / boolean at src/compiler/parser.ts:7346 | returned to caller | one proven runtime family |
| src/compiler/parser.ts:7363 | lookAhead | (inferred, no return annotation) / boolean at src/compiler/parser.ts:7351 | returned to caller | one proven runtime family |
| src/compiler/parser.ts:7377 | lookAhead | (inferred, no return annotation) / boolean at src/compiler/parser.ts:7366 | returned to caller | one proven runtime family |
| src/compiler/parser.ts:7705 | lookAhead | boolean / boolean at src/compiler/parser.ts:7723 | truthiness (condition) | one proven runtime family |
| src/compiler/parser.ts:7757 | lookAhead | SyntaxKind / SyntaxKind at src/compiler/parser.ts:2207 | comparison: lookAhead(nextToken) === SyntaxKind.OpenParenToken | one proven runtime family |
| src/compiler/parser.ts:7758 | tryParse | (inferred, no return annotation) / LiteralExpression \| undefined at src/compiler/parser.ts:7758 | returned to caller | checked result specialization |
| src/compiler/parser.ts:7766 | tryParse | (inferred, no return annotation) / ConstructorDeclaration \| undefined at src/compiler/parser.ts:7766 | returned to caller | checked result specialization |
| src/compiler/parser.ts:7987 | tryParse | (inferred, no return annotation) / boolean at src/compiler/parser.ts:2758 | truthiness (!result); truthiness (condition) | one proven runtime family |
| src/compiler/parser.ts:7991 | lookAhead | (inferred, no return annotation) / boolean at src/compiler/parser.ts:8365 | truthiness (condition) | one proven runtime family |
| src/compiler/parser.ts:8077 | lookAhead | (inferred, no return annotation) / boolean at src/compiler/parser.ts:8365 | truthiness (condition) | one proven runtime family |
| src/compiler/parser.ts:8193 | lookAhead | (inferred, no return annotation) / boolean at src/compiler/parser.ts:2968 | returned to caller | one proven runtime family |
| src/compiler/parser.ts:8257 | tryParse | TypeNode \| undefined / TypeNode \| undefined at src/compiler/parser.ts:4523 | truthiness (short circuit); value stored/passed or discarded; see source context | checked result specialization |
| src/compiler/parser.ts:8358 | lookAhead | (inferred, no return annotation) / boolean at src/compiler/parser.ts:8361 | returned to caller | one proven runtime family |
| src/compiler/parser.ts:8398 | lookAhead | (inferred, no return annotation) / boolean at src/compiler/parser.ts:7557 | truthiness (short circuit); truthiness (condition) | one proven runtime family |
| src/compiler/parser.ts:8406 | lookAhead | (inferred, no return annotation) / boolean at src/compiler/parser.ts:7553 | truthiness (!result); truthiness (condition) | one proven runtime family |
| src/compiler/parser.ts:8911 | scanRange | (inferred, no return annotation) / JSDoc at src/compiler/parser.ts:8915 | value stored/passed or discarded; see source context; returned to caller | one proven runtime family |
| src/compiler/parser.ts:9061 | lookAhead | boolean / boolean at src/compiler/parser.ts:9046 | truthiness (condition) | one proven runtime family |
| src/compiler/parser.ts:9072 | lookAhead | boolean / boolean at src/compiler/parser.ts:9046 | truthiness (condition) | one proven runtime family |
| src/compiler/parser.ts:9227 | resetTokenState | (no callback) / (no callback) | value stored/passed or discarded; see source context | rewind, void result |
| src/compiler/parser.ts:9310 | tryParse | (inferred, no return annotation) / "link" \| "linkcode" \| "linkplain" \| undefined at src/compiler/parser.ts:9346 | value stored/passed or discarded; see source context; truthiness (!result); truthiness (condition); comparison: linkType === "link"; truthiness (conditional); comparison: linkType === "linkcode"; truthiness (conditional) | checked result specialization |
| src/compiler/parser.ts:9429 | lookAhead | (inferred, no return annotation) / "link" \| "linkcode" \| "linkplain" \| undefined at src/compiler/parser.ts:9346 | truthiness (!result); truthiness (condition) | checked result specialization |
| src/compiler/parser.ts:9451 | tryParse | (inferred, no return annotation) / false \| JSDocTemplateTag \| JSDocParameterTag \| JSDocPropertyTag \| JSDocTypeTag \| JSDocThisTag at src/compiler/parser.ts:9451 | truthiness (condition); comparison: child.kind === SyntaxKind.JSDocParameterTag; truthiness (short circuit); comparison: child.kind === SyntaxKind.JSDocPropertyTag; value stored/passed or discarded; see source context; comparison: child.kind === SyntaxKind.JSDocTemplateTag; value stored/passed or discarded; see source context | checked result specialization |
| src/compiler/parser.ts:9487 | lookAhead | (inferred, no return annotation) / boolean at src/compiler/parser.ts:9487 | value stored/passed or discarded; see source context; truthiness (conditional) | one proven runtime family |
| src/compiler/parser.ts:9526 | resetTokenState | (no callback) / (no callback) | value stored/passed or discarded; see source context | rewind, void result |
| src/compiler/parser.ts:9624 | tryParse | (inferred, no return annotation) / false \| JSDocTemplateTag \| JSDocPropertyTag \| JSDocTypeTag at src/compiler/parser.ts:9624 | truthiness (condition); comparison: child.kind === SyntaxKind.JSDocTemplateTag; comparison: child.kind === SyntaxKind.JSDocTypeTag; value stored/passed or discarded; see source context; value stored/passed or discarded; see source context | checked result specialization |
| src/compiler/parser.ts:9695 | tryParse | (inferred, no return annotation) / JSDocTemplateTag \| JSDocParameterTag at src/compiler/parser.ts:9695 | truthiness (condition); comparison: child.kind === SyntaxKind.JSDocTemplateTag; value stored/passed or discarded; see source context; value stored/passed or discarded; see source context | checked result specialization; assertion hides a runtime alternative |
| src/compiler/parser.ts:9707 | tryParse | (inferred, no return annotation) / JSDocReturnTag \| undefined at src/compiler/parser.ts:9707 | value stored/passed or discarded; see source context; returned to caller | checked result specialization |
| src/compiler/scanner.ts:2602 | scanRange | (inferred, no return annotation) / void at src/compiler/scanner.ts:2602 | value stored/passed or discarded; see source context | one proven runtime family |
| src/compiler/scanner.ts:3978 | speculationHelper | T / T at src/compiler/scanner.ts:3977 | returned to caller | checked result specialization |
| src/compiler/scanner.ts:3982 | speculationHelper | T / T at src/compiler/scanner.ts:3981 | returned to caller | checked result specialization |
| src/compiler/scanner.ts:3996 | resetTokenState | (no callback) / (no callback) | value stored/passed or discarded; see source context | rewind, void result |

## Supplemental scanner rescans

These rewind/reinterpret the current token; they take no speculative callback. This includes parser wrappers and calls to them. `setTextPos` has zero calls (only an interface declaration and an exported alias of resetTokenState). setText initialization is not a speculative callback.

| Site | Call | Result type | Consumer |
|---|---|---|---|
| src/compiler/parser.ts:2225 | reScanGreaterToken | SyntaxKind | returned to caller |
| src/compiler/parser.ts:2229 | reScanSlashToken | SyntaxKind | returned to caller |
| src/compiler/parser.ts:2233 | reScanTemplateToken | SyntaxKind | returned to caller |
| src/compiler/parser.ts:2237 | reScanLessThanToken | SyntaxKind | returned to caller |
| src/compiler/parser.ts:2241 | reScanHashToken | SyntaxKind | returned to caller |
| src/compiler/parser.ts:2665 | reScanInvalidIdentifier | SyntaxKind | comparison: scanner.reScanInvalidIdentifier() === SyntaxKind.Identifier |
| src/compiler/parser.ts:3715 | reScanTemplateToken | SyntaxKind | value stored/passed or discarded; see source context |
| src/compiler/parser.ts:3741 | reScanTemplateToken | SyntaxKind | value stored/passed or discarded; see source context |
| src/compiler/parser.ts:3792 | reScanLessThanToken | SyntaxKind | comparison: reScanLessThanToken() === SyntaxKind.LessThanToken |
| src/compiler/parser.ts:4605 | reScanAsteriskEqualsToken | SyntaxKind | value stored/passed or discarded; see source context |
| src/compiler/parser.ts:4611 | reScanQuestionToken | SyntaxKind | value stored/passed or discarded; see source context |
| src/compiler/parser.ts:5128 | reScanGreaterToken | SyntaxKind | truthiness (condition) |
| src/compiler/parser.ts:5613 | reScanGreaterToken | SyntaxKind | value stored/passed or discarded; see source context |
| src/compiler/parser.ts:6171 | reScanJsxToken | JsxTokenSyntaxKind | value stored/passed or discarded; see source context |
| src/compiler/parser.ts:6510 | reScanTemplateToken | SyntaxKind | value stored/passed or discarded; see source context |
| src/compiler/parser.ts:6568 | reScanLessThanToken | SyntaxKind | comparison: reScanLessThanToken() !== SyntaxKind.LessThanToken |
| src/compiler/parser.ts:6574 | reScanGreaterToken | SyntaxKind | comparison: reScanGreaterToken() !== SyntaxKind.GreaterThanToken |
| src/compiler/parser.ts:6612 | reScanTemplateToken | SyntaxKind | value stored/passed or discarded; see source context |
| src/compiler/parser.ts:6650 | reScanSlashToken | SyntaxKind | comparison: reScanSlashToken() === SyntaxKind.RegularExpressionLiteral |
| src/compiler/parser.ts:8828 | reScanHashToken | SyntaxKind | value stored/passed or discarded; see source context |
| src/compiler/parser.ts:9337 | reScanHashToken | SyntaxKind | value stored/passed or discarded; see source context |

## Text reset calls

These can reposition the scanner but are initialization/range setup, not speculative callbacks.

| Site | Call | Result |
|---|---|---|
| src/compiler/parser.ts:1772 | scanner.setText(sourceText) | void |
| src/compiler/parser.ts:1783 | scanner.setText("") | void |
| src/compiler/parser.ts:8793 | scanner.setText(content, start, length) | void |
| src/compiler/scanner.ts:1058 | setText(text, start, length) | void |
| src/compiler/scanner.ts:3962 | setText(text, start, length) | void |
