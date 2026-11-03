package hostconfiguration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

const switchName = "ADAMIC_HOST_NODE_WITHOUT_INSPECTOR"

func TestHostConfigurationReaderUsesDeclaredNode(t *testing.T) {
	t.Parallel()
	if input := os.Getenv("ADAMIC_HOST_CONFIGURATION_PROBE"); input != "" {
		var row struct {
			Configuration *Configuration `json:"configuration"`
			Node          struct {
				Stdout string `json:"stdout"`
			} `json:"node"`
		}
		data, err := os.ReadFile(input)
		if err != nil {
			t.Fatal(err)
		}
		if err := json.Unmarshal(data, &row); err != nil {
			t.Fatal(err)
		}
		node := Require(t, 1, row.Configuration)
		fixture, err := filepath.Abs("testdata/configuration.a")
		if err != nil {
			t.Fatal(err)
		}
		runner, err := filepath.Abs("../../oracle/node.mjs")
		if err != nil {
			t.Fatal(err)
		}
		out, err := exec.Command(node, "--disable-warning=ExperimentalWarning", runner, fixture).CombinedOutput()
		if err != nil || string(out) != row.Node.Stdout {
			t.Fatalf("configured Node mismatch: %v: got %q want %q", err, out, row.Node.Stdout)
		}
		return
	}
	node, err := exec.LookPath("node")
	if err != nil {
		t.Fatal(err)
	}
	binary, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile("testdata/status.json")
	if err != nil {
		t.Fatal(err)
	}
	var rows []map[string]any
	if err := json.Unmarshal(data, &rows); err != nil {
		t.Fatal(err)
	}
	invoke := func(label string, row map[string]any, selected string) (string, error) {
		t.Helper()
		raw, err := json.Marshal(row)
		if err != nil {
			t.Fatal(err)
		}
		input := filepath.Join(t.TempDir(), "status.json")
		if err := os.WriteFile(input, raw, 0600); err != nil {
			t.Fatal(err)
		}
		command := exec.Command(binary, "-test.run=^TestHostConfigurationReaderUsesDeclaredNode$", "-test.count=1", "-test.v", "-test.timeout=15s")
		command.Env = append(os.Environ(), "ADAMIC_HOST_CONFIGURATION_PROBE="+input, switchName+"="+selected)
		out, err := command.CombinedOutput()
		t.Logf("%s: %s", label, out)
		return string(out), err
	}
	// A stock-Node launcher labels the test input, not an inspector implementation.
	// Removing the field selects stock Node and the real byte comparison must fail.
	launcher := filepath.Join(t.TempDir(), "configured-node")
	if err := os.WriteFile(launcher, []byte(fmt.Sprintf("#!/bin/sh\nADAMIC_HOST_CONFIGURATION_PROOF=configured exec '%s' \"$@\"\n", strings.ReplaceAll(node, "'", "'\\''"))), 0700); err != nil {
		t.Fatal(err)
	}
	configured := map[string]any{"configuration": rows[0]["configuration"], "node": map[string]string{"stdout": "configured\n"}}
	out, err := invoke("labelled Node selected", configured, launcher)
	if err != nil {
		t.Fatalf("configured control: %v\n%s", err, out)
	}
	delete(configured, "configuration")
	out, err = invoke("configuration field deleted", configured, launcher)
	if err == nil || !strings.Contains(out, "configured Node mismatch") || !strings.Contains(out, `got "stock\n" want "configured\n"`) {
		t.Fatalf("deleted-field mutant survived: %v\n%s", err, out)
	}
	out, err = invoke("unset", rows[0], "")
	if err != nil || !strings.Contains(out, "row 1 needs node-without-inspector; set "+switchName) || !strings.Contains(out, "--- SKIP:") {
		t.Fatalf("want named skip: %v\n%s", err, out)
	}
	out, err = invoke("stock Node selected", rows[0], node)
	if err != nil || !strings.Contains(out, "--- PASS:") {
		t.Fatalf("selected Node: %v\n%s", err, out)
	}
	bad := map[string]any{"configuration": map[string]any{"name": "unknown"}, "node": map[string]string{"stdout": "stock\n"}}
	out, err = invoke("unknown", bad, node)
	if err == nil || !strings.Contains(out, "unknown host configuration") {
		t.Fatalf("unknown recipe silently accepted: %v\n%s", err, out)
	}

}

func TestHostConfigurationResolversAgree(t *testing.T) {
	t.Parallel()
	data, err := os.ReadFile("testdata/status.json")
	if err != nil {
		t.Fatal(err)
	}
	for _, script := range []struct {
		program   string
		arguments []string
	}{
		{"python3", []string{"-B", "-c", `import json,sys; from configuration import node_for; print(node_for(json.load(sys.stdin)[0],1))`}},
		{"node", []string{"-e", `const {nodeFor}=require('./configuration.cjs');const fs=require('fs');console.log(nodeFor(JSON.parse(fs.readFileSync(0))[0],1));`}},
	} {
		cmd := exec.Command(script.program, script.arguments...)
		cmd.Stdin = bytes.NewReader(data)
		cmd.Env = append(os.Environ(), switchName+"=")
		out, err := cmd.CombinedOutput()
		if err != nil || !strings.Contains(string(out), "SKIP: row 1 needs node-without-inspector; set "+switchName) {
			t.Fatalf("%s: %v %s", script.program, err, out)
		}
	}
}

// The script readers print the same named skip; Go asserts their actual entry
// points, including which Node command runs, rather than trusting a skipped row.
func TestHostConfigurationScriptReaders(t *testing.T) {
	t.Parallel()
	for _, command := range [][]string{{"python3", "-B", "test_readers.py", "-v"}, {"node", "readers_test.cjs"}} {
		t.Run(command[0], func(t *testing.T) {
			ctx, cancel := context.WithTimeout(t.Context(), 20*time.Second)
			defer cancel()
			out, err := exec.CommandContext(ctx, command[0], command[1:]...).CombinedOutput()
			if err != nil {
				t.Fatalf("reader proofs: %v\n%s", err, out)
			}
			t.Logf("%s", out)
		})
	}
}
