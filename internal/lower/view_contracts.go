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

func (lowering *lowering) strictViewContract(node *ast.Node, target *checker.Type) (ir.ViewContractID, error) {
	if lowering.result.ViewContractTypes == nil {
		lowering.result.ViewContractTypes = map[int]ir.ViewContractID{}
	}
	if id := lowering.result.ViewContractTypes[int(target.Id())]; id != 0 {
		return id, nil
	}
	if target.Flags()&checker.TypeFlagsUndefined != 0 || lowering.phantomUndefined(target) {
		id := ir.ViewContractID(len(lowering.result.ViewContracts) + 1)
		lowering.result.ViewContracts = append(
			lowering.result.ViewContracts,
			ir.ViewContract{Kind: ir.ViewUndefined, Name: "undefined", Undefined: true, Of: ir.Object},
		)
		lowering.result.ViewContractTypes[int(target.Id())] = id
		return id, nil
	}
	if target.Flags()&checker.TypeFlagsNull != 0 {
		id := ir.ViewContractID(len(lowering.result.ViewContracts) + 1)
		lowering.result.ViewContracts = append(
			lowering.result.ViewContracts,
			ir.ViewContract{Kind: ir.ViewNull, Name: "null", Null: true, Of: ir.Object},
		)
		lowering.result.ViewContractTypes[int(target.Id())] = id
		return id, nil
	}
	build := func(child *checker.Type) (ir.ViewContractID, error) { return lowering.viewContract(node, child) }
	// Retain the strict adapter boundary for the array lane on compiler/views-v4.
	if target.Flags()&checker.TypeFlagsObject != 0 && (lowering.isArrayType(target) || checker.IsTupleType(target)) {
		if viewArrayContractHook == nil {
			return 0, lowering.notYet(node, "an array checked-view contract")
		}
		return viewArrayContractHook(lowering, node, target, build)
	}
	if lowering.callableViewContract(target) {
		if viewCallableContractHook == nil {
			return 0, &Refused{
				Where: lowering.program.Where(node),
				What:  "a checked view with a callable contract",
				Fix:   "prove the callable body rather than asserting its signature",
			}
		}
		return viewCallableContractHook(lowering, node, target, build)
	}

	represented := target
	if base := lowering.phantomBase(target); base != nil {
		represented = base
	}
	of, known := lowering.representation(represented)
	// representation() knows an object intersection only when every field is a scalar
	// (objectIntersection). A plain one is an object whatever its fields hold, so its contract takes
	// that from its members, each held as an object, and its fields from its refined properties.
	if !known && represented == target && lowering.plainObjectIntersection(target) {
		of, known = ir.Object, true
		for _, member := range target.Types() {
			if held, isKnown := lowering.representation(member); !isKnown || held != ir.Object {
				known = false
			}
		}
	}
	if !known {
		return 0, lowering.notYet(node, "checked-view representation for "+lowering.checker.TypeToString(target))
	}
	contract := ir.ViewContract{
		Undefined: lowering.includesUndefined(target),
		Name:      lowering.checker.TypeToString(target),
		Of:        of,
	}
	if isClassInstance(target) {
		if declaration := lowering.classNodeFor(target); declaration != nil {
			contract.Nominal = lowering.program.Where(declaration) + ":" + contract.Name
			contract.NominalBases = lowering.viewNominalBases(target, map[*checker.Type]bool{})
		}
	}
	if target.Flags()&checker.TypeFlagsUndefined != 0 || lowering.phantomUndefined(target) {
		contract.Kind = ir.ViewUndefined
	} else if interfaceScalar(represented) && of != ir.Union {
		contract.Kind = ir.ViewScalar
		contract.Allowed = lowering.viewContractLiterals(represented)
	} else {
		contract.Kind = ir.ViewObject
	}

	id := ir.ViewContractID(len(lowering.result.ViewContracts) + 1)
	lowering.result.ViewContracts = append(lowering.result.ViewContracts, contract)
	lowering.result.ViewContractTypes[int(target.Id())] = id

	if contract.Kind == ir.ViewObject || (contract.Kind == ir.ViewUnion && of == ir.Object) {
		for _, property := range lowering.checker.GetPropertiesOfType(target) {
			child, buildError := build(lowering.checker.GetTypeOfSymbol(property))
			if buildError != nil {
				return 0, buildError
			}

			contract.Fields = append(contract.Fields, ir.ViewFieldContract{
				Name:     property.Name,
				Contract: child,
				Optional: property.Flags&ast.SymbolFlagsOptional != 0,
				Readonly: lowering.checker.IsReadonlySymbol(property),
			})
		}
	}

	lowering.result.ViewContracts[int(id)-1] = contract
	return id, nil
}

func (lowering *lowering) viewContractLiterals(target *checker.Type) []ir.ViewLiteral {
	// A whole numeric enum admits numbers outside its declared members.
	if lowering.openNumericEnumType(target) {
		return nil
	}
	if target.Flags()&checker.TypeFlagsUnion != 0 {
		var allowed []ir.ViewLiteral
		for _, member := range target.Types() {
			if member.Flags()&checker.TypeFlagsUndefined != 0 {
				continue
			}
			values := lowering.viewContractLiterals(member)
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
		return []ir.ViewLiteral{{Of: ir.Boolean, Boolean: lowering.checker.TypeToString(target) == "true"}}
	}
	return nil
}

func (lowering *lowering) viewNominalBases(target *checker.Type, seen map[*checker.Type]bool) []string {
	if seen[target] {
		return nil
	}
	seen[target] = true
	var names []string
	for _, base := range lowering.classBases(target) {
		if declaration := lowering.classNodeFor(base); declaration != nil {
			names = append(names, lowering.program.Where(declaration)+":"+lowering.checker.TypeToString(base))
			names = append(names, lowering.viewNominalBases(base, seen)...)
		}
	}
	return names
}
