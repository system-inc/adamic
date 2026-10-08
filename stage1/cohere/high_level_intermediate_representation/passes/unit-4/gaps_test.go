package unit4

import (
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestArrayNeverGap(t *testing.T) {
	t.Parallel()
	root, err := filepath.Abs("../../../../..")
	if err != nil {
		t.Fatal(err)
	}
	for _, fixture := range []struct{ file, want string }{{"array-never.a", "0\n"}, {"array-never-coalesce.a", "1\n"}} {
		entry := filepath.Join(root, "stage1/cohere/high_level_intermediate_representation/passes/unit-4/gaps", fixture.file)
		if got := string(unit4Run(t, root, "node", "--no-warnings", "oracle/node.mjs", entry)); got != fixture.want {
			t.Fatalf("Node %s: %q", fixture.file, got)
		}
		for _, backend := range []string{"build", "js"} {
			args := []string{"run", "./cmd/adamic", backend, entry}
			if backend == "build" {
				args = append(args, "-o", filepath.Join(t.TempDir(), "gap"))
			}
			command := exec.Command("go", args...)
			command.Dir = root
			output, err := command.CombinedOutput()
			if err == nil || !strings.Contains(string(output), "stage 0 can't lower an array of never yet") {
				t.Fatalf("%s %s: gap closed or changed: %v %s", fixture.file, backend, err, output)
			}
			t.Logf("%s %s: %s", fixture.file, backend, strings.TrimSpace(string(output)))
		}
	}
}
