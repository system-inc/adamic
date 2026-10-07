package cloud

import (
	"encoding/json"
	"flag"
	"os"
	"path/filepath"
	"testing"
)

var updateModuleTargets = flag.Bool("update-module-targets", false, "record reviewed stage1 Go invocations and imports")

func TestCohereModuleTargets(t *testing.T) {
	data, err := os.ReadFile("cohere-module-targets.json")
	if err != nil {
		t.Fatal(err)
	}
	var targets moduleTargets
	if err = json.Unmarshal(data, &targets); err != nil {
		t.Fatal(err)
	}
	calls, imports, err := stage1GoInputs("..")
	if err != nil {
		t.Fatal(err)
	}
	if *updateModuleTargets {
		targets.Calls = calls
		targets.Imports = imports
		encoded, err := json.MarshalIndent(targets, "", "  ")
		if err != nil {
			t.Fatal(err)
		}
		if err = os.WriteFile("cohere-module-targets.json", append(encoded, '\n'), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if err = checkModuleTargets(targets, calls, imports); err != nil {
		t.Fatal(err)
	}
}

func TestModuleTargetCheckCatchesNewInvocation(t *testing.T) {
	root := t.TempDir()
	directory := filepath.Join(root, "stage1")
	if err := os.Mkdir(directory, 0o755); err != nil {
		t.Fatal(err)
	}
	file := filepath.Join(directory, "new_test.go")
	if err := os.WriteFile(file, []byte("package probe\nfunc probe() { execute(\"go\", \"test\", \"./internal/new-oracle\") }\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	calls, imports, err := stage1GoInputs(root)
	if err != nil {
		t.Fatal(err)
	}
	if err := checkModuleTargets(moduleTargets{}, calls, imports); err == nil {
		t.Fatal("new Go oracle target escaped coverage check")
	}
}
