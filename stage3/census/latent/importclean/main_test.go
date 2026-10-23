package main

import (
	"bytes"
	"context"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// The removed function body was the only use of errors, like Lower's async pass.
func TestProduct_RemovedImportBuilds(t *testing.T) {
	source := []byte(`package fixture
import ("errors"; f "fmt"; _ "net/http")
func lower() error { return errors.New(f.Sprint("production")) }
`)
	fs := token.NewFileSet()
	file, err := parser.ParseFile(fs, "fixture.go", source, parser.ParseComments)
	if err != nil {
		t.Fatal(err)
	}
	// Model the no-output rewrite: the production body is replaced by a measurement error.
	replacement, err := parser.ParseFile(fs, "body.go", `package fixture
 func lower() error { return f.Errorf("measurement only") }`, 0)
	if err != nil {
		t.Fatal(err)
	}
	file.Decls[1].(*ast.FuncDecl).Body = replacement.Decls[0].(*ast.FuncDecl).Body
	var rewritten bytes.Buffer
	if err := format.Node(&rewritten, fs, file); err != nil {
		t.Fatal(err)
	}
	output, err := clean(rewritten.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(output), `f "fmt"`) || !strings.Contains(string(output), `_ "net/http"`) {
		t.Fatal("used or side-effect import lost")
	}
	directory := t.TempDir()
	if err := os.WriteFile(filepath.Join(directory, "fixture.go"), output, 0600); err != nil {
		t.Fatal(err)
	}
	log, err := os.Create(filepath.Join(directory, "build.log"))
	if err != nil {
		t.Fatal(err)
	}
	defer log.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, "go", "build", "-o", filepath.Join(directory, "fixture.a"), filepath.Join(directory, "fixture.go"))
	command.Stdout = log
	command.Stderr = log
	if err := command.Run(); err != nil {
		data, _ := os.ReadFile(log.Name())
		t.Fatalf("build: %v\n%s", err, data)
	}
}
