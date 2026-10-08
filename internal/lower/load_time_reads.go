package lower

import (
	"fmt"
	"sync"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/microsoft/TypeScript/tsc/shim/compiler"
	"github.com/system-inc/cohere/rule_runner"
)

const loadTimeReadRule = "nexus/correctness-no-import-cycle-load-time-read"

type Finding = rule_runner.Finding

type ruleRunner interface {
	RunRule(*compiler.Program, *checker.Checker, []*ast.SourceFile, string) ([]Finding, error)
}
type cohereRuleRunner struct{}

func (cohereRuleRunner) RunRule(program *compiler.Program, typeChecker *checker.Checker, files []*ast.SourceFile, name string) ([]Finding, error) {
	return rule_runner.RunRule(program, typeChecker, files, name)
}

var loadTimeReadRunner ruleRunner = cohereRuleRunner{}
var loadTimeReadLock sync.Mutex

// The rule inventories hazardous cycle reads, but a finding does not reject the
// program. The actual entry's ESM order can prove the provider finished; every
// remaining global read keeps the existing runtime dead-zone check.
func (l *lowering) loadTimeReads(files []*ast.SourceFile) error {
	loadTimeReadLock.Lock()
	defer loadTimeReadLock.Unlock()
	findings, err := loadTimeReadRunner.RunRule(l.program.CompilerProgram(), l.checker, files, loadTimeReadRule)
	if err != nil {
		return fmt.Errorf("lower: import-cycle rule: %w", err)
	}
	l.cycleReadFindings = findings
	l.proveModuleReads(files)
	return nil
}

func (l *lowering) proveModuleReads(files []*ast.SourceFile) {
	l.provenModuleReads = map[*ast.Node]bool{}
	positions := map[*ast.SourceFile]int{}
	for index, file := range files {
		positions[file] = index
	}
	for _, file := range files {
		var visit func(*ast.Node) bool
		visit = func(node *ast.Node) bool {
			if node == nil || ast.IsFunctionLike(node) {
				return false
			}
			switch node.Kind {
			case ast.KindImportDeclaration, ast.KindExportDeclaration:
				return false
			case ast.KindHeritageClause:
				if node.AsHeritageClause().Token == ast.KindExtendsKeyword && ast.IsClassLike(node.Parent) {
					for _, element := range node.AsHeritageClause().Types.Nodes {
						visit(element.AsExpressionWithTypeArguments().Expression)
					}
				}
				return false
			case ast.KindExpressionWithTypeArguments:
				visit(node.AsExpressionWithTypeArguments().Expression)
				return false
			case ast.KindPropertyDeclaration:
				if !ast.HasStaticModifier(node) {
					return false
				}
			}
			if ast.IsTypeNode(node) {
				return false
			}
			if ast.IsIdentifier(node) || ast.IsPropertyAccessExpression(node) {
				symbol := l.checker.GetSymbolAtLocation(node)
				if ast.IsIdentifier(node) && node.Parent != nil && ast.IsShorthandPropertyAssignment(node.Parent) {
					symbol = l.checker.GetShorthandAssignmentValueSymbol(node.Parent)
				}
				if symbol != nil {
					symbol = l.checker.SkipAlias(symbol)
					if symbol != nil && symbol.ValueDeclaration != nil {
						declaration := symbol.ValueDeclaration
						provider := ast.GetSourceFileOfNode(declaration)
						index, present := positions[provider]
						if (declaration.Kind == ast.KindFunctionDeclaration && declaration.Parent != nil && declaration.Parent.Kind == ast.KindSourceFile) || (present && provider != file && index < positions[file]) {
							l.provenModuleReads[node] = true
						}
					}
				}
			}
			node.ForEachChild(visit)
			return false
		}
		visit(file.AsNode())
	}
}

func (l *lowering) checkedModuleRead(node *ast.Node, local int) bool {
	// A regular enum has runtime storage even in an acyclic module. Check its
	// actual read, rather than delaying every construction until all enums exist.
	symbol := l.symbol(node)
	enum := symbol != nil && symbol.ValueDeclaration != nil && symbol.ValueDeclaration.Kind == ast.KindEnumDeclaration
	return !l.provenModuleReads[node] && (l.checked(local) || enum)
}
