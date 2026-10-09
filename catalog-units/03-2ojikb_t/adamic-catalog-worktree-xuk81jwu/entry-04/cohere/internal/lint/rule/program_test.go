package rule

import (
	"fmt"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/compiler"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
)

// fileNamed is a file a rule may ask about, by name.
type fileNamed string

func (name fileNamed) FileName() tspath.RootedFilePath { return tspath.RootedFilePath(name) }
func (name fileNamed) PathKey() tspath.PathKey         { return tspath.PathKey(name) }

// refusalOf calls one method and returns the view's refusal, or "" when the view let the call through.
// A call the view admits reaches an empty program and may panic there, which is not a refusal.
func refusalOf(call func()) (refusal string) {
	defer func() {
		if recovered := recover(); recovered != nil {
			if message := fmt.Sprint(recovered); strings.Contains(message, "declares ProgramReads") {
				refusal = message
			}
		}
	}()
	call()
	return ""
}

// ctx.Program refuses every read a rule did not declare, by name, and admits every read it did
// (#b8k3bp6). Each method is tried under a declaration of nothing and under one of its own kind only,
// so a method checking the wrong kind, or none, fails here by name.
func TestAProgramViewAdmitsExactlyTheDeclaredReads(t *testing.T) {
	t.Parallel()
	methods := []struct {
		name string
		read ProgramRead
		call func(Program)
	}{
		{"Options", ReadsCompilerOptions, func(view Program) { view.Options() }},
		{"GetCurrentDirectory", ReadsCompilerOptions, func(view Program) { view.GetCurrentDirectory() }},
		{"UseCaseSensitiveFileNames", ReadsCompilerOptions, func(view Program) { view.UseCaseSensitiveFileNames() }},
		{"IsSourceFileDefaultLibrary", ReadsDefaultLibrary, func(view Program) { view.IsSourceFileDefaultLibrary("lib.d.ts") }},
		{"DefaultLibraryPath", ReadsDefaultLibrary, func(view Program) { view.DefaultLibraryPath() }},
		{"GetSourceFileForResolvedModule", ReadsOtherFiles, func(view Program) { view.GetSourceFileForResolvedModule("other.ts") }},
		{"SourceFiles", ReadsOtherFiles, func(view Program) { view.SourceFiles() }},
		{"GetSourceFile", ReadsOtherFiles, func(view Program) { view.GetSourceFile("other.ts") }},
		{"FS", ReadsOtherFiles, func(view Program) { view.FS() }},
		{"DesignSystemFS", ReadsDesignSystem, func(view Program) { view.DesignSystemFS() }},
	}
	program := new(compiler.Program)
	for _, method := range methods {
		t.Run(method.name, func(t *testing.T) {
			t.Parallel()
			refused := refusalOf(func() { method.call(ViewProgram(program, nil, Rule{Name: "probe"})) })
			if !strings.Contains(refused, "rule probe called Program."+method.name) || !strings.Contains(refused, programReadNames[method.read]) {
				t.Errorf("an undeclared %s was not refused by name and kind: %q", method.name, refused)
			}
			for _, other := range []ProgramRead{ReadsCompilerOptions, ReadsDefaultLibrary, ReadsModuleResolution, ReadsOtherFiles, ReadsDesignSystem} {
				refused := refusalOf(func() { method.call(ViewProgram(program, nil, Rule{Name: "probe", ProgramReads: other})) })
				if admitted := refused == ""; admitted != (other == method.read) {
					t.Errorf("%s under a declaration of %s: admitted %v, want %v", method.name, other, admitted, other == method.read)
				}
			}
		})
	}
}

// Resolving the linted file's own imports is a module-resolution read; resolving another file's is a
// read of other files, which no per-file key covers.
func TestResolvingAnotherFilesImportsIsAReadOfOtherFiles(t *testing.T) {
	t.Parallel()
	program := new(compiler.Program)
	resolution := Rule{Name: "probe", ProgramReads: ReadsModuleResolution}
	refused := refusalOf(func() { ViewProgram(program, nil, resolution).ResolveModule(fileNamed("other.ts"), nil) })
	if !strings.Contains(refused, "ReadsOtherFiles") {
		t.Errorf("resolving another file's imports under ReadsModuleResolution alone was not refused: %q", refused)
	}
	both := Rule{Name: "probe", ProgramReads: ReadsModuleResolution | ReadsOtherFiles}
	if refused := refusalOf(func() { ViewProgram(program, nil, both).ResolveModule(fileNamed("other.ts"), nil) }); refused != "" {
		t.Errorf("resolving another file's imports was refused although ReadsOtherFiles is declared: %q", refused)
	}
}

// A program that could not be built is a nil Program, so the rules' `ctx.Program == nil` checks keep
// meaning what they meant.
func TestANilProgramViewsAsNil(t *testing.T) {
	t.Parallel()
	if ViewProgram(nil, nil, Rule{Name: "probe", ProgramReads: ReadsOtherFiles}) != nil {
		t.Error("a nil program gave a non-nil view")
	}
	var slot ProgramViewSlot
	if slot.Point(nil, nil, Rule{Name: "probe", ProgramReads: ReadsOtherFiles}) != nil {
		t.Error("a nil program gave a non-nil view from a slot")
	}
	slot.Release()
}

// endedOf calls one method and returns the view's "file ended" panic, or "" when the view let the call through.
func endedOf(call func()) (ended string) {
	defer func() {
		if recovered := recover(); recovered != nil {
			if message := fmt.Sprint(recovered); strings.Contains(message, "after the file it was given ended") {
				ended = message
			}
		}
	}()
	call()
	return ""
}

// A slot's view is good for its file only (#9xfg09f). Once released, every read through it panics by the
// rule's name, including the reads a rule declared, so a rule that kept its Program past the file fails
// loudly instead of reading through the file the slot was pointed at next. The view a rule kept from one
// file stays ended while the next file runs, because the slot alternates between two views.
func TestASlotViewKeptPastItsFileFailsOnEveryRead(t *testing.T) {
	t.Parallel()
	program := new(compiler.Program)
	everything := Rule{Name: "keeps-its-program", ProgramReads: ReadsCompilerOptions | ReadsDefaultLibrary | ReadsModuleResolution | ReadsOtherFiles | ReadsDesignSystem}
	reads := map[string]func(Program){
		"Options":                        func(view Program) { view.Options() },
		"GetCurrentDirectory":            func(view Program) { view.GetCurrentDirectory() },
		"UseCaseSensitiveFileNames":      func(view Program) { view.UseCaseSensitiveFileNames() },
		"IsSourceFileDefaultLibrary":     func(view Program) { view.IsSourceFileDefaultLibrary("lib.d.ts") },
		"DefaultLibraryPath":             func(view Program) { view.DefaultLibraryPath() },
		"ResolveModule":                  func(view Program) { view.ResolveModule(fileNamed("other.ts"), nil) },
		"GetSourceFileForResolvedModule": func(view Program) { view.GetSourceFileForResolvedModule("other.ts") },
		"SourceFiles":                    func(view Program) { view.SourceFiles() },
		"GetSourceFile":                  func(view Program) { view.GetSourceFile("other.ts") },
		"FS":                             func(view Program) { view.FS() },
		"DesignSystemFS":                 func(view Program) { view.DesignSystemFS() },
		"Identity":                       func(view Program) { view.Identity() },
	}

	var slot ProgramViewSlot
	first := slot.Point(program, nil, everything)
	if ended := endedOf(func() { first.Identity() }); ended != "" {
		t.Fatalf("a view refused a read during its own file: %s", ended)
	}
	slot.Release()
	second := slot.Point(program, nil, everything)

	for name, read := range reads {
		ended := endedOf(func() { read(first) })
		if ended == "" {
			t.Errorf("Program.%s read through a view whose file had ended", name)
			continue
		}
		if !strings.Contains(ended, everything.Name) || !strings.Contains(ended, name) {
			t.Errorf("Program.%s's refusal names neither the rule nor the read: %s", name, ended)
		}
	}
	if ended := endedOf(func() { second.Identity() }); ended != "" {
		t.Errorf("the next file's view was ended along with the last one: %s", ended)
	}
	slot.Release()
	if ended := endedOf(func() { second.Identity() }); ended == "" {
		t.Error("the second file's view still read after its own release")
	}
}
