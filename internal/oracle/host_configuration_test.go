package oracle

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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
		for _, mode := range []string{"unset", "selected", "unknown"} {
			t.Run(reader+"/"+mode, func(t *testing.T) {
				var rows []map[string]json.RawMessage
				if err := json.Unmarshal(original, &rows); err != nil {
					t.Fatal(err)
				}
				index := 0
				if reader == "newline" {
					index = 22
				}
				rows[index]["configuration"] = model[0]["configuration"]
				if mode == "unknown" {
					rows[index]["configuration"] = json.RawMessage(`{"name":"unknown-build"}`)
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
					selected = node
				}
				cmd := exec.Command(binary, "-test.run=^TestHostConfigurationOracleReaders$", "-test.count=1", "-test.v", "-test.timeout=30s")
				cmd.Env = append(os.Environ(), "ADAMIC_HOST_CONFIGURATION_ORACLE_INPUT="+input, "ADAMIC_HOST_CONFIGURATION_ORACLE_READER="+reader, "ADAMIC_HOST_NODE_WITHOUT_INSPECTOR="+selected)
				out, err := cmd.CombinedOutput()
				t.Logf("%s", out)
				switch mode {
				case "unset":
					if err != nil || !strings.Contains(string(out), "needs node-without-inspector; set ADAMIC_HOST_NODE_WITHOUT_INSPECTOR") || !strings.Contains(string(out), "--- SKIP:") {
						t.Fatalf("want named skip: %v\n%s", err, out)
					}
				case "selected":
					if err != nil || strings.Contains(string(out), "--- SKIP:") {
						t.Fatalf("want selected-Node pass: %v\n%s", err, out)
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
