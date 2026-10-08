package checker

import (
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestGlobalSource(t *testing.T) {
	for _, row := range []struct {
		name, extension, source, strict string
		want                            []string
	}{
		{"script", ".ts", "var x=1;", "true", []string{"1", "global-source", "0", "0", "1"}},
		{"module", ".ts", "export {};", "true", []string{"1", "global-source", "1", "0", "1"}},
		{"sloppy-option", ".ts", "var x=1;", "false", []string{"1", "global-source", "0", "0", "0"}},
		{"javascript", ".js", "var x=1;", "true", []string{"1", "global-source", "0", "1", "1"}},
	} {
		t.Run(row.name, func(t *testing.T) {
			directory := t.TempDir()
			config, file := filepath.Join(directory, "tsconfig.json"), filepath.Join(directory, "input"+row.extension)
			for path, text := range map[string]string{config: `{"compilerOptions":{"strict":` + row.strict + `,"allowJs":true,"moduleDetection":"auto"}}`, file: row.source} {
				if err := os.WriteFile(path, []byte(text), 0600); err != nil {
					t.Fatal(err)
				}
			}
			p, err := Open(config, []string{file})
			if err != nil {
				t.Fatal(err)
			}
			sf := p.Compiler.GetSourceFile(tspath.RootedFilePathFromAbsolute(file))
			wire, err := p.Inspect(file, uint64(sf.Pos()), uint64(sf.End()), "SourceFile", "global-source")
			if err != nil {
				t.Fatal(err)
			}
			if got := decodedFields(t, wire); !reflect.DeepEqual(got, row.want) {
				t.Fatalf("got %q want %q", got, row.want)
			}
			if _, err := p.Inspect(file, uint64(sf.Pos()), uint64(sf.End()), "SourceFile", "global-source\nextra"); err == nil {
				t.Fatal("accepted global-source suffix")
			}
			identifier := sf.Statements.Nodes[0]
			if _, err := p.Inspect(file, uint64(identifier.Pos()), uint64(identifier.End()), identifier.Kind.String()[4:], "global-source"); err == nil {
				t.Fatal("accepted non-source-file global-source")
			}
		})
	}
}
