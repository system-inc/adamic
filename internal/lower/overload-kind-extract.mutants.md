# An overload's Extract by kind, checked at each call: mutants (#0kkpw2e set 1)

Each mutant is one exact edit on this head, applied alone, run, then restored from `HEAD`. The pins are
internal/lower/overload_kind_extract_test.go: `TestOverloadKindExtractIsChecked` (x1, tsc's
injectClassNamedEvaluationHelperBlockIfMissing shape, agreeing with Node natively and on the JavaScript backend),
`TestOverloadKindExtractLieStopsAtTheCall`, `TestOverloadKindExtractStaysRefusedInA` and
`TestOverloadKindExtractUnsoundTagsStayRefused` (shared value, kind union, getter kind).

## 1. The distinct-value test dropped: two members whose kinds share a value pass

File: internal/lower/overload_kind_extract.go

Old:

```go
		value := kind.AsLiteralType().Value()
		if values[value] {
			return false // Another member carries this value, so the kind can't say which member a value is.
		}
		values[value] = true
```

New:

```go
		value := kind.AsLiteralType().Value()
		values[value] = true
```

Expected: `TestOverloadKindExtractUnsoundTagsStayRefused/shared_value` reds: Beta and Gamma are both 2 in one
enum, tsc gives them one literal type, and with nothing asking whether another member carries Beta's value the
overload is accepted narrow, so the row no longer stops at its declaration (it should lower, its one call
checked against Beta | Gamma). The kind union and getter kind rows stay green (a union of
literals isn't an enum literal, and a getter's kind is an accessor), and so do x1, its lie and its .a twin.
