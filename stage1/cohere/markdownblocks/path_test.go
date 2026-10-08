package markdownblocks

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/system-inc/adamic/internal/native"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

// Parallel execution reserves memory: this compares full per-node path observations and sanitized mutant output.
func TestMarkdownAstPath(t *testing.T) {
	parallelMarkdownMemory(t, 5)
	root, err := filepath.Abs(repository)
	if err != nil {
		t.Fatal(err)
	}
	inputs, files := blockCorpus(t, root, "whitespace")
	// This file sweep is small beside the mandatory generated and fixed checks.
	selection := selectMarkdownFiles(t, root, inputs, files, 1)
	if selection.Sample {
		t.Log(selection.Log(t.Name()))
	}
	inputs, files = selectedMarkdownInputs(inputs, files, selection)
	for depth := 1; depth <= 32; depth++ {
		text := strings.Repeat("> ", depth) + "- a *b*\n"
		inputs = append(inputs, auditInput{Name: fmt.Sprintf("generated/path/depth%d", depth), Text: text})
	}
	dir := t.TempDir()
	var batch bytes.Buffer
	encoder := json.NewEncoder(&batch)
	for _, item := range inputs {
		if err := encoder.Encode(item); err != nil {
			t.Fatal(err)
		}
	}
	cases := filepath.Join(dir, "cases.jsonl")
	write(t, cases, batch.Bytes())
	cohere := filepath.Join(root, "cohere")
	mainPath := filepath.Join(cohere, "cmd/adamic_path/main.go")
	driver, err := filepath.Abs("testdata/path_go.go")
	if err != nil {
		t.Fatal(err)
	}
	bridge, err := filepath.Abs("testdata/path_bridge.go")
	if err != nil {
		t.Fatal(err)
	}
	astBridge, err := filepath.Abs("testdata/ast_bridge.go")
	if err != nil {
		t.Fatal(err)
	}
	pathFacts, err := filepath.Abs("testdata/path_facts.go")
	if err != nil {
		t.Fatal(err)
	}
	overlay, err := json.Marshal(map[string]any{"Replace": map[string]string{mainPath: driver, filepath.Join(cohere, "internal/format/markdown/adamic_path_facts.go"): pathFacts, filepath.Join(cohere, "internal/format/markdown/adamic_path.go"): bridge, filepath.Join(cohere, "internal/format/markdown/adamic_ast.go"): astBridge}})
	if err != nil {
		t.Fatal(err)
	}
	overlayPath := filepath.Join(dir, "overlay.json")
	write(t, overlayPath, overlay)
	goBinary := filepath.Join(dir, "go-path")
	build := bounded(t, "go", "build", "-overlay="+overlayPath, "-o", goBinary, mainPath)
	build.Dir = cohere
	if output, err := combinedOutput(build); err != nil {
		t.Fatalf("Go path %v %s", err, output)
	}
	nativeCases := filepath.Join(dir, "native.txt")
	want := pathExecute(t, nil, goBinary, cases, nativeCases)
	clean(t, "actual Go AstPath", want.run)
	if keep := os.Getenv("ADAMIC_PATH_KEEP"); keep != "" {
		if err := os.MkdirAll(keep, 0755); err != nil {
			t.Fatal(err)
		}
		data, err := os.ReadFile(nativeCases)
		if err != nil {
			t.Fatal(err)
		}
		write(t, filepath.Join(keep, "native.txt"), data)
		copyPathOutput(t, want.file, filepath.Join(keep, "want.txt"))
	}
	fork := os.Getenv("ADAMIC_MARKDOWNBLOCKS_FORK")
	if fork == "" {
		fork = filepath.Join(cohere, "internal/format/prettier/bundles")
	}
	installed, err := os.ReadFile(filepath.Join(fork, "standalone.js"))
	if err != nil {
		t.Fatal(err)
	}
	pinned, err := os.ReadFile(filepath.Join(cohere, "internal/format/prettier/bundles/standalone.js"))
	if err != nil {
		t.Fatal(err)
	}
	equal(t, "pinned standalone", installed, pinned)
	gapGo := execute(t, nil, goBinary, "--key-gap")
	clean(t, "Go key witness", gapGo)
	equal(t, "Go key witness", gapGo.stdout, []byte("key=\"\" present=false\n"))
	gapFork := execute(t, nil, "node", "testdata/path_library.mjs", fork, "--key-gap")
	clean(t, "fork key witness", gapFork)
	equal(t, "fork key witness", gapFork.stdout, []byte("key=0 type=number\n"))
	main, err := filepath.Abs("testdata/path_probe.ts")
	if err != nil {
		t.Fatal(err)
	}
	program := lowered(t, main)
	binary := nativeBinary(t, native.C(program), true)
	var environment []string
	if runtime.GOOS == "linux" {
		environment = []string{"ASAN_OPTIONS=detect_leaks=0"}
	}
	answer := pathExecute(t, environment, binary, nativeCases)
	// Compare runners sequentially so their complete outputs do not accumulate together.
	for _, side := range []struct {
		name string
		run  func() pathOutput
	}{
		{"actual Go AstPath from transport", func() pathOutput { return pathExecute(t, nil, goBinary, "--facts", nativeCases) }},
		{"actual original AstPath", func() pathOutput { return pathExecute(t, nil, "node", "testdata/path_library.mjs", fork, nativeCases) }},
		{"native", func() pathOutput { return answer }},
		{"source Node", func() pathOutput { return pathNode(t, main, nativeCases) }},
		{"backend", func() pathOutput { return pathBackend(t, program, nativeCases) }},
	} {
		result := side.run()
		clean(t, side.name, result.run)
		pathEqual(t, side.name, result.file, want.file, inputs)
		removePathOutput(t, result.file)
	}
	pathLeaks(t, program, binary, nativeCases)

	for _, m := range []struct{ name, from, to string }{
		{"map callback result", "results.push(callback(path, index, array));", "results.push(callback(path, index, array) + 1);"},
		{"array frame", "this.at(-3)", "this.at(-5)"},
		{"ancestor order", "index -= 2", "index -= 4"},
		{"call restore", "this.truncate(length);\n        return result;", "this.truncate(length + 2);\n        return result;"},
	} {
		t.Run(m.name, func(t *testing.T) {
			scratch := t.TempDir()
			if err := os.Mkdir(filepath.Join(scratch, "testdata"), 0755); err != nil {
				t.Fatal(err)
			}
			for _, file := range []string{"astArena.ts", "astWalk.ts", "astProtocol.ts", "codec.ts", "astPath.ts", "pathObserve.ts", "testdata/path_probe.ts"} {
				data, err := os.ReadFile(file)
				if err != nil {
					t.Fatal(err)
				}
				if file == "astPath.ts" {
					if !strings.Contains(string(data), m.from) {
						t.Fatal("mutation anchor")
					}
					data = []byte(strings.Replace(string(data), m.from, m.to, 1))
				}
				write(t, filepath.Join(scratch, file), data)
			}
			result := pathNode(t, filepath.Join(scratch, "testdata/path_probe.ts"), nativeCases)
			clean(t, m.name, result.run)
			offset, _, equal := pathDifference(t, result.file, want.file)
			if equal {
				t.Fatal("survived")
			}
			t.Logf("caught by path bytes, first difference%d", offset)
			removePathOutput(t, result.file)
		})
	}
	fast := filepath.Join(dir, "fast")
	if err := native.Build(native.C(program), fast, native.Options{}); err != nil {
		t.Fatal(err)
	}
	for _, side := range []struct {
		name, command string
		args          []string
	}{
		{"actual Go AstPath", goBinary, []string{"--facts", nativeCases}},
		{"native AstPath", fast, []string{nativeCases}},
		{"actual original AstPath", "node", []string{"testdata/path_library.mjs", fork, nativeCases}},
	} {
		start := time.Now()
		for i := 0; i < 3; i++ {
			result := pathExecute(t, nil, side.command, side.args...)
			clean(t, side.name, result.run)
			pathEqual(t, side.name, result.file, want.file, inputs)
			removePathOutput(t, result.file)
		}
		t.Logf("%s %.1f documents/s, three runs, full path observation I/O, AST parsing excluded", side.name, float64(3*len(inputs))/time.Since(start).Seconds())
	}
	t.Logf("%d contexts, %d physical files, all per-node getter bytes and parent-node call/map/each/parent restore bytes agree; %d output bytes", len(inputs), files, pathOutputSize(t, want.file))
}
