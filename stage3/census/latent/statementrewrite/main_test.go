package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/buildcache"
)

func TestMissingStatementFailsWithMethodName(test *testing.T) {
	test.Parallel()
	root := test.TempDir()
	input := filepath.Join(root, "source.go")
	output := filepath.Join(root, "output.go")
	if writeError := os.WriteFile(
		input,
		[]byte("package lower\nfunc (l *lowering) another() {}\n"),
		0600,
	); writeError != nil {
		test.Fatal(writeError)
	}
	log, createError := os.Create(filepath.Join(root, "rewrite.log"))
	if createError != nil {
		test.Fatal(createError)
	}
	command := exec.Command(statementrewrite(test), input, output)
	command.Stdout = log
	command.Stderr = log
	runError := command.Run()
	log.Close()
	text, _ := os.ReadFile(filepath.Join(root, "rewrite.log"))
	if runError == nil ||
		!strings.Contains(string(text), "lowering.statement: expected exactly one function, found 0") {
		test.Fatalf("missing statement mutant survived: %v %s", runError, text)
	}
	if _, statError := os.Stat(output); !os.IsNotExist(statError) {
		test.Fatal("rewrite published partial output")
	}
}

// rewritten runs the command on source and returns its output file's text (empty if none) and its log.
func rewritten(test *testing.T, source string) (string, string, error) {
	test.Helper()
	root := test.TempDir()
	input := filepath.Join(root, "source.go")
	output := filepath.Join(root, "output.go")
	if writeError := os.WriteFile(input, []byte(source), 0600); writeError != nil {
		test.Fatal(writeError)
	}
	return rewrittenFile(test, input, output)
}

func rewrittenFile(test *testing.T, input, output string) (string, string, error) {
	test.Helper()
	log, createError := os.Create(output + ".log")
	if createError != nil {
		test.Fatal(createError)
	}
	command := exec.Command(statementrewrite(test), input, output)
	command.Stdout = log
	command.Stderr = log
	runError := command.Run()
	log.Close()
	text, _ := os.ReadFile(output + ".log")
	result, _ := os.ReadFile(output)
	return string(result), string(text), runError
}

// The signature's types are the contract, not its names: the named results Diagnostics' debug source lines need
// (f2cf1ebb, #mdaed6m) rewrite as the unnamed form does, and the raw method keeps them.
func TestStatementRewriteAcceptsNamedAndUnnamedResults(test *testing.T) {
	test.Parallel()
	for name, signature := range map[string]string{
		"unnamed": "(node *ast.Node) ([]ir.Statement, error)",
		"named":   "(node *ast.Node) (result []ir.Statement, err error)",
	} {
		test.Run(name, func(test *testing.T) {
			test.Parallel()
			output, log, rewrittenError := rewritten(
				test,
				"package lower\nfunc (l *lowering) statement"+signature+" { return nil, nil }\n",
			)
			if rewrittenError != nil {
				test.Fatalf("%s signature refused: %v %s", name, rewrittenError, log)
			}
			if !strings.Contains(output, "func (l *lowering) latentStatementRaw"+signature) {
				test.Fatalf("the raw method lost its signature:\n%s", output)
			}
			if !strings.Contains(output, "return l.latentStatement(node)") {
				test.Fatalf("no latent wrapper:\n%s", output)
			}
		})
	}
}

// Any other shape still fails closed, naming the signature it met.
func TestStatementRewriteRefusesOtherSignatures(test *testing.T) {
	test.Parallel()
	for name, signature := range map[string]string{
		"one result":        "(node *ast.Node) []ir.Statement",
		"another type":      "(node *ast.Node) ([]ir.Expression, error)",
		"two parameters":    "(node *ast.Node, depth int) ([]ir.Statement, error)",
		"another parameter": "(node ast.Node) ([]ir.Statement, error)",
	} {
		test.Run(name, func(test *testing.T) {
			test.Parallel()
			output, log, rewrittenError := rewritten(
				test,
				"package lower\nfunc (l *lowering) statement"+signature+" { panic(0) }\n",
			)
			if rewrittenError == nil || !strings.Contains(log, "lowering.statement: unsupported signature") {
				test.Fatalf("%s survived: %v %s", name, rewrittenError, log)
			}
			if output != "" {
				test.Fatal("rewrite published partial output")
			}
		})
	}
}

// Main's own lowering.statement rewrites, so a signature change in internal/lower reds here, at its author's change,
// not at the next meter run.
func TestStatementRewriteReadsMainStatements(test *testing.T) {
	test.Parallel()
	// Reads: internal/lower/statements.go
	input, absolutePathError := filepath.Abs("../../../../internal/lower/statements.go")
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	output, log, rewrittenFileError := rewrittenFile(test, input, filepath.Join(test.TempDir(), "statements.go"))
	if rewrittenFileError != nil || !strings.Contains(output, "func (l *lowering) latentStatementRaw(") {
		test.Fatalf("main's lowering.statement doesn't rewrite: %v %s", rewrittenFileError, log)
	}
}

// statementrewrite is this package's command, a keyed product (GOWORK=off, as go run built it), so no test runs go.
func statementrewrite(test testing.TB) string {
	test.Helper()
	return buildcache.GoBuild(test, "statementrewrite", "./stage3/census/latent/statementrewrite", nil, "GOWORK=off")
}

func TestProduct_Statementrewrite(test *testing.T) {
	test.Parallel()
	statementrewrite(test)
}
