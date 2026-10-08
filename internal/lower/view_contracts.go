package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
	"reflect"
)

// Lane 2 registers these adapters from its new files. Until then an unsupported
// family stays NotYet/Refused, rather than silently lowering an incomplete contract.
type viewContractBuilder func(*checker.Type) (ir.ViewContractID, error)
type viewContractHook func(*lowering, *ast.Node, *checker.Type, viewContractBuilder) (ir.ViewContractID, error)

var viewArrayContractHook viewContractHook
var viewCallableContractHook viewContractHook

func (l *lowering) strictViewContract(node *ast.Node, target *checker.Type) (ir.ViewContractID, error) {
	if l.result.ViewContractTypes == nil {
		l.result.ViewContractTypes = map[int]ir.ViewContractID{}
	}
	if id := l.result.ViewContractTypes[int(target.Id())]; id != 0 {
		return id, nil
	}
	if target.Flags()&checker.TypeFlagsNull != 0 {
		id := ir.ViewContractID(len(l.result.ViewContracts) + 1)
		l.result.ViewContracts = append(l.result.ViewContracts, ir.ViewContract{Kind: ir.ViewNull, Name: "null", Null: true, Of: ir.Union})
		l.result.ViewContractTypes[int(target.Id())] = id
		return id, nil
	}
	if target.Flags()&checker.TypeFlagsUndefined != 0 {
		id := ir.ViewContractID(len(l.result.ViewContracts) + 1)
		l.result.ViewContracts = append(l.result.ViewContracts, ir.ViewContract{Kind: ir.ViewUndefined, Name: "undefined", Undefined: true, Of: ir.Object})
		l.result.ViewContractTypes[int(target.Id())] = id
		return id, nil
	}
	build := func(child *checker.Type) (ir.ViewContractID, error) { return l.viewContract(node, child) }
	// Pure phantom brands retain their existing erased contract path.
	// Real NodeArray own fields still require the array adapter.
	if l.viewArrayBase(target) != nil && l.phantomArrayBase(target) == nil || checker.IsTupleType(target) {
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
	if id, handled, err := l.viewOptionalArrayContract(node, target); handled {
		return id, err
	}
	if l.includesUndefined(target) {
		present := l.checker.GetNonNullableType(target)
		if of, known := l.representation(present); known && of == ir.Object {
			id, err := build(present)
			if err != nil {
				return 0, err
			}
			contract := l.result.ViewContracts[id-1]
			contract.Undefined = true
			contract.Name = l.checker.TypeToString(target)
			optional := ir.ViewContractID(len(l.result.ViewContracts) + 1)
			l.result.ViewContracts = append(l.result.ViewContracts, contract)
			l.result.ViewContractTypes[int(target.Id())] = optional
			return optional, nil
		}
	}
	of, known := l.representation(target)
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
	scalar := target
	if base := l.phantomBase(target); base != nil {
		scalar = base
	}
	if target.Flags()&checker.TypeFlagsUndefined != 0 {
		contract.Kind = ir.ViewUndefined
	} else if interfaceScalar(scalar) && of != ir.Union {
		contract.Kind = ir.ViewScalar
		contract.Allowed = l.viewContractLiterals(scalar)
	} else {
		contract.Kind = ir.ViewObject
	}
	// A union retains each member contract; common fields are not a certificate
	// for the other fields of any selected member.
	if target.Flags()&checker.TypeFlagsUnion != 0 && (!interfaceScalar(target) || of == ir.Union) {
		contract.Kind = ir.ViewUnion
	}
	id := ir.ViewContractID(len(l.result.ViewContracts) + 1)
	l.result.ViewContracts = append(l.result.ViewContracts, contract)
	l.result.ViewContractTypes[int(target.Id())] = id
	if contract.Kind == ir.ViewUnion {
		for _, member := range target.Types() {
			child, err := build(member)
			if err != nil {
				return 0, err
			}
			contract.Members = append(contract.Members, child)
		}
	}
	if contract.Kind == ir.ViewObject || (contract.Kind == ir.ViewUnion && of == ir.Object) {
		for _, property := range l.checker.GetPropertiesOfType(target) {
			child, err := build(l.checker.GetTypeOfSymbol(property))
			if err != nil {
				return 0, err
			}
			// Signature certification is demanded by reads, and failure is metadata
			// until the shared flow establishes a read may receive this view.
			if l.callableViewContract(l.checker.GetTypeOfSymbol(property)) {
				if err := l.viewCallableFieldUses(node, target, property); err != nil {
					l.result.ViewContracts[child-1].Unsupported = "callable"
				}
			}
			contract.Fields = append(contract.Fields, ir.ViewFieldContract{Name: property.Name, Contract: child, Optional: property.Flags&ast.SymbolFlagsOptional != 0, Readonly: l.checker.IsReadonlySymbol(property)})
		}
	}
	if contract.Kind == ir.ViewUnion && contract.Of == ir.Object {
		tagged := false
		for _, field := range contract.Fields {
			child := l.result.ViewContracts[field.Contract-1]
			tagged = tagged || !field.Optional && child.Kind == ir.ViewScalar && len(child.Allowed) != 0
		}
		if !tagged {
			contract.Unsupported = "untagged object union"
		}
	}
	l.result.ViewContracts[int(id)-1] = contract
	return id, nil
}

func (l *lowering) viewContractLiterals(target *checker.Type) []ir.ViewLiteral {
	if base := l.phantomBase(target); base != nil {
		return l.viewContractLiterals(base)
	}
	// A whole numeric enum admits numbers outside its declared members.
	if l.openNumericEnumType(target) {
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
