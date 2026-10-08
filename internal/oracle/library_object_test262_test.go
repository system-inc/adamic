package oracle

import (
	"bufio"
	"compress/gzip"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// Preserve the adapted test262 programs, not just the aggregate pass count.
// Each added pass must also agree in the release and JavaScript backends and
// finish without leaks. These captures do not join the positive fixture glob.
func TestObjectNewlyPassingProgramsAgreeWithNode(t *testing.T) {
	file, err := os.Open(filepath.Join(repository, "docs/library-object/unit3/new-passes.jsonl.gz"))
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	reader, err := gzip.NewReader(file)
	if err != nil {
		t.Fatal(err)
	}
	defer reader.Close()
	scanner := bufio.NewScanner(reader)
	scanner.Buffer(make([]byte, 4096), 4<<20)
	count := 0
	for scanner.Scan() {
		var capture struct {
			Path    string `json:"path"`
			Program string `json:"program"`
		}
		if err := json.Unmarshal(scanner.Bytes(), &capture); err != nil {
			t.Fatal(err)
		}
		count++
		t.Run(capture.Path, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "program.a")
			if err := os.WriteFile(path, []byte(capture.Program), 0600); err != nil {
				t.Fatal(err)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			node := onNode(t, path)
			sanitized, binary := natively(t, program)
			if difference := disagreement(node, sanitized); difference != "" {
				t.Errorf("native: %s: Node stdout %q native stdout %q", difference, node.stdout, sanitized.stdout)
			}
			if release := released(t, program); disagreement(node, release) != "" {
				t.Errorf("release: %s", disagreement(node, release))
			}
			if javascript := onJavaScriptBackend(t, program); disagreement(node, javascript) != "" {
				t.Errorf("JavaScript: %s: Node stdout %q backend stdout %q", disagreement(node, javascript), node.stdout, javascript.stdout)
			}
			if node.exitCode == 0 {
				if leaked := leaks(t, program, binary); leaked != "" {
					t.Errorf("leaks: %s", leaked)
				}
			}
		})
	}
	if err := scanner.Err(); err != nil {
		t.Fatal(err)
	}
	if count == 0 {
		t.Fatal("no newly passing programs were checked")
	}
	t.Logf("%d newly passing programs compared with Node in all backends", count)
}
