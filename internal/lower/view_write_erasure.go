package lower

import (
	"github.com/system-inc/adamic/internal/ir"
	"reflect"
)

// A tag fact belongs to one binding and is killed by a call or a reassignment.
// Structural discriminants alone are not runtime class/slot certificates.
func eraseTagViewWrites(program *ir.Program) {
	var rewrite func([]ir.Statement, map[int]ir.InstanceOf) []ir.Statement
	clone := func(facts map[int]ir.InstanceOf) map[int]ir.InstanceOf {
		out := map[int]ir.InstanceOf{}
		for key, value := range facts {
			out[key] = value
		}
		return out
	}
	rewrite = func(body []ir.Statement, facts map[int]ir.InstanceOf) []ir.Statement {
		for index, statement := range body {
			switch statement.(type) {
			case ir.If, ir.Loop, ir.Try, ir.Block:
			default:
				statement = eraseTagViewReads(program, statement, facts)
				body[index] = statement
			}
			switch node := statement.(type) {
			case ir.If:
				if readinessHasCalls(node.Condition) {
					clear(facts)
				}
				yes := clone(facts)
				if tag, ok := node.Condition.(ir.InstanceOf); ok {
					if receiver, ok := tag.Value.(ir.Read); ok {
						yes[receiver.Local] = tag
					}
				}
				node.Then = rewrite(node.Then, yes)
				node.Else = rewrite(node.Else, clone(facts))
				body[index] = node
			case ir.Block:
				node.Body = rewrite(node.Body, clone(facts))
				body[index] = node
			case ir.Loop:
				node.Body = rewrite(node.Body, map[int]ir.InstanceOf{})
				node.Update = rewrite(node.Update, map[int]ir.InstanceOf{})
				body[index] = node
			case ir.SetProperty:
				if node.WriteContract != 0 {
					if receiver, ok := node.Object.(ir.Read); ok {
						if tag, ok := facts[receiver.Local]; ok && tagSlotsAccept(program, tag, node) {
							node.WriteProven = true
						}
					}
					if trappedViewReceiver(program, node.Object) {
						node.WriteProven = true
					}
				}
				body[index] = node
			}
			invalidates := readinessHasCalls(statement)
			walk(statement, func(value any) bool {
				switch value.(type) {
				case ir.Call, ir.CallClosure, ir.Assign, ir.If, ir.Loop, ir.Try:
					invalidates = true
				}
				return !invalidates
			})
			if invalidates {
				clear(facts)
			}
		}
		return body
	}
	program.Main = rewrite(program.Main, map[int]ir.InstanceOf{})
	for index := range program.Functions {
		program.Functions[index].Body = rewrite(program.Functions[index].Body, map[int]ir.InstanceOf{})
	}
}

func tagSlotsAccept(program *ir.Program, tag ir.InstanceOf, write ir.SetProperty) bool {
	allowed := ir.ScalarWriteContracts(program, write.WriteContract)
	found := false
	for index, class := range program.Classes {
		id := index + 1
		matches := id == tag.Class
		if !tag.Exact {
			for base := class.Base; base != 0; base = program.Classes[base-1].Base {
				matches = matches || base == tag.Class
			}
		}
		if !matches {
			continue
		}
		found = true
		compatible := false
		for _, field := range class.Fields {
			if field.Name == write.Name {
				for _, id := range allowed {
					compatible = compatible || id == field.Contract && id != 0
				}
			}
		}
		if !compatible {
			return false
		}
	}
	return found
}

// Only the source expression itself is covered. A never nested in its logical
// type is not a trap or an unreachable proof. The generated helper must end in a
// direct panic, with no preceding control flow capable of returning a value.
func trappedViewReceiver(program *ir.Program, expression ir.Expression) bool {
	switch value := expression.(type) {
	case ir.CheckedCast:
		return trappedViewReceiver(program, value.Value)
	case ir.Defined:
		return trappedViewReceiver(program, value.Value)
	case ir.Call:
		// A virtual base body is not a proof about its overrides.
		for _, target := range program.CallTargets(value) {
			function := program.Functions[target]
			if len(function.Body) == 0 {
				return false
			}
			if _, ok := function.Body[len(function.Body)-1].(ir.Panic); !ok {
				return false
			}
			for _, statement := range function.Body[:len(function.Body)-1] {
				if _, ok := statement.(ir.Evaluate); !ok {
					return false
				}
			}
		}
		return true
	}
	return false
}

// Erase semantic read checks only when a dominating nominal tag proves both the
// slot contract and the exact physical representation. Readiness is independent.
func eraseTagViewReads(program *ir.Program, statement ir.Statement, facts map[int]ir.InstanceOf) ir.Statement {
	if readinessHasCalls(statement) {
		return statement
	}
	var transform func(reflect.Value) reflect.Value
	transform = func(value reflect.Value) reflect.Value {
		if !value.IsValid() {
			return value
		}
		switch value.Kind() {
		case reflect.Interface:
			if value.IsNil() {
				return value
			}
			mapped := transform(value.Elem())
			node := mapped.Interface()
			if property, ok := node.(ir.Property); ok && property.View != "" {
				if receiver, ok := property.Object.(ir.Read); ok {
					if tag, ok := facts[receiver.Local]; ok && tagReadSlot(program, tag, property) {
						property.View = ""
						property.ViewAllowed = nil
						node = property
					}
				}
			}
			out := reflect.New(value.Type()).Elem()
			out.Set(reflect.ValueOf(node))
			return out
		case reflect.Struct:
			out := reflect.New(value.Type()).Elem()
			for i := 0; i < value.NumField(); i++ {
				out.Field(i).Set(transform(value.Field(i)))
			}
			return out
		case reflect.Slice:
			if value.IsNil() {
				return value
			}
			out := reflect.MakeSlice(value.Type(), value.Len(), value.Len())
			for i := 0; i < value.Len(); i++ {
				out.Index(i).Set(transform(value.Index(i)))
			}
			return out
		}
		return value
	}
	return transform(reflect.ValueOf(statement)).Interface().(ir.Statement)
}

func tagReadSlot(program *ir.Program, tag ir.InstanceOf, property ir.Property) bool {
	target := property.ViewContract
	if target == 0 {
		target = program.ViewContractTypes[property.ViewTypeID]
	}
	if target == 0 {
		return false
	}
	found := false
	for index, class := range program.Classes {
		matches := index+1 == tag.Class
		if !tag.Exact {
			for base := class.Base; base != 0; base = program.Classes[base-1].Base {
				matches = matches || base == tag.Class
			}
		}
		if !matches {
			continue
		}
		found = true
		compatible := false
		for _, field := range class.Fields {
			if field.Name == property.Name && field.Value.Type() == property.Type() {
				for _, id := range ir.ScalarWriteContracts(program, field.Contract) {
					compatible = compatible || id == target
				}
			}
		}
		if !compatible {
			return false
		}
	}
	return found
}
