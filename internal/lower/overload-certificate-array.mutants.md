# Certificate A for arrayFrom, under a per-call gate: mutants (#5fws6q8)

Each mutant is one exact edit on this head, applied alone, run, then restored from `HEAD`. The pins are
internal/lower/overload_certificate_array_test.go: `TestOverloadCertificateArrayFromLowers` (tsc's arrayFrom,
both overloads called directly, agreeing with Node natively and on the JavaScript backend),
`TestOverloadCertificateArrayNegativesStayRefused` (Correctness's six array negatives as she wrote them, and five
with only their loop moved onto a parameter), `TestOverloadCertificateArrayPrototypeRewriteStaysRefused`,
`TestOverloadCertificateArrayGate` (parenthesized, optional, import alias and namespace entries direct; comma,
spread and tagged template entries NotYet at the call) and `TestOverloadCertificateArrayEntryRefusesANonCall`.

## 1. The gate answers true for every call: erasure per declaration

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

Expected: `TestOverloadCertificateArrayGate/comma_callee`, `/spread_argument` and `/tagged_template` red:
overloadCertificateEntry answers true for each, and lowered whole none stops at the gate, so each either lowers
or stops in some other road's words (the comma's callee at overloadSetValueStop, "an overloaded function as a
value", which the gate must not lean on). The four direct rows, the non-call row (an identifier isn't a call),
the arrayFrom pin and the negatives stay green.

## 2. The certificate accepts a push of an aliased local: the alias test dropped

File: internal/lower/overload_certificate_array.go

Old:

```go
		if overloadArrayWritten(found.node) {
			return shape, false // An index, length or whole write puts a value no push proved (array-index-write).
		}
		return shape, false // Any other read is an alias, a second view that can write what no push proved (array-alias).
	}
```

New:

```go
		if overloadArrayWritten(found.node) {
			return shape, false // An index, length or whole write puts a value no push proved (array-index-write).
		}
	}
```

Expected: `TestOverloadCertificateArrayNegativesStayRefused/alias_over_a_parameter` reds: `const alias = out`
is a read that is neither a push's receiver nor the return, and with nothing refusing it the census accepts
bad by the certificate, so the row no longer stops at its declaration (it stops later at the call, the
implementation's (string | number)[] in another representation than string[], or lowers). Her own alias
program stays green (its loop is over a literal), and so do the index write, capture, getter and default
rows (a write, a closure, a loop body that isn't pushes, a default), the pin and the gate rows.

## 3. The shorthand arm dropped: a shorthand property's name read as the property

File: internal/lower/overload_certificate_array.go

Old:

```go
func (l *lowering) referencedSymbol(identifier *ast.Node) *ast.Symbol {
	if identifier.Parent != nil && identifier.Parent.Kind == ast.KindShorthandPropertyAssignment {
		return l.checker.GetShorthandAssignmentValueSymbol(identifier.Parent)
	}
	return l.symbol(identifier)
}
```

New:

```go
func (l *lowering) referencedSymbol(identifier *ast.Node) *ast.Symbol {
	return l.symbol(identifier)
}
```

Expected: `TestOverloadCertificateArrayNegativesStayRefused/shorthand_alias_over_a_parameter` reds, and
`/spread_shorthand_alias_over_a_parameter` and `/mapper_written_through_a_destructuring_assignment` with it:
`{ out }`'s name, and `({ map } = other)`'s, resolve to the object literal's property, so the walk sees neither
the alias nor the parameter write, the census accepts bad by the certificate, and each row no longer stops at its
declaration (it stops later at the call, or lowers). Every other row stays green. referencedSymbol is shared with the
node-test proof, so this edit reds that proof's two shorthand rows too (overload-parse-tree-node.mutants.md, m3).
