package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
	"reflect"
)

// Family adapters build child descriptors through the shared lazy builder.
type viewContractBuilder func(*checker.Type) (ir.ViewContractID, error)
type viewContractHook func(*lowering, *ast.Node, *checker.Type, viewContractBuilder) (ir.ViewContractID, error)

// Reserved for the array adapter carried by compiler/views-v4.
// This base routes array descriptors through unionAggregateContract first.
var viewArrayContractHook viewContractHook

// Reserved attachment point for V4 standalone callable contracts.
// Current callable obligations are recorded lazily before strict construction.
var viewCallableContractHook viewContractHook

func (l *lowering) strictViewContract(node *ast.Node, target *checker.Type) (ir.ViewContractID, error) {
	if l.result.ViewContractTypes == nil {
		l.result.ViewContractTypes = map[int]ir.ViewContractID{}
	}
	if id := l.result.ViewContractTypes[int(target.Id())]; id != 0 {
		return id, nil
	}
	if target.Flags()&checker.TypeFlagsUndefined != 0 || l.phantomUndefined(target) {
		id := ir.ViewContractID(len(l.result.ViewContracts) + 1)
		l.result.ViewContracts = append(l.result.ViewContracts, ir.ViewContract{Kind: ir.ViewUndefined, Name: "undefined", Undefined: true, Of: ir.Object})
		l.result.ViewContractTypes[int(target.Id())] = id
		return id, nil
	}
	if target.Flags()&checker.TypeFlagsNull != 0 {
		id := ir.ViewContractID(len(l.result.ViewContracts) + 1)
		l.result.ViewContracts = append(l.result.ViewContracts, ir.ViewContract{Kind: ir.ViewNull, Name: "null", Null: true, Of: ir.Object})
		l.result.ViewContractTypes[int(target.Id())] = id
		return id, nil
	}
	if target.Flags()&checker.TypeFlagsNonPrimitive != 0 {
		id := ir.ViewContractID(len(l.result.ViewContracts) + 1)
		l.result.ViewContractTypes[int(target.Id())] = id
		l.result.ViewContracts = append(l.result.ViewContracts, ir.ViewContract{Kind: ir.ViewObject, Name: "object", Of: ir.Union})
		return id, nil
	}
	build := func(child *checker.Type) (ir.ViewContractID, error) { return l.viewContract(node, child) }
	// Retain the strict adapter boundary for the array lane on compiler/views-v4.
	if target.Flags()&checker.TypeFlagsObject != 0 && (l.isArrayType(target) || checker.IsTupleType(target)) {
		if viewArrayContractHook == nil {
			return 0, l.notYet(node, "an array checked-view contract")
		}
		return viewArrayContractHook(l, node, target, build)
	}
	if l.callableViewContract(target) {
		if viewCallableContractHook == nil {
			return 0, &Refused{Where: l.program.Where(node), What: "a checked view with a callable contract", Fix: "prove the callable body rather than asserting its signature"}
		}
		return viewCallableContractHook(l, node, target, build)
	}

	represented := target
	if base := l.phantomBase(target); base != nil {
		represented = base
	}
	of, known := l.representation(represented)
	// representation() knows an object intersection only when every field is a scalar
	// (objectIntersection). A plain one is an object whatever its fields hold, so its contract takes
	// that from its members, each held as an object, and its fields from its refined properties.
	if !known && represented == target && l.plainObjectIntersection(target) {
		of, known = ir.Object, true
		for _, member := range target.Types() {
			if held, isKnown := l.representation(member); !isKnown || held != ir.Object {
				known = false
			}
		}
	}
	if !known {
		return 0, l.notYet(node, "checked-view representation for "+l.checker.TypeToString(target))
	}
	contract := ir.ViewContract{Undefined: l.includesUndefined(target), Name: l.checker.TypeToString(target), Of: of}
	if isClassInstance(target) {
		if declaration := l.classNodeFor(target); declaration != nil {
			contract.Nominal = l.program.Where(declaration) + ":" + contract.Name
			contract.NominalBases = l.viewNominalBases(target, map[*checker.Type]bool{})
		}
	}
	if target.Flags()&checker.TypeFlagsUndefined != 0 || l.phantomUndefined(target) {
		contract.Kind = ir.ViewUndefined
	} else if interfaceScalar(represented) && of != ir.Union {
		contract.Kind = ir.ViewScalar
		contract.Allowed = l.viewContractLiterals(represented)
	} else {
		contract.Kind = ir.ViewObject
	}

	id := ir.ViewContractID(len(l.result.ViewContracts) + 1)
	l.result.ViewContracts = append(l.result.ViewContracts, contract)
	l.result.ViewContractTypes[int(target.Id())] = id

	if contract.Kind == ir.ViewObject || (contract.Kind == ir.ViewUnion && of == ir.Object) {
		for _, property := range l.checker.GetPropertiesOfType(target) {
			child, err := build(l.checker.GetTypeOfSymbol(property))
			if err != nil {
				return 0, err
			}

			contract.Fields = append(contract.Fields, ir.ViewFieldContract{Name: property.Name, Contract: child, Optional: property.Flags&ast.SymbolFlagsOptional != 0, Readonly: l.checker.IsReadonlySymbol(property)})
		}
	}

	l.result.ViewContracts[int(id)-1] = contract
	return id, nil
}

// viewStaticsAdmitted is whether a class constructor may stand for contract's object (#s03ndgw). A field
// Object's prototype names (toString, constructor and the rest, and the __proto__ family the library
// doesn't declare, objectPrototypeName) leaves it false. So does an optional field Function's interface
// names (call, apply, bind, name, length, prototype and the rest), asked of the checker by name with
// GetPropertyOfType, which finds Object's members too: a class that doesn't declare it reads Function's
// on Node, where an absent optional field would read undefined. A required one is admitted: the run time
// walk (view_unions_untagged.c's static_own_slot, readiness.go's adamicViewStaticField) finds it only as
// a data static a class on the chain wrote, and name, length and prototype only as the class's own, so
// Function's own never stands in for it. A field whose finished contract is missing, unknown or
// unsupported, an unsupported contract, one with no field, or a library without Function to ask leaves
// it false.
func (l *lowering) viewStaticsAdmitted(contract ir.ViewContract) bool {
	function := l.checker.GetGlobalSymbol("Function", ast.SymbolFlagsInterface, nil)
	if function == nil || contract.Unsupported != "" || len(contract.Fields) == 0 {
		return false
	}
	functionType := l.checker.GetDeclaredTypeOfSymbol(function)
	for _, field := range contract.Fields {
		if field.Contract <= 0 || int(field.Contract) > len(l.result.ViewContracts) {
			return false
		}
		if held := l.result.ViewContracts[field.Contract-1]; held.Kind == ir.ViewUnknown || held.Unsupported != "" {
			return false
		}
		if l.objectPrototypeName(field.Name) || field.Optional && l.checker.GetPropertyOfType(functionType, field.Name) != nil {
			return false
		}
	}
	return true
}

// finishViewStaticsAdmitted sets StaticsAdmitted on every contract once the table is finished, after
// lowering's last pass, so viewStaticsAdmitted reads each field's settled contract. Set inside a
// contract's own field loop, a field whose contract is a back edge to an ancestor still being built
// would read that ancestor's placeholder, and stay true once the ancestor was made Unknown or
// Unsupported. It's an object's (strictViewContract) or an object union's (buildUnionReadContract,
// which StaticsAdmittedReceiver reads for a write), never a fixed tuple's, whose positions no class
// constructor holds.
func (l *lowering) finishViewStaticsAdmitted() {
	for index, contract := range l.result.ViewContracts {
		object := contract.Kind == ir.ViewObject && !contract.FixedTuple
		union := contract.Kind == ir.ViewUnion && contract.Of == ir.Object
		l.result.ViewContracts[index].StaticsAdmitted = (object || union) && l.viewStaticsAdmitted(contract)
	}
}

func (l *lowering) viewContractLiterals(target *checker.Type) []ir.ViewLiteral {
	// A whole numeric enum admits numbers outside its declared members, a one-member enum's included.
	if l.tagHoldsAnyNumber(target) {
		return nil
	}
	if target.Flags()&checker.TypeFlagsUnion != 0 {
		var allowed []ir.ViewLiteral
		for _, member := range target.Types() {
			if member.Flags()&checker.TypeFlagsUndefined != 0 {
				continue
			}
			values := l.viewContractLiterals(member)
			if len(values) == 0 {
				return nil
			}
			allowed = append(allowed, values...)
		}
		return allowed
	}
	flags := target.Flags()
	switch {
	case flags&checker.TypeFlagsStringLiteral != 0:
		value, ok := target.AsLiteralType().Value().(string)
		if ok {
			return []ir.ViewLiteral{{Of: ir.String, String: value}}
		}
	case flags&checker.TypeFlagsNumberLiteral != 0:
		value := reflect.ValueOf(target.AsLiteralType().Value())
		if value.Kind() == reflect.Float64 {
			return []ir.ViewLiteral{{Of: ir.Number, Number: value.Float()}}
		}
	case flags&checker.TypeFlagsBooleanLiteral != 0:
		return []ir.ViewLiteral{{Of: ir.Boolean, Boolean: l.checker.TypeToString(target) == "true"}}
	}
	return nil
}

func (l *lowering) viewNominalBases(target *checker.Type, seen map[*checker.Type]bool) []string {
	if seen[target] {
		return nil
	}
	seen[target] = true
	var names []string
	for _, base := range l.classBases(target) {
		if declaration := l.classNodeFor(base); declaration != nil {
			names = append(names, l.program.Where(declaration)+":"+l.checker.TypeToString(base))
			names = append(names, l.viewNominalBases(base, seen)...)
		}
	}
	return names
}
