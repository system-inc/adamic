package load

import (
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/core"
)

func TestProjectTypeScriptExtensionImports(t *testing.T) {
	t.Parallel()
	main, err := os.ReadFile("testdata/project-ts-imports/main.a")
	if err != nil {
		t.Fatal(err)
	}
	value, err := os.ReadFile("testdata/project-ts-imports/x.a")
	if err != nil {
		t.Fatal(err)
	}
	// Materialize the Adamic fixture as TypeScript inside an ordinary project.
	paths := writeProgram(t,
		[2]string{"main.ts", strings.ReplaceAll(string(main), "./x.a", "./x.ts")},
		[2]string{"x.ts", string(value)},
		[2]string{"tsconfig.json", `{"compilerOptions":{"composite":true,"strict":true,"noUncheckedIndexedAccess":true,"target":"es2022","lib":["es2022"],"types":[],"module":"nodenext","moduleResolution":"nodenext","allowImportingTsExtensions":false,"noEmit":false},"files":["main.ts","x.ts"]}`},
	)
	for _, roots := range [][]string{paths[:1], paths[:2]} {
		program, err := Load(roots)
		if err != nil {
			t.Fatal(err)
		}
		for _, diagnostic := range program.CompilerProgram().GetProgramDiagnostics() {
			t.Fatalf("invalid project import options: %s", program.formatDiagnostic(diagnostic))
		}
		options := program.CompilerProgram().Options()
		if !options.AllowImportingTsExtensions.IsTrue() || !options.NoEmit.IsTrue() {
			t.Fatalf("Adamic checker import options missing: %+v", options)
		}
		if options.Module != core.ModuleKindNodeNext || options.ModuleResolution != core.ModuleResolutionKindNodeNext || options.Target != core.ScriptTargetES2022 || !reflect.DeepEqual(options.Lib, []string{"lib.es2022.d.ts"}) || len(options.Types) != 0 || !options.Strict.IsTrue() || !options.NoUncheckedIndexedAccess.IsTrue() || !options.Composite.IsTrue() {
			t.Fatalf("project options changed: %+v", options)
		}
		if len(program.Files()) != len(roots) {
			t.Fatal("unrequested implementation roots exposed")
		}
	}
}
