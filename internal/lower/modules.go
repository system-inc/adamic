// Functions and declarations moved unchanged from lower.go.
package lower

import (
	"slices"
	"sort"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
)

// moduleOrder is the order the program's modules run in, ECMAScript's: each module's imports first,
// depth-first in the order they're written, then the module itself. The prelude's 'adamic' module
// has no body to run. A module already in progress is the cycle back edge and is skipped.
//
// It's answered once per lowering and entry (moduleOrdered), its error with it: 25 callers ask, and
// plainReceiver asks once a receiver, which on a cyclic program reran proveModuleReads over the whole
// program each time (#20pb346: tsc's entry census 776 to 1,324 s, dumpTypes 5 to 106 s). The answer
// is the same each time for one program and entry: ESMModuleOrder and dynamicImportGap read only the
// program's syntax and the checker's symbols. Its side effects, cyclicModules and provenModuleReads,
// are made by the first answer and written nowhere else, so a held answer leaves them as they were.
// A different entry is answered afresh and replaces the held one.
func (lowering *lowering) moduleOrder(entry *ast.SourceFile) ([]*ast.SourceFile, error) {
	if held := lowering.moduleOrdered; held != nil && held.entry == entry {
		return slices.Clone(held.order), held.err
	}
	order, orderModulesError := lowering.orderModules(entry)
	lowering.moduleOrdered = &moduleOrdering{entry: entry, order: order, err: orderModulesError}
	return slices.Clone(order), orderModulesError
}

// moduleOrdering is moduleOrder's answer for one entry: the order, or the error that refused it.
type moduleOrdering struct {
	entry *ast.SourceFile
	order []*ast.SourceFile
	err   error
}

// orderModules computes moduleOrder's answer and makes its side effects; moduleOrder holds it.
func (lowering *lowering) orderModules(entry *ast.SourceFile) ([]*ast.SourceFile, error) {
	lowering.moduleOrders++
	order, cyclic := ESMModuleOrder(lowering.checker, entry)
	// Dynamic imports are not lowered. Refuse them before imported types can
	// request source identities during declaration or initializer lowering.
	for _, module := range order {
		if dynamicImportGapError := lowering.dynamicImportGap(module); dynamicImportGapError != nil {
			return nil, dynamicImportGapError
		}
	}
	lowering.cyclicModules = cyclic
	if cyclic {
		lowering.proveModuleReads(order)
	}
	return order, nil
}

// ESMModuleOrder is shared with the source-ledger audit (cycleledger) so it measures the
// compiler's actual scheduling proof, rather than replaying observed Node order.
func ESMModuleOrder(typeChecker *checker.Checker, entry *ast.SourceFile) ([]*ast.SourceFile, bool) {
	order := []*ast.SourceFile{}
	state := map[*ast.SourceFile]int{} // 1 while its imports are being visited, 2 once placed
	cyclic := false
	var visit func(module *ast.SourceFile)
	visit = func(module *ast.SourceFile) {
		switch state[module] {
		case 1:
			cyclic = true
			return
		case 2:
			return
		}
		state[module] = 1
		for _, statement := range module.Statements.Nodes {
			if statement.Kind != ast.KindImportDeclaration && statement.Kind != ast.KindExportDeclaration {
				continue
			}
			if !runtimeModuleStatement(statement) {
				continue
			}
			specifier := statement.ModuleSpecifier()
			if specifier == nil {
				continue
			}
			target := typeChecker.GetSymbolAtLocation(specifier)
			if target == nil || len(target.Declarations) == 0 || target.Declarations[0].Kind != ast.KindSourceFile {
				// 'adamic' is an ambient module in the prelude: nothing to run.
				continue
			}
			imported := target.Declarations[0].AsSourceFile()
			if load.IsPrelude(imported) || load.IsLibrary(imported) {
				continue
			}
			visit(imported)
		}
		state[module] = 2
		order = append(order, module)
	}
	visit(entry)
	return order, cyclic
}

// declareModule registers the module's globals and functions before anything is lowered, so a
// function can call one declared below it and read a global declared below it, as JavaScript
// allows. Every function's signature is written, from the checker, before any body is lowered, so a
// call knows what the function it calls returns wherever that function is declared, and two
// functions can call each other. Then each body is lowered.
func (lowering *lowering) declareModule(statements []*ast.Node) error {
	statements = namespaceDeclarations(statements)
	declarations := []*ast.Node{}
	for _, statement := range statements {
		switch statement.Kind {
		case ast.KindVariableStatement:
			if lowering.hostBindingDeclared(statement) {
				continue // the host's names are the host's (host_bindings.go)
			}
			list := statement.AsVariableStatement().DeclarationList
			if list.Flags&ast.NodeFlagsBlockScoped == 0 && !assertionVarList(list) && !namespaceVariable(list) {
				continue // statements() refuses var where it stands
			}
			for _, declaration := range list.AsVariableDeclarationList().Declarations.Nodes {
				// A module's const [a, b] = tuple, or { x, y } = object, declares globals too, each name
				// its own.
				if declaration.Name().Kind == ast.KindArrayBindingPattern ||
					declaration.Name().Kind == ast.KindObjectBindingPattern {
					for _, binding := range declaration.Name().AsBindingPattern().Elements.Nodes {
						if skipped(binding) || !ast.IsIdentifier(binding.Name()) {
							continue
						}
						local, declareLocalError := lowering.declareLocal(binding.Name())
						if declareLocalError != nil {
							return declareLocalError
						}
						lowering.result.Locals[local].Global = true
						lowering.result.Locals[local].Hoisted = list.Flags&ast.NodeFlagsBlockScoped == 0
						if namespaceVariable(list) && list.Flags&ast.NodeFlagsBlockScoped == 0 {
							lowering.result.Locals[local].NamespaceVar = true
							lowering.result.Locals[local].NamespaceState = true
						}
					}
					continue
				}
				if ast.IsIdentifier(declaration.Name()) {
					local, declareLocalError := lowering.declareLocal(declaration.Name())
					if declareLocalError != nil {
						return declareLocalError
					}
					lowering.result.Locals[local].Global = true
					lowering.result.Locals[local].Hoisted = list.Flags&ast.NodeFlagsBlockScoped == 0
					if namespaceVariable(list) {
						lowering.result.Locals[local].NamespaceVar = list.Flags&ast.NodeFlagsBlockScoped == 0
						lowering.result.Locals[local].NamespaceState = declaration.Initializer() == nil ||
							lowering.result.Locals[local].NamespaceVar
					}
				}
			}
		case ast.KindEnumDeclaration:
			// Every enum with an object is declared here, before any function body lowers: a const enum
			// a namespace cast reaches inside a function, as each of debug.ts's formatters does, reads
			// its binding before the declaring module's statements are lowered (#52zja36).
			built, enumObjectBuiltError := lowering.enumObjectBuilt(statement)
			if enumObjectBuiltError != nil {
				return enumObjectBuiltError
			}
			if built {
				local, enumLocalError := lowering.enumLocal(statement)
				if enumLocalError != nil {
					return enumLocalError
				}
				lowering.result.Locals[local].Global = true
			}
		case ast.KindClassDeclaration:
			if lowering.classes == nil {
				lowering.classes = map[*ast.Symbol]*ast.Node{}
			}
			lowering.classes[lowering.symbol(statement.Name())] = statement
			if lowering.needsStatics(statement) {
				if lowering.staticGlobals == nil {
					lowering.staticGlobals = map[*ast.Symbol]int{}
				}
				local, declareLocalError := lowering.declareLocal(statement.Name())
				if declareLocalError != nil {
					return declareLocalError
				}
				lowering.result.Locals[local].Global = true
				lowering.staticStorage(statement)
			}
		case ast.KindFunctionDeclaration:
			if generatorDeclaration(statement) && !loweredGenerator(statement) {
				// export default function* () has no name to declare it by (#qk8rztp).
				return lowering.generatorNotYet(statement, generatorForm(statement))
			}
			if statement.Body() == nil && lowering.censusImplementation(statement) != nil {
				continue
			}
			if censusOverloadsError := lowering.censusOverloads(statement); censusOverloadsError != nil {
				return censusOverloadsError
			}
			symbol := lowering.symbol(statement.Name())
			if len(statement.TypeParameters()) > 0 {
				// A generic function is lowered once per instantiation, where it's called (generic.go).
				if lowering.generics == nil {
					lowering.generics = map[*ast.Symbol]*ast.Node{}
				}
				lowering.generics[symbol] = statement
				continue
			}
			if lowering.functions == nil {
				lowering.functions = map[*ast.Symbol]int{}
			}
			lowering.functions[symbol] = len(lowering.result.Functions)
			lowering.result.Functions = append(lowering.result.Functions, ir.Function{Name: statement.Name().Text()})
			declarations = append(declarations, statement)
		}
	}
	for _, declaration := range declarations {
		if signatureError := lowering.signature(
			lowering.functions[lowering.symbol(declaration.Name())],
			declaration,
			-1,
		); signatureError != nil {
			return signatureError
		}
	}
	for _, statement := range statements {
		if statement.Kind == ast.KindClassDeclaration && lowering.needsStatics(statement) {
			if _, staticInstanceError := lowering.staticInstance(statement); staticInstanceError != nil {
				return staticInstanceError
			}
		}
	}
	for _, declaration := range declarations {
		if functionBodyError := lowering.functionBody(declaration); functionBodyError != nil {
			return functionBodyError
		}
	}
	return nil
}

// recordExports preserves the entry's public names after module flattening. Aliases and
// reexports resolve through the checker, never by matching a function's spelling.
func (lowering *lowering) recordExports(entry *ast.SourceFile) {
	exports := lowering.checker.GetExportsOfModule(entry.AsNode().Symbol())
	sort.Slice(exports, func(left, right int) bool { return exports[left].Name < exports[right].Name })
	for _, exported := range exports {
		if lowering.checker.GetTypeOnlyAliasDeclaration(exported) != nil {
			continue
		}
		target := exported
		if target.Flags&ast.SymbolFlagsAlias != 0 {
			target = lowering.checker.GetAliasedSymbol(target)
		}
		target = lowering.checker.GetExportSymbolOfSymbol(target)
		if function, ok := lowering.functions[target]; ok {
			lowering.result.Exports = append(
				lowering.result.Exports,
				ir.Export{Name: exported.Name, Function: function},
			)
		} else if local, ok := lowering.locals[target]; ok && lowering.result.Locals[local].Type == ir.Closure {
			lowering.result.Exports = append(
				lowering.result.Exports,
				ir.Export{Name: exported.Name, Function: -1, Local: local},
			)
		}
	}
}

// Explicit type imports and exports have no ESM evaluation edge. With verbatimModuleSyntax,
// even an empty named import after inline type stripping still evaluates its module.
func runtimeModuleStatement(statement *ast.Node) bool {
	if statement.Kind == ast.KindImportDeclaration {
		clause := statement.AsImportDeclaration().ImportClause
		return clause == nil || !clause.IsTypeOnly()
	}
	return !statement.AsExportDeclaration().IsTypeOnly
}
