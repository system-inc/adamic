package high_level_intermediate_representation

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFileFunctionCacheOracle(t *testing.T) {
	t.Parallel()
	root, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	lane := filepath.Join(root, "stage1/cohere/high_level_intermediate_representation")
	temp := t.TempDir()
	upstream := filepath.Join(root, "cohere/internal/lint/ecmascript/high_level_intermediate_representation")
	replacements := map[string]string{
		filepath.Join(upstream, "stage1_hir_checkpoint_test.go"): filepath.Join(lane, "replay/oracle_test.go"),
		filepath.Join(upstream, "stage1_hir_inputs_test.go"):     filepath.Join(lane, "replay/inputs_test.go"),
	}
	for _, name := range []string{"oracle", "census", "cache"} {
		replacements[filepath.Join(upstream, "stage1_hir_"+name+"_test.go")] = filepath.Join(lane, "testdata", name+"_test.go")
	}
	provider, err := os.ReadFile(filepath.Join(root, "bridge/tsgo/checker/symbol_graph.go"))
	if err != nil {
		t.Fatal(err)
	}
	providerPath := filepath.Join(temp, "symbol_graph_test.go")
	if err := os.WriteFile(providerPath, []byte(strings.Replace(string(provider), "package checker", "package high_level_intermediate_representation", 1)), 0600); err != nil {
		t.Fatal(err)
	}
	replacements[filepath.Join(upstream, "stage1_hir_symbol_graph_test.go")] = providerPath

	data, err := json.Marshal(map[string]any{"Replace": replacements})
	if err != nil {
		t.Fatal(err)
	}
	overlay := filepath.Join(temp, "overlay.json")
	if err := os.WriteFile(overlay, data, 0600); err != nil {
		t.Fatal(err)
	}
	command(t, filepath.Join(root, "cohere"), []string{"GOWORK=" + filepath.Join(root, "cohere/go.work"), "HIR_CACHE_ORACLE=" + temp}, "go", "test", "-v", "-count=1", "-tags=lintoracle", "-overlay", overlay, "-run=^TestStage1FileCacheOracle$", "./internal/lint/ecmascript/high_level_intermediate_representation")
	want, err := os.ReadFile(filepath.Join(temp, "go.dump"))
	if err != nil {
		t.Fatal(err)
	}
	manifest := filepath.Join(temp, "manifest.tsv")
	driver := filepath.Join(lane, "cache_main.ts")
	node := command(t, root, nil, "node", "--no-warnings", "oracle/node.mjs", driver, manifest)
	binary := filepath.Join(temp, "cache")
	command(t, root, nil, "go", "run", "./cmd/adamic", "build", driver, "-o", binary)
	native := command(t, root, nil, binary, manifest)
	for name, got := range map[string][]byte{"Node": node, "native": native} {
		if !bytes.Equal(got, want) {
			t.Fatalf("%s cache oracle differs: %s", name, firstDifference(got, want))
		}
	}
	config := filepath.Join(temp, "tsconfig.json")
	if err := os.WriteFile(config, []byte(`{"compilerOptions":{"strict":true,"target":"ES2022","jsx":"preserve","moduleDetection":"force","types":[]}}`), 0600); err != nil {
		t.Fatal(err)
	}
	archive := filepath.Join(temp, "tsgo.a")
	command(t, root, nil, "go", "build", "-buildmode=c-archive", "-o", archive, "./bridge/tsgo/archive")
	live := filepath.Join(temp, "live-cache")
	command(t, root, nil, "go", "run", "./cmd/adamic", "build", filepath.Join(lane, "cache_live_main.ts"), "-o", live, "--tsgo", archive)
	if got := command(t, root, nil, live, config, manifest); !bytes.Equal(got, want) {
		t.Fatalf("live resident-checker ForFunction differs: %s", firstDifference(got, want))
	}
	t.Log("live native resident-checker ForFunction matches Go without a replay snapshot")

	t.Logf("%d function cache entries match Go, repeated identities and checker gating", strings.Count(string(want), "identity 1 unchecked 1"))
	// A fresh lowering on each hit executes successfully but loses Go's object identity.
	mutant, err := os.MkdirTemp(filepath.Dir(lane), "hir-cache-mutant-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(mutant)
	files, err := filepath.Glob(filepath.Join(lane, "*.ts"))
	if err != nil {
		t.Fatal(err)
	}
	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		if filepath.Base(file) == "cache.ts" {
			from := "if(this.functions.has(key)) { return this.functions.get(key); }"
			if strings.Count(string(data), from) != 1 {
				t.Fatal("cache mutant anchor moved")
			}
			data = []byte(strings.Replace(string(data), from, "if(this.functions.has(key)) { return lowerParsedFunction(this.parser, index, this.source, this.symbols); }", 1))
		}
		if err := os.WriteFile(filepath.Join(mutant, filepath.Base(file)), data, 0600); err != nil {
			t.Fatal(err)
		}
	}
	driver = filepath.Join(mutant, "cache_main.ts")
	node = command(t, root, nil, "node", "--no-warnings", "oracle/node.mjs", driver, manifest)
	binary = filepath.Join(temp, "cache-mutant")
	command(t, root, nil, "go", "run", "./cmd/adamic", "build", driver, "-o", binary)
	native = command(t, root, nil, binary, manifest)
	if !bytes.Equal(node, native) || bytes.Equal(node, want) || !bytes.Contains(node, []byte("identity 0 unchecked 1")) {
		t.Fatal("cache identity mutant survived or differed between backends")
	}
	t.Log("fresh-lowering cache-hit mutant executes and disagrees with Go on both backends")
}
