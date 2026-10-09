package program

import (
	"context"
	"sync"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/compiler"
)

// FusedCheck is a cold run's type check done inside the walk: each file is checked on the goroutine that walks
// it, just before its rules run, rather than by the compiler's whole-program check on goroutines of its own
// (#679s763).
//
// # Why
//
// The whole-program check and the walk want the same checkers. Run beside each other, they took turns at
// every checker's lock and doubled the goroutines asking for cores: measured on ahra, the check slowed from
// 1.6s alone to 2.0 to 2.2s beside the walk and format, and became the run's critical path. Fused, one
// goroutine owns a checker at a time, and the rules read types the check computed a moment earlier.
//
// # Why the diagnostics are the same
//
// The compiler's whole-program check is itself a goroutine per checker, checking that checker's files one at
// a time and collecting each file's diagnostics as it finishes it (checkerPool.forEachCheckerGroupDo). A
// file checked alone, on the checker that owns it, is the same call on the same checker. What can differ is
// the order a checker sees its files in, and TestEveryFileWalkedOnAForeignCheckerFindsTheSame and
// CheckReusing's own per-file assembly already hold the diagnostics to not depending on it.
//
// # A stolen file is checked on the thief's checker
//
// A worker whose own group is done walks files taken from another group on its own checker (walkQueue), and
// checks them there too, rather than queueing for the owner's checker. Queued, the thieves waited 1.28s
// summed on cold ahra, most of it in the walk's last 300ms, with the cores idle (#tcbrgdx). Deferring the
// stolen checks until after the walk only moved that tail, since one owner then checked its stolen files
// alone. The diagnostics are the same whichever checker computes them, because a type in a generic cycle
// is computed from a canonical root rather than from whatever the checker met first (#rt071h3);
// TestEveryFileWalkedOnAForeignCheckerFindsTheSame checks every file on a foreign checker at 4 and 16
// checkers against one checker alone.
//
// The parts are assembled as AllDiagnosticParts returns them: syntactic diagnostics alone if there are any,
// else bind diagnostics, then every file's semantic diagnostics sorted and deduplicated together, as the
// compiler's whole-program call does.
type FusedCheck struct {
	mutex   sync.Mutex
	checked map[*ast.SourceFile][]*ast.Diagnostic
}

// NewFusedCheck is an empty check, for a graph whose walk will do its checking.
func NewFusedCheck() *FusedCheck {
	return &FusedCheck{checked: map[*ast.SourceFile][]*ast.Diagnostic{}}
}

// checkFile checks one file, unless it was already checked, on the checker that owns checkerFile: the file
// itself, or for a stolen file, a file of the thief's own group. The checker's lock is taken and released
// inside the call, so it must not be held: the walk calls this before it takes a checker for the file's
// rules, never while holding one.
func (f *FusedCheck) checkFile(ctx context.Context, g *Graph, sourceFile *ast.SourceFile, checkerFile *ast.SourceFile) {
	f.mutex.Lock()
	_, done := f.checked[sourceFile]
	f.mutex.Unlock()
	if done {
		return
	}
	var diagnostics []*ast.Diagnostic
	if checkerFile == sourceFile {
		diagnostics = g.Program.GetSemanticDiagnostics(ctx, sourceFile)
	} else {
		diagnostics = semanticDiagnosticsOn(ctx, g.Program, sourceFile, checkerFile)
	}
	f.mutex.Lock()
	f.checked[sourceFile] = diagnostics
	f.mutex.Unlock()
}

// semanticDiagnosticsOn is GetSemanticDiagnostics with the checker that owns checkerFile instead of the one
// that owns sourceFile: the same skip, the same call on the checker, the same filter and sort.
func semanticDiagnosticsOn(ctx context.Context, program *compiler.Program, sourceFile *ast.SourceFile, checkerFile *ast.SourceFile) []*ast.Diagnostic {
	if program.SkipTypeChecking(sourceFile, false) {
		return nil
	}
	fileChecker, release := program.GetTypeCheckerForFileExclusive(ctx, checkerFile)
	diagnostics := compiler.Program_getSemanticDiagnosticsWithChecker(program, ctx, fileChecker, sourceFile)
	release()
	return compiler.FilterAndSortDiagnostics(diagnostics)
}

// Checked is how many files have been checked so far.
func (f *FusedCheck) Checked() int {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	return len(f.checked)
}

// DiagnosticParts is the check's parts, as AllDiagnosticParts would have returned them. A program file the
// walk did not check (a third-party declaration, a file the config ignores, any file when no walk ran) is
// checked here, on its own checker, a goroutine per checker.
func (f *FusedCheck) DiagnosticParts(ctx context.Context, g *Graph) TypeDiagnosticParts {
	syntactic := g.Program.GetSyntacticDiagnostics(ctx, nil)
	if len(syntactic) > 0 {
		return TypeDiagnosticParts{Syntactic: syntactic}
	}
	bind := g.Program.GetBindDiagnostics(ctx, nil)

	programFiles := g.Program.GetSourceFiles()
	var unchecked []*ast.SourceFile
	f.mutex.Lock()
	for _, sourceFile := range programFiles {
		if _, done := f.checked[sourceFile]; !done {
			unchecked = append(unchecked, sourceFile)
		}
	}
	f.mutex.Unlock()
	for index, diagnostics := range g.semanticDiagnosticsOf(ctx, unchecked) {
		f.checked[unchecked[index]] = diagnostics
	}

	var semantic []*ast.Diagnostic
	for _, sourceFile := range programFiles {
		semantic = append(semantic, f.checked[sourceFile]...)
	}
	return TypeDiagnosticParts{Bind: bind, Semantic: compiler.SortAndDeduplicateDiagnostics(semantic)}
}
