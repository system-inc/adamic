package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// Stringification is safe only after a complete primitive/nullish dictionary
// selector. Object, array and callable ToPrimitive obligations remain refused.
func (l *lowering) primitiveDictionaryStringValue(value ir.Expression) bool {
	read, ok := value.(ir.Property)
	if call, record := value.(ir.RecordCall); record && call.Method == "get" && call.DictionaryRead != nil {
		read, ok = *call.DictionaryRead, true
	}
	if !ok || read.DictionaryKey == nil || read.Of != ir.Union {
		return false
	}
	kinds, complete := ir.DictionaryReadKinds(l.result, read.ViewContract)
	if !complete {
		return false
	}
	for _, kind := range kinds {
		switch kind {
		case "string", "number", "boolean", "null", "undefined":
		default:
			return false
		}
	}
	return len(kinds) != 0
}

// Restrict this adapter to rich dictionaries whose declared union has complete
// open scalar alternatives. Unknown/any/generic arms cannot authorize selection.
func (l *lowering) primitiveDictionaryCandidate(target *checker.Type) bool {
	if target.Flags()&checker.TypeFlagsUnion == 0 {
		return false
	}
	var number, stringType, boolean, reference bool
	for _, member := range target.Types() {
		flags := member.Flags()
		if flags&(checker.TypeFlagsAny|checker.TypeFlagsUnknown|checker.TypeFlagsTypeParameter) != 0 {
			return false
		}
		number = number || flags&checker.TypeFlagsNumber != 0
		stringType = stringType || flags&checker.TypeFlagsString != 0
		boolean = boolean || flags&(checker.TypeFlagsBoolean|checker.TypeFlagsBooleanLiteral) != 0
		reference = reference || flags&checker.TypeFlagsObject != 0
	}
	return number && stringType && boolean && reference
}

func (l *lowering) dictionaryPrimitiveReadContract(node *ast.Node, target *checker.Type) (ir.ViewContractID, bool) {
	if !l.primitiveDictionaryCandidate(target) {
		return 0, false
	}
	selected := ir.ViewContract{Kind: ir.ViewUnion, Of: ir.Union, Name: l.checker.TypeToString(target)}
	for _, member := range target.Types() {
		if member.Flags()&(checker.TypeFlagsStringLike|checker.TypeFlagsNumberLike|checker.TypeFlagsBooleanLike|checker.TypeFlagsNull|checker.TypeFlagsUndefined) == 0 {
			continue
		}
		id, err := l.viewContract(node, member)
		if err != nil {
			return 0, false
		}
		selected.Members = append(selected.Members, id)
	}
	id := ir.ViewContractID(len(l.result.ViewContracts) + 1)
	l.result.ViewContracts = append(l.result.ViewContracts, selected)
	if !ir.PrimitiveDictionaryReadCertificate(l.result, ir.Property{DictionaryPrimitive: true, DictionaryKey: ir.Undefined{}, ViewContract: id}) {
		l.result.ViewContracts = l.result.ViewContracts[:len(l.result.ViewContracts)-1]
		return 0, false
	}
	return id, true
}
