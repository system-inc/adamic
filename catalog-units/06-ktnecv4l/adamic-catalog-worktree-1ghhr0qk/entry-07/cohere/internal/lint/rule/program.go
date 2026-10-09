package rule

import (
	"fmt"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/compiler"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/microsoft/TypeScript/tsc/shim/vfs"
)

// ProgramRead is what a rule reads through ctx.Program beyond the file it was handed, declared so the
// findings cache can tell which reads its key already covers (#b8k3bp6).
//
// It replaced one flag, ReadsProgram, that said "reaches the program" and so made 45 rules uncacheable,
// about a third of the rule time on ahra, when most of them read nothing the key does not already
// hold. The kinds are split by what covers them:
//
//	ReadsCompilerOptions   the tsconfig's options and the run's directory: the findings key hashes
//	                       the tsconfig chain and the root
//	ReadsDefaultLibrary    whether a declaration lives in a lib file: the libs are embedded in the
//	                       binary, which is in the key
//	ReadsModuleResolution  where this file's own imports resolve: the type fingerprint's edges come
//	                       from resolution, and every file it can reach is in the closure or the
//	                       global component
//	ReadsOtherFiles        anything else: the program's file list, a file by name or by resolution,
//	                       another file's imports, the file system. Nothing per file covers it, so a
//	                       rule declaring it is never cached
//	ReadsDesignSystem      the files the Tailwind design system is built from: its entry point, the
//	                       stylesheets it imports, and the paths it probed and did not find. Read
//	                       through DesignSystemFS, which records every one, so the set is observed
//	                       rather than listed and can key a cache of its own (#pyhm2t2)
//
// A rule reads ctx.Program through a view that refuses every method outside its declaration, with a
// panic naming the rule and the read. The walk contains that panic to the rule and names it, so a rule
// cannot read more than it declared without every run, fixture and crash guard saying so. Declaring
// more than a rule reads costs a cache miss and nothing else, the same asymmetry NeedsTypeChecker has.
type ProgramRead uint8

const (
	ReadsCompilerOptions ProgramRead = 1 << iota
	ReadsDefaultLibrary
	ReadsModuleResolution
	ReadsOtherFiles
	ReadsDesignSystem
)

// programReadNames names each read in a refusal.
var programReadNames = map[ProgramRead]string{
	ReadsCompilerOptions:  "ReadsCompilerOptions",
	ReadsDefaultLibrary:   "ReadsDefaultLibrary",
	ReadsModuleResolution: "ReadsModuleResolution",
	ReadsOtherFiles:       "ReadsOtherFiles",
	ReadsDesignSystem:     "ReadsDesignSystem",
}

// String lists the reads a declaration holds, for refusals and tests.
func (reads ProgramRead) String() string {
	names := []string{}
	for _, read := range []ProgramRead{ReadsCompilerOptions, ReadsDefaultLibrary, ReadsModuleResolution, ReadsOtherFiles, ReadsDesignSystem} {
		if reads&read != 0 {
			names = append(names, programReadNames[read])
		}
	}
	if len(names) == 0 {
		return "nothing"
	}
	return strings.Join(names, " | ")
}

// ResolvedModule is what a rule may read of where one of its file's imports resolved.
type ResolvedModule struct {
	ResolvedFileName        string
	IsExternalLibraryImport bool
	// PackageName is the resolved package's name, or empty when the import is not a package's.
	PackageName string
}

// IsResolved reports whether the import resolved to a file.
func (resolved ResolvedModule) IsResolved() bool {
	return resolved.ResolvedFileName != ""
}

// ProgramIdentity tells one program from another without handing out the program, so a rule may cache
// what it derived from a run and notice the next run, and cannot reach past its declaration through it.
type ProgramIdentity struct {
	program *compiler.Program
}

// Program is ctx.Program: the program, as far as a rule declared it reads it. Each method says which
// read it is.
type Program interface {
	// ReadsCompilerOptions.
	Options() *core.CompilerOptions
	GetCurrentDirectory() string
	UseCaseSensitiveFileNames() bool

	// ReadsDefaultLibrary.
	IsSourceFileDefaultLibrary(path tspath.PathKey) bool
	DefaultLibraryPath() string

	// ReadsModuleResolution for the file the rule was handed, ReadsOtherFiles for any other.
	ResolveModule(file ast.HasFileName, specifier *ast.StringLiteralLike) ResolvedModule

	// ReadsOtherFiles.
	GetSourceFileForResolvedModule(fileName string) *ast.SourceFile
	SourceFiles() []*ast.SourceFile
	GetSourceFile(fileName string) *ast.SourceFile
	FS() vfs.FS

	// ReadsDesignSystem: a file system that records every path it is asked about and refuses writes.
	DesignSystemFS() *RecordingFS

	// Identity reads nothing.
	Identity() ProgramIdentity
}

// ViewProgram gives one rule its view of program while it runs on sourceFile: every method outside the
// rule's ProgramReads panics, naming the rule and the read it did not declare. A nil program is a nil
// Program, so `ctx.Program == nil` still says the program could not be built.
func ViewProgram(program *compiler.Program, sourceFile *ast.SourceFile, subject Rule) Program {
	if program == nil {
		return nil
	}
	return &programView{program: program, sourceFile: sourceFile, ruleName: subject.Name, reads: subject.ProgramReads}
}

// ProgramViewSlot is one rule's view of the program, kept by a walk worker and pointed at each file it
// dispatches in turn. The walk used to build every rule's view afresh for each file, 101 MB of views on a
// cold ahra run (#9xfg09f); a slot per rule per worker is a few hundred kilobytes for the whole run.
//
// A view is good for its file and no longer. Release ends it, and from then on every read through it
// panics, naming the rule. The slot alternates between two views, so a Program a rule kept from one file
// is still the ended view while the next file runs, and reading it fails loudly rather than reading through
// whatever file the slot was pointed at next. A view kept across two or more files is not caught. What it
// would read is the same program under the same rule's declared reads, so only ResolveModule's "is this
// another file" check could answer differently.
type ProgramViewSlot struct {
	views [2]programView
	next  int
}

// Point aims the slot at sourceFile for subject and returns the view as the rule's Program. A nil program
// is a nil Program, as ViewProgram's is.
func (slot *ProgramViewSlot) Point(program *compiler.Program, sourceFile *ast.SourceFile, subject Rule) Program {
	if program == nil {
		return nil
	}
	view := &slot.views[slot.next]
	slot.next = 1 - slot.next
	*view = programView{program: program, sourceFile: sourceFile, ruleName: subject.Name, reads: subject.ProgramReads}
	return view
}

// Release ends the file the slot was last pointed at: every read through that view panics from now on.
func (slot *ProgramViewSlot) Release() {
	view := &slot.views[1-slot.next]
	if view.program == nil {
		return
	}
	view.ended = true
	view.program = nil
}

type programView struct {
	program    *compiler.Program
	sourceFile *ast.SourceFile
	ruleName   string
	reads      ProgramRead

	// ended is set once the file the view was given is over (ProgramViewSlot.Release).
	ended bool
}

// live panics if the view's file is over.
func (view *programView) live(method string) {
	if view.ended {
		panic(fmt.Sprintf("rule %s called Program.%s after the file it was given ended; a Program is good for its file only, so keep what it answered rather than the Program",
			view.ruleName, method))
	}
}

// require panics unless the view is live and the rule declared read.
func (view *programView) require(read ProgramRead, method string) {
	view.live(method)
	if view.reads&read == 0 {
		panic(fmt.Sprintf("rule %s called Program.%s, which is %s, and declares ProgramReads %s; declare the read so the findings cache does not replay a stale verdict",
			view.ruleName, method, programReadNames[read], view.reads))
	}
}

func (view *programView) Options() *core.CompilerOptions {
	view.require(ReadsCompilerOptions, "Options")
	return view.program.Options()
}

func (view *programView) GetCurrentDirectory() string {
	view.require(ReadsCompilerOptions, "GetCurrentDirectory")
	return view.program.GetCurrentDirectory().AsString()
}

func (view *programView) UseCaseSensitiveFileNames() bool {
	view.require(ReadsCompilerOptions, "UseCaseSensitiveFileNames")
	return view.program.UseCaseSensitiveFileNames()
}

func (view *programView) IsSourceFileDefaultLibrary(path tspath.PathKey) bool {
	view.require(ReadsDefaultLibrary, "IsSourceFileDefaultLibrary")
	return view.program.IsSourceFileDefaultLibrary(path)
}

func (view *programView) DefaultLibraryPath() string {
	view.require(ReadsDefaultLibrary, "DefaultLibraryPath")
	return view.program.Host().DefaultLibraryPath().AsString()
}

func (view *programView) ResolveModule(file ast.HasFileName, specifier *ast.StringLiteralLike) ResolvedModule {
	// Another file's imports are not in this file's fingerprint.
	if view.sourceFile == nil || file.FileName() != view.sourceFile.FileName() {
		view.require(ReadsOtherFiles, "ResolveModule on another file")
	} else {
		view.require(ReadsModuleResolution, "ResolveModule")
	}
	resolved := view.program.GetResolvedModuleFromModuleSpecifier(file, specifier)
	if resolved == nil || !resolved.IsResolved() {
		return ResolvedModule{}
	}
	return ResolvedModule{
		ResolvedFileName:        resolved.ResolvedFileName.AsString(),
		IsExternalLibraryImport: resolved.IsExternalLibraryImport,
		PackageName:             resolved.PackageId.Name,
	}
}

func (view *programView) GetSourceFileForResolvedModule(fileName string) *ast.SourceFile {
	view.require(ReadsOtherFiles, "GetSourceFileForResolvedModule")
	// Upstream's own takes the compiler's resolved module, which a rule never holds (it holds this
	// package's ResolvedModule). This is the lookup it made by name before typed file paths: the file,
	// else the file a project reference redirects the name to.
	name := tspath.RootedFilePath(fileName)
	if file := view.program.GetSourceFile(name); file != nil {
		return file
	}
	if redirect := view.program.GetParseFileRedirect(name); redirect != "" {
		return view.program.GetSourceFile(redirect)
	}
	return nil
}

func (view *programView) SourceFiles() []*ast.SourceFile {
	view.require(ReadsOtherFiles, "SourceFiles")
	return view.program.SourceFiles()
}

func (view *programView) GetSourceFile(fileName string) *ast.SourceFile {
	view.require(ReadsOtherFiles, "GetSourceFile")
	return view.program.GetSourceFile(tspath.RootedFilePath(fileName))
}

func (view *programView) FS() vfs.FS {
	view.require(ReadsOtherFiles, "FS")
	return view.program.Host().FS()
}

func (view *programView) DesignSystemFS() *RecordingFS {
	view.require(ReadsDesignSystem, "DesignSystemFS")
	recorder := NewRecordingFS(view.program.Host().FS())
	recordDesignSystemFS(view.program, recorder)
	return recorder
}

func (view *programView) Identity() ProgramIdentity {
	view.live("Identity")
	return ProgramIdentity{program: view.program}
}

// FingerprintPath is how a ProgramFingerprint names a file or directory: relative to the program's
// current directory, so the hash holds when the same tree is checked out somewhere else, and the same
// across fixtures built in different directories. Reading the current directory is ReadsCompilerOptions,
// so a rule calling this declares it, which costs nothing: the findings key already hashes the root.
//
// A path under the current directory is named by slicing off the directory and its separator, which
// allocates nothing. tspath's general answer splits both paths into components and joins them again, and
// a fingerprint names every file of the program, once per computation: on an edit run of ahra that was
// 1.6 GB (#9prvp67). The slice is the string tspath gives for a descendant, which
// TestFingerprintPathSlicesWhatTspathWouldJoin holds over every file of this repository, both ways a host
// can treat case. On a case-insensitive host a program's paths are canonical, lowercased, while the
// directory keeps its case, so the directory is compared as tspath lowers it; only an ASCII directory takes
// the slice, since ToFileNameLowerCase treats a few other characters specially. Anything else, a path
// outside the directory or a directory spelled otherwise, takes tspath's answer.
func FingerprintPath(program Program, path string) string {
	directory := program.GetCurrentDirectory()
	if inside, under := fingerprintPathUnder(directory, path, program.UseCaseSensitiveFileNames()); under {
		return inside
	}
	return tspath.GetRelativePathFromDirectory(directory, path, caseSensitivity(program.UseCaseSensitiveFileNames()))
}

// fingerprintPathUnder is the part of path below directory, when path is a descendant and the bytes say so:
// exactly, or on a case-insensitive host against the directory lowered as tspath lowers ASCII.
func fingerprintPathUnder(directory string, path string, caseSensitive bool) (string, bool) {
	if directory == "" || len(path) <= len(directory) {
		return "", false
	}
	for index := 0; index < len(directory); index++ {
		character := directory[index]
		if character >= 0x80 {
			return "", false
		}
		if !caseSensitive && 'A' <= character && character <= 'Z' {
			character += 'a' - 'A'
		}
		if path[index] != character {
			return "", false
		}
	}
	rest := path[len(directory):]
	if strings.HasSuffix(directory, "/") {
		return rest, true
	}
	if inside, separated := strings.CutPrefix(rest, "/"); separated && inside != "" {
		return inside, true
	}
	return "", false
}

// caseSensitivity is a program's answer to UseCaseSensitiveFileNames as the CaseSensitivity tspath takes.
func caseSensitivity(caseSensitive bool) tspath.CaseSensitivity {
	if caseSensitive {
		return tspath.CaseSensitive
	}
	return tspath.CaseInsensitive
}
