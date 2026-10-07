package lower

import (
	"fmt"
	"sync"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/microsoft/TypeScript/tsc/shim/compiler"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/system-inc/adamic/internal/load"
)

const loadTimeReadRule = "nexus/correctness-no-import-cycle-load-time-read"

// Finding is the public runner's diagnostic shape. Keep the adapter boundary here.
type Finding struct {
	Rule, File   string
	Line, Column int
	Message      string
}

type ruleRunner interface {
	RunRule(program *compiler.Program, checker *checker.Checker, files []*ast.SourceFile, ruleName string) ([]Finding, error)
}

// TODO #xjdce2d: replace this implementation with the public cohere runner once its
// package path lands. Never copy cohere's internal rule. Its current rule does not
// follow calls, so verify indirect load-time reads before removing the interim guard.
var loadTimeReadRunner ruleRunner = conservativeLoadTimeReads{}
var loadTimeReadLock sync.Mutex

func (l *lowering) loadTimeReads(files []*ast.SourceFile) error {
	// The public runner caches one program's graph process-wide. Calls must be serial.
	loadTimeReadLock.Lock()
	defer loadTimeReadLock.Unlock()
	findings, err := loadTimeReadRunner.RunRule(l.program.CompilerProgram(), l.checker, files, loadTimeReadRule)
	if err != nil {
		return fmt.Errorf("lower: import-cycle rule: %w", err)
	}
	if len(findings) == 0 {
		return nil
	}
	finding := findings[0]
	for _, file := range files {
		if file.FileName() == finding.File {
			finding.File = l.program.FileName(file)
			break
		}
	}
	return &Refused{
		Where: fmt.Sprintf("%s:%d:%d", finding.File, finding.Line, finding.Column),
		What:  finding.Message + " [" + finding.Rule + "]",
		Fix:   "defer the read until module evaluation finishes or break the import cycle (#xjdce2d)",
	}
}

type conservativeLoadTimeReads struct{}

// The interim deliberately checks the whole reachable cyclic program, not just one
// component. Every imported value used at top level is refused, including functions
// and already initialized exporters. Deferred imports are allowed only when module
// bodies cannot invoke user code: declarations and literal console output. This
// closes indirect reads through local calls, callbacks, getters and coercions without
// pretending to have cohere's analysis. It is intentionally more restrictive.
func (conservativeLoadTimeReads) RunRule(program *compiler.Program, typeChecker *checker.Checker, files []*ast.SourceFile, ruleName string) ([]Finding, error) {
	if ruleName != loadTimeReadRule {
		panic("lower: unsupported load-time-read rule " + ruleName)
	}
	var findings []Finding
	add := func(file *ast.SourceFile, node *ast.Node, message string) {
		line, column := scanner.GetLineAndCharacterOfPosition(file, scanner.GetTokenPosOfNode(node, file, false))
		findings = append(findings, Finding{Rule: ruleName, File: file.FileName(), Line: line + 1, Column: column + 1, Message: message})
	}
	deferredImport := false
	var visit func(*ast.Node, bool, *ast.SourceFile) bool
	visit = func(node *ast.Node, deferred bool, file *ast.SourceFile) bool {
		if node == nil {
			return false
		}
		switch node.Kind {
		case ast.KindImportDeclaration, ast.KindExportDeclaration, ast.KindInterfaceDeclaration, ast.KindTypeAliasDeclaration:
			return false
		case ast.KindHeritageClause:
			if node.AsHeritageClause().Token != ast.KindExtendsKeyword || !ast.IsClassLike(node.Parent) {
				return false
			}
			for _, element := range node.AsHeritageClause().Types.Nodes {
				visit(element.AsExpressionWithTypeArguments().Expression, deferred, file)
			}
			return false
		case ast.KindExpressionWithTypeArguments:
			visit(node.AsExpressionWithTypeArguments().Expression, deferred, file)
			return false
		}
		if ast.IsTypeNode(node) {
			return false
		}
		if ast.IsFunctionLike(node) {
			deferred = true
		}
		if node.Kind == ast.KindPropertyDeclaration && !ast.HasStaticModifier(node) {
			deferred = true
		}
		if ast.IsIdentifier(node) {
			symbol := typeChecker.GetSymbolAtLocation(node)
			if parent := node.Parent; parent != nil && ast.IsShorthandPropertyAssignment(parent) {
				symbol = typeChecker.GetShorthandAssignmentValueSymbol(parent)
			}
			if symbol != nil && symbol.Flags&ast.SymbolFlagsAlias != 0 {
				for _, declaration := range symbol.Declarations {
					switch declaration.Kind {
					case ast.KindImportSpecifier, ast.KindImportClause, ast.KindNamespaceImport:
						if deferred {
							deferredImport = true
						} else {
							add(file, node, "a conservative refusal of an imported binding read during cyclic module evaluation")
						}
					}
				}
			}
		}
		node.ForEachChild(func(child *ast.Node) bool { return visit(child, deferred, file) })
		return false
	}
	for _, file := range files {
		visit(file.AsNode(), false, file)
	}
	if deferredImport {
		for _, file := range files {
			for _, statement := range file.Statements.Nodes {
				if !inertModuleStatement(typeChecker, statement) {
					add(file, statement, "module evaluation may invoke code that reads an import in a cyclic program (conservative interim)")
				}
			}
		}
	}
	_ = program // The real runner uses program identity and its module resolver.
	return findings, nil
}

func literalModuleValue(node *ast.Node) bool {
	if node == nil {
		return true
	}
	switch node.Kind {
	case ast.KindStringLiteral, ast.KindNoSubstitutionTemplateLiteral, ast.KindNumericLiteral, ast.KindTrueKeyword, ast.KindFalseKeyword, ast.KindNullKeyword:
		return true
	}
	return false
}

func inertModuleStatement(typeChecker *checker.Checker, node *ast.Node) bool {
	switch node.Kind {
	case ast.KindImportDeclaration, ast.KindExportDeclaration, ast.KindFunctionDeclaration, ast.KindInterfaceDeclaration, ast.KindTypeAliasDeclaration, ast.KindEmptyStatement:
		return true
	case ast.KindVariableStatement:
		for _, declaration := range node.AsVariableStatement().DeclarationList.AsVariableDeclarationList().Declarations.Nodes {
			if !ast.IsIdentifier(declaration.Name()) || !literalModuleValue(declaration.AsVariableDeclaration().Initializer) {
				return false
			}
		}
		return true
	case ast.KindClassDeclaration:
		// Reject executable class shapes until the public analysis is available.
		safe := true
		node.ForEachChild(func(child *ast.Node) bool {
			if child.Kind == ast.KindHeritageClause || child.Kind == ast.KindDecorator {
				safe = false
			}
			return false
		})
		for _, member := range node.AsClassDeclaration().Members.Nodes {
			if member.Name() != nil && member.Name().Kind == ast.KindComputedPropertyName {
				safe = false
			}
			if member.Kind == ast.KindClassStaticBlockDeclaration {
				safe = false
			}
			if member.Kind == ast.KindPropertyDeclaration && ast.HasStaticModifier(member) && !literalModuleValue(member.AsPropertyDeclaration().Initializer) {
				safe = false
			}
		}
		return safe
	case ast.KindExpressionStatement:
		expression := node.AsExpressionStatement().Expression
		if expression.Kind != ast.KindCallExpression {
			return literalModuleValue(expression)
		}
		call := expression.AsCallExpression()
		if call.Expression.Kind != ast.KindPropertyAccessExpression {
			return false
		}
		access := call.Expression.AsPropertyAccessExpression()
		if !ast.IsIdentifier(access.Expression) || access.Expression.Text() != "console" || (access.Name().Text() != "log" && access.Name().Text() != "error") {
			return false
		}
		symbol := typeChecker.GetSymbolAtLocation(access.Expression)
		if symbol == nil || len(symbol.Declarations) == 0 || !load.IsPrelude(ast.GetSourceFileOfNode(symbol.Declarations[0])) {
			return false
		}
		return len(call.Arguments.Nodes) == 1 && literalModuleValue(call.Arguments.Nodes[0])
	}
	return false
}
