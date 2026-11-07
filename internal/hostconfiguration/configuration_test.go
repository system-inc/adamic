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
	"syscall"
	"testing"
	"time"
)

const switchName = "ADAMIC_HOST_NODE_WITHOUT_INSPECTOR"

func TestHostConfigurationReaderUsesDeclaredNode(test *testing.T) {
	test.Parallel()
	if input := os.Getenv("ADAMIC_HOST_CONFIGURATION_PROBE"); input != "" {
		var row struct {
			Configuration *Configuration `json:"configuration"`
			Node          struct {
				Stdout string `json:"stdout"`
			} `json:"node"`
		}
		data, readError := os.ReadFile(input)
		if readError != nil {
			test.Fatal(readError)
		}
		if decodeError := json.Unmarshal(data, &row); decodeError != nil {
			test.Fatal(decodeError)
		}
		if why := Skip(1, row.Configuration); why != "" {
			test.Skipf(
				"configuration reader row %d: the node step needs node-without-inspector; set ADAMIC_HOST_NODE_WITHOUT_INSPECTOR",
				1,
			)
		}
		node := Require(test, 1, row.Configuration)
		fixture, absolutePathError := filepath.Abs("testdata/configuration.a")
		if absolutePathError != nil {
			test.Fatal(absolutePathError)
		}
		runner, absolutePathError := filepath.Abs("../../oracle/node.mjs")
		if absolutePathError != nil {
			test.Fatal(absolutePathError)
		}
		out, combinedOutputError := exec.Command(
			node,
			"--disable-warning=ExperimentalWarning",
			runner,
			fixture,
		).CombinedOutput()
		if combinedOutputError != nil || string(out) != row.Node.Stdout {
			test.Fatalf("configured Node mismatch: %v: got %q want %q", combinedOutputError, out, row.Node.Stdout)
		}
		return
	}
	node, lookPathError := exec.LookPath("node")
	if lookPathError != nil {
		test.Fatal(lookPathError)
	}
	binary, executableError := os.Executable()
	if executableError != nil {
		test.Fatal(executableError)
	}
	data, readError := os.ReadFile("testdata/status.json")
	if readError != nil {
		test.Fatal(readError)
	}
	var rows []map[string]any
	if decodeError := json.Unmarshal(data, &rows); decodeError != nil {
		test.Fatal(decodeError)
	}
	invoke := func(label string, row map[string]any, selected string) (string, error) {
		test.Helper()
		raw, encodeError := json.Marshal(row)
		if encodeError != nil {
			test.Fatal(encodeError)
		}
		input := filepath.Join(test.TempDir(), "status.json")
		if writeError := os.WriteFile(input, raw, 0600); writeError != nil {
			test.Fatal(writeError)
		}
		command := exec.Command(
			binary,
			"-test.run=^TestHostConfigurationReaderUsesDeclaredNode$",
			"-test.count=1",
			"-test.v",
			"-test.timeout=15s",
		)
		command.Env = append(os.Environ(), "ADAMIC_HOST_CONFIGURATION_PROBE="+input, switchName+"="+selected)
		out, combinedOutputError := command.CombinedOutput()
		test.Logf("%s: %s", label, out)
		return string(out), combinedOutputError
	}
	// A stock-Node launcher labels the test input, not an inspector implementation.
	// Removing the field selects stock Node and the real byte comparison must fail.
	launcher := nodeStub(test, node, "configured")
	configured := map[string]any{
		"configuration": rows[0]["configuration"],
		"node":          map[string]string{"stdout": "configured\n"},
	}
	out, invokeError := invoke("labelled Node selected", configured, launcher)
	if invokeError != nil {
		test.Fatalf("configured control: %v\n%s", invokeError, out)
	}
	delete(configured, "configuration")
	out, invokeError = invoke("configuration field deleted", configured, launcher)
	if invokeError == nil ||
		!strings.Contains(out, "configured Node mismatch") ||
		!strings.Contains(out, `got "stock\n" want "configured\n"`) {
		test.Fatalf("deleted-field mutant survived: %v\n%s", invokeError, out)
	}
	out, invokeError = invoke("unset", rows[0], "")
	if invokeError != nil || !strings.Contains(
		out,
		"configuration reader row 1: the node step needs node-without-inspector; set "+switchName,
	) || !strings.Contains(out, "--- SKIP:") {
		test.Fatalf("want named skip: %v\n%s", invokeError, out)
	}
	out, invokeError = invoke("verified stub selected", rows[0], nodeStub(test, node, ""))
	if invokeError != nil || !strings.Contains(out, "--- PASS:") {
		test.Fatalf("selected Node: %v\n%s", invokeError, out)
	}
	bad := map[string]any{
		"configuration": map[string]any{"name": "unknown"},
		"node":          map[string]string{"stdout": "stock\n"},
	}
	out, invokeError = invoke("unknown", bad, node)
	if invokeError == nil || !strings.Contains(out, "unknown host configuration") {
		test.Fatalf("unknown recipe silently accepted: %v\n%s", invokeError, out)
	}

	// Skip must never turn metadata errors or a wrong binary into a passing skip.
	wrongMetadata := map[string]any{
		"configuration": map[string]any{"name": "node-without-inspector", "node": "wrong"},
		"node":          map[string]string{"stdout": "stock\n"},
	}
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
		out, invokeError := invoke(failure.name, failure.row, failure.binary)
		if invokeError == nil || !strings.Contains(out, failure.want) || strings.Contains(out, "--- SKIP:") {
			test.Fatalf("%s must fail, never skip: %v\n%s", failure.name, invokeError, out)
		}
	}

}

func TestHostConfigurationResolversAgree(test *testing.T) {
	test.Parallel()
	data, readError := os.ReadFile("testdata/status.json")
	if readError != nil {
		test.Fatal(readError)
	}
	for _, script := range []struct {
		program   string
		arguments []string
	}{
		{"python3", []string{
			"-B",
			"-c",
			`import json,sys; from configuration import node_for; print(node_for(json.load(sys.stdin)[0],1))`,
		}},
		{"node", []string{
			"-e",
			`const {nodeFor}=require('./configuration.cjs');const fs=require('fs');console.log(nodeFor(JSON.parse(fs.readFileSync(0))[0],1));`,
		}},
	} {
		cmd := exec.Command(script.program, script.arguments...)
		cmd.Stdin = bytes.NewReader(data)
		cmd.Env = append(os.Environ(), switchName+"=")
		out, combinedOutputError := cmd.CombinedOutput()
		if combinedOutputError != nil ||
			!strings.Contains(string(out), "SKIP: row 1 needs node-without-inspector; set "+switchName) {
			test.Fatalf("%s: %v %s", script.program, combinedOutputError, out)
		}
	}
}

// The script readers print the same named skip; Go asserts their actual entry
// points, including which Node command runs, rather than trusting a skipped row.
func TestHostConfigurationScriptReaders(test *testing.T) {
	test.Parallel()
	for _, command := range [][]string{{"python3", "-B", "test_readers.py", "-v"}, {"node", "readers_test.cjs"}} {
		test.Run(command[0], func(test *testing.T) {
			ctx, cancel := context.WithTimeout(test.Context(), 20*time.Second)
			defer cancel()
			cmd := exec.CommandContext(ctx, command[0], command[1:]...)
			if command[0] == "node" {
				// setup.sh pins tsc and its adjacent API; no npm install lives in a runner's tree.
				tsc, lookPathError := exec.LookPath("tsc")
				if lookPathError != nil {
					test.Fatal(lookPathError)
				}
				tsc, resolveSymlinksError := filepath.EvalSymlinks(tsc)
				if resolveSymlinksError != nil {
					test.Fatal(resolveSymlinksError)
				}
				api := filepath.Join(filepath.Dir(tsc), "../lib/typescript.js")
				cmd.Env = append(os.Environ(), "CENSUS_TYPESCRIPT="+api)
			}
			out, combinedOutputError := cmd.CombinedOutput()
			if combinedOutputError != nil {
				test.Fatalf("reader proofs: %v\n%s", combinedOutputError, out)
			}
			test.Logf("%s", out)
		})
	}
}

// Test-only launcher: it supplies the probe line, then runs stock Node for the
// mechanism fixture. Production resolution never accepts stock Node itself.
func nodeStub(test *testing.T, node, label string) string {
	test.Helper()
	path := filepath.Join(test.TempDir(), "node-stub")
	text := fmt.Sprintf(
		"#!/bin/sh\nif [ \"$1\" = -p ]; then echo 'v24.19.0 false'; exit 0; fi\nADAMIC_HOST_CONFIGURATION_PROOF='%s' exec '%s' \"$@\"\n",
		label,
		strings.ReplaceAll(node, "'", "'\\''"),
	)
	writeExecutable(test, path, text)
	return path
}

func TestHostConfigurationRejectsStockNode(test *testing.T) {
	test.Parallel()
	node, lookPathError := exec.LookPath("node")
	if lookPathError != nil {
		test.Fatal(lookPathError)
	}
	cfg := &Configuration{
		Name:      "node-without-inspector",
		Node:      "v24.19.0",
		Configure: "--without-inspector",
		Ruling:    "#cn4yapb",
	}
	_, _, resolveError := Resolve(1, cfg, func(string) string { return node })
	if resolveError == nil ||
		!strings.Contains(resolveError.Error(), node) ||
		!strings.Contains(resolveError.Error(), "v24.19.0 true") ||
		!strings.Contains(resolveError.Error(), "v24.19.0 false") {
		test.Fatalf("stock Node must fail its capability probe: %v", resolveError)
	}
	test.Log(resolveError)
}

func TestHostConfigurationProbeIsCached(test *testing.T) {
	test.Parallel()
	cfg := &Configuration{
		Name:      "node-without-inspector",
		Node:      "v24.19.0",
		Configure: "--without-inspector",
		Ruling:    "#cn4yapb",
	}
	for _, line := range []string{"v24.19.0 false", "wrong"} {
		dir := test.TempDir()
		binary := filepath.Join(dir, "node")
		calls := filepath.Join(dir, "calls")
		text := fmt.Sprintf("#!/bin/sh\necho probe >> '%s'\necho '%s'\n", calls, line)
		writeExecutable(test, binary, text)
		for i := 0; i < 2; i++ {
			_, _, resolveError := Resolve(i+1, cfg, func(string) string { return binary })
			if (resolveError == nil) != (line == "v24.19.0 false") {
				test.Fatal(resolveError)
			}
		}
		data, readError := os.ReadFile(calls)
		if readError != nil {
			test.Fatal(readError)
		}
		if string(data) != "probe\n" {
			test.Fatalf("probe wasn't cached: %q", data)
		}
	}
}

func TestHostConfigurationScriptProbes(test *testing.T) {
	test.Parallel()
	data, readError := os.ReadFile("testdata/status.json")
	if readError != nil {
		test.Fatal(readError)
	}
	node, lookPathError := exec.LookPath("node")
	if lookPathError != nil {
		test.Fatal(lookPathError)
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
		{"node", []string{
			"-e",
			`const {nodeFor}=require('./configuration.cjs');const fs=require('fs');const row=JSON.parse(fs.readFileSync(0))[0];for(let i=0;i<2;i++){try{nodeFor(row,i+1);}catch(error){console.log(String(error));if(process.env.PROBE_TEST_MODE==='stock')process.exit(1);}}`,
		}},
	}
	for _, script := range scripts {
		for _, mode := range []string{"stock", "success", "failure"} {
			test.Run(script.binary+"/"+mode, func(test *testing.T) {
				selected := node
				calls := filepath.Join(test.TempDir(), "calls")
				if mode != "stock" {
					selected = filepath.Join(test.TempDir(), "node")
					line := "v24.19.0 false"
					if mode == "failure" {
						line = "wrong"
					}
					text := fmt.Sprintf("#!/bin/sh\necho probe >> '%s'\necho '%s'\n", calls, line)
					writeExecutable(test, selected, text)
				}
				cmd := exec.Command(script.binary, script.args...)
				cmd.Stdin = bytes.NewReader(data)
				cmd.Env = append(os.Environ(), switchName+"="+selected, "PROBE_TEST_MODE="+mode)
				out, combinedOutputError := cmd.CombinedOutput()
				if mode == "stock" {
					if combinedOutputError == nil ||
						!strings.Contains(string(out), "v24.19.0 true") ||
						!strings.Contains(string(out), "v24.19.0 false") {
						test.Fatalf("stock Node wasn't rejected: %v %s", combinedOutputError, out)
					}
					return
				}
				if combinedOutputError != nil {
					test.Fatalf("probe test: %v %s", combinedOutputError, out)
				}
				contents, readError := os.ReadFile(calls)
				if readError != nil {
					test.Fatal(readError)
				}
				if string(contents) != "probe\n" {
					test.Fatalf("probe not cached: %q", contents)
				}
				if mode == "failure" && !strings.Contains(string(out), "wrong") {
					test.Fatalf("mismatch wasn't reported: %s", out)
				}
			})
		}
	}
}

// writeExecutable writes a script the test then executes. A parallel test's fork
// inherits every descriptor open at that moment until its child execs, and an exec of
// a file still open for writing fails with ETXTBSY. So forks are held off until the
// writable descriptor is closed; a rename alone keeps the same inode.
func writeExecutable(test *testing.T, path, script string) {
	test.Helper()
	var temporary string
	defer func() {
		if temporary != "" {
			os.Remove(temporary)
		}
	}()
	func() {
		syscall.ForkLock.Lock()
		defer syscall.ForkLock.Unlock()
		file, temporaryFileError := os.CreateTemp(filepath.Dir(path), ".executable-*")
		if temporaryFileError != nil {
			test.Fatal(temporaryFileError)
		}
		temporary = file.Name()
		if _, writeStringError := file.WriteString(script); writeStringError != nil {
			file.Close()
			test.Fatal(writeStringError)
		}
		if chmodError := file.Chmod(0700); chmodError != nil {
			file.Close()
			test.Fatal(chmodError)
		}
		if closeError := file.Close(); closeError != nil {
			test.Fatal(closeError)
		}
	}()
	if renameError := os.Rename(temporary, path); renameError != nil {
		test.Fatal(renameError)
	}
	temporary = ""
}
