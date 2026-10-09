package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

// Finite Partial<Record<K,V>> values use the dictionary from their creation.
// Optionality describes entry presence, not an undefined value in every slot.
// In particular, GetNonMissingTypeOfSymbol preserves explicit undefined in V.
func (l *lowering) finitePartialRecordElement(t *checker.Type) *checker.Type {
	if t == nil || t.Flags()&checker.TypeFlagsObject == 0 {
		return nil
	}
	partial := l.recordUtilityAlias(t, "Partial")
	if partial == nil {
		return nil
	}
	alias := partial.Alias()
	if alias == nil || alias.Symbol().Name != "Partial" || !l.librarySymbol(alias.Symbol()) || len(alias.TypeArguments()) != 1 {
		return nil
	}
	base := l.concrete(alias.TypeArguments()[0])
	base = l.recordUtilityAlias(base, "Record")
	if base == nil {
		return nil
	}
	record := base.Alias()
	if record == nil || record.Symbol().Name != "Record" || !l.librarySymbol(record.Symbol()) || len(record.TypeArguments()) != 2 {
		return nil
	}
	keys := l.concrete(record.TypeArguments()[0])
	members := []*checker.Type{keys}
	if keys.Flags()&checker.TypeFlagsUnion != 0 {
		members = keys.Types()
	}
	for _, key := range members {
		if key.Flags()&(checker.TypeFlagsStringLiteral|checker.TypeFlagsNumberLiteral) == 0 {
			return nil
		}
	}
	properties := l.checker.GetPropertiesOfType(t)
	if len(properties) == 0 || len(l.checker.GetIndexInfosOfType(t)) != 0 || len(l.checker.GetSignaturesOfType(t, checker.SignatureKindCall)) != 0 || len(l.checker.GetSignaturesOfType(t, checker.SignatureKindConstruct)) != 0 {
		return nil
	}
	value := l.concrete(record.TypeArguments()[1])
	for _, property := range properties {
		if property.Flags&ast.SymbolFlagsOptional == 0 || l.checker.IsReadonlySymbol(property) || l.concrete(l.checker.GetNonMissingTypeOfSymbol(property)) != value {
			return nil
		}
	}
	return value
}

// Resolve program aliases through the checker, retaining the library origin proof.
// Generic alias arguments must be substituted before inspecting the mapped type.
func (l *lowering) recordUtilityAlias(t *checker.Type, name string) *checker.Type {
	seen := map[*checker.Type]bool{}
	for t != nil && !seen[t] {
		seen[t] = true
		alias := t.Alias()
		if alias == nil {
			return nil
		}
		symbol := alias.Symbol()
		if symbol.Name == name && l.librarySymbol(symbol) {
			return t
		}
		if len(symbol.Declarations) != 1 || symbol.Declarations[0].Kind != ast.KindTypeAliasDeclaration {
			return nil
		}
		declaration := symbol.Declarations[0].AsTypeAliasDeclaration()
		next := l.checker.GetTypeAtLocation(declaration.Type)
		parameters := symbol.Declarations[0].TypeParameters()
		arguments := alias.TypeArguments()
		if len(parameters) != len(arguments) {
			return nil
		}
		if len(parameters) != 0 {
			sources := make([]*checker.Type, len(parameters))
			for i, parameter := range parameters {
				sources[i] = l.checker.GetTypeAtLocation(parameter.Name())
			}
			next = instantiateType(l.checker, next, newTypeMapper(sources, arguments))
		}
		t = l.concrete(next)
	}
	return nil
}
