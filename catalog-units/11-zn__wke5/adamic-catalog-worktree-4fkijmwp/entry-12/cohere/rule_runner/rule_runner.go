// Package rule_runner runs one of cohere's rules in another program's process, over a program and checker
// that program already holds, so a rule lives once and is never copied (#xjdce2d, #s6z97k8).
//
// Adamic is the caller it was built for: its compiler holds a typescript-go Program and checker through
// cohere's shims, and asks a rule about them rather than re-implementing it in its lowering. The package is
// public, outside internal/, and moves nothing: Go allows a package under github.com/system-inc/cohere to
// import cohere's internal packages, and a module that replaces github.com/system-inc/cohere imports this one.
//
// # The allowlist
//
// Only the rules named in allowed can be run. Every rule added there becomes part of a contract another
// program builds on, so the surface grows by a decision in this file, never by a rule merely existing. A
// rule not on it is refused with NotAllowedError.
//
// # One program at a time
//
// A rule that derives a fact about the whole program caches it process-wide, keyed by the program's
// identity: correctness-no-import-cycle-load-time-read builds its import graph once per program. Calling
// RunRule for one program, then another, rebuilds it. Calling it for two programs at once from two
// goroutines keeps only one of their graphs at a time, so each call rebuilds, and while both are correct,
// neither is fast. Run one program at a time.
package rule_runner

import (
	"fmt"
	"sort"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/microsoft/TypeScript/tsc/shim/compiler"
	"github.com/system-inc/cohere/internal/lint/linter"
	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/internal/lint/rules/adamic"
	"github.com/system-inc/cohere/internal/lint/rules/nexus"
)

// allowed is every rule another program may run, by its registered name. See the package comment.
var allowed = map[string]rule.Rule{
	// For Adamic's stage 3, which accepts import cycles in module order and refuses only a read at load.
	nexus.CorrectnessNoImportCycleLoadTimeRead.Name: nexus.CorrectnessNoImportCycleLoadTimeRead,
	// For Adamic's lowering, which refused these holes in a copy of its own until cohere's rules took every one
	// that is real (#xjdce2d step 4, ruled by @system_cohere). What Adamic can't represent stays its own refusal.
	adamic.InvariantMutable.Name: adamic.InvariantMutable,
	adamic.NominalClass.Name:     adamic.NominalClass,
}

// Finding is one finding, named as cohere's own `--json` output names it: the rule, the file, the 1-based
// line and byte column of where it starts, and its message. Message is the rule's description as written,
// so it may hold a newline where a rule writes a second paragraph; cohere's printers join those into one
// line, and a caller printing one finding per line should too.
type Finding struct {
	Rule    string
	File    string
	Line    int
	Column  int
	Message string
}

// NotAllowedError is a request for a rule that isn't on the allowlist, whether or not cohere has the rule.
type NotAllowedError struct {
	Rule string
}

func (err NotAllowedError) Error() string {
	return fmt.Sprintf("rule_runner: %q is not a rule another program may run; the allowlist in rule_runner.go decides which are", err.Rule)
}

// RunRule runs the rule named ruleName over files, which belong to program, with typeChecker answering its
// type questions, and returns what it found, in file order and then source order.
//
// The rule runs as cohere's walk runs it: through the program view its declared reads allow, with a file
// cache shared across the file, and one traversal per file. Suppression comments do not apply: a caller
// running a rule to decide what its compiler accepts reads every finding, and an `eslint-disable` comment is
// a statement to a linter, not to a compiler.
//
// typeChecker must be one the caller holds exclusively for the call, as the walk holds a file's checker. A
// rule that panics is reported as an error naming the rule and the file, never as an empty result.
func RunRule(program *compiler.Program, typeChecker *checker.Checker, files []*ast.SourceFile, ruleName string) (findings []Finding, err error) {
	subject, isAllowed := allowed[ruleName]
	if !isAllowed {
		return nil, NotAllowedError{Rule: ruleName}
	}
	if program == nil || typeChecker == nil {
		return nil, fmt.Errorf("rule_runner: %s needs a program and a checker, and was given program %t and checker %t", ruleName, program != nil, typeChecker != nil)
	}
	for _, sourceFile := range files {
		diagnostics, fileErr := runOnFile(sourceFile, subject, program, typeChecker)
		if fileErr != nil {
			return nil, fileErr
		}
		sort.SliceStable(diagnostics, func(left int, right int) bool {
			return diagnostics[left].Range.Pos() < diagnostics[right].Range.Pos()
		})
		for _, diagnostic := range diagnostics {
			fileName, line, column := diagnostic.Location()
			findings = append(findings, Finding{
				Rule:    diagnostic.RuleName,
				File:    fileName,
				Line:    line,
				Column:  column,
				Message: diagnostic.Message.Description,
			})
		}
	}
	return findings, nil
}

// runOnFile runs subject on one file, turning a panic into an error that names the rule and the file.
func runOnFile(sourceFile *ast.SourceFile, subject rule.Rule, program *compiler.Program, typeChecker *checker.Checker) (diagnostics []rule.Diagnostic, err error) {
	defer func() {
		if recovered := recover(); recovered != nil {
			err = fmt.Errorf("rule_runner: %s panicked on %s: %v", subject.Name, sourceFile.FileName(), recovered)
		}
	}()
	return linter.LintFile(sourceFile, []rule.Rule{subject}, program, typeChecker), nil
}
