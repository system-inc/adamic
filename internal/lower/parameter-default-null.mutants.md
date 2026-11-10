# Mutant for the parameter default's undefined test

Applied to internal/lower/functions.go. Replace

		prologue = append(prologue, ir.Declare{Local: parameter.local, Value: ir.Conditional{Condition: ir.IsUndefined{Value: incoming}, WhenTrue: fallback, WhenNot: present, Of: of}})

with

		_ = present
		prologue = append(prologue, ir.Declare{Local: parameter.local, Value: ir.Coalesce{Value: incoming, Fallback: fallback, Of: of}})

Then `ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestParameterDefaultNullAgrees' -count=1 -v`: number_null.a, string_null.a and effect_on_undefined_only.a go red (a passed null takes the default, as on main be2242c1), and maybe_control.a stays green.

# Mutant (b): unknown, any and open types read as holding no null

Replace

		holdsNull := l.includesNull(concreteType) || concreteType.Flags()&(checker.TypeFlagsUnknown|checker.TypeFlagsAny|checker.TypeFlagsInstantiable) != 0

with

		holdsNull := l.includesNull(concreteType)

Then `ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestParameterDefaultNullAgrees/unknown_param.a' -count=1 -v`: unknown_param.a goes red, since f(null) takes the default (Compiler's must on #1mw3c4x). If unknown_param.a doesn't lower at all, it stops before the default, and this mutant has no witness: say so.
