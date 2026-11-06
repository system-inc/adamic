package oracle

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
)

func TestHostConfigurationOracleReaders(t *testing.T) {
	t.Parallel()
	if input := os.Getenv("ADAMIC_HOST_CONFIGURATION_ORACLE_INPUT"); input != "" {
		if os.Getenv("ADAMIC_HOST_CONFIGURATION_ORACLE_READER") == "newline" {
			nodeProcessHostNewLine(t, input)
		} else {
			hostStatusFile(t, 0, false, input)
		}
		return
	}
	binary, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	node, err := exec.LookPath("node")
	if err != nil {
		t.Fatal(err)
	}
	stubDirectory := t.TempDir()
	stub := filepath.Join(stubDirectory, "node-stub")
	calls := filepath.Join(stubDirectory, "calls")
	writeExecutable(t, stub, "#!/bin/sh\nif [ \"$1\" = -p ]; then echo 'v24.19.0 false'; exit 0; fi\nprintf '%s\\n' \"$@\" >> '"+calls+"'\nexec '"+node+"' \"$@\"\n")
	original, err := os.ReadFile(filepath.Join(repository, "stage3/fixtures/host/status.json"))
	if err != nil {
		t.Fatal(err)
	}
	template, err := os.ReadFile(filepath.Join(repository, "internal/hostconfiguration/testdata/status.json"))
	if err != nil {
		t.Fatal(err)
	}
	var model []map[string]json.RawMessage
	if err := json.Unmarshal(template, &model); err != nil {
		t.Fatal(err)
	}
	for _, reader := range []string{"status", "newline"} {
		for _, mode := range []string{"unset", "selected", "unknown", "wrong-recorded"} {
			t.Run(reader+"/"+mode, func(t *testing.T) {
				var rows []map[string]json.RawMessage
				if err := json.Unmarshal(original, &rows); err != nil {
					t.Fatal(err)
				}
				index := 0
				if reader == "status" {
					rows[0], rows[22] = rows[22], rows[0]
				}
				if reader == "newline" {
					index = 22
				}
				rows[index]["configuration"] = model[0]["configuration"]
				if mode == "unknown" {
					rows[index]["configuration"] = json.RawMessage(`{"name":"unknown-build"}`)
				}
				if mode == "wrong-recorded" {
					var expected map[string]any
					if err := json.Unmarshal(rows[index]["node"], &expected); err != nil {
						t.Fatal(err)
					}
					expected["stdout"] = "wrong recorded bytes\n"
					rows[index]["node"], _ = json.Marshal(expected)
				}
				encoded, err := json.Marshal(rows)
				if err != nil {
					t.Fatal(err)
				}
				input := filepath.Join(t.TempDir(), "status.json")
				if err := os.WriteFile(input, encoded, 0600); err != nil {
					t.Fatal(err)
				}
				selected := ""
				if mode == "selected" {
					selected = stub
				}
				cmd := exec.Command(binary, "-test.run=^TestHostConfigurationOracleReaders$", "-test.count=1", "-test.v", "-test.timeout=30s")
				cmd.Env = append(os.Environ(), "ADAMIC_HOST_CONFIGURATION_ORACLE_INPUT="+input, "ADAMIC_HOST_CONFIGURATION_ORACLE_READER="+reader, "ADAMIC_HOST_NODE_WITHOUT_INSPECTOR="+selected)
				if err := os.WriteFile(calls, nil, 0600); err != nil {
					t.Fatal(err)
				}
				out, err := cmd.CombinedOutput()
				t.Logf("%s", out)
				if mode == "unset" || mode == "selected" {
					for _, backend := range []string{"native"} {
						if !strings.Contains(string(out), "--- PASS: TestHostConfigurationOracleReaders/"+backend) {
							t.Fatalf("%s comparison didn't run: %s", backend, out)
						}
					}
				}
				if mode == "unset" && !strings.Contains(string(out), "--- SKIP: TestHostConfigurationOracleReaders/javascript") {
					t.Fatalf("JavaScript needs the configured binary: %s", out)
				}
				if mode == "selected" && !strings.Contains(string(out), "--- PASS: TestHostConfigurationOracleReaders/javascript") {
					t.Fatalf("configured JavaScript did not run: %s", out)
				}
				if mode == "selected" {
					data, err := os.ReadFile(calls)
					if err != nil || strings.Count(string(data), "oracle/node.mjs") < 2 {
						t.Fatalf("source and generated JavaScript must both use the configured binary: %v %s", err, data)
					}
				}
				switch mode {
				case "unset":
					if err != nil || !strings.Contains(string(out), "needs node-without-inspector; set ADAMIC_HOST_NODE_WITHOUT_INSPECTOR") || !strings.Contains(string(out), "--- SKIP:") {
						t.Fatalf("want named skip: %v\n%s", err, out)
					}
				case "selected":
					if err != nil || strings.Contains(string(out), "--- SKIP:") {
						t.Fatalf("want selected-Node pass: %v\n%s", err, out)
					}
				case "wrong-recorded":
					if err == nil || !strings.Contains(string(out), "--- FAIL: TestHostConfigurationOracleReaders/native") {
						t.Fatalf("recorded bytes mismatch not caught natively: %v\n%s", err, out)
					}
				case "unknown":
					if err == nil || !strings.Contains(string(out), "unknown host configuration") {
						t.Fatalf("want unknown-name failure: %v\n%s", err, out)
					}
				}
			})
		}
	}
}

// writeExecutable writes a script the test then executes. A parallel test's fork
// inherits every descriptor open at that moment until its child execs, and an exec of
// a file still open for writing fails with ETXTBSY. So forks are held off until the
// writable descriptor is closed; a rename alone keeps the same inode.
func writeExecutable(t *testing.T, path, script string) {
	t.Helper()
	var temporary string
	defer func() {
		if temporary != "" {
			os.Remove(temporary)
		}
	}()
	func() {
		syscall.ForkLock.Lock()
		defer syscall.ForkLock.Unlock()
		file, err := os.CreateTemp(filepath.Dir(path), ".executable-*")
		if err != nil {
			t.Fatal(err)
		}
		temporary = file.Name()
		if _, err := file.WriteString(script); err != nil {
			file.Close()
			t.Fatal(err)
		}
		if err := file.Chmod(0700); err != nil {
			file.Close()
			t.Fatal(err)
		}
		if err := file.Close(); err != nil {
			t.Fatal(err)
		}
	}()
	if err := os.Rename(temporary, path); err != nil {
		t.Fatal(err)
	}
	temporary = ""
}
