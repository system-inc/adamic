package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestMissingStatementFailsWithMethodName(t *testing.T) {
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
	command := exec.Command("go", "run", ".", input, output)
	command.Env = append(os.Environ(), "GOWORK=off")
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
