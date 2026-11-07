package ir

import (
	"encoding/json"
	"os/exec"
	"slices"
	"testing"
)

func TestProcessEnvironmentPrototypeKeysMatchNode(test *testing.T) {
	test.Parallel()
	data, combinedOutputError := exec.Command(
		"node",
		"-p",
		"JSON.stringify(Object.getOwnPropertyNames(Object.prototype))",
	).CombinedOutput()
	if combinedOutputError != nil {
		test.Fatalf("Node prototype inventory: %v: %s", combinedOutputError, data)
	}
	var keys []string
	if decodeError := json.Unmarshal(data, &keys); decodeError != nil {
		test.Fatal(decodeError)
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
