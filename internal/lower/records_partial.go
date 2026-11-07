package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

// Finite Partial<Record<K,V>> values use the dictionary from their creation.
// Optionality describes entry presence, not an undefined value in every slot.
// In particular, GetNonMissingTypeOfSymbol preserves explicit undefined in V.
func (lowering *lowering) finitePartialRecordElement(t *checker.Type) *checker.Type {
	if t == nil || t.Flags()&checker.TypeFlagsObject == 0 {
		return nil
	}
	partial := lowering.recordUtilityAlias(t, "Partial")
	if partial == nil {
		return nil
	}
	alias := partial.Alias()
	if alias == nil ||
		alias.Symbol().Name != "Partial" ||
		!lowering.librarySymbol(alias.Symbol()) ||
		len(alias.TypeArguments()) != 1 {
		return nil
	}
	base := lowering.concrete(alias.TypeArguments()[0])
	base = lowering.recordUtilityAlias(base, "Record")
	if base == nil {
		return nil
	}
	record := base.Alias()
	if record == nil ||
		record.Symbol().Name != "Record" ||
		!lowering.librarySymbol(record.Symbol()) ||
		len(record.TypeArguments()) != 2 {
		return nil
	}
	keys := lowering.concrete(record.TypeArguments()[0])
	members := []*checker.Type{keys}
	if keys.Flags()&checker.TypeFlagsUnion != 0 {
		members = keys.Types()
	}
	for _, key := range members {
		if key.Flags()&(checker.TypeFlagsStringLiteral|checker.TypeFlagsNumberLiteral) == 0 {
			return nil
		}
	}
	properties := lowering.checker.GetPropertiesOfType(t)
	if len(properties) == 0 ||
		len(lowering.checker.GetIndexInfosOfType(t)) != 0 ||
		len(lowering.checker.GetSignaturesOfType(t, checker.SignatureKindCall)) != 0 ||
		len(lowering.checker.GetSignaturesOfType(t, checker.SignatureKindConstruct)) != 0 {
		return nil
	}
	value := lowering.concrete(record.TypeArguments()[1])
	for _, property := range properties {
		if property.Flags&ast.SymbolFlagsOptional == 0 ||
			lowering.checker.IsReadonlySymbol(property) ||
			lowering.concrete(lowering.checker.GetNonMissingTypeOfSymbol(property)) != value {
			return nil
		}
	}
	return value
}

// pickedRecordBase is the record a Pick of its literal keys is stored as: the view adds no storage, so
// its picked fields are the record's named slots, a record passes into it unchanged, and anything
// stored as the Pick is stored as the whole record, its named contracts included.
func (lowering *lowering) pickedRecordBase(t *checker.Type) *checker.Type {
	if t == nil || t.Flags()&checker.TypeFlagsObject == 0 {
		return nil
	}
	picked := lowering.recordUtilityAlias(t, "Pick")
	if picked == nil {
		return nil
	}
	alias := picked.Alias()
	if alias == nil ||
		alias.Symbol().Name != "Pick" ||
		!lowering.librarySymbol(alias.Symbol()) ||
		len(alias.TypeArguments()) != 2 {
		return nil
	}
	keys := lowering.concrete(alias.TypeArguments()[1])
	members := []*checker.Type{keys}
	if keys.Flags()&checker.TypeFlagsUnion != 0 {
		members = keys.Types()
	}
	for _, key := range members {
		if key.Flags()&checker.TypeFlagsStringLiteral == 0 {
			return nil
		}
	}
	base := lowering.concrete(alias.TypeArguments()[0])
	if lowering.recordElement(base) == nil {
		return nil
	}
	return base
}

// Resolve program aliases through the checker, retaining the library origin proof.
// Generic alias arguments must be substituted before inspecting the mapped type.
func (lowering *lowering) recordUtilityAlias(t *checker.Type, name string) *checker.Type {
	seen := map[*checker.Type]bool{}
	for t != nil && !seen[t] {
		seen[t] = true
		alias := t.Alias()
		if alias == nil {
			return nil
		}
		symbol := alias.Symbol()
		if symbol.Name == name && lowering.librarySymbol(symbol) {
			return t
		}
		if len(symbol.Declarations) != 1 || symbol.Declarations[0].Kind != ast.KindTypeAliasDeclaration {
			return nil
		}
		declaration := symbol.Declarations[0].AsTypeAliasDeclaration()
		next := lowering.checker.GetTypeAtLocation(declaration.Type)
		parameters := symbol.Declarations[0].TypeParameters()
		arguments := alias.TypeArguments()
		if len(parameters) != len(arguments) {
			return nil
		}
		if len(parameters) != 0 {
			sources := make([]*checker.Type, len(parameters))
			for i, parameter := range parameters {
				sources[i] = lowering.checker.GetTypeAtLocation(parameter.Name())
			}
			next = instantiateType(lowering.checker, next, newTypeMapper(sources, arguments))
		}
		t = lowering.concrete(next)
	}
	return nil
}
