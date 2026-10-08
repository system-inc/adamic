package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
	"strings"
)

// Lazy classification calls this named dictionary adapter. It describes read
// obligations independently of a producer or writable-slot certificate.
var viewDictionaryContractHook viewContractHook = internDictionaryViewContract

func internDictionaryViewContract(l *lowering, node *ast.Node, target *checker.Type, build viewContractBuilder) (ir.ViewContractID, error) {
	if l.checker.IsArrayType(target) || checker.IsTupleType(target) || build == nil {
		return 0, l.notYet(node, "a string-key dictionary contract")
	}
	var element *checker.Type
	for _, index := range l.checker.GetIndexInfosOfType(target) {
		if index.KeyType().Flags()&checker.TypeFlagsString != 0 {
			element = index.ValueType()
		} else {
			// Numeric/template/symbol index constraints cannot silently disappear.
			return 0, l.notYet(node, "a dictionary with a non-string index contract")
		}
	}
	if element == nil {
		element = l.finitePartialRecordElement(target)
	}
	if element == nil {
		return 0, l.notYet(node, "a dictionary without a string index contract")
	}
	if l.result.ViewContractTypes == nil {
		l.result.ViewContractTypes = map[int]ir.ViewContractID{}
	}
	if id := l.result.ViewContractTypes[int(target.Id())]; id != 0 {
		return id, nil
	}
	before := len(l.result.ViewContracts)
	id := ir.ViewContractID(before + 1)
	of := ir.Object
	if l.recordElement(target) != nil {
		of = ir.Record
	}
	contract := ir.ViewContract{Kind: ir.ViewDictionary, Of: of, Name: l.checker.TypeToString(target)}
	// Reserve before building children: recursive record/object graphs share ids.
	l.result.ViewContracts = append(l.result.ViewContracts, contract)
	l.result.ViewContractTypes[int(target.Id())] = id
	rollback := func(err error) (ir.ViewContractID, error) {
		l.result.ViewContracts = l.result.ViewContracts[:before]
		for key, child := range l.result.ViewContractTypes {
			if int(child) > before {
				delete(l.result.ViewContractTypes, key)
			}
		}
		return 0, err
	}
	var err error
	contract.Element, err = build(element)
	if err != nil {
		return rollback(err)
	}
	if contract.Element <= 0 || int(contract.Element) > len(l.result.ViewContracts) {
		return rollback(l.notYet(node, "an unavailable dictionary element contract"))
	}
	for _, property := range l.checker.GetPropertiesOfType(target) {
		child, err := build(l.checker.GetTypeOfSymbol(property))
		if err != nil {
			return rollback(err)
		}
		if child <= 0 || int(child) > len(l.result.ViewContracts) {
			return rollback(l.notYet(node, "an unavailable dictionary named field contract"))
		}
		contract.Fields = append(contract.Fields, ir.ViewFieldContract{Name: property.Name, Contract: child, Optional: property.Flags&ast.SymbolFlagsOptional != 0, Readonly: l.checker.IsReadonlySymbol(property)})
	}
	l.result.ViewContracts[id-1] = contract
	return id, nil
}

// stringDictionary selects only the ordinary string index contract. Other key
// domains remain deferred obligations and are never silently widened.
func (l *lowering) stringDictionary(target *checker.Type) bool {
	if l.regexDictionaryReceiver(target) {
		return false
	}
	if target.Flags()&checker.TypeFlagsObject == 0 || l.checker.IsArrayType(target) || checker.IsTupleType(target) {
		return false
	}
	infos := l.checker.GetIndexInfosOfType(target)
	return len(infos) == 1 && infos[0].KeyType().Flags()&checker.TypeFlagsString != 0
}

func (l *lowering) dictionaryRead(node *ast.Node, object, key ir.Expression) (ir.Expression, error) {
	if key.Type() != ir.String {
		return nil, l.notYet(node, "a dictionary key that is not a string")
	}
	declared := l.concrete(l.checker.GetTypeAtLocation(node))
	id, err := l.viewContract(node, declared)
	if err != nil {
		return nil, err
	}
	if _, ok := ir.DictionaryReadKinds(l.result, id); !ok {
		return nil, l.lazyReadRefusal(node, sourceExpression(node), "dictionary element "+l.checker.TypeToString(declared))
	}
	if l.includesUndefined(declared) && len(l.viewLiterals(declared)) != 0 && !dictionaryWholeBoolean(l.viewLiterals(declared)) {
		return nil, l.lazyReadRefusal(node, sourceExpression(node), "optional finite dictionary element")
	}
	fields, err := l.viewSchema(node, declared)
	if err != nil {
		return nil, err
	}
	if l.result.CheckedFields == nil {
		l.result.CheckedFields = map[string]bool{}
	}
	for field := range fields {
		l.result.CheckedFields[field] = true
	}
	// Actual cast origins govern demand; a record read is not itself a cast.
	of, err := l.typeOf(node)
	if err != nil {
		return nil, err
	}
	return ir.Property{Object: object, DictionaryKey: key, Of: of, View: sourceExpression(node), ViewWhere: l.program.Where(node), ViewType: l.checker.TypeToString(declared), ViewContract: id, ViewTypeID: int(declared.Id()), ViewAllowed: dictionaryReadLiterals(l.viewLiterals(declared))}, nil
}

// Resolved string-index signatures carry generic arguments that have no named
// parameter/property to infer from. Recursive dictionary graphs stay unmapped
// here until their independent binder mapping is available.
func (l *lowering) inferDictionaryIndexTypes(declared, instantiated *checker.Type, into map[*checker.Type]*checker.Type) {
	if !l.stringDictionary(declared) || !l.stringDictionary(instantiated) {
		return
	}
	wanted := l.checker.GetIndexInfosOfType(declared)[0].ValueType()
	given := l.checker.GetIndexInfosOfType(instantiated)[0].ValueType()
	if l.stringDictionary(l.checker.GetNonNullableType(wanted)) || wanted == declared {
		return
	}
	l.inferTypes(wanted, given, into)
}

// Read descriptors cannot authorize a write into a differently represented
// source slot. Until record production/write dispatch is available, refuse the
// actual write rather than reinterpret a scalar table through its target type.
func (l *lowering) dictionaryWriteRefusal(node *ast.Node) error {
	if !ast.IsAssignmentTarget(node) {
		return nil
	}
	var receiver *ast.Node
	switch node.Kind {
	case ast.KindPropertyAccessExpression:
		receiver = node.AsPropertyAccessExpression().Expression
	case ast.KindElementAccessExpression:
		receiver = node.AsElementAccessExpression().Expression
	default:
		return nil
	}
	if l.regexDictionaryReceiver(l.checker.GetTypeAtLocation(receiver)) {
		return nil
	}
	if l.processPath(receiver) == "process.env" {
		return nil
	}
	if symbol := l.checker.GetSymbolAtLocation(receiver); symbol != nil {
		for _, declaration := range symbol.Declarations {
			if declaration.Kind == ast.KindVariableDeclaration {
				initializer := declaration.AsVariableDeclaration().Initializer
				if initializer != nil && initializer.Kind == ast.KindAsExpression && l.recordElement(l.checker.GetTypeAtLocation(initializer.AsAsExpression().Expression)) == nil && l.recordElement(l.checker.GetTypeAtLocation(initializer)) != nil {
					return l.notYet(node, "a dictionary write without a producer storage certificate")
				}
			}
		}
	}
	if l.recordElement(l.checker.GetTypeAtLocation(receiver)) != nil {
		return nil
	}
	if l.stringDictionary(l.checker.GetNonNullableType(l.concrete(l.checker.GetTypeAtLocation(receiver)))) {
		return l.notYet(node, "a dictionary write without a producer storage certificate")
	}
	return nil
}

func dictionaryWholeBoolean(values []ir.Expression) bool {
	if len(values) != 2 {
		return false
	}
	a, ok := values[0].(ir.BooleanConstant)
	b, other := values[1].(ir.BooleanConstant)
	return ok && other && a.Value != b.Value
}
func dictionaryReadLiterals(values []ir.Expression) []ir.Expression {
	if dictionaryWholeBoolean(values) {
		return nil
	}
	return values
}

// Unsupported storage refuses at the producer, independently of lazy cast admission.
func (l *lowering) dictionaryProducerRefusal(node *ast.Node) error {
	if node.Kind != ast.KindObjectLiteralExpression {
		return nil
	}
	if contextual := l.checker.GetContextualType(node, checker.ContextFlagsNone); contextual != nil && contextual.Flags()&checker.TypeFlagsObject != 0 && enumObjectSymbol(contextual) == nil && len(l.checker.GetIndexInfosOfType(contextual)) != 0 && l.recordElement(contextual) == nil {
		return l.notYet(node, "a value of type "+l.checker.TypeToString(contextual)+" with an unsupported index signature or named dictionary members")
	}
	return nil
}

// RegExp named groups retain the regex lane's write refusal and storage adapter.
func (l *lowering) regexDictionaryReceiver(t *checker.Type) bool {
	for _, info := range l.checker.GetIndexInfosOfType(l.checker.GetNonNullableType(t)) {
		if declaration := info.Declaration(); declaration != nil {
			file := ast.GetSourceFileOfNode(declaration)
			if load.IsLibrary(file) && strings.Contains(string(file.AsSourceFile().FileName()), ".regexp.") {
				return true
			}
		}
	}
	return false
}

// An erased structural source can appear assignable to a dictionary because
// it declares no keys. That relation does not certify the source storage.
func (l *lowering) dictionaryCastNeedsView(source, target *checker.Type) bool {
	return source.Flags()&checker.TypeFlagsObject != 0 && target.Flags()&checker.TypeFlagsObject != 0 &&
		!isClassInstance(source) && !isClassInstance(target) && l.recordElement(source) == nil &&
		(l.stringDictionary(target) || l.finitePartialRecordElement(target) != nil) &&
		l.checker.IsTypeAssignableTo(target, source)
}
