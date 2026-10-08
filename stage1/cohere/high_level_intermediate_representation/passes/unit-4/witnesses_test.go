package unit4

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRetainedWitnesses(t *testing.T) {
	t.Parallel()
	root, err := filepath.Abs("../../../../..")
	if err != nil {
		t.Fatal(err)
	}
	lane := filepath.Join(root, "stage1/cohere/high_level_intermediate_representation/passes/unit-4")
	for _, name := range []string{"primitive", "reactive", "effects", "effects-nested"} {
		before := filepath.Join(lane, "testdata", name+".before.checkpoint")
		data, err := os.ReadFile(before)
		if err != nil {
			t.Fatal(err)
		}
		header := strings.Split(strings.Split(string(data), "\n")[1], "\t")
		key := header[1]
		manifest := key + "\t" + before + "\t" + filepath.Join(lane, "testdata", name+".after.checkpoint") + "\n"
		path := filepath.Join(t.TempDir(), "manifest.tsv")
		if err := os.WriteFile(path, []byte(manifest), 0600); err != nil {
			t.Fatal(err)
		}
		entry := "main.a"
		if strings.HasPrefix(name, "effects") {
			entry = "effects_main.a"
		}
		unit4Compare(t, unit4Run(t, root, "node", "--no-warnings", "oracle/node.mjs", filepath.Join(lane, entry), path), manifest)
	}
}
