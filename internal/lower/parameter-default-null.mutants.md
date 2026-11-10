# Mutant for the parameter default's undefined test

Applied to internal/lower/functions.go. Replace

		prologue = append(prologue, ir.Declare{Local: parameter.local, Value: ir.Conditional{Condition: ir.IsUndefined{Value: incoming}, WhenTrue: fallback, WhenNot: present, Of: of}})

with

		_ = present
		prologue = append(prologue, ir.Declare{Local: parameter.local, Value: ir.Coalesce{Value: incoming, Fallback: fallback, Of: of}})

Then `ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestParameterDefaultNullAgrees' -count=1 -v`: number_null.a, string_null.a and effect_on_undefined_only.a go red (a passed null takes the default, as on main be2242c1), and maybe_control.a stays green.
