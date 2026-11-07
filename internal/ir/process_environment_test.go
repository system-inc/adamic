package ir

import (
	"encoding/json"
	"os/exec"
	"slices"
	"testing"
)

func TestProcessEnvironmentPrototypeKeysMatchNode(test *testing.T) {
	test.Parallel()
	data, err := exec.Command(
		"node",
		"-p",
		"JSON.stringify(Object.getOwnPropertyNames(Object.prototype))",
	).CombinedOutput()
	if err != nil {
		test.Fatalf("Node prototype inventory: %v: %s", err, data)
	}
	var keys []string
	if err := json.Unmarshal(data, &keys); err != nil {
		test.Fatal(err)
	}
	if !slices.Equal(keys, NodeEnvironmentPrototypeKeys[:]) {
		test.Fatalf(
			"Node Object.prototype keys changed: got %q, generated %q; regenerate with the pinned Node: node oracle/generate-process-environment-keys.mjs",
			keys,
			NodeEnvironmentPrototypeKeys,
		)
	}
}

func TestProcessEnvironmentHostReadCannotThrow(test *testing.T) {
	test.Parallel()
	if (NodeFSFile{Operation: "host_env", Of: String}).MayThrow() {
		test.Fatal("missing values are undefined; inherited keys panic, not throw")
	}
}
