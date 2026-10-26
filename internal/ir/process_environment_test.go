package ir

import (
	"encoding/json"
	"os/exec"
	"slices"
	"testing"
)

func TestProcessEnvironmentPrototypeKeysMatchNode(t *testing.T) {
	t.Parallel()
	data, err := exec.Command("node", "-p", "JSON.stringify(Object.getOwnPropertyNames(Object.prototype))").CombinedOutput()
	if err != nil {
		t.Fatalf("Node prototype inventory: %v: %s", err, data)
	}
	var keys []string
	if err := json.Unmarshal(data, &keys); err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(keys, NodeEnvironmentPrototypeKeys[:]) {
		t.Fatalf("Node Object.prototype keys changed: got %q, generated %q; regenerate with the pinned Node: node oracle/generate-process-environment-keys.mjs", keys, NodeEnvironmentPrototypeKeys)
	}
}

func TestProcessEnvironmentHostReadCannotThrow(t *testing.T) {
	t.Parallel()
	if (NodeFSFile{Operation: "host_env", Of: String}).MayThrow() {
		t.Fatal("missing values are undefined; inherited keys panic, not throw")
	}
}
