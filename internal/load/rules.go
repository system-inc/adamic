package load

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/cohere/rule_runner"
)

// RuleFindings uses the compiler and checker already held by the caller. The caller
// must keep the checker acquired until the rule has finished.
func (p *Program) RuleFindings(typeChecker *checker.Checker, files []*ast.SourceFile, name string) ([]rule_runner.Finding, error) {
	return rule_runner.RunRule(p.compiler, typeChecker, files, name)
}
