package program

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/compiler"
	"github.com/microsoft/TypeScript/tsc/shim/incremental"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/zeebo/xxh3"
)

// Export signatures: what an importer can see of a file through types.
//
// A file's signature is the hash of its declaration output, the .d.ts the compiler would write for it, and
// of the diagnostics that output carries. An edit inside a function body leaves it unchanged; an edit to
// an exported type, an inferred return type, or a JSDoc comment on an export changes it. It is the same
// definition the compiler's own incremental builder uses to decide which importers to re-check, computed
// the same way, so the two agree: measured on ahra 2026-10-02, 3,741 of 3,741 signatures computed here
// equal the ones the compiler's buildinfo stores for the same bytes.
//
// Computing one means emitting a file's declarations, which needs the checker. Every project file costs
// about 500ms on ahra, so only files whose bytes changed since their signature was recorded are computed,
// and every other file carries its recorded signature forward. A file with no signature recorded, from the
// table or the compiler's build info, falls back to its version, the hash of its bytes: the compiler's own
// conservative default, which keys the file's shape on its content, so an edit to it re-runs its
// importers' shape-keyed rules rather than replaying them. It never replays anything stale.
//
// That fallback is also the bound. Declaration emit has no ceiling: on www-ahra-ai one ESLint config, whose
// export's type is inferred from typescript-eslint's and the plugins' objects, ran 9 minutes and grew to
// 17 GB without finishing, and tsgo's own --declaration emit does the same on it (#5txm9gg). It cannot be
// cancelled part way, and it runs on the checker the walk takes next, so a time limit around it would not
// be sound. So a file is emitted up front only when its recorded signature is a real one, computed by an
// emit that once finished: the edit of a known file, the case shapes exist for. A first run on a project
// with no build info emits nothing, and says how many shapes it keyed on content (ContentKeyedShapes).
//
// # What a signature does not cover, and the syntax hash that does
//
// Anything an importer reads off the file's syntax rather than its types. `async f(): Promise<T>` and
// `f(): Promise<T>` emit the same declaration, and so do a function whose return type was written and one
// whose return type was inferred to the same thing. A rule reaches that syntax through a symbol's
// declarations, and 128 of 164 type-aware rules can reach a declaration, measured 2026-10-02.
//
// So a file's shape is its signature and its Syntax: the hash of its text with every function body cut
// out. Modifiers, annotations, initializers, parameters, decorators, JSDoc and pragma comments outside
// bodies are all in it, so a rule reading any of them off an imported declaration sees a change. What
// neither half covers is a function body's own text, and only a rule that reads into an imported
// function's body needs more than the shape. See SignatureFingerprints.
//
// Whitespace and plain comments are not in it. A rule can reach a plain comment in another file only by
// reading that file's text, which the guard counts as reading into a body (TestRulesClaimShapesOnlyWhere
// TheScanAllowsIt), and positions were never in the shape, since a body edit moves every declaration below
// it. Keeping them made a comment appended to a file look like an export edit, and every type-aware rule
// ran again on every importer (#zqsdzbq, found by @system_cohere).

// SignatureEntry is one file's recorded shape, with the version of the bytes it was computed from.
type SignatureEntry struct {
	Version   string
	Signature string

	// Syntax is the hash of the file's text with its function bodies elided. Computed from the parse
	// alone, so it costs no checker and is filled in for any entry that lacks it, a seeded one included.
	Syntax string

	// Abandoned is set on a content-keyed entry whose emit was still running at GraduateSignatures' deadline, so
	// these bytes are never emitted again. Edited bytes start a fresh entry without it.
	Abandoned bool
}

// FileVersion is a file's version as the compiler defines it: the xxh3 hash of its text, in hex.
func FileVersion(text string) string {
	sum := xxh3.HashString128(text).Bytes()
	return hex.EncodeToString(sum[:])
}

// Signatures returns every project file's signature, keyed by file name, and how many were computed rather
// than carried forward from previous.
func (g *Graph) Signatures(ctx context.Context, previous map[string]SignatureEntry) (map[string]SignatureEntry, int) {
	projectFiles := g.ProjectFiles()
	signatures := make(map[string]SignatureEntry, len(projectFiles))
	stale := []*ast.SourceFile{}
	contentKeyed := 0
	for _, sourceFile := range projectFiles {
		version := FileVersion(sourceFile.Text())
		if recorded, found := previous[sourceFile.FileName().AsString()]; found && recorded.Version == version && recorded.Signature != "" {
			// A syntax of another definition is recomputed rather than carried: kept, it would differ from the
			// next edit's and move the shape once for nothing. A syntax that is the version stands for a file
			// whose whole text is its shape, and stays.
			if recorded.Syntax == "" || recorded.Syntax != recorded.Version && !strings.HasPrefix(recorded.Syntax, syntaxDefinition) {
				recorded.Syntax = elidedBodiesVersion(sourceFile)
			}
			signatures[sourceFile.FileName().AsString()] = recorded
			continue
		}
		// A declaration file or a JSON module has no declaration output of its own. The compiler uses the
		// version for both, and so does this. Its whole text is its shape, so its syntax is its version too.
		if sourceFile.IsDeclarationFile || ast.IsJsonSourceFile(sourceFile) {
			signatures[sourceFile.FileName().AsString()] = SignatureEntry{Version: version, Signature: version, Syntax: version}
			continue
		}
		// No real signature recorded, so none is computed: see the bound above. Its whole text is its shape.
		if recorded, found := previous[sourceFile.FileName().AsString()]; !found || recorded.Signature == "" || recorded.Signature == recorded.Version {
			signatures[sourceFile.FileName().AsString()] = SignatureEntry{Version: version, Signature: version, Syntax: version}
			contentKeyed++
			continue
		}
		stale = append(stale, sourceFile)
	}
	g.ContentKeyedShapes = countContentKeyed(projectFiles, signatures)
	if len(stale) == 0 {
		return signatures, contentKeyed
	}

	var mutex sync.Mutex
	g.signatureEmits += len(stale)
	g.Program.Emit(ctx, compiler.EmitOptions{
		TargetSourceFiles: stale,
		EmitOnly:          compiler.EmitOnlyBuilderSignature,
		WriteFile: func(fileName tspath.RootedFilePath, text string, data *compiler.WriteFileData) error {
			if data == nil || data.SourceFile == nil {
				return nil
			}
			signature := declarationSignature(data.SourceFile, text, data)
			entry := SignatureEntry{
				Version:   FileVersion(data.SourceFile.Text()),
				Signature: signature,
				Syntax:    elidedBodiesVersion(data.SourceFile),
			}
			mutex.Lock()
			signatures[data.SourceFile.FileName().AsString()] = entry
			mutex.Unlock()
			return nil
		},
	})

	// A file the emit did not write for has no signature to compare, so its version stands in, which
	// can only over-invalidate.
	for _, sourceFile := range stale {
		if _, found := signatures[sourceFile.FileName().AsString()]; !found {
			version := FileVersion(sourceFile.Text())
			signatures[sourceFile.FileName().AsString()] = SignatureEntry{Version: version, Signature: version, Syntax: version}
		}
	}
	g.ContentKeyedShapes = countContentKeyed(projectFiles, signatures)
	return signatures, len(stale) + contentKeyed
}

// beforeSignatureEmit, when set, runs before GraduateSignatures emits a file. Only a test sets it, to plant a
// file whose emit outruns the deadline.
var beforeSignatureEmit func(fileName string)

// GraduateSignatures gives files whose shape is keyed on content a real declaration signature, emitting them
// one at a time until every one is done or deadline passes, and updates shapes in place. It returns how many
// graduated, and the file it abandoned, "" when it abandoned none (#9knyr86).
//
// Signatures never emits a file without a real recorded signature, because an emit has no ceiling and cannot
// be cancelled (#5txm9gg), so a project with no build info keyed every shape on content forever. Here the run's
// verdict is already out, so an emit nobody waits on can be left behind at the deadline: the caller writes the
// table and exits, which ends it. The file it was on is marked Abandoned, and is not tried again while its
// bytes are the same, but only when that emit had run for abandonAfter: a file the deadline caught a moment
// after it began is innocent (ahra's slowest emit is 116ms, its median 0.1ms), and the next run tries it first,
// with the whole budget. A file the deadline stopped before it began is simply tried by the next run.
//
// The emits run on a goroutine of their own, and only this function touches shapes, so whatever the goroutine
// finishes after the deadline is dropped.
func (g *Graph) GraduateSignatures(shapes map[string]SignatureEntry, deadline time.Time, abandonAfter time.Duration) (graduated int, abandoned string) {
	var candidates []*ast.SourceFile
	for _, sourceFile := range g.ProjectFiles() {
		if sourceFile.IsDeclarationFile || ast.IsJsonSourceFile(sourceFile) {
			continue
		}
		entry, found := shapes[sourceFile.FileName().AsString()]
		if !found || entry.Signature != entry.Version || entry.Abandoned || entry.Version != FileVersion(sourceFile.Text()) {
			continue
		}
		candidates = append(candidates, sourceFile)
	}
	if len(candidates) == 0 {
		return 0, ""
	}

	type emitted struct {
		fileName string
		entry    SignatureEntry
	}
	type emitting struct {
		fileName string
		started  time.Time
	}
	results := make(chan emitted, len(candidates))
	var inProgress atomic.Pointer[emitting]
	// Read once, here, so a test restoring the hook never races the goroutine it left behind.
	hook := beforeSignatureEmit
	go func() {
		defer close(results)
		for _, sourceFile := range candidates {
			fileName := sourceFile.FileName()
			inProgress.Store(&emitting{fileName: fileName.AsString(), started: time.Now()})
			if hook != nil {
				hook(fileName.AsString())
			}
			var entry *SignatureEntry
			g.Program.Emit(context.Background(), compiler.EmitOptions{
				TargetSourceFiles: []*ast.SourceFile{sourceFile},
				EmitOnly:          compiler.EmitOnlyBuilderSignature,
				WriteFile: func(_ tspath.RootedFilePath, text string, data *compiler.WriteFileData) error {
					if data != nil && data.SourceFile != nil {
						entry = &SignatureEntry{
							Version:   FileVersion(data.SourceFile.Text()),
							Signature: declarationSignature(data.SourceFile, text, data),
							Syntax:    elidedBodiesVersion(data.SourceFile),
						}
					}
					return nil
				},
			})
			inProgress.Store(nil)
			// A file the emit wrote nothing for keeps its version as its signature, as Signatures does, and is
			// tried again next run.
			if entry != nil {
				results <- emitted{fileName.AsString(), *entry}
			}
		}
	}()

	timer := time.NewTimer(time.Until(deadline))
	defer timer.Stop()
	for {
		select {
		case result, open := <-results:
			if !open {
				return graduated, ""
			}
			shapes[result.fileName] = result.entry
			graduated++
		case <-timer.C:
			// What finished before now is kept, the file in progress included if it finished while this read.
			current := inProgress.Load()
			for drained := false; !drained; {
				select {
				case result, open := <-results:
					if !open {
						return graduated, ""
					}
					shapes[result.fileName] = result.entry
					graduated++
				default:
					drained = true
				}
			}
			if current == nil || time.Since(current.started) < abandonAfter {
				return graduated, ""
			}
			if entry := shapes[current.fileName]; entry.Signature == entry.Version {
				entry.Abandoned = true
				shapes[current.fileName] = entry
				return graduated, current.fileName
			}
			return graduated, ""
		}
	}
}

// countContentKeyed is how many of the project's source files have their version for a signature: no
// declaration output stands behind their shape. Declaration files and JSON modules are left out, since
// their version is their signature by definition.
func countContentKeyed(projectFiles []*ast.SourceFile, signatures map[string]SignatureEntry) int {
	count := 0
	for _, sourceFile := range projectFiles {
		if sourceFile.IsDeclarationFile || ast.IsJsonSourceFile(sourceFile) {
			continue
		}
		if entry := signatures[sourceFile.FileName().AsString()]; entry.Signature == entry.Version {
			count++
		}
	}
	return count
}

// syntaxDefinition prefixes every syntax elidedBodiesVersion computes, and moves when what it covers does.
// 2: whitespace and plain comments outside bodies are left out.
const syntaxDefinition = "syntax2:"

// elidedBodiesVersion is the version of a file's text with the body of every function, method, accessor,
// constructor, arrow function and class static block cut out, outermost first, and with the whitespace and
// plain comments between tokens left out. An edit inside a body, or to a plain comment or a blank line,
// leaves it unchanged; an edit anywhere else changes it.
//
// Trivia is found from the parse, never by scanning the text on its own: each node's leading trivia runs
// from its full start to its first token, which the parser fixed, so a regular expression or a string
// holding "//" cannot be mistaken for a comment. Trivia between tokens that start no node, before a closing
// brace say, is kept, which can only make the shape move more often.
func elidedBodiesVersion(sourceFile *ast.SourceFile) string {
	text := sourceFile.Text()
	type cut struct {
		start, end  int
		replacement string
	}
	var cuts []cut
	triviaAt := map[int]bool{}
	noteTrivia := func(position int) {
		if triviaAt[position] {
			return
		}
		triviaAt[position] = true
		tokenStart := scanner.SkipTrivia(text, position)
		if tokenStart <= position || tokenStart > len(text) {
			return
		}
		cuts = append(cuts, cut{position, tokenStart, keptTrivia(text, position, tokenStart)})
	}
	var visit func(node *ast.Node) bool
	visit = func(node *ast.Node) bool {
		if ast.IsFunctionLikeDeclaration(node) || ast.IsClassStaticBlockDeclaration(node) {
			if body := node.Body(); body != nil {
				// A body's position includes the trivia before it, so a comment between a signature and its
				// brace is cut with the body, and an edit to it is as invisible to the shape as one inside.
				// No rule reads such a comment off an imported declaration; one that did would need the
				// content fingerprint anyway.
				noteTrivia(node.Pos())
				cuts = append(cuts, cut{body.Pos(), body.End(), "{}"})
				triviaAt[body.Pos()] = true
				node.ForEachChild(func(child *ast.Node) bool {
					if child.End() <= body.Pos() {
						visit(child)
					}
					return false
				})
				return false
			}
		}
		noteTrivia(node.Pos())
		node.ForEachChild(visit)
		return false
	}
	sourceFile.AsNode().ForEachChild(visit)

	sort.Slice(cuts, func(first, second int) bool { return cuts[first].start < cuts[second].start })
	var elided strings.Builder
	kept := 0
	for _, each := range cuts {
		if each.start < kept {
			// Inside a range already cut, which only a body can be: its trivia went with it.
			continue
		}
		elided.WriteString(text[kept:each.start])
		elided.WriteString(each.replacement)
		kept = each.end
	}
	elided.WriteString(text[kept:])
	return syntaxDefinition + FileVersion(elided.String())
}

// keptTrivia is what a stretch of trivia contributes to the syntax: the comments a rule or the compiler can
// read without the text, JSDoc and pragmas, and one separator so two tokens never run together. The
// separator is a line break when the stretch held one, since a line break can decide how a statement ends,
// and a space otherwise; at the start or the end of the file, with no token on one side, it is nothing. A
// stretch holding anything but whitespace and comments, a conflict marker or a shebang, is kept whole.
//
// The stretch is walked here rather than through the scanner's comment ranges, which follow the language's
// convention that a comment on the same line as the token before it belongs to that token, and so never
// list it. The stretch is known to be trivia, from the parse, so walking it alone is safe.
func keptTrivia(text string, start int, end int) string {
	var comments strings.Builder
	lineBreak := false
	for position := start; position < end; {
		rest := text[position:end]
		switch {
		case rest[0] == ' ' || rest[0] == '\t' || rest[0] == '\v' || rest[0] == '\f':
			position++
		case rest[0] == '\n' || rest[0] == '\r':
			lineBreak = true
			position++
		case strings.HasPrefix(rest, "//"):
			length := strings.IndexAny(rest, "\r\n")
			if length < 0 {
				length = len(rest)
			}
			if comment := rest[:length]; isReadableComment(comment) {
				comments.WriteString(comment)
				comments.WriteString("\n")
			}
			position += length
		case strings.HasPrefix(rest, "/*"):
			closing := strings.Index(rest[2:], "*/")
			if closing < 0 {
				return text[start:end]
			}
			comment := rest[:closing+4]
			if isReadableComment(comment) {
				comments.WriteString(comment)
				comments.WriteString("\n")
			}
			if strings.ContainsAny(comment, "\r\n") {
				lineBreak = true
			}
			position += len(comment)
		default:
			return text[start:end]
		}
	}
	separator := " "
	switch {
	case start == 0 || end == len(text):
		separator = ""
	case lineBreak:
		separator = "\n"
	}
	return separator + comments.String()
}

// isReadableComment reports a comment something reads without scanning text: JSDoc, which the parser attaches
// to declarations, and the pragmas and directives the compiler acts on.
func isReadableComment(comment string) bool {
	switch {
	case strings.HasPrefix(comment, "/**") && comment != "/**/":
		return true
	case strings.HasPrefix(comment, "///"):
		return true
	}
	rest := strings.TrimLeft(strings.TrimPrefix(strings.TrimPrefix(comment, "//"), "/*"), " \t*")
	return strings.HasPrefix(rest, "@") || strings.HasPrefix(rest, "#")
}

// SeedSignatures returns previous with an entry added for every project file it lacks, taken from the
// signatures the compiler's own buildinfo recorded. A first run of the table then starts from what the
// compiler already computed instead of computing every signature (about 500ms on ahra) or falling back to
// versions, which would make each file's first edit invalidate its importers once for nothing.
//
// Only real shapes are taken. An entry whose signature is its version, or that has none, says the compiler
// never computed a shape for that file, and the file is left to be computed here.
func (g *Graph) SeedSignatures(previous map[string]SignatureEntry) map[string]SignatureEntry {
	seeded := make(map[string]SignatureEntry, len(previous))
	for name, entry := range previous {
		seeded[name] = entry
	}
	if g.CompilerHost == nil || g.Config == nil {
		return seeded
	}
	buildInfoFileName := g.Config.GetBuildInfoFileName()
	if buildInfoFileName == "" {
		return seeded
	}
	buildInfo := incremental.NewBuildInfoReader(g.CompilerHost).ReadBuildInfo(g.Config)
	if buildInfo == nil || len(buildInfo.FileNames) != len(buildInfo.FileInfos) {
		return seeded
	}
	directory := tspath.GetDirectoryPath(buildInfoFileName.AsString())
	recorded := make(map[tspath.PathKey]SignatureEntry, len(buildInfo.FileNames))
	for index, name := range buildInfo.FileNames {
		info := buildInfo.FileInfos[index].GetFileInfo()
		if info == nil || info.Signature() == "" || info.Signature() == info.Version() {
			continue
		}
		recorded[g.pathFor(tspath.GetNormalizedAbsolutePath(string(name), tspath.RootedDirectoryPath(directory)))] = SignatureEntry{Version: info.Version(), Signature: info.Signature()}
	}
	for _, sourceFile := range g.ProjectFiles() {
		if _, found := seeded[sourceFile.FileName().AsString()]; found {
			continue
		}
		if entry, found := recorded[sourceFile.PathKey()]; found {
			seeded[sourceFile.FileName().AsString()] = entry
		}
	}
	return seeded
}

// declarationSignature is the compiler's computeSignatureWithDiagnostics, which is unexported: the
// declaration text up to any source-map comment, then each diagnostic rendered as the compiler renders it,
// hashed.
func declarationSignature(sourceFile *ast.SourceFile, text string, data *compiler.WriteFileData) string {
	var builder strings.Builder
	if data.SourceMapUrlPos != -1 {
		builder.WriteString(text[:data.SourceMapUrlPos])
	} else {
		builder.WriteString(text)
	}
	for _, diagnostic := range data.Diagnostics {
		writeSignatureDiagnostic(diagnostic, sourceFile, &builder)
	}
	return FileVersion(builder.String())
}

func writeSignatureDiagnostic(diagnostic *ast.Diagnostic, sourceFile *ast.SourceFile, builder *strings.Builder) {
	if diagnostic == nil {
		return
	}
	builder.WriteString("\n")
	if diagnostic.File() != sourceFile {
		builder.WriteString(tspath.EnsurePathIsNonModuleName(tspath.GetRelativePathFromDirectory(
			tspath.GetDirectoryPath(string(sourceFile.PathKey())),
			string(diagnostic.File().PathKey()),
			tspath.CaseInsensitive,
		)))
	}
	if diagnostic.File() != nil {
		fmt.Fprintf(builder, "(%d,%d): ", diagnostic.Pos(), diagnostic.Len())
	}
	builder.WriteString(diagnostic.Category().Name())
	fmt.Fprintf(builder, "%d: ", diagnostic.Code())
	builder.WriteString(string(diagnostic.MessageKey()))
	builder.WriteString("\n")
	for _, argument := range diagnostic.MessageArgs() {
		builder.WriteString(argument)
		builder.WriteString("\n")
	}
	for _, chained := range diagnostic.MessageChain() {
		writeSignatureDiagnostic(chained, sourceFile, builder)
	}
	for _, related := range diagnostic.RelatedInformation() {
		writeSignatureDiagnostic(related, sourceFile, builder)
	}
}

// SignatureFingerprints gives each project file a hash of everything its types can depend on, like
// TypeFingerprints, except that what it hashes of every OTHER file in its import closure is that file's
// signature rather than its bytes. A body edit to a file imported by half the tree changes its own
// fingerprint and nobody else's.
//
// The file's own bytes are still in its fingerprint, all of them: a rule reads its own file's syntax.
// The global component is TypeFingerprints' own, bytes and all, since a global declaration changes types
// everywhere without being imported.
//
// It is sound only for a rule that sees other files through types. A rule that reads an imported
// declaration's syntax stays on TypeFingerprints.
func (g *Graph) SignatureFingerprints(signatures map[string]SignatureEntry) map[tspath.PathKey][sha256.Size]byte {
	projectFiles := g.ProjectFiles()
	global, edges, contents, resolutions := g.typeGraph()

	shapes := make(map[tspath.PathKey][sha256.Size]byte, len(projectFiles))
	for _, sourceFile := range projectFiles {
		entry, found := signatures[sourceFile.FileName().AsString()]
		shape := entry.Signature + "\x00" + entry.Syntax
		if !found || entry.Signature == "" || entry.Syntax == "" {
			// No shape means the bytes stand in, as they would for a file never computed.
			shape = "version " + FileVersion(sourceFile.Text())
		}
		// Where its imports resolve is part of what a file shows its importers; see TypeFingerprints.
		shapes[sourceFile.PathKey()] = withResolutions(sha256.Sum256([]byte(shape)), resolutions[sourceFile.PathKey()])
	}
	components := fingerprintComponents(g.Anchor, projectFiles, edges, shapes, global)

	fingerprints := make(map[tspath.PathKey][sha256.Size]byte, len(projectFiles))
	for _, sourceFile := range projectFiles {
		hash := sha256.New()
		component := components[sourceFile.PathKey()]
		content := contents[sourceFile.PathKey()]
		hash.Write(component[:])
		hash.Write(content[:])
		var sum [sha256.Size]byte
		copy(sum[:], hash.Sum(nil))
		fingerprints[sourceFile.PathKey()] = sum
	}
	return fingerprints
}
