package program_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/microsoft/TypeScript/tsc/shim/vfs/osvfs"

	"github.com/system-inc/cohere/internal/types/program"
)

// memoryFixture is a project that exercises each thing a build asks of its filesystem: the tsconfig and its
// extends, the include glob over nested directories, a relative import, a package resolved from
// node_modules through its package.json, the bundled libs laid over it, and a type error in each of two
// files, so the diagnostics compared below are real ones.
var memoryFixture = map[string]string{
	"tsconfig.base.json":                `{"compilerOptions": {"target": "ES2022", "module": "esnext", "moduleResolution": "bundler", "strict": true, "noEmit": true}}`,
	"tsconfig.json":                     `{"extends": "./tsconfig.base.json", "include": ["src/**/*.ts"]}`,
	"src/main.ts":                       "import { shared } from './nested/shared';\nimport { greet } from 'greeter';\nexport const total: number = shared + greet('x');\nexport const later = Promise.resolve(shared);\n",
	"src/nested/shared.ts":              "export const shared: number = 'not a number';\n",
	"node_modules/greeter/package.json": `{"name": "greeter", "types": "./index.d.ts"}`,
	"node_modules/greeter/index.d.ts":   "export declare function greet(name: string): string;\n",
	"notes/ignored.ts":                  "export const outside: number = 'never included';\n",
}

// inMemory is the fixture's files under a root, by absolute slash path, as a MemoryFS holds them.
func inMemory(root string, files map[string]string) map[string]string {
	memory := map[string]string{}
	for name, contents := range files {
		memory[filepath.ToSlash(filepath.Join(root, name))] = contents
	}
	return memory
}

// buildReport is what a build says: the files it holds, which it calls its own, and every diagnostic with
// its file and position, sorted.
func buildReport(t *testing.T, options program.Options) (files []string, ours []string, diagnostics []string) {
	t.Helper()
	graph, err := program.Build(options)
	if err != nil {
		t.Fatalf("building: %v", err)
	}
	root := options.CurrentDirectory
	relative := func(fileName string) string {
		return strings.TrimPrefix(filepath.ToSlash(fileName), filepath.ToSlash(root)+"/")
	}
	for _, sourceFile := range graph.SourceFiles() {
		files = append(files, relative(sourceFile.FileName().AsString()))
	}
	for _, sourceFile := range graph.ProjectFiles() {
		ours = append(ours, relative(sourceFile.FileName().AsString()))
	}
	background := context.Background()
	for _, diagnostic := range append(graph.AllDiagnostics(background), graph.ConfigDiagnostics(background)...) {
		location := "(no file)"
		if diagnostic.File() != nil {
			location = fmt.Sprintf("%s@%d", relative(diagnostic.File().FileName().AsString()), diagnostic.Loc().Pos())
		}
		diagnostics = append(diagnostics, fmt.Sprintf("%s TS%d", location, diagnostic.Code()))
	}
	slices.Sort(files)
	slices.Sort(ours)
	slices.Sort(diagnostics)
	return files, ours, diagnostics
}

// A program built from a MemoryFS holds the files, calls the same ones its own, and reports the same
// diagnostics at the same positions as the program built from the same files on disk. The control: a
// MemoryFS missing the imported file reports differently, so the comparison is one that can fail, and the
// fixture's two planted type errors are reported by both, so it compares real diagnostics.
func TestAMemoryProgramReportsWhatTheDiskProgramDoes(t *testing.T) {
	t.Parallel()
	// Resolved first: the compiler reads a node_modules package by its real path, and a temporary directory
	// on macOS sits behind the /tmp symbolic link, which nothing in memory is.
	directory, err := filepath.EvalSymlinks(writeProject(t, memoryFixture))
	if err != nil {
		t.Fatal(err)
	}
	diskFiles, diskOurs, diskDiagnostics := buildReport(t, program.Options{ConfigFileName: "tsconfig.json", CurrentDirectory: directory})

	// The memory build sits at the same absolute root, so every file name compares as it is; nothing is
	// read from that directory, which a second root proves below.
	memoryFiles, memoryOurs, memoryDiagnostics := buildReport(t, program.Options{
		ConfigFileName: "tsconfig.json", CurrentDirectory: directory, FileSystem: program.NewMemoryFS(inMemory(directory, memoryFixture)),
	})
	if !slices.Equal(diskFiles, memoryFiles) {
		t.Fatalf("the memory program holds different files:\ndisk   %v\nmemory %v", diskFiles, memoryFiles)
	}
	if !slices.Equal(diskOurs, memoryOurs) || !slices.Equal(diskOurs, []string{"src/main.ts", "src/nested/shared.ts"}) {
		t.Fatalf("the programs' own files differ, or are not the two included:\ndisk   %v\nmemory %v", diskOurs, memoryOurs)
	}
	if !slices.Equal(diskDiagnostics, memoryDiagnostics) {
		t.Fatalf("the memory program reports differently:\ndisk   %v\nmemory %v", diskDiagnostics, memoryDiagnostics)
	}
	if want := 2; len(diskDiagnostics) != want {
		t.Fatalf("the fixture's %d planted type errors came back as %v", want, diskDiagnostics)
	}

	// A root that exists only in memory builds the same report: the build read nothing from disk.
	elsewhere := "/cohere-memory-fixture"
	_, elsewhereOurs, elsewhereDiagnostics := buildReport(t, program.Options{
		ConfigFileName: "tsconfig.json", CurrentDirectory: elsewhere, FileSystem: program.NewMemoryFS(inMemory(elsewhere, memoryFixture)),
	})
	if !slices.Equal(elsewhereOurs, diskOurs) || !slices.Equal(elsewhereDiagnostics, diskDiagnostics) {
		t.Fatalf("a root only memory holds reports differently:\nown %v %v\ndisk %v %v", elsewhereOurs, elsewhereDiagnostics, diskOurs, diskDiagnostics)
	}

	missing := inMemory(directory, memoryFixture)
	delete(missing, filepath.ToSlash(filepath.Join(directory, "src/nested/shared.ts")))
	_, _, missingDiagnostics := buildReport(t, program.Options{
		ConfigFileName: "tsconfig.json", CurrentDirectory: directory, FileSystem: program.NewMemoryFS(missing),
	})
	if slices.Equal(missingDiagnostics, diskDiagnostics) {
		t.Fatalf("control: a memory program missing an imported file reported what the whole one did: %v", missingDiagnostics)
	}
}

// A MemoryFS answers as a directory tree does: each directory lists its files and subdirectories once,
// sorted, the root included, and nothing is there that was not given. Writes fail rather than vanish.
func TestAMemoryFSListsItsTree(t *testing.T) {
	t.Parallel()
	memory := program.NewMemoryFS(map[string]string{
		"/project/a.ts":          "a",
		"/project/sub/b.ts":      "b",
		"/project/sub/c.ts":      "c",
		"/project/sub/deep/d.ts": "d",
		"/top.ts":                "top",
	})
	if entries := memory.GetAccessibleEntries("/"); !slices.Equal(entries.Files, []string{"top.ts"}) || !slices.Equal(entries.Directories, []string{"project"}) {
		t.Errorf("the root lists %+v", entries)
	}
	if entries := memory.GetAccessibleEntries("/project/sub"); !slices.Equal(entries.Files, []string{"b.ts", "c.ts"}) || !slices.Equal(entries.Directories, []string{"deep"}) {
		t.Errorf("a directory with two files and a subdirectory lists %+v", entries)
	}
	if !memory.DirectoryExists("/project/sub/deep") || memory.DirectoryExists("/project/a.ts") || memory.FileExists("/project/sub") {
		t.Error("a file and a directory were confused")
	}
	if contents, found := memory.ReadFile("/project/sub/../a.ts"); !found || contents != "a" {
		t.Errorf("an unclean path read %q, %t", contents, found)
	}
	if _, found := memory.ReadFile("/project/missing.ts"); found || memory.Stat("/project/missing.ts") != nil {
		t.Error("a file that was never given was found")
	}
	if information := memory.Stat("/project/sub/b.ts"); information == nil || information.IsDir() || information.Size() != 1 {
		t.Errorf("a file's stat is %+v", information)
	}
	if err := memory.WriteFile("/project/new.ts", "x"); err == nil || memory.FileExists("/project/new.ts") {
		t.Error("a write to the read-only filesystem succeeded")
	}
}

// MemoryFS reads a file's bytes as the disk does, case by case against osvfs, the filesystem every real build
// reads. A fixture holds bytes as they would sit in a file, and the compiler must see what it would see
// there, so the corpus rows about byte order marks and irregular whitespace mean the same from memory.
func TestMemoryFSReadsBytesAsTheDiskDoes(t *testing.T) {
	t.Parallel()

	directory := t.TempDir()
	for _, testCase := range []struct {
		name  string
		bytes string
	}{
		{"empty", ""},
		{"utf-8", "export const a = 1;\n"},
		{"utf-8 with a byte order mark", "\xEF\xBB\xBFexport const a = 1;\n"},
		{"a byte order mark alone", "\xEF\xBB\xBF"},
		{"utf-16 little endian", "\xFF\xFEa\x00=\x00\xe9\x00\n\x00"},
		{"utf-16 big endian", "\xFE\xFF\x00a\x00=\x00\xe9\x00\n"},
		{"utf-16 with an odd trailing byte", "\xFF\xFEa\x00b"},
		{"utf-16 with a lone surrogate", "\xFF\xFE\x00\xd8a\x00"},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			onDisk := filepath.ToSlash(filepath.Join(directory, strings.ReplaceAll(testCase.name, " ", "-")+".ts"))
			if err := os.WriteFile(onDisk, []byte(testCase.bytes), 0o644); err != nil {
				t.Fatal(err)
			}
			fromDisk, diskOk := osvfs.FS().ReadFile(tspath.RootedFilePath(onDisk))
			fromMemory, memoryOk := program.NewMemoryFS(map[string]string{"/fixture/a.ts": testCase.bytes}).ReadFile("/fixture/a.ts")
			if fromMemory != fromDisk || memoryOk != diskOk {
				t.Fatalf("memory read %q (%v), the disk %q (%v)", fromMemory, memoryOk, fromDisk, diskOk)
			}
		})
	}
}
