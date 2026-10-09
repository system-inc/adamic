package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"
	"unsafe"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/system-inc/cohere/internal/format/formatoptions"
	"github.com/system-inc/cohere/internal/format/native"
	"github.com/system-inc/cohere/internal/types/program"
)

// The speculation formats a file from the program's own copy, reading nothing, when the content pack's stat
// says that copy is still the file's, and reads the disk when it is not (#q6dey77). Both halves on a real
// program read through a real pack:
//
// An unchanged file's kept result is the program's text itself, the same bytes in the same memory, so it was
// never read again. Rereading every file to compare it with the program's was 51 MB on a cold ahra run.
//
// A file edited after the program read it, the walk's read, is read from the disk, its result is of the new
// bytes, and keepable discards it, since those are not the bytes the walk's findings are about. The edit is
// formatted, so the speculation keeps whichever copy it formats and only the stat can tell it the program's
// copy is stale: without the stat check, the stale copy is kept as formatted and the edit goes unnoticed.
//
// Both with a content pack and without one: a `--no-cache` run has none, and the program then keeps the stats
// it read under itself. The 51 MB was measured on that run.
func TestTheSpeculationFormatsTheProgramsCopyOnlyWhileItIsStillTheFiles(t *testing.T) {
	t.Parallel()
	for _, withPack := range []bool{true, false} {
		t.Run(map[bool]string{true: "with a content pack", false: "without one"}[withPack], func(t *testing.T) {
			t.Parallel()
			speculationFormatsTheProgramsCopyOnlyWhileItIsStillTheFiles(t, withPack)
		})
	}
}

func speculationFormatsTheProgramsCopyOnlyWhileItIsStillTheFiles(t *testing.T, withPack bool) {
	// Resolved, so the names are the program's own: on macOS the temporary directory is behind a symbolic link.
	directory, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	unchanged := filepath.Join(directory, "Unchanged.ts")
	edited := filepath.Join(directory, "Edited.ts")
	for _, fileName := range []string{unchanged, edited} {
		if err := os.WriteFile(fileName, []byte("export const alpha = 1;\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	// Unchanged too, but behind a UTF-8 byte order mark, which the program's copy has dropped (#hkv1hgp).
	marked := filepath.Join(directory, "Marked.ts")
	if err := os.WriteFile(marked, append([]byte{0xEF, 0xBB, 0xBF}, "export const beta = 1;\n"...), 0o644); err != nil {
		t.Fatal(err)
	}
	configuration := `{"compilerOptions":{"strict":true,"module":"esnext","target":"esnext","moduleResolution":"bundler"},"include":["**/*.ts"]}`
	if err := os.WriteFile(filepath.Join(directory, "tsconfig.json"), []byte(configuration), 0o644); err != nil {
		t.Fatal(err)
	}
	var pack *program.ContentPack
	if withPack {
		if pack, err = program.OpenContentPack(t.TempDir()); err != nil {
			t.Fatal(err)
		}
	}
	graph, err := program.Build(program.Options{ConfigFileName: "tsconfig.json", CurrentDirectory: directory,
		ContentPack: pack})
	if err != nil {
		t.Fatal(err)
	}
	// Bound, as the speculation requires of a tree before it takes one.
	_, release := graph.Program.GetTypeCheckerForFile(t.Context(), graph.Program.GetSourceFile(tspath.RootedFilePath(unchanged)))
	release()

	// The edit lands after the program's read: new bytes, and an mtime a second later, so no clock tick can
	// leave its stat equal to the one the pack recorded.
	editedText := "export const alpha = 2;\n"
	if err := os.WriteFile(edited, []byte(editedText), 0o644); err != nil {
		t.Fatal(err)
	}
	later := time.Now().Add(time.Second)
	if err := os.Chtimes(edited, later, later); err != nil {
		t.Fatal(err)
	}

	var programs programOffer
	programs.offer(graph)
	formatter := native.Formatter{Options: formatoptions.Default()}
	transform := func(fileName string, text string, parsed *ast.SourceFile) (string, error) {
		return formatter.FormatParsed(fileName, text, parsed)
	}
	speculation := speculateFormatOn([]string{unchanged, edited, marked}, transform, 1, 2, &programs)
	speculation.wait()

	for _, fileName := range []string{unchanged, edited, marked} {
		if sourceFile := graph.Program.GetSourceFile(tspath.RootedFilePath(fileName)); sourceFile == nil || !sourceFile.IsBound() {
			t.Fatalf("%s is not bound, so the speculation never offered it the program's copy", filepath.Base(fileName))
		}
	}
	attempt, kept := speculation.attempts[unchanged]
	if !kept || attempt.err != nil || attempt.result.Changed {
		t.Fatalf("the unchanged file was not kept as formatted: kept %v, %+v", kept, attempt)
	}
	if unsafe.StringData(attempt.result.Text) != unsafe.StringData(graph.Program.GetSourceFile(tspath.RootedFilePath(unchanged)).Text()) {
		t.Errorf("the unchanged file's result is not the program's copy, so it was read again")
	}

	if read := speculation.read[edited]; read != editedText {
		t.Errorf("the edited file was formatted from %q, not from the bytes now on disk", read)
	}
	// The marked file is judged on its bytes on disk, mark and all, never on the program's copy without it.
	if read, kept := speculation.read[marked]; kept && read == graph.Program.GetSourceFile(tspath.RootedFilePath(marked)).Text() {
		t.Errorf("the marked file was formatted from the program's copy, which is not the file's bytes")
	}

	keepable := speculation.keepable(nil, graph)
	if _, has := keepable[edited]; has {
		t.Errorf("the edited file's attempt was kept, pairing the walk's findings with formatting of other bytes")
	}
	if _, has := keepable[unchanged]; !has {
		t.Errorf("the unchanged file's attempt was not kept")
	}
}
