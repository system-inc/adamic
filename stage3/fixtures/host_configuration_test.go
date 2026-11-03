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

func TestHostConfigurationFixtureReaders(t *testing.T) {
	t.Parallel()
	repository, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	node, err := exec.LookPath("node")
	if err != nil {
		t.Fatal(err)
	}
	original, err := os.ReadFile("host/status.json")
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
	for _, reader := range []string{"execute", "manifest", "directories"} {
		for _, mode := range []string{"unset", "selected", "unknown"} {
			t.Run(reader+"/"+mode, func(t *testing.T) {
				root := t.TempDir()
				// Metadata readers see the whole union, including configured entries.
				err := filepath.WalkDir(".", func(path string, entry fs.DirEntry, err error) error {
					if err != nil {
						return err
					}
					target := filepath.Join(root, path)
					if entry.IsDir() {
						return os.MkdirAll(target, 0700)
					}
					if entry.Name() != "status.json" && filepath.Ext(path) != ".a" && filepath.Ext(path) != ".ts" {
						return nil
					}
					data, err := os.ReadFile(path)
					if err != nil {
						return err
					}
					return os.WriteFile(target, data, 0600)
				})
				if err != nil {
					t.Fatal(err)
				}
				var rows []map[string]json.RawMessage
				if err := json.Unmarshal(original, &rows); err != nil {
					t.Fatal(err)
				}
				rows[0]["configuration"] = model[0]["configuration"]
				if mode == "unknown" {
					rows[0]["configuration"] = json.RawMessage(`{"name":"unknown-build"}`)
				}
				encoded, err := json.Marshal(rows)
				if err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filepath.Join(root, "host/status.json"), encoded, 0600); err != nil {
					t.Fatal(err)
				}
				selector := "^TestFixturesHost$/^host$/^01_readFile_utf8.a$"
				if reader == "manifest" {
					selector = "^TestFixtureShardManifest$"
				}
				if reader == "directories" {
					selector = "^TestFixtureDirectoriesHaveTopLevelTests$"
				}
				cmd := exec.Command(executable, "-test.run="+selector, "-test.count=1", "-test.v", "-test.timeout=30s", "-fixtures="+root)
				selected := ""
				if mode == "selected" {
					selected = node
				}
				cmd.Env = append(os.Environ(), "ADAMIC_HOST_NODE_WITHOUT_INSPECTOR="+selected)
				out, err := cmd.CombinedOutput()
				t.Logf("%s", out)
				if mode == "unknown" {
					if err == nil || !strings.Contains(string(out), "unknown host configuration") {
						t.Fatalf("want unknown-name failure: %v\n%s", err, out)
					}
					return
				}
				if err != nil {
					t.Fatalf("configured reader: %v\n%s", err, out)
				}
				if mode == "unset" && !strings.Contains(string(out), "row 1 needs node-without-inspector; set ADAMIC_HOST_NODE_WITHOUT_INSPECTOR") {
					t.Fatalf("missing named skip:\n%s", out)
				}
				if mode == "selected" && strings.Contains(string(out), "needs node-without-inspector") {
					t.Fatalf("selected recipe skipped:\n%s", out)
				}
				if reader == "manifest" && !strings.Contains(string(out), "host/01_readFile_utf8.a shard=") {
					t.Fatalf("metadata dropped configured row:\n%s", out)
				}
			})
		}
	}
}
