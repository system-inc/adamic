package fixtures

import (
	"encoding/json"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestHostConfigurationFixtureReaders(test *testing.T) {
	test.Parallel()
	repository, absolutePathError := filepath.Abs("../..")
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	// Fetch the existing product before starting a 30-second child proof. A runner
	// supplies it, while a fresh local proof may need to compile the oracle first.
	fixtureOracleHook(test, repository)
	executable, executableError := os.Executable()
	if executableError != nil {
		test.Fatal(executableError)
	}
	node, lookPathError := exec.LookPath("node")
	if lookPathError != nil {
		test.Fatal(lookPathError)
	}
	stub := filepath.Join(test.TempDir(), "node-stub")
	if writeError := os.WriteFile(
		stub,
		[]byte("#!/bin/sh\nif [ \"$1\" = -p ]; then echo 'v24.19.0 false'; exit 0; fi\nexec '"+node+"' \"$@\"\n"),
		0700,
	); writeError != nil {
		test.Fatal(writeError)
	}
	original, readError := os.ReadFile("host/status.json")
	if readError != nil {
		test.Fatal(readError)
	}
	template, readError := os.ReadFile(filepath.Join(repository, "internal/hostconfiguration/testdata/status.json"))
	if readError != nil {
		test.Fatal(readError)
	}
	var model []map[string]json.RawMessage
	if decodeError := json.Unmarshal(template, &model); decodeError != nil {
		test.Fatal(decodeError)
	}
	for _, reader := range []string{"execute", "manifest", "directories"} {
		modes := []string{"unset", "selected", "unknown"}
		if reader == "execute" {
			modes = append(modes, "wrong-recorded")
		}
		for _, mode := range modes {
			test.Run(reader+"/"+mode, func(test *testing.T) {
				root := test.TempDir()
				// Metadata readers see the whole union, including configured entries.
				walkError := filepath.WalkDir(".", func(path string, entry fs.DirEntry, pathError error) error {
					if pathError != nil {
						return pathError
					}
					target := filepath.Join(root, path)
					if entry.IsDir() {
						return os.MkdirAll(target, 0700)
					}
					if entry.Name() != "status.json" && filepath.Ext(path) != ".a" && filepath.Ext(path) != ".ts" {
						return nil
					}
					data, readError := os.ReadFile(path)
					if readError != nil {
						return readError
					}
					return os.WriteFile(target, data, 0600)
				})
				if walkError != nil {
					test.Fatal(walkError)
				}
				var rows []map[string]json.RawMessage
				if decodeError := json.Unmarshal(original, &rows); decodeError != nil {
					test.Fatal(decodeError)
				}
				rows[0] = make(map[string]json.RawMessage, len(model[0]))
				for key, value := range model[0] {
					rows[0][key] = value
				}
				proof, readError := os.ReadFile(
					filepath.Join(repository, "internal/hostconfiguration/testdata/configuration.a"),
				)
				if readError != nil {
					test.Fatal(readError)
				}
				if writeError := os.WriteFile(
					filepath.Join(root, "host/configuration.a"),
					proof,
					0600,
				); writeError != nil {
					test.Fatal(writeError)
				}
				if mode == "unknown" {
					rows[0]["configuration"] = json.RawMessage(`{"name":"unknown-build"}`)
				}
				if mode == "wrong-recorded" {
					rows[0]["node"] = json.RawMessage(`{"stdout":"wrong recorded bytes\n","stderr":"","exit":0}`)
				}
				encoded, encodeError := json.Marshal(rows)
				if encodeError != nil {
					test.Fatal(encodeError)
				}
				if writeError := os.WriteFile(
					filepath.Join(root, "host/status.json"),
					encoded,
					0600,
				); writeError != nil {
					test.Fatal(writeError)
				}
				selector := "^TestFixturesHost$/^host$/^configuration.a$"
				if reader == "manifest" {
					selector = "^TestFixtureShardManifest$"
				}
				if reader == "directories" {
					selector = "^TestFixtureDirectoriesHaveTopLevelTests$"
				}
				command := exec.Command(
					executable,
					"-test.run="+selector,
					"-test.count=1",
					"-test.v",
					"-test.timeout=30s",
					"-fixtures="+root,
				)
				selected := ""
				if mode == "selected" {
					selected = stub
				}
				command.Env = append(os.Environ(), "ADAMIC_HOST_NODE_WITHOUT_INSPECTOR="+selected)
				out, combinedOutputError := command.CombinedOutput()
				test.Logf("%s", out)
				if mode == "wrong-recorded" {
					if combinedOutputError == nil ||
						!strings.Contains(string(out), "--- FAIL: TestFixturesHost/host/configuration.a/native") {
						test.Fatalf("recorded bytes mismatch wasn't caught: %v\n%s", combinedOutputError, out)
					}
					return
				}
				if mode == "unknown" {
					if combinedOutputError == nil || !strings.Contains(string(out), "unknown host configuration") {
						test.Fatalf("want unknown-name failure: %v\n%s", combinedOutputError, out)
					}
					return
				}
				if combinedOutputError != nil {
					test.Fatalf("configured reader: %v\n%s", combinedOutputError, out)
				}
				if reader == "execute" {
					for _, backend := range []string{"native", "stage0"} {
						if !strings.Contains(string(out), "--- PASS: TestFixturesHost/host/configuration.a/"+backend) {
							test.Fatalf("%s comparison didn't run: %s", backend, out)
						}
					}
				}
				if reader == "execute" && strings.Contains(string(out), "/configuration.a/javascript") {
					test.Fatalf("Compiles row unexpectedly ran the JavaScript backend: %s", out)
				}
				if mode == "unset" {
					want := "row 1 needs node-without-inspector; set ADAMIC_HOST_NODE_WITHOUT_INSPECTOR"
					if reader == "execute" {
						want = "fixture row 1: the node step needs node-without-inspector; set ADAMIC_HOST_NODE_WITHOUT_INSPECTOR"
					}
					if !strings.Contains(string(out), want) {
						test.Fatalf("missing named skip:\n%s", out)
					}
				}
				if mode == "selected" && strings.Contains(string(out), "needs node-without-inspector") {
					test.Fatalf("selected recipe skipped:\n%s", out)
				}
				if reader == "manifest" && !strings.Contains(string(out), "host/configuration.a shard=") {
					test.Fatalf("metadata dropped configured row:\n%s", out)
				}
			})
		}
	}
}
