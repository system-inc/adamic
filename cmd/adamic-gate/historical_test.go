package main

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestHistoricalRequiredInputAuthority(t *testing.T) {
	root := t.TempDir()
	directory := filepath.Join(root, "cmd/adamic-gate")
	if err := os.MkdirAll(directory, 0700); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(directory, "required_environment.go")
	if err := os.WriteFile(path, []byte("package main\nvar requiredGateVariables=[]string{\"ADAMIC_ORACLE_WASI\",\"ADAMIC_TEST_WASI\"}\n"), 0600); err != nil {
		t.Fatal(err)
	}
	got, err := requiredGateVariables(root)
	if err != nil || !reflect.DeepEqual(got, []string{"ADAMIC_ORACLE_WASI", "ADAMIC_TEST_WASI"}) {
		t.Fatal(got, err)
	}
	if err := os.WriteFile(path, []byte("package main\nvar requiredGateVariables=dynamic()\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := requiredGateVariables(root); err == nil {
		t.Fatal("accepted dynamic historical authority")
	}
}
