package lower

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"unicode/utf16"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/bundled"
	"github.com/microsoft/TypeScript/tsc/shim/compiler"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/microsoft/TypeScript/tsc/shim/tsoptions"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/microsoft/TypeScript/tsc/shim/vfs/cachedvfs"
	"github.com/microsoft/TypeScript/tsc/shim/vfs/osvfs"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
)

// Descriptor admission and production compilation are distinct measurements.
// This audit never bypasses load diagnostics or exposes partial executable IR.
func TestLazyViewCensus(t *testing.T) {
	root := os.Getenv("LAZY_VIEW_CENSUS_ROOT")
	if root == "" {
		t.Skip("set LAZY_VIEW_CENSUS_ROOT, LAZY_VIEW_CENSUS_SEEDS and LAZY_VIEW_CENSUS_OUTPUT")
	}
	input, err := os.ReadFile(os.Getenv("LAZY_VIEW_CENSUS_SEEDS"))
	if err != nil {
		t.Fatal(err)
	}
	type site struct {
		File, Text, Kind string
		Start, End       int
	}
	var seeds []site
	if err := json.Unmarshal(input, &seeds); err != nil {
		t.Fatal(err)
	}
	if len(seeds) != 2936 {
		t.Fatalf("want 2936 exact sites, got %d", len(seeds))
	}
	selected := map[string]map[[2]int]site{}
	kinds := map[string]int{}
	for _, seed := range seeds {
		if selected[seed.File] == nil {
			selected[seed.File] = map[[2]int]site{}
		}
		selected[seed.File][[2]int{seed.Start, seed.End}] = seed
		kinds[seed.Kind]++
	}
	configPath := filepath.Join(root, "src/compiler/tsconfig.json")
	config, diagnostics := tsoptions.GetParsedCommandLineOfConfigFile(tspath.RootedFilePathFromAbsolute(configPath), nil, nil, osvfs.FS(), nil)
	if config == nil || len(diagnostics) != 0 || len(config.Errors) != 0 {
		t.Fatal("config diagnostics", diagnostics)
	}
	fs := cachedvfs.From(bundled.WrapFS(osvfs.FS()))
	checked := compiler.NewProgram(compiler.ProgramOptions{Config: config, Host: compiler.NewCachedFSCompilerHost(fs, bundled.LibPath(), nil, nil, nil)})
	scratch := filepath.Join(t.TempDir(), "formatter.a")
	if err := os.WriteFile(scratch, []byte("export {};"), 0600); err != nil {
		t.Fatal(err)
	}
	formatter, err := load.Load([]string{scratch})
	if err != nil {
		t.Fatal(err)
	}
	interned := map[string]int{}
	matched := 0
	var roots []string
	for _, file := range checked.GetSourceFiles() {
		if file.IsDeclarationFile || !strings.HasPrefix(string(file.FileName()), root+"/") {
			continue
		}
		roots = append(roots, string(file.FileName()))
		relative, err := filepath.Rel(root, string(file.FileName()))
		if err != nil {
			t.Fatal(err)
		}
		positions := selected[filepath.ToSlash(relative)]
		if len(positions) == 0 {
			continue
		}
		// The immutable JavaScript ledger uses UTF-16 offsets; the native parser
		// uses bytes. Match both endpoints after converting at rune boundaries.
		offsets := map[int]int{}
		units := 0
		for position, character := range file.Text() {
			offsets[position] = units
			units += utf16.RuneLen(character)
		}
		offsets[len(file.Text())] = units
		checker, release := checked.GetTypeCheckerForFile(context.Background(), file)
		audit := &lowering{checker: checker, program: formatter, result: &ir.Program{}}
		var visit ast.Visitor
		visit = func(node *ast.Node) bool {
			if node.Kind == ast.KindAsExpression {
				position := scanner.GetTokenPosOfNode(node, file, false)
				if seed, ok := positions[[2]int{offsets[position], offsets[node.End()]}]; ok {
					if file.Text()[position:node.End()] != seed.Text {
						t.Fatalf("text changed at %s:%d", relative, position)
					}
					_, err := audit.viewSchema(node, checker.GetTypeAtLocation(node))
					if err != nil {
						t.Fatalf("descriptor admission %s:%d: %v", relative, position, err)
					}
					interned[seed.Kind]++
					matched++
				}
			}
			node.ForEachChild(visit)
			return false
		}
		file.AsNode().ForEachChild(visit)
		release()
	}
	if matched != len(seeds) {
		t.Fatalf("matched %d of %d", matched, len(seeds))
	}
	loaded, failure := load.Load(roots)
	compiled := map[string]int{"tagged": 0, "untagged": 0}
	stage := "checker"
	if failure == nil {
		stage = "lowering"
		_, failure = Lower(context.Background(), loaded)
		if failure == nil {
			compiled = kinds
			stage = "compiled"
		}
	}
	witness := ""
	if failure != nil {
		witness = failure.Error()
		if len(witness) > 16000 {
			witness = witness[:16000]
		}
	}
	output := struct {
		Sites                      int            `json:"sites"`
		DescriptorInterned         map[string]int `json:"descriptor_interned"`
		ProductionCompiled         map[string]int `json:"production_compiled"`
		ProductionStop, Diagnostic string
		Limits                     string
	}{matched, interned, compiled, stage, witness, "Descriptors are admission metadata, not successful cast lowering. Production counts require the unchanged whole compiler project to pass Adamic loading and lowering. Family counts are separate stock read-demand witnesses, not allocation-flow proofs."}
	data, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(os.Getenv("LAZY_VIEW_CENSUS_OUTPUT"), append(data, '\n'), 0600); err != nil {
		t.Fatal(err)
	}
	t.Logf("sites=%d descriptor interned=%v production compiled=%v stop=%s", matched, interned, compiled, stage)
}
