package program

import (
	"context"
	"crypto/sha256"
	"reflect"
	"sync"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/microsoft/TypeScript/tsc/shim/compiler"
	"github.com/microsoft/TypeScript/tsc/shim/core"
)

// The types phase's cache: each project file's semantic diagnostics, replayed while the file's shape
// fingerprint is unchanged, so an edit re-checks only the edited file and the files its shape change
// reaches (#zqsdzbq, from test's benchmark e932bc1a: a one-file comment edit on ahra re-checked all 3,862
// files, 2.05s, nearly all of an edit run).
//
// # Why the shape fingerprint is the right key
//
// A file's semantic diagnostics are a function of its own text and of the types it can see: its import
// closure and everything global. SignatureFingerprints hashes exactly that: the file's own bytes, every
// file in its closure by its shape (its declaration output and its syntax with function bodies cut out),
// where its imports resolved, and the global component, bytes and all (installed declarations, scripts,
// anything that augments the global scope). An edit inside an imported function's body moves the
// importer's diagnostics only through the function's inferred type, which is part of its declaration
// output, so it moves the importer's key too. That is the same bar the compiler's own incremental builder
// holds a file's signature to.
//
// # What is never replayed
//
// A file whose diagnostics carry related information or a chain located in another file: a body edit there
// can move the position without moving the shape. A diagnostic whose text the compiler repopulates from
// program state. And everything, whenever the run that recorded the section saw a global diagnostic: those
// are produced as a side effect of checking files, so a run that skipped the files could not reproduce
// them. Syntactic and bind diagnostics are not cached at all; they are cheap and every run computes them.
//
// # What the fingerprint cannot see
//
// The compiler options. Turning on strict moves no file's text, shape or resolutions, and every file it
// newly fails would replay as clean (#hfv0ae3, proven against 3bef97f). So the section carries a key over
// what the caller knows the options come from, and a section under another key replays nothing.

// Version 2 added the key; a version 1 section, which has none, is never replayed.
const typesSectionVersion = 2

// TypesSection is the types phase's section of the cache table.
type TypesSection struct {
	Version int

	// Key is what the diagnostics depend on beyond the files the fingerprints cover: the tsconfig and
	// everything it extends. See NewTypeDiagnosticsReuse.
	Key [sha256.Size]byte

	// GlobalsClean is whether the run that recorded it saw no global diagnostics. Only then can a run that
	// skips files report what a run that checked them all would.
	GlobalsClean bool

	// Entries is each project file's semantic diagnostics, by file name, under the shape fingerprint they
	// were computed against. A file with none has an entry with none.
	Entries map[string]TypesEntry
}

// TypesEntry is one file's semantic diagnostics and the fingerprint they hold under.
type TypesEntry struct {
	Fingerprint [sha256.Size]byte
	Diagnostics []StoredDiagnostic
}

// StoredDiagnostic is a compiler diagnostic in the file it belongs to, as the compiler's own build info
// stores one: location, code, category, message key and arguments or external text, chain and related
// information. NoFile marks a part the compiler gave no file.
type StoredDiagnostic struct {
	Pos, End           int
	Code               int32
	Category           int64
	MessageKey         string
	MessageArgs        []string
	Source             string
	MessageText        string
	ReportsUnnecessary bool
	ReportsDeprecated  bool
	SkippedOnNoEmit    bool
	NoFile             bool
	Chain              []StoredDiagnostic
	Related            []StoredDiagnostic
}

// TypeDiagnosticsReuse is one run's use of the section: what was stored, and what this run records.
type TypeDiagnosticsReuse struct {
	stored *TypesSection
	key    [sha256.Size]byte

	mutex    sync.Mutex
	recorded *TypesSection
	replayed int

	// replayedFiles is the files whose semantic diagnostics this run took from the section, by file name, so
	// a run scoped to some of the program can say how many of those it replayed (ReplayedAmong).
	replayedFiles map[string]struct{}
}

// NewTypeDiagnosticsReuse wraps what the table held, which may be nil, for a run whose compiler options
// hash to key. Only a section recorded under the same key is replayed, and what this run records carries it.
func NewTypeDiagnosticsReuse(stored *TypesSection, key [sha256.Size]byte) *TypeDiagnosticsReuse {
	return &TypeDiagnosticsReuse{stored: stored, key: key}
}

// Unchanged reports whether what this run recorded is what the table already holds: every file replayed
// under the same key, and the global verdict the same. Writing it back would rewrite the table for nothing.
func (r *TypeDiagnosticsReuse) Unchanged() bool {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	if r.recorded == nil || r.stored == nil {
		return r.recorded == nil
	}
	return r.stored.Key == r.recorded.Key && r.stored.GlobalsClean == r.recorded.GlobalsClean &&
		r.replayed == len(r.recorded.Entries) && len(r.recorded.Entries) == len(r.stored.Entries)
}

// Replayed is how many files' semantic diagnostics this run took from the section.
func (r *TypeDiagnosticsReuse) Replayed() int {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	return r.replayed
}

// ReplayedAmong is how many of files had their semantic diagnostics taken from the section. The check
// covers the whole program whatever the run's scope, so Replayed counts files a scoped run never reports
// on, and a line saying "3955 of 1 files" read as nonsense (#rxqptqp).
func (r *TypeDiagnosticsReuse) ReplayedAmong(files []*ast.SourceFile) int {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	count := 0
	for _, sourceFile := range files {
		if _, replayed := r.replayedFiles[sourceFile.FileName().AsString()]; replayed {
			count++
		}
	}
	return count
}

// Recorded is the section this run leaves for the next, nil when it recorded nothing.
func (r *TypeDiagnosticsReuse) Recorded() *TypesSection {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	return r.recorded
}

// FinishGlobals says whether this run saw any global diagnostic. Until it is told none, what it recorded
// may only be read as a full check: a run that saw one records nothing to replay.
func (r *TypeDiagnosticsReuse) FinishGlobals(clean bool) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	if r.recorded != nil {
		r.recorded.GlobalsClean = clean
	}
}

// TypeDiagnosticParts is a check's syntactic, bind and semantic diagnostics, kept apart so the semantic
// ones can be recorded per file. With any syntactic diagnostic the check stops there, as the compiler's own
// does, and the other two are empty.
type TypeDiagnosticParts struct {
	Syntactic []*ast.Diagnostic
	Bind      []*ast.Diagnostic
	Semantic  []*ast.Diagnostic
}

// All is the parts in the order AllDiagnostics returns them.
func (p TypeDiagnosticParts) All() []*ast.Diagnostic {
	if len(p.Syntactic) > 0 {
		return p.Syntactic
	}
	return append(append([]*ast.Diagnostic{}, p.Bind...), p.Semantic...)
}

// CheckReusing checks the project files whose shape fingerprint moved since the section was recorded and
// replays the rest. It reports false when it cannot, and the caller runs the full check and hands its parts
// to RecordFull: no section, a section under other compiler options or from a run that saw global
// diagnostics, no shapes this run, or so
// much changed that the full check, which spreads every checker over every file, is the faster way.
func (g *Graph) CheckReusing(ctx context.Context, reuse *TypeDiagnosticsReuse) (TypeDiagnosticParts, bool) {
	stored := reuse.stored
	if stored == nil || stored.Version != typesSectionVersion || stored.Key != reuse.key || !stored.GlobalsClean || g.Shapes == nil {
		return TypeDiagnosticParts{}, false
	}
	syntactic := g.Program.GetSyntacticDiagnostics(ctx, nil)
	if len(syntactic) > 0 {
		return TypeDiagnosticParts{Syntactic: syntactic}, true
	}

	fingerprints := g.SignatureFingerprints(g.Shapes)
	projectFiles := g.ProjectFiles()
	var replayed []*ast.Diagnostic
	var toCheck []*ast.SourceFile
	entries := make(map[string]TypesEntry, len(projectFiles))
	replayedFiles := map[string]struct{}{}
	for _, sourceFile := range projectFiles {
		entry, found := stored.Entries[sourceFile.FileName().AsString()]
		if found && entry.Fingerprint == fingerprints[sourceFile.PathKey()] {
			for _, diagnostic := range entry.Diagnostics {
				replayed = append(replayed, diagnostic.diagnostic(sourceFile))
			}
			entries[sourceFile.FileName().AsString()] = entry
			replayedFiles[sourceFile.FileName().AsString()] = struct{}{}
			continue
		}
		toCheck = append(toCheck, sourceFile)
	}
	if len(toCheck) > len(projectFiles)/2 {
		return TypeDiagnosticParts{}, false
	}

	bind := g.Program.GetBindDiagnostics(ctx, nil)
	checked := g.semanticDiagnosticsOf(ctx, toCheck)
	semantic := replayed
	for index, sourceFile := range toCheck {
		semantic = append(semantic, checked[index]...)
		if entry, keepable := typesEntryFor(sourceFile, fingerprints[sourceFile.PathKey()], checked[index]); keepable {
			entries[sourceFile.FileName().AsString()] = entry
		}
	}

	// A check cancelled part way can hand back a partial list, which recorded as a file's whole verdict would
	// replay a file as clean. Whatever finished before the cancel is still exactly the list for its key. A
	// retired graph's check describes bytes the fixer replaced, and its graph's successor records instead;
	// asked under the mutex, so a retired check finishing late can never overwrite what the successor
	// recorded.
	//
	// The replayed files too: a retired check finishing after its successor would otherwise leave the old
	// graph's count on the types line.
	reuse.mutex.Lock()
	if !g.Retired() {
		reuse.replayed, reuse.replayedFiles = len(replayedFiles), replayedFiles
	}
	if ctx.Err() == nil && !g.Retired() {
		reuse.recorded = &TypesSection{Version: typesSectionVersion, Key: reuse.key, Entries: entries}
	}
	reuse.mutex.Unlock()
	return TypeDiagnosticParts{Bind: bind, Semantic: compiler.SortAndDeduplicateDiagnostics(semantic)}, true
}

// RecordFull records a full check's semantic diagnostics per project file, for the next run to replay.
func (g *Graph) RecordFull(ctx context.Context, reuse *TypeDiagnosticsReuse, parts TypeDiagnosticParts) {
	if len(parts.Syntactic) > 0 || g.Shapes == nil || ctx.Err() != nil {
		return
	}
	fingerprints := g.SignatureFingerprints(g.Shapes)
	byFile := map[*ast.SourceFile][]*ast.Diagnostic{}
	for _, diagnostic := range parts.Semantic {
		byFile[diagnostic.File()] = append(byFile[diagnostic.File()], diagnostic)
	}
	entries := make(map[string]TypesEntry, len(g.ProjectFiles()))
	for _, sourceFile := range g.ProjectFiles() {
		if entry, keepable := typesEntryFor(sourceFile, fingerprints[sourceFile.PathKey()], byFile[sourceFile]); keepable {
			entries[sourceFile.FileName().AsString()] = entry
		}
	}
	reuse.mutex.Lock()
	// Asked here, under the mutex, for the reason CheckReusing gives: a retired graph never records.
	if !g.Retired() {
		reuse.recorded = &TypesSection{Version: typesSectionVersion, Key: reuse.key, Entries: entries}
	}
	reuse.mutex.Unlock()
}

// semanticDiagnosticsOf checks each file on the checker that owns it, a goroutine per checker, so a handful
// of files costs a handful of checks and many files still spread across every checker.
func (g *Graph) semanticDiagnosticsOf(ctx context.Context, files []*ast.SourceFile) [][]*ast.Diagnostic {
	results := make([][]*ast.Diagnostic, len(files))
	groups := map[*checker.Checker][]int{}
	for index, sourceFile := range files {
		owner, release := g.Program.GetTypeCheckerForFile(ctx, sourceFile)
		release()
		groups[owner] = append(groups[owner], index)
	}
	var waitGroup sync.WaitGroup
	for _, indices := range groups {
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			for _, index := range indices {
				results[index] = g.Program.GetSemanticDiagnostics(ctx, files[index])
			}
		}()
	}
	waitGroup.Wait()
	return results
}

// typesEntryFor stores a file's semantic diagnostics, or reports false when any of them cannot be replayed
// faithfully: located in another file, or repopulated from program state.
func typesEntryFor(sourceFile *ast.SourceFile, fingerprint [sha256.Size]byte, diagnostics []*ast.Diagnostic) (TypesEntry, bool) {
	entry := TypesEntry{Fingerprint: fingerprint, Diagnostics: make([]StoredDiagnostic, 0, len(diagnostics))}
	for _, diagnostic := range diagnostics {
		stored, keepable := storeDiagnostic(diagnostic, sourceFile)
		if !keepable {
			return TypesEntry{}, false
		}
		entry.Diagnostics = append(entry.Diagnostics, stored)
	}
	return entry, true
}

func storeDiagnostic(diagnostic *ast.Diagnostic, sourceFile *ast.SourceFile) (StoredDiagnostic, bool) {
	if diagnostic.RepopulateInfo() != nil || (diagnostic.File() != nil && diagnostic.File() != sourceFile) {
		return StoredDiagnostic{}, false
	}
	stored := StoredDiagnostic{
		Pos:                diagnostic.Pos(),
		End:                diagnostic.End(),
		Code:               diagnostic.Code(),
		Category:           int64(diagnostic.Category()),
		MessageKey:         string(diagnostic.MessageKey()),
		MessageArgs:        diagnostic.MessageArgs(),
		Source:             diagnostic.Source(),
		MessageText:        diagnostic.MessageText(),
		ReportsUnnecessary: diagnostic.ReportsUnnecessary(),
		ReportsDeprecated:  diagnostic.ReportsDeprecated(),
		SkippedOnNoEmit:    diagnostic.SkippedOnNoEmit(),
		NoFile:             diagnostic.File() == nil,
	}
	// A diagnostic with neither a key nor external text is an ad hoc message, which nothing here can rebuild.
	if stored.MessageKey == "" && stored.MessageText == "" {
		return StoredDiagnostic{}, false
	}
	for _, part := range diagnostic.MessageChain() {
		chained, keepable := storeDiagnostic(part, sourceFile)
		if !keepable {
			return StoredDiagnostic{}, false
		}
		stored.Chain = append(stored.Chain, chained)
	}
	for _, part := range diagnostic.RelatedInformation() {
		related, keepable := storeDiagnostic(part, sourceFile)
		if !keepable {
			return StoredDiagnostic{}, false
		}
		stored.Related = append(stored.Related, related)
	}
	return stored, true
}

// diagnostic rebuilds the compiler diagnostic, the way the compiler rebuilds one from its build info.
func (stored StoredDiagnostic) diagnostic(sourceFile *ast.SourceFile) *ast.Diagnostic {
	file := sourceFile
	if stored.NoFile {
		file = nil
	}
	var chain, related []*ast.Diagnostic
	for _, part := range stored.Chain {
		chain = append(chain, part.diagnostic(sourceFile))
	}
	for _, part := range stored.Related {
		related = append(related, part.diagnostic(sourceFile))
	}
	// The shim does not name the compiler's key and category types, so each is set through a value of its
	// own type, taken from a diagnostic's accessor.
	var zero ast.Diagnostic
	key, category := zero.MessageKey(), zero.Category()
	reflect.ValueOf(&key).Elem().SetString(stored.MessageKey)
	reflect.ValueOf(&category).Elem().SetInt(stored.Category)
	diagnostic := ast.NewDiagnosticFromSerialized(file, core.NewTextRange(stored.Pos, stored.End), stored.Code, category, key,
		stored.MessageArgs, chain, related, stored.ReportsUnnecessary, stored.ReportsDeprecated, stored.SkippedOnNoEmit)
	if stored.Source != "" || stored.MessageText != "" {
		diagnostic.SetExternalData(stored.Source, stored.MessageText)
	}
	return diagnostic
}
