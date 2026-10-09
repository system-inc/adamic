package high_level_intermediate_representation

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/rule"
)

func moduleGlobalLoad(node *ast.Node, symbol *ast.Symbol) *LoadGlobal {
	binding := rule.ImportBindingOf(node, symbol)
	return &LoadGlobal{
		Name:        node.Text(),
		BindingKind: globalBindingKinds[binding.Kind],
		Source:      binding.Source,
		Imported:    binding.Imported,
	}
}

// globalBindingKinds maps how rule.ImportBindingOf found a name bound to the lowering's own kinds.
var globalBindingKinds = map[rule.ImportBindingKind]GlobalBindingKind{
	rule.ImportBindingGlobal:      GlobalBindingKindGlobal,
	rule.ImportBindingModuleLocal: GlobalBindingKindModuleLocal,
	rule.ImportBindingDefault:     GlobalBindingKindImportDefault,
	rule.ImportBindingNamespace:   GlobalBindingKindImportNamespace,
	rule.ImportBindingSpecifier:   GlobalBindingKindImportSpecifier,
}
