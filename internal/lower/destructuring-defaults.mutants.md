# Mutants for destructuring defaults (#smqr4vn)

The pins are `TestDestructuringDefaultsAgree` (internal/oracle/destructuring_defaults_test.go) over
internal/oracle/testdata/destructuring_defaults/, each lowering and agreeing with Node natively, sanitized and
on JavaScript. Run one with:

```
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -count=1 -run 'TestDestructuringDefaultsAgree/<program>'
```

## (a) The default evaluated eagerly

internal/lower/collections.go, `destructuredDefault`. Old:

```go
	chosen := ir.Conditional{Condition: ir.IsUndefined{Value: held}, WhenTrue: fallback, WhenNot: present, Of: of}
```

New:

```go
	chosen := ir.Effects{Body: []ir.Statement{ir.Evaluate{Value: fallback}}, Result: ir.Conditional{Condition: ir.IsUndefined{Value: held}, WhenTrue: fallback, WhenNot: present, Of: of}}
```

Red: `TestDestructuringDefaultsAgree/default_effect.a` on all three backends. fallback runs for a present
field too, and twice for an absent one, so pick prints 1, 103, 105, 0 and a count of 6 where Node prints 1,
101, 102, 0 and 2, and label's order reads a,b,b,c,c where Node's reads b,c.

## (b) ?? in place of the undefined test

internal/lower/collections.go, `destructuredDefault`. Old:

```go
	chosen := ir.Conditional{Condition: ir.IsUndefined{Value: held}, WhenTrue: fallback, WhenNot: present, Of: of}
```

New (present is otherwise unused, which Go refuses):

```go
	_ = present
	chosen := ir.Coalesce{Value: held, Fallback: fallback, Of: of}
```

Live witness: `TestDestructureUnionAgrees/null/native`, `/sanitized` and
`/javascript`, over internal/oracle/testdata/destructure_union_null.a. Union fields
now reach the undefined-only guard. A present null stays null on Node but takes
the default under this mutant. The fixture's adjacent `.mutant.md` spells out the
undefined-or-null condition mutant that was run and caught on all three backends.

## (c) mayHoldNull narrowed to includesNull

internal/lower/expression.go, `mayHoldNull`. Old:

```go
	return l.includesNull(concrete) || concrete.Flags()&(checker.TypeFlagsUnknown|checker.TypeFlagsAny|checker.TypeFlagsInstantiable) != 0
```

New:

```go
	return l.includesNull(concrete)
```

There is no destructuring witness for this narrower guard: unknown is held as
Union and deliberately bypasses mayHoldNull because its null sentinel is distinct.
The former unknown refusal is now an agreement in
internal/oracle/testdata/destructure_union_former_stops.a. The parameter-default
witness remains #1mw3c4x's unknown_param.a.
