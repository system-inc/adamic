# Structure JSX return helpers

Three separate Adamic .a files port isJsxValue(nodes, index), jsxAfterParentheses(nodes, index) and returnArgumentLooksLikeJsx(nodes, index). Each takes the handed node's numeric index in an immutable flat parser projection. The numbers are node identities, not numeric kinds. model.a carries only boolean classifications and direct child indices. No rule dispatch or registration is authored here.

isJsxValue accepts exactly JSX elements, fragments and self-closing elements. jsxAfterParentheses unwraps any number of ParenthesizedExpression nodes, then applies isJsxValue. returnArgumentLooksLikeJsx first unwraps parentheses; it accepts direct JSX, or one conditional expression whose immediate true or false branch is JSX after unwrapping its parentheses. It does not recurse into nested conditionals, inspect the condition, unwrap assertions or search arbitrary children. Every helper returns false for -1, representing Go nil.

A caller supplies actual parser kind classifications and the ParenthesizedExpression.Expression / ConditionalExpression.WhenTrue / WhenFalse edges. The valid domain is an acyclic immutable AST, mutually consistent flags, valid node indices and non-nil children under parenthesized expressions. An out-of-range node index fails explicitly rather than silently becoming a clean verdict. Arbitrary malformed projections and cyclic trees are outside this contract. This adapter is not an independent parser.

The owned oracle parses every captured consumer fixture string with pinned typescript-go and queries every resulting node plus nil against the unchanged original helpers. Controls include all three JSX kinds, conditional arms/condition/nested conditionals, parenthesis depth 128, as/satisfies/non-null/type assertions, logical expressions, arrays, object literals and bare returns. The type-assertion control is parsed as TS; other sources as TSX. Recovery trees from captured strings are retained. Node source, emitted JavaScript and sanitized native consume identical factual AST shapes; no Go result is passed as an input to the helpers.

With the setup environment sourced:

```
ADAMIC_GATE_UNCACHED=1 go test ./stage1/cohere/lint/helpers/slot05/batch32 -count=1 -v -timeout=20m > /tmp/lint05-batch32-final-helpers.log 2>&1
```

Set ADAMIC_SLOT05_BATCH32_EVIDENCE to an existing directory to retain generated raw cases, actual Go answers and coverage. Published evidence is compressed losslessly. This is helper parity on Go's actual parser geometry; whole-rule findings and Adamic's independently parsed AST remain separate integration work.
