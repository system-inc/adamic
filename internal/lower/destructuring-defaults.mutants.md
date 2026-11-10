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

No live witness. Every field whose type holds null and undefined is held as a union with a null member,
which destructureFrom stops before destructuredDefault is reached ("a destructured name held otherwise than
its field", pinned by internal/lower's TestDestructuringDefaultNullStops), and a reference field whose null
and undefined are held alike stops in destructuredDefault itself. So the undefined test is never asked to
tell a present null from undefined today. Whoever admits a null-holding field's union to destructuring
holds a present null to Node first (Node prints null where ?? takes the default), and this mutant must turn
that pin red.

## (c) mayHoldNull narrowed to includesNull

internal/lower/expression.go, `mayHoldNull`. Old:

```go
	return l.includesNull(concrete) || concrete.Flags()&(checker.TypeFlagsUnknown|checker.TypeFlagsAny|checker.TypeFlagsInstantiable) != 0
```

New:

```go
	return l.includesNull(concrete)
```

Red: MEASURED_C
