package ir

import "reflect"

// Cast locations are diagnostic metadata, never runtime certificates.
type ViewCastLocation struct {
	Value Expression
	Where string
}

func ViewDiagnosticLabel(program *Program, expression string) string {
	if location := program.ViewReadCastLocations[expression]; expression != "" && location != "" {
		return expression + " (" + location + ")"
	}
	return expression
}

// Decorate only diagnostic metadata at emission. Executable operands and the
// original IR read markers remain unchanged, including for omission mutants.
func ViewDiagnosticExpression(program *Program, expression Expression) Expression {
	if len(program.ViewReadCastLocations) == 0 {
		return expression
	}
	var decorate func(reflect.Value) reflect.Value
	decorate = func(value reflect.Value) reflect.Value {
		if value.Kind() == reflect.Pointer {
			if value.IsNil() {
				return value
			}
			result := reflect.New(value.Type().Elem())
			result.Elem().Set(decorate(value.Elem()))
			return result
		}
		if value.Kind() != reflect.Struct {
			return value
		}
		result := reflect.New(value.Type()).Elem()
		result.Set(value)
		if field := result.FieldByName("View"); field.IsValid() && field.Kind() == reflect.String {
			field.SetString(ViewDiagnosticLabel(program, field.String()))
		}
		// Checked callable dispatch can consume a Property directly instead
		// of routing its callee through the backend's ordinary value hook.
		if closure := result.FieldByName("Closure"); closure.IsValid() && closure.Kind() == reflect.Interface && !closure.IsNil() && closure.Elem().Type() == reflect.TypeOf(Property{}) {
			closure.Set(decorate(closure.Elem()))
		}
		for _, name := range []string{"ViewRead", "DictionaryRead"} {
			if field := result.FieldByName(name); field.IsValid() {
				field.Set(decorate(field))
			}
		}
		return result
	}
	return decorate(reflect.ValueOf(expression)).Interface().(Expression)
}
