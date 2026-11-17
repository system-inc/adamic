# Certificate B for arrayToMap, under certificate A's per-call gate: mutants (#5fws6q8)

Each mutant is one exact edit on this head, applied alone, run, then restored from `HEAD`. The pins are
internal/lower/overload_certificate_map_test.go: `TestOverloadCertificateArrayToMapLowers` (tsc's arrayToMap whole,
each of its four overloads called directly, agreeing with Node natively and on the JavaScript backend),
`TestOverloadCertificateArrayToMapFourthOverloadServed` (the four-overload set with overload 4 called lowers,
its V2 witnessed as U by censusInferRemainder), `TestOverloadCertificateArrayToMapFourthOverloadElementSetStaysRefused`
(overload 4 over a body that sets the element is refused for its result by the certificate),
`TestOverloadCertificateMapInstanceStorage` and, in
overload_certificate_array_test.go, `TestOverloadCertificateArrayInstanceStorage` (Lower's probe J-h37mw8d:
calls whose mapper changes the element's type, each stopping at the call in the storage condition's words or
agreeing with Node natively), `TestOverloadCertificateMapNegativesStayRefused` (Correctness's five
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

## 5. The remainder read dropped: overload 4's V2 is bottom again

File: internal/lower/census_small.go

Old:

```go
				l.censusInferRemainder(pair[0], pair[1], served, inferred)
```

New:

```go
				_, _ = pair, served
```

Expected: `TestOverloadCertificateArrayToMapFourthOverloadServed` reds, refused at overload 4's parameter
makeValue as before the witness: V2 is bottom, the implementation's makeValue is (value: T) => T, and
(value: T) => U doesn't serve it. So does every row lowering tsc's whole set, since the census refuses the set
at overload 4 whichever overload a call resolves to: each `TestOverloadCertificateArrayToMapLowers` subtest,
and the comma and spread rows of `TestOverloadCertificateMapGate` (refused before the gate's words). The
element-set negative reds too, refused at the parameter rather than its result, which is the row's point: the
witness gets overload 4 to the certificate, and the certificate refuses it. The Map negatives, the identity
rows and A's rows stay green.

## 6. The storage condition dropped: a direct call is erased whatever its instance stores

File: internal/lower/overload_certificate_array.go

Old:

```go
		return l.overloadCertificateStorage(node, overload, resolved)
```

New:

```go
		return nil
```

Expected: `TestOverloadCertificateArrayInstanceStorage/string_to_number` reds: the instance builds
(string | number)[] of boxed unions, the call reads number[], and natively it prints 1.0,1.0 where Node prints
4.0,2.0 (J-h37mw8d), so the row's native check fails. Every `TestOverloadCertificateMapInstanceStorage` row reds
too, each stopping inside the instance where no Map of a union lowers, in words other than the storage
condition's. The T = U rows, `TestOverloadCertificateArrayFromLowers` and `TestOverloadCertificateArrayToMapLowers`,
stay green: their instances store what the call reads.
