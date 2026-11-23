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
does on this head. m3 and m4 are the edits that would let them through.

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

Expected: `TestOverloadCertificateFilterNegativesStayRefused/m1_outer_alias` reds: `const result = frozen` counts as
fresh, its only reference is the return, so the census accepts bad by the certificate and the row no longer is
refused for its result. "m1 second parameter" stays green, since its return is a parameter other than the first,
never a const, and "m2 written parameter" stays green on its write. The pins stay green.

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

Expected: `TestOverloadCertificateFilterNegativesStayRefused/m2_written_parameter` reds: `array = frozen` is a write
nothing refuses, so the census accepts bad by the certificate. Both m1 rows stay green, and so do the pins.
