# Certificate B for arrayToMap, under certificate A's per-call gate: mutants (#5fws6q8)

Each mutant is one exact edit on this head, applied alone, run, then restored from `HEAD`. The pins are
internal/lower/overload_certificate_map_test.go: `TestOverloadCertificateArrayToMapLowers` (tsc's arrayToMap,
overloads 1 to 3 each called directly, agreeing with Node natively and on the JavaScript backend),
`TestOverloadCertificateArrayToMapFourthOverloadStops` (all four overloads stop at overload 4's parameter
census, before the certificate is asked), `TestOverloadCertificateMapNegativesStayRefused` (Correctness's five
Map negatives as she wrote them, and eight canonical rows each standing on its own hazard),
`TestOverloadCertificateMapIdentityDefaultAccepted`, `TestOverloadCertificateMapPrototypeRewriteStaysRefused` and
`TestOverloadCertificateMapGate` (a direct entry; comma and spread entries NotYet at the call). Mutants 3 and 4
edit the walk and the gate certificate A shares, so they are A's mutants 2 and 1 and red A's rows too.

## 1. The identity proof skipped: any default stands for the identity

File: internal/lower/overload_certificate_map.go

Old:

```go
		if initializer == nil || !l.overloadIdentityDefault(initializer) {
```

New:

```go
		if initializer == nil {
```

Expected: `TestOverloadCertificateMapNegativesStayRefused/rebound_identity` and `/default_other_than_identity`
red: an omitted mapper runs the default, and with its binding unproven the census accepts bad by the element
relating to the promised value, so neither row stops at its declaration (it stops later at the call, the
implementation's Map<string, string | number> in another representation than Map<string, string>, or lowers).
`/supplied_undefined_mapper_over_identity` stays green, its element a number against a promised string.

## 2. The mapped overload's default exclusion dropped: a present mapper means the default can't run

File: internal/lower/overload_certificate_map.go

Old:

```go
		defaults = ast.HasQuestionToken(declared[shape.mapper]) || l.censusHasUndefined(mapper)
```

New: the line deleted, so a mapper the overload declares is never taken to run the default.

Expected: `TestOverloadCertificateMapNegativesStayRefused/supplied_undefined_mapper_over_identity` reds: its
overload types makeValue as ((value: number) => string) | undefined, whose function part answers the promised
string, so the census accepts bad, though `bad([0], v => "a", undefined)` runs identity and sets 0
(map-default-undefined). Every other row stays green, the identity rows among them, since their overload omits
the mapper.

## 3. The alias test dropped (certificate A's mutant 2, in the shared walk)

File: internal/lower/overload_certificate_array.go

Old:

```go
		return false // Any other read is an alias or a mutator no certificate judged, writing what nothing proved (array-alias, map-alias, map-wrong-set).
```

New: the line deleted.

Expected: `TestOverloadCertificateMapNegativesStayRefused/alias_over_a_parameter` and
`/wrong_set_over_a_parameter` red: `const alias = result` and a set after the loop are references to the local
that are neither the judged set's receiver nor the return, and with nothing refusing them the census accepts
bad. Her own map-alias and map-wrong-set stay green (a for-of over a literal), and so do the capture, getter,
default and identity rows. A's alias rows red with them.

## 4. The gate answers true for every call (certificate A's mutant 1)

File: internal/lower/overload_certificate_array.go

Old:

```go
func (l *lowering) overloadCertificateEntry(node *ast.Node, overload *ast.Node) (erase bool) {
```

New:

```go
func (l *lowering) overloadCertificateEntry(node *ast.Node, overload *ast.Node) (erase bool) {
	if node != nil && (node.Kind == ast.KindCallExpression || node.Kind == ast.KindNewExpression || node.Kind == ast.KindTaggedTemplateExpression) {
		return true
	}
```

Expected: `TestOverloadCertificateMapGate/comma_callee` and `/spread_argument` red: overloadCertificateEntry
answers true for each, and lowered whole neither stops at the gate in its words. The direct row, the lowering
pins and the negatives stay green. A's comma, spread and tagged template rows red with them.
