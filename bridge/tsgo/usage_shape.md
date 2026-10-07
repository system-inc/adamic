# usage-shape

Wave 05's checker question exposes raw type topology for native type-parameter usage counting. It accepts function-like declarations, type parameters and property declarations; suffixes and other node kinds fail. No lint verdict or usage count is computed in Go.

The existing length-framed UTF-16 field protocol starts with version `1`, question `usage-shape`, root ID and record count. IDs are local to this response, one-based; zero means absent. Identity is reserved before recursion. Each record contains:

1. Type flags, object flags, type-parameter declaration start and end, declaration filename.
2. Constraint ID, default type ID, array flag, tuple flag, readonly tuple flag, symbol name.
3. Nine count-prefixed ID lists: alias arguments, component types, reference arguments, property types, numeric/string index types, signatures, signature parameters, signature type parameters, and signature return types.

Signature records have zero flags and carry their receiver/parameters, type parameters and return or predicate type. Declaration spans use the checker's source positions. The Adamic decoder validates framing and ID bounds. The native traversal owns visitation limits, multiplicity, witnesses, diagnostics and suggestions.

`checker/facts.go` registers the question with one added switch line. The existing unsupported-question refusal remains intact, including its independent mutant. This shared registration is intentionally kept on one physical line as required by the wave instructions; gofmt would expand it to two lines. New Go files are gofmt-clean.
