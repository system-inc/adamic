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

func TestHostConfigurationOracleReaders(test *testing.T) {
	test.Parallel()
	if input := os.Getenv("ADAMIC_HOST_CONFIGURATION_ORACLE_INPUT"); input != "" {
		if os.Getenv("ADAMIC_HOST_CONFIGURATION_ORACLE_READER") == "newline" {
			nodeProcessHostNewLine(test, input)
		} else {
			hostStatusFile(test, 0, false, input)
		}
		return
	}
	binary, executableError := os.Executable()
	if executableError != nil {
		test.Fatal(executableError)
	}
	node, lookPathError := exec.LookPath("node")
	if lookPathError != nil {
		test.Fatal(lookPathError)
	}
	stubDirectory := test.TempDir()
	stub := filepath.Join(stubDirectory, "node-stub")
	calls := filepath.Join(stubDirectory, "calls")
	writeExecutable(
		test,
		stub,
		"#!/bin/sh\nif [ \"$1\" = -p ]; then echo 'v24.19.0 false'; exit 0; fi\nprintf '%s\\n' \"$@\" >> '"+
			calls+
			"'\nexec '"+
			node+
			"' \"$@\"\n",
	)
	original, readError := os.ReadFile(filepath.Join(repository, "stage3/fixtures/host/status.json"))
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
	for _, reader := range []string{"status", "newline"} {
		for _, mode := range []string{"unset", "selected", "unknown", "wrong-recorded"} {
			test.Run(reader+"/"+mode, func(test *testing.T) {
				var rows []map[string]json.RawMessage
				if decodeError := json.Unmarshal(original, &rows); decodeError != nil {
					test.Fatal(decodeError)
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
					if decodeError := json.Unmarshal(rows[index]["node"], &expected); decodeError != nil {
						test.Fatal(decodeError)
					}
					expected["stdout"] = "wrong recorded bytes\n"
					rows[index]["node"], _ = json.Marshal(expected)
				}
				encoded, encodeError := json.Marshal(rows)
				if encodeError != nil {
					test.Fatal(encodeError)
				}
				input := filepath.Join(test.TempDir(), "status.json")
				if writeError := os.WriteFile(input, encoded, 0600); writeError != nil {
					test.Fatal(writeError)
				}
				selected := ""
				if mode == "selected" {
					selected = stub
				}
				cmd := exec.Command(
					binary,
					"-test.run=^TestHostConfigurationOracleReaders$",
					"-test.count=1",
					"-test.v",
					"-test.timeout=30s",
				)
				cmd.Env = append(
					os.Environ(),
					"ADAMIC_HOST_CONFIGURATION_ORACLE_INPUT="+input,
					"ADAMIC_HOST_CONFIGURATION_ORACLE_READER="+reader,
					"ADAMIC_HOST_NODE_WITHOUT_INSPECTOR="+selected,
				)
				if writeError := os.WriteFile(calls, nil, 0600); writeError != nil {
					test.Fatal(writeError)
				}
				out, combinedOutputError := cmd.CombinedOutput()
				test.Logf("%s", out)
				if mode == "unset" || mode == "selected" {
					for _, backend := range []string{"native"} {
						if !strings.Contains(string(out), "--- PASS: TestHostConfigurationOracleReaders/"+backend) {
							test.Fatalf("%s comparison didn't run: %s", backend, out)
						}
					}
				}
				if mode == "unset" &&
					!strings.Contains(string(out), "--- SKIP: TestHostConfigurationOracleReaders/javascript") {
					test.Fatalf("JavaScript needs the configured binary: %s", out)
				}
				if mode == "selected" &&
					!strings.Contains(string(out), "--- PASS: TestHostConfigurationOracleReaders/javascript") {
					test.Fatalf("configured JavaScript did not run: %s", out)
				}
				if mode == "selected" {
					data, readError := os.ReadFile(calls)
					if readError != nil || strings.Count(string(data), "oracle/node.mjs") < 2 {
						test.Fatalf(
							"source and generated JavaScript must both use the configured binary: %v %s",
							readError,
							data,
						)
					}
				}
				switch mode {
				case "unset":
					if combinedOutputError != nil || !strings.Contains(
						string(out),
						"needs node-without-inspector; set ADAMIC_HOST_NODE_WITHOUT_INSPECTOR",
					) || !strings.Contains(string(out), "--- SKIP:") {
						test.Fatalf("want named skip: %v\n%s", combinedOutputError, out)
					}
				case "selected":
					if combinedOutputError != nil || strings.Contains(string(out), "--- SKIP:") {
						test.Fatalf("want selected-Node pass: %v\n%s", combinedOutputError, out)
					}
				case "wrong-recorded":
					if combinedOutputError == nil ||
						!strings.Contains(string(out), "--- FAIL: TestHostConfigurationOracleReaders/native") {
						test.Fatalf("recorded bytes mismatch not caught natively: %v\n%s", combinedOutputError, out)
					}
				case "unknown":
					if combinedOutputError == nil || !strings.Contains(string(out), "unknown host configuration") {
						test.Fatalf("want unknown-name failure: %v\n%s", combinedOutputError, out)
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
