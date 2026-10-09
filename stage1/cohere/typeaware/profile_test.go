package typeaware

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func profileSourceMutant(h *harness, stage0, archive, name, from, to string) string {
	h.t.Helper()
	directory := filepath.Join(h.directory, name+"-source")
	if err := os.MkdirAll(directory, 0755); err != nil {
		h.t.Fatal(err)
	}
	sourceDirectory := filepath.Join(h.repository, "stage1/cohere/typeaware")
	files, err := filepath.Glob(filepath.Join(sourceDirectory, "*.ts"))
	if err != nil {
		h.t.Fatal(err)
	}
	for _, path := range files {
		data, err := os.ReadFile(path)
		if err != nil {
			h.t.Fatal(err)
		}
		source := string(data)
		if filepath.Base(path) == "shadow.ts" {
			if strings.Count(source, from) != 1 {
				h.t.Fatalf("nonunique mutant %s", name)
			}
			source = strings.Replace(source, from, to, 1)
		}
		source = strings.ReplaceAll(source, "../../typescript/", filepath.Join(h.repository, "stage1/typescript")+"/")
		source = strings.ReplaceAll(source, "../lint/", filepath.Join(h.repository, "stage1/cohere/lint")+"/")
		if err := os.WriteFile(filepath.Join(directory, filepath.Base(path)), []byte(source), 0600); err != nil {
			h.t.Fatal(err)
		}
	}
	return h.build(stage0, name, filepath.Join(directory, "volume_suite.ts"), archive, false)
}

func TestShadowIndexMissingBinding_000(t *testing.T) {
	t.Parallel()
	started := time.Now()
	h := volumeProfileHarness(t)
	stage0 := volumeProfileStage0(h)
	archive := h.archive("checker", "", false)
	t.Logf("setup: %.3fs", time.Since(started).Seconds())
	mutant := volumeProfileNative(h, stage0, archive, false, "typeaware volume missing-binding namedScopes wrong", func(local *harness) string {
		return profileSourceMutant(local, stage0, archive, "missing-binding", "this.namedScopes.set(binding.name, fresh);", "this.namedScopes.set('wrong', fresh);")
	}, "this.namedScopes.set(binding.name, fresh);", "this.namedScopes.set('wrong', fresh);")
	source := h.write("input.ts", "const x=1;\nexport {};\n")
	manifest := h.write("manifest", source+"\n")
	got := h.run("missing-binding-run", exec.Command(mutant, filepath.Join(h.repository, "stage1/cohere/typeaware/testdata/tsconfig.json"), manifest))
	if code, ok := got.err.(*exec.ExitError); !ok || code.ExitCode() != 70 || !bytes.Contains(got.stderr, []byte("missing binding index")) {
		t.Fatalf("missing binding index was not refused: %v %s", got.err, got.stderr)
	}
	t.Log("omitted binding key: compiled mutant panics 70, missing binding index; cooked=false")
}
