package tsprinter

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/stage1/cohere/tsprinter/oracles"

	"github.com/system-inc/adamic/internal/load"
)

// A family's lowered product names everything its lowering reads: every repository source the loader reads for main.ts
// and statementsMain.ts, imports' imports included, is under printerGateLoweredFiles, and nothing it declares holds
// cohere/TypeScript/tsc/testdata or names cohere whole. A bundled library and the prelude are compiled into the binary,
// so the compiler's inputs key them. Keying builds nothing. Mutant: stage1/typescript dropped from the files (the
// parser and scanner the port imports go unkeyed).
func TestPrinterGateLoweredKeyCoversWhatItReads(t *testing.T) {
	t.Parallel()
	root, err := filepath.Abs(oracles.Repository)
	if err != nil {
		t.Fatal(err)
	}
	declared := printerGateLoweredFiles()
	for _, input := range declared {
		if baselines := "cohere/TypeScript/tsc/testdata"; baselines == input || strings.HasPrefix(baselines, input+"/") {
			t.Fatalf("the lowered products' key declares %s, which holds %s", input, baselines)
		}
	}
	for _, entry := range []string{"main.ts", "statementsMain.ts"} {
		program, err := load.Load([]string{filepath.Join(root, "stage1/cohere/tsprinter", entry)})
		if err != nil {
			t.Fatalf("%s: %v", entry, err)
		}
		sources, compiledIn := map[string]bool{}, 0
		for _, file := range program.CompilerProgram().GetSourceFiles() {
			name := string(file.FileName())
			if strings.HasPrefix(name, "bundled://") || strings.HasPrefix(name, "/adamic-prelude/") {
				compiledIn++
				continue
			}
			relative, err := filepath.Rel(root, name)
			if err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
				t.Fatalf("%s reads %s, outside the repository", entry, name)
			}
			sources[filepath.ToSlash(relative)] = true
		}
		if gaps := compilerInputGaps(declared, sources); len(gaps) > 0 {
			t.Fatalf("lowering %s reads files the lowered product's key misses: %s", entry, strings.Join(gaps, ", "))
		}
		t.Logf("%s: %d repository source files, each keyed, and %d compiled in", entry, len(sources), compiledIn)
	}
}
