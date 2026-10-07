package main

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestInventoryEngine(t *testing.T) {
	root, err := filepath.Abs("../../../..")
	if err != nil {
		t.Fatal(err)
	}
	scratch := t.TempDir()
	cohere := filepath.Join(root, "cohere")
	virtual := filepath.Join(cohere, "adamic_inventory.go")
	virtualTest := filepath.Join(cohere, "adamic_inventory_test.go")
	overlay, err := json.Marshal(map[string]any{"Replace": map[string]string{virtual: filepath.Join(root, "stage1/cohere/lint/inventory/testdata/engine.go"), virtualTest: filepath.Join(root, "stage1/cohere/lint/inventory/testdata/engine_test.go")}})
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(scratch, "overlay.json")
	if err := os.WriteFile(path, overlay, 0644); err != nil {
		t.Fatal(err)
	}
	logPath := filepath.Join(scratch, "engine.log")
	log, err := os.Create(logPath)
	if err != nil {
		t.Fatal(err)
	}
	command := exec.Command("go", "test", "-overlay="+path, "-count=1", "-v", virtual, virtualTest)
	command.Dir = cohere
	command.Stdout = log
	command.Stderr = log
	err = command.Run()
	log.Close()
	data, readErr := os.ReadFile(logPath)
	if readErr != nil {
		t.Fatal(readErr)
	}
	t.Logf("%s", data)
	if err != nil {
		t.Fatal(err)
	}
}
