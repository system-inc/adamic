// Functions and declarations moved unchanged from lower.go.
package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
)

// moduleOrder is the order the program's modules run in, ECMAScript's: each module's imports first,
// depth-first in the order they're written, then the module itself. The prelude's 'adamic' module
// has no body to run. A module already in progress is the cycle back edge and is skipped.
func (l *lowering) moduleOrder(entry *ast.SourceFile) ([]*ast.SourceFile, error) {
	order, cyclic := esmModuleOrder(l.checker, entry)
	l.cyclicModules = cyclic
	if cyclic {
		if err := l.loadTimeReads(order); err != nil {
			return nil, err
		}
	}
	return order, nil
}

// esmModuleOrder is shared with the source-ledger audit so it measures the
// compiler's actual scheduling proof, rather than replaying observed Node order.
func esmModuleOrder(typeChecker *checker.Checker, entry *ast.SourceFile) ([]*ast.SourceFile, bool) {
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
func (l *lowering) declareModule(statements []*ast.Node) error {
	declarations := []*ast.Node{}
	for _, statement := range statements {
		switch statement.Kind {
		case ast.KindVariableStatement:
			list := statement.AsVariableStatement().DeclarationList
			if list.Flags&ast.NodeFlagsBlockScoped == 0 && !assertionVarList(list) {
				continue // statements() refuses var where it stands
			}
			for _, declaration := range list.AsVariableDeclarationList().Declarations.Nodes {
				if l.nodeRequireBinding(declaration) {
					continue
				}
				// A module's const [a, b] = tuple, or { x, y } = object, declares globals too, each name
				// its own.
				if declaration.Name().Kind == ast.KindArrayBindingPattern || declaration.Name().Kind == ast.KindObjectBindingPattern {
					for _, binding := range declaration.Name().AsBindingPattern().Elements.Nodes {
						if skipped(binding) || !ast.IsIdentifier(binding.Name()) {
							continue
						}
						local, err := l.declareLocal(binding.Name())
						if err != nil {
							return err
						}
						l.result.Locals[local].Global = true
						l.result.Locals[local].Hoisted = list.Flags&ast.NodeFlagsBlockScoped == 0
					}
					continue
				}
				if ast.IsIdentifier(declaration.Name()) {
					local, err := l.declareLocal(declaration.Name())
					if err != nil {
						return err
					}
					l.result.Locals[local].Global = true
					l.result.Locals[local].Hoisted = list.Flags&ast.NodeFlagsBlockScoped == 0
				}
			}
		case ast.KindEnumDeclaration:
			if !ast.HasSyntacticModifier(statement, ast.ModifierFlagsConst) {
				local, err := l.enumLocal(statement)
				if err != nil {
					return err
				}
				l.result.Locals[local].Global = true
			}
		case ast.KindClassDeclaration:
			if l.classes == nil {
				l.classes = map[*ast.Symbol]*ast.Node{}
			}
			l.classes[l.symbol(statement.Name())] = statement
			if l.needsStatics(statement) {
				if l.staticGlobals == nil {
					l.staticGlobals = map[*ast.Symbol]int{}
				}
				local, err := l.declareLocal(statement.Name())
				if err != nil {
					return err
				}
				l.result.Locals[local].Global = true
				l.staticStorage(statement)
			}
		case ast.KindFunctionDeclaration:
			symbol := l.symbol(statement.Name())
			if len(statement.TypeParameters()) > 0 {
				// A generic function is lowered once per instantiation, where it's called (generic.go).
				if l.generics == nil {
					l.generics = map[*ast.Symbol]*ast.Node{}
				}
				l.generics[symbol] = statement
				continue
			}
			if l.functions == nil {
				l.functions = map[*ast.Symbol]int{}
			}
			l.functions[symbol] = len(l.result.Functions)
			l.result.Functions = append(l.result.Functions, ir.Function{Name: statement.Name().Text()})
			declarations = append(declarations, statement)
		}
	}
	for _, declaration := range declarations {
		if err := l.signature(l.functions[l.symbol(declaration.Name())], declaration, -1); err != nil {
			return err
		}
	}
	for _, statement := range statements {
		if statement.Kind == ast.KindClassDeclaration && l.needsStatics(statement) {
			if _, err := l.staticInstance(statement); err != nil {
				return err
			}
		}
	}
	for _, declaration := range declarations {
		if err := l.functionBody(declaration); err != nil {
			return err
		}
	}
	return nil
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
