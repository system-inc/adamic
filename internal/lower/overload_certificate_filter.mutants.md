# Certificate D for filter, under certificate A's per-call gate: mutants (#4z7y6pq)

Each mutant is one exact edit on this head, applied alone, run, then restored from `HEAD`. The pins are
internal/lower/overload_certificate_filter_test.go: `TestOverloadCertificateFilterKeptArgumentComesBack` and
`TestOverloadCertificateFilterDroppedResultIsFresh` (tsc's filter whole, overload 2 called directly, its argument
handed back and a fresh slice, each agreeing with Node natively and on the JavaScript backend),
`TestOverloadCertificateFilterMaybeArray` (overload 6, with undefined and with an array),
`TestOverloadCertificateFilterCensus` (overloads 2 and 6 served, 3 and 7 refused for their result),
`TestOverloadCertificateFilterPredicateOverloadsStayRefused` (overloads 3 and 7 stopped at their call) and
`TestOverloadCertificateFilterNegativesStayRefused`.

m1 and m2 are programs, not edits: the negative rows themselves. m1 returns an alias the body didn't build, a
module's readonly array through a const ("m1 outer alias") and the second parameter ("m1 second parameter"); m2
writes the parameter before returning it ("m2 written parameter"). Each must stay refused for its result, and
does on this head. m3, m3name and m4 are the edits that would let them through, and m5 the storage condition's
reading that overload 2's calls need.

## m3. The freshness test dropped: every const that doesn't hold an element is fresh

File: internal/lower/overload_certificate_filter.go

Old:

```go
				case initializer.Kind == ast.KindArrayLiteralExpression && len(initializer.AsArrayLiteralExpression().Elements.Nodes) == 0, l.overloadFilterSlice(initializer, array):
```

New:

```go
				default:
```

Observed (J-5fnqd1q): `TestOverloadCertificateFilterNegativesStayRefused/m1_outer_alias` reds, `const result = frozen`
counted fresh, its only reference the return, so the census accepts bad ("want overload 1 of bad refused by the
census, got <nil>"). The pins red with it: filter's own `const len = array.length` counts as fresh too, and its
read in `i < len` is neither a push's receiver nor a return, so filter's body fails the shape and overloads 2 and 6
are refused again (`KeptArgumentComesBack`, `DroppedResultIsFresh`, `MaybeArray`, `Census`). "m1 second parameter"
and "m2 written parameter" stay green.

## m3name. The freshness test widened: a const initialized to a name is fresh

File: internal/lower/overload_certificate_filter.go

Old:

```go
				case initializer.Kind == ast.KindArrayLiteralExpression && len(initializer.AsArrayLiteralExpression().Elements.Nodes) == 0, l.overloadFilterSlice(initializer, array):
```

New:

```go
				case initializer.Kind == ast.KindArrayLiteralExpression && len(initializer.AsArrayLiteralExpression().Elements.Nodes) == 0, l.overloadFilterSlice(initializer, array), ast.IsIdentifier(initializer):
```

Observed (J-5fnqd1q): only `TestOverloadCertificateFilterNegativesStayRefused/m1_outer_alias` reds, so that row is
held by the freshness test alone. Every pin and the other negatives stay green.

## m4. The parameter's reads unjudged: overloadFilterParameterRead answers true

File: internal/lower/overload_certificate_filter.go

Old:

```go
func (l *lowering) overloadFilterParameterRead(reference *ast.Node, array *ast.Symbol) bool {
```

New:

```go
func (l *lowering) overloadFilterParameterRead(reference *ast.Node, array *ast.Symbol) bool {
	return true
```

Observed (J-5fnqd1q): only `TestOverloadCertificateFilterNegativesStayRefused/m2_written_parameter` reds: `array =
frozen` is a write nothing refuses, so the census accepts bad. Both m1 rows and every pin stay green.

## m5. The storage condition's undefined arm not read through

File: internal/lower/overload_certificate_array.go

Old:

```go
	if storedHeld.IsReference() {
```

New:

```go
	if false && storedHeld.IsReference() {
```

Observed (J-5fnqd1q): `TestOverloadCertificateFilterKeptArgumentComesBack` and `DroppedResultIsFresh` red, each
stopped at its call: "a certified overload whose implementation stores readonly number[] | undefined where the call
reads number[]". `MaybeArray` stays green, its two sides both unions, and every certificate A and B row stays green.
