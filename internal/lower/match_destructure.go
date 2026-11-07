package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// matchArray reports whether a destructured type is a RegExp result, which is held as an array of
// strings whose absent captures read as undefined.
func (lowering *lowering) matchArray(destructured *checker.Type) bool {
	return lowering.isLibraryType(
		lowering.checker.GetNonNullableType(lowering.concrete(destructured)),
		"RegExpExecArray",
		"RegExpMatchArray",
	)
}

// destructureMatch declares the names of const [, major, minor = "0"] = match from the local held,
// which holds a RegExp result. On an array nobody has changed the iterator of, ECMAScript reads the
// indexes in order, so each name is held[index] read as match[index] reads, which is undefined for an
// absent capture and for an index past the end; a hole skips its index; and a default is evaluated
// only when its read is undefined, after the reads before it and before the reads after it. Rest
// elements and nested patterns stay not yet.
func (lowering *lowering) destructureMatch(pattern *ast.Node, held int) ([]ir.Statement, error) {
	statements := []ir.Statement{}
	for index, binding := range pattern.AsBindingPattern().Elements.Nodes {
		if binding.Kind == ast.KindOmittedExpression || binding.Name() == nil {
			continue
		}
		declared := binding.AsBindingElement()
		if declared.DotDotDotToken != nil {
			return nil, lowering.notYet(binding, "a rest element when destructuring a RegExp match")
		}
		if !ast.IsIdentifier(binding.Name()) {
			return nil, lowering.notYet(binding, "a nested pattern when destructuring a RegExp match")
		}
		value := ir.Expression(ir.ArrayIndex{
			Array:   ir.Read{Local: held, Of: ir.Array},
			Index:   ir.NumberConstant{Value: float64(index)},
			Element: ir.String,
		})
		var fallback ir.Expression
		if declared.Initializer != nil {
			// Lowered before the name is declared, so the default can't see the name it initializes.
			var expressionError error
			if fallback, expressionError = lowering.expression(declared.Initializer); expressionError != nil {
				return nil, expressionError
			}
		}
		local, declareLocalError := lowering.declareLocal(binding.Name())
		if declareLocalError != nil {
			return nil, declareLocalError
		}
		of := lowering.result.Locals[local].Type
		if of != ir.String {
			return nil, lowering.notYet(binding, "a RegExp match name held as a "+typeName(of))
		}
		if fallback != nil {
			fallback = fit(fallback, of)
			if fallback.Type() != of {
				return nil, lowering.notYet(binding, "a RegExp match default of another representation than its name")
			}
			// A destructuring default fires on undefined only, and ir.Coalesce is ??, which fires on null too. They
			// agree here because a match array's elements are a string or undefined, never null; a matchArray that
			// admits a type whose elements can be null needs its own test here.
			value = ir.Coalesce{Value: value, Fallback: fallback, Of: of}
		}
		statements = append(statements, lowering.initializeLocal(local, value)...)
	}
	return statements, nil
}
