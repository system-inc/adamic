package indexedwitness

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

func TestOptionalReturnAttribution(t *testing.T) {
	data, err := os.ReadFile("D037.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture witness
	if err := json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	directory := t.TempDir()
	path := filepath.Join(directory, "main.ts")
	source := strings.ReplaceAll(fixture.Source, "ABSENT", "true")
	// The standalone audit has only the project libraries, without the loader console overlay.
	source = source[:strings.Index(source, "console.log")]
	write(t, path, source)
	config := filepath.Join(directory, "tsconfig.json")
	write(t, config, `{"compilerOptions":{"strict":true,"noUncheckedIndexedAccess":false,"lib":["es2024"],"module":"esnext","noEmit":true},"files":["main.ts"]}`)
	report, err := load.AuditProjectOptions(context.Background(), config)
	if err != nil {
		t.Fatal(err)
	}
	if len(report.ProjectErrors) != 0 {
		t.Fatalf("ordinary errors: %v", report.ProjectErrors)
	}
	attributed := false
	for _, site := range report.Sites {
		if site.Line == 11 {
			for _, option := range site.Options {
				if option == "noUncheckedIndexedAccess" {
					attributed = true
				}
			}
		}
	}
	if !attributed {
		t.Fatalf("optional result must retain indexed option attribution: %+v", report.Sites)
	}
	// Removing the overloads makes absence an ordinary declared return contract.
	source = strings.ReplaceAll(source, "function getSourceFileOfNode(node: Declaration): SourceFile;\n", "")
	source = strings.ReplaceAll(source, "function getSourceFileOfNode(node: Declaration | undefined): SourceFile | undefined;\n", "")
	write(t, path, source)
	if _, err := load.Load([]string{path}); err == nil || !strings.Contains(err.Error(), "not assignable") {
		t.Fatalf("ordinary optional return must remain an error: %v", err)
	}
}

func TestDifferentOverloadStorageUsesAreaAdapter(t *testing.T) {
	directory := t.TempDir()
	path := filepath.Join(directory, "main.ts")
	write(t, filepath.Join(directory, "tsconfig.json"), `{"compilerOptions":{"strict":true,"noUncheckedIndexedAccess":false,"lib":["es2024"],"module":"esnext","noEmit":true},"files":["main.ts"]}`)
	write(t, path, "function read(value: number): number;\nfunction read(value: number | undefined): number | undefined;\nfunction read(value: number | undefined): number | undefined { return value; }\nconsole.log(`${read(7)}`);\n")
	checked, err := load.Load([]string{path})
	if err != nil {
		t.Fatal(err)
	}
	program, err := lower.Lower(context.Background(), checked)
	if err != nil {
		t.Fatal(err)
	}
	binary := filepath.Join(directory, "native")
	if err := native.Build(native.C(program), binary, native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	expected := run("node", "--eval", "function read(value) { return value; } console.log(`${read(7)}`);")
	if expected.code != 0 || expected.stdout != "7\n" {
		t.Fatalf("Node oracle: %+v", expected)
	}
	if got := run(binary); got != expected {
		t.Fatalf("area overload adapter: %+v, want %+v", got, expected)
	}
}
