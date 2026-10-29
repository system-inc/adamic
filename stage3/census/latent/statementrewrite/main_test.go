package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/buildcache"
)

func TestMissingStatementFailsWithMethodName(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	input := filepath.Join(root, "source.go")
	output := filepath.Join(root, "output.go")
	if err := os.WriteFile(input, []byte("package lower\nfunc (l *lowering) another() {}\n"), 0600); err != nil {
		t.Fatal(err)
	}
	log, err := os.Create(filepath.Join(root, "rewrite.log"))
	if err != nil {
		t.Fatal(err)
	}
	command := exec.Command(statementrewrite(t), input, output)
	command.Stdout = log
	command.Stderr = log
	err = command.Run()
	log.Close()
	text, _ := os.ReadFile(filepath.Join(root, "rewrite.log"))
	if err == nil || !strings.Contains(string(text), "lowering.statement: expected exactly one function, found 0") {
		t.Fatalf("missing statement mutant survived: %v %s", err, text)
	}
	if _, err := os.Stat(output); !os.IsNotExist(err) {
		t.Fatal("rewrite published partial output")
	}
}

// rewritten runs the command on source and returns its output file's text (empty if none) and its log.
func rewritten(t *testing.T, source string) (string, string, error) {
	t.Helper()
	root := t.TempDir()
	input := filepath.Join(root, "source.go")
	output := filepath.Join(root, "output.go")
	if err := os.WriteFile(input, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	return rewrittenFile(t, input, output)
}

func rewrittenFile(t *testing.T, input, output string) (string, string, error) {
	t.Helper()
	log, err := os.Create(output + ".log")
	if err != nil {
		t.Fatal(err)
	}
	command := exec.Command(statementrewrite(t), input, output)
	command.Stdout = log
	command.Stderr = log
	err = command.Run()
	log.Close()
	text, _ := os.ReadFile(output + ".log")
	result, _ := os.ReadFile(output)
	return string(result), string(text), err
}

// The signature's types are the contract, not its names: the named results Diagnostics' debug source lines need
// (f2cf1ebb, #mdaed6m) rewrite as the unnamed form does, and the raw method keeps them.
func TestStatementRewriteAcceptsNamedAndUnnamedResults(t *testing.T) {
	t.Parallel()
	for name, signature := range map[string]string{
		"unnamed": "(node *ast.Node) ([]ir.Statement, error)",
		"named":   "(node *ast.Node) (result []ir.Statement, err error)",
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			output, log, err := rewritten(t, "package lower\nfunc (l *lowering) statement"+signature+" { return nil, nil }\n")
			if err != nil {
				t.Fatalf("%s signature refused: %v %s", name, err, log)
			}
			if !strings.Contains(output, "func (l *lowering) latentStatementRaw"+signature) {
				t.Fatalf("the raw method lost its signature:\n%s", output)
			}
			if !strings.Contains(output, "return l.latentStatement(node)") {
				t.Fatalf("no latent wrapper:\n%s", output)
			}
		})
	}
}

// Any other shape still fails closed, naming the signature it met.
func TestStatementRewriteRefusesOtherSignatures(t *testing.T) {
	t.Parallel()
	for name, signature := range map[string]string{
		"one result":        "(node *ast.Node) []ir.Statement",
		"another type":      "(node *ast.Node) ([]ir.Expression, error)",
		"two parameters":    "(node *ast.Node, depth int) ([]ir.Statement, error)",
		"another parameter": "(node ast.Node) ([]ir.Statement, error)",
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			output, log, err := rewritten(t, "package lower\nfunc (l *lowering) statement"+signature+" { panic(0) }\n")
			if err == nil || !strings.Contains(log, "lowering.statement: unsupported signature") {
				t.Fatalf("%s survived: %v %s", name, err, log)
			}
			if output != "" {
				t.Fatal("rewrite published partial output")
			}
		})
	}
}

// Main's own lowering.statement rewrites, so a signature change in internal/lower reds here, at its author's change,
// not at the next meter run.
func TestStatementRewriteReadsMainStatements(t *testing.T) {
	t.Parallel()
	// Reads: internal/lower/statements.go
	input, err := filepath.Abs("../../../../internal/lower/statements.go")
	if err != nil {
		t.Fatal(err)
	}
	output, log, err := rewrittenFile(t, input, filepath.Join(t.TempDir(), "statements.go"))
	if err != nil || !strings.Contains(output, "func (l *lowering) latentStatementRaw(") {
		t.Fatalf("main's lowering.statement doesn't rewrite: %v %s", err, log)
	}
}

// statementrewrite is this package's command, a keyed product (GOWORK=off, as go run built it), so no test runs go.
func statementrewrite(t testing.TB) string {
	t.Helper()
	return buildcache.GoBuild(t, "statementrewrite", "./stage3/census/latent/statementrewrite", nil, "GOWORK=off")
}

func TestProduct_Statementrewrite(t *testing.T) {
	t.Parallel()
	statementrewrite(t)
}
