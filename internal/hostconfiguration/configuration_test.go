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
		if why := Skip(1, row.Configuration); why != "" {
			t.Skip(why)
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
	launcher := nodeStub(t, node, "configured")
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
	out, err = invoke("verified stub selected", rows[0], nodeStub(t, node, ""))
	if err != nil || !strings.Contains(out, "--- PASS:") {
		t.Fatalf("selected Node: %v\n%s", err, out)
	}
	bad := map[string]any{"configuration": map[string]any{"name": "unknown"}, "node": map[string]string{"stdout": "stock\n"}}
	out, err = invoke("unknown", bad, node)
	if err == nil || !strings.Contains(out, "unknown host configuration") {
		t.Fatalf("unknown recipe silently accepted: %v\n%s", err, out)
	}

	// Skip must never turn metadata errors or a wrong binary into a passing skip.
	wrongMetadata := map[string]any{"configuration": map[string]any{"name": "node-without-inspector", "node": "wrong"}, "node": map[string]string{"stdout": "stock\n"}}
	for _, failure := range []struct {
		name   string
		row    map[string]any
		binary string
		want   string
	}{
		{"unknown without binary", bad, "", "unknown host configuration"},
		{"metadata without binary", wrongMetadata, "", "invalid metadata"},
		{"stock binary", rows[0], node, "v24.19.0 true"},
	} {
		out, err := invoke(failure.name, failure.row, failure.binary)
		if err == nil || !strings.Contains(out, failure.want) || strings.Contains(out, "--- SKIP:") {
			t.Fatalf("%s must fail, never skip: %v\n%s", failure.name, err, out)
		}
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

// Test-only launcher: it supplies the probe line, then runs stock Node for the
// mechanism fixture. Production resolution never accepts stock Node itself.
func nodeStub(t *testing.T, node, label string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "node-stub")
	text := fmt.Sprintf("#!/bin/sh\nif [ \"$1\" = -p ]; then echo 'v24.19.0 false'; exit 0; fi\nADAMIC_HOST_CONFIGURATION_PROOF='%s' exec '%s' \"$@\"\n", label, strings.ReplaceAll(node, "'", "'\\''"))
	if err := os.WriteFile(path, []byte(text), 0700); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestHostConfigurationRejectsStockNode(t *testing.T) {
	t.Parallel()
	node, err := exec.LookPath("node")
	if err != nil {
		t.Fatal(err)
	}
	cfg := &Configuration{Name: "node-without-inspector", Node: "v24.19.0", Configure: "--without-inspector", Ruling: "#cn4yapb"}
	_, _, err = Resolve(1, cfg, func(string) string { return node })
	if err == nil || !strings.Contains(err.Error(), node) || !strings.Contains(err.Error(), "v24.19.0 true") || !strings.Contains(err.Error(), "v24.19.0 false") {
		t.Fatalf("stock Node must fail its capability probe: %v", err)
	}
	t.Log(err)
}

func TestHostConfigurationProbeIsCached(t *testing.T) {
	t.Parallel()
	cfg := &Configuration{Name: "node-without-inspector", Node: "v24.19.0", Configure: "--without-inspector", Ruling: "#cn4yapb"}
	for _, line := range []string{"v24.19.0 false", "wrong"} {
		dir := t.TempDir()
		binary := filepath.Join(dir, "node")
		calls := filepath.Join(dir, "calls")
		text := fmt.Sprintf("#!/bin/sh\necho probe >> '%s'\necho '%s'\n", calls, line)
		if err := os.WriteFile(binary, []byte(text), 0700); err != nil {
			t.Fatal(err)
		}
		for i := 0; i < 2; i++ {
			_, _, err := Resolve(i+1, cfg, func(string) string { return binary })
			if (err == nil) != (line == "v24.19.0 false") {
				t.Fatal(err)
			}
		}
		data, err := os.ReadFile(calls)
		if err != nil {
			t.Fatal(err)
		}
		if string(data) != "probe\n" {
			t.Fatalf("probe wasn't cached: %q", data)
		}
	}
}

func TestHostConfigurationScriptProbes(t *testing.T) {
	t.Parallel()
	data, err := os.ReadFile("testdata/status.json")
	if err != nil {
		t.Fatal(err)
	}
	node, err := exec.LookPath("node")
	if err != nil {
		t.Fatal(err)
	}
	scripts := []struct {
		binary string
		args   []string
	}{
		{"python3", []string{"-B", "-c", `import json,sys,os
from configuration import node_for
row=json.load(sys.stdin)[0]
for i in range(2):
 try: node_for(row,i+1)
 except ValueError as error:
  print(error)
  if os.environ['PROBE_TEST_MODE']=='stock': sys.exit(1)
`}},
		{"node", []string{"-e", `const {nodeFor}=require('./configuration.cjs');const fs=require('fs');const row=JSON.parse(fs.readFileSync(0))[0];for(let i=0;i<2;i++){try{nodeFor(row,i+1);}catch(error){console.log(String(error));if(process.env.PROBE_TEST_MODE==='stock')process.exit(1);}}`}},
	}
	for _, script := range scripts {
		for _, mode := range []string{"stock", "success", "failure"} {
			t.Run(script.binary+"/"+mode, func(t *testing.T) {
				selected := node
				calls := filepath.Join(t.TempDir(), "calls")
				if mode != "stock" {
					selected = filepath.Join(t.TempDir(), "node")
					line := "v24.19.0 false"
					if mode == "failure" {
						line = "wrong"
					}
					text := fmt.Sprintf("#!/bin/sh\necho probe >> '%s'\necho '%s'\n", calls, line)
					if err := os.WriteFile(selected, []byte(text), 0700); err != nil {
						t.Fatal(err)
					}
				}
				cmd := exec.Command(script.binary, script.args...)
				cmd.Stdin = bytes.NewReader(data)
				cmd.Env = append(os.Environ(), switchName+"="+selected, "PROBE_TEST_MODE="+mode)
				out, err := cmd.CombinedOutput()
				if mode == "stock" {
					if err == nil || !strings.Contains(string(out), "v24.19.0 true") || !strings.Contains(string(out), "v24.19.0 false") {
						t.Fatalf("stock Node wasn't rejected: %v %s", err, out)
					}
					return
				}
				if err != nil {
					t.Fatalf("probe test: %v %s", err, out)
				}
				contents, err := os.ReadFile(calls)
				if err != nil {
					t.Fatal(err)
				}
				if string(contents) != "probe\n" {
					t.Fatalf("probe not cached: %q", contents)
				}
				if mode == "failure" && !strings.Contains(string(out), "wrong") {
					t.Fatalf("mismatch wasn't reported: %s", out)
				}
			})
		}
	}
}
