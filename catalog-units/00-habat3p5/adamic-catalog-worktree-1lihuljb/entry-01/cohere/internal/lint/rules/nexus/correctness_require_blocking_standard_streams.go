package nexus

import (
	"crypto/sha256"
	"slices"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/system-inc/cohere/internal/lint/ecmascript/control_flow_graph"
	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/policy"
)

const correctnessRequireBlockingStandardStreamsId = "exitBeforeBlockingStandardStreams"

// The rule's message and the options it picks from, whose wording lives in
// `policy/messages/correctness-require-blocking-standard-streams.json`.
var (
	correctnessRequireBlockingStandardStreamsText         = policy.MessageOf("nexus/correctness-require-blocking-standard-streams", correctnessRequireBlockingStandardStreamsId)
	correctnessRequireBlockingStandardStreamsShebang      = correctnessRequireBlockingStandardStreamsText.Option("processReason", "shebang")
	correctnessRequireBlockingStandardStreamsUnimported   = correctnessRequireBlockingStandardStreamsText.Option("processReason", "unimported")
	correctnessRequireBlockingStandardStreamsOneExit      = correctnessRequireBlockingStandardStreamsText.Option("exitCount", "one")
	correctnessRequireBlockingStandardStreamsSeveralExits = correctnessRequireBlockingStandardStreamsText.Option("exitCount", "several")
)

// CorrectnessRequireBlockingStandardStreams reports a file that runs as a process and can reach
// `process.exit()` after a write to stdout or stderr without having called Nexus's
// `blockStandardStreams()` first.
//
//	invalid: #!/usr/bin/env tsx
//	         async function main() { if(!name) { console.error('usage'); process.exit(1); } }
//	         main().catch(function(error) { console.error(error); process.exit(1); });
//	invalid: // a forked worker no file imports
//	         if(!specFilePath) { console.error('usage'); process.exit(2); }
//	valid:   async function main() { blockStandardStreams(); ... } main().catch(...);
//	valid:   blockStandardStreams(); if(!specFilePath) { console.error('usage'); process.exit(2); }
//	valid:   runCommandLineInterface(new AhraCommandLineInterface());
//	valid:   // a library file other files import, whatever its functions do
//
// # Where it came from
//
// Task `#gjs78wz`, ruled by `@system_cohere`. Node writes to a pipe asynchronously on POSIX, and
// `process.exit()` does not wait for the queue: measured on Node 24 on macOS, 1 MiB written and then
// `process.exit(0)`, piped into a slow reader, delivered 65,536 bytes, exit status 0, no error.
// `nexus/correctness-no-process-exit-after-output` found the hazard per exit, 1,012 times on ahra, and
// each fix is a `process.exitCode` and a `return` placed by hand. Nexus's `source/system/StandardStreams.ts`
// fixes every exit of a process at once by making stdout and stderr blocking, and
// `runCommandLineInterface` calls it before any command runs, so the per-exit rule is registered off and
// the check moves here: to the entry, where the one fix belongs. The real sites were the fifteen
// entries in the task's `ExitEntries.tsv` that block nothing, among them `ahra facets`
// (`modules/facets/FacetsCommandLineInterface.ts`, its own `main().catch(...)` with seven exits), the
// forked `modules/data/DataConversionShardWorker.ts`, Structure's spawned `StructureDoctor.ts`,
// `StructureAnalyzer.ts` and `StructureLintEngineParity.ts`, and the `main().catch(...)` scripts under
// `modules/phi/social/scripts/`.
//
// # What makes a file an entry
//
// What the file proves about itself, and one fact about the program:
//
//   - **A `#!` line.** The author declares the file executable. It is an entry even when other files
//     import it too, because run directly it is the process's first module.
//   - **Nothing in the program imports it.** Then its top-level code only ever runs as the first
//     module of a process: a script run by path, a forked or spawned child, or a file a tool loads by
//     convention. An import is any `import` or `export ... from` not written `type`, an
//     `import x = require()`, and a dynamic `import()` or `require()` of a string literal.
//
// A file without a `#!` that something imports is declined whole: its top-level code may run inside
// another entry's process after that entry blocked, and the AST of this file cannot say which. Names
// play no part: `main` is not special, and neither is `CommandLineInterface` in a file name.
//
// # What the process runs
//
// The code that runs because the file was loaded, followed by symbol inside this file: the module's
// top level, every function of this file it calls (by an identifier that resolves to a function
// declaration or a `const` bound to a function), every immediately invoked function, and every
// function handed to a call as an argument (`main().catch(function(error) { ... })`,
// `process.on('SIGINT', onInterrupt)`), transitively. An exported function no load-time code reaches
// is not counted: it runs in whichever process calls it. A call into another file is not followed;
// an exit there is a missed finding here.
//
// # The exit
//
// The condition `nexus/correctness-no-process-exit-after-output` reports, reused from it: a call of
// `NodeJS.Process.exit` reached, on some path through its function's control-flow graph, after a
// write to stdout or stderr in the same function. A write in one function and an exit in another
// (`showHelp(); process.exit(0);`) is a missed finding, as it is there.
//
// # Blocking first
//
// A file whose code, and the code of every file it imports, transitively, never reaches Nexus's
// `source/system/StandardStreams.ts` cannot block (the program index below decides that once). Every
// exit the process runs is then reported as soon as one qualifies.
//
// A file that can reach it is walked in order, and anything that may block ends the walk's path:
//
//   - **A call that may block**: one resolving by symbol to Nexus's `blockStandardStreams` (by its
//     declaration's name and file, `source/system/StandardStreams.ts`), or to any function whose body
//     may call one, transitively and by symbol (`runCommandLineInterface` is one, and so is a `main`
//     whose first statement blocks). A callee the checker cannot resolve, a `declare`d function or a
//     value of unknown origin called in a file that can reach the streams, a dynamic `import()`, and
//     `require()` all count as blocking, and so does handing a function that may block to any call.
//     A call into a file that cannot reach the streams never blocks, whatever it is.
//   - **An `await`**, a `for await` and an `await using`, in a file that blocks anywhere: the code
//     after it runs on a later turn, after anything else the file starts.
//   - **A callback or a generator**, in such a file, is taken to start blocked: it may run after any
//     later statement.
//   - **The files it imports.** Each runs its top-level code before this file's body. If one that can
//     reach the streams makes a call at load that may block, the whole file is taken as blocked.
//
// Each of these can only cost a finding. So a file that blocks first is silent, and a file that
// blocks too late is reported when an exit is reached before the block on some path: a usage guard
// that prints and exits above `runCommandLineInterface(...)`, or a `main` that writes and exits before
// its own `blockStandardStreams()`.
//
// Two things are not seen, and either could make a finding false. `blockStandardStreams` stored in a
// variable, a property or a registry and called through it: proving no such store exists would mean
// reading every file's references, and on ahra both Nexus functions are only ever called directly.
// And code the program does not hold that runs first in the process, a `node --import` preload that
// blocks: ahra's entries run under `tsx`, whose loader does not.
//
// # The finding
//
// One per entry, at its first qualifying exit in source order, naming how many there are. The fix is
// one call at the entry, not one per exit.
//
// # Cost
//
// The program index (which files are imported, which can reach the streams) is built once per run,
// under a package lock, as `correctness-no-import-cycle-load-time-read` builds its graph: one
// resolution per import specifier and no checker. The rule reads other files, so it declares
// `ReadsOtherFiles`, and the findings cache replays it on a file while its bytes, its type fingerprint
// and the program fingerprint hold (#5yfbdse). A file that never spells `exit` is declined before the
// index is asked for.
//
// # No fix
//
// Where the call belongs is the author's choice: the first statement of the module, or of the `main`
// it runs, and above a guard that must not run unblocked. A fixer would pick one blindly.
var CorrectnessRequireBlockingStandardStreams = rule.Rule{
	Name: "nexus/correctness-require-blocking-standard-streams",

	// Symbol identity is what tells `NodeJS.Process.exit` and the global console from look-alikes,
	// Nexus's `blockStandardStreams` from any other function of that name, and a call of this file's
	// own function from a call of an import.
	NeedsTypeChecker: true,

	// The index reads every file's imports and their resolutions, this file's among them. The run's
	// directory, which the fingerprint names files relative to.
	ProgramReads: rule.ReadsCompilerOptions | rule.ReadsModuleResolution | rule.ReadsOtherFiles,

	// See correctnessRequireBlockingStandardStreamsFingerprint for what a file's verdict reads.
	ProgramFingerprint: correctnessRequireBlockingStandardStreamsFingerprint,

	// Only a file that calls `process.exit` can hold a finding, and most files do not.
	NoListener: rule.NoListenerDeclinesIrrelevantFiles,

	Run: func(ctx rule.Context, options any) rule.Listeners {
		if ctx.TypeChecker == nil || ctx.Program == nil || ctx.SourceFile == nil {
			return nil
		}
		// Every exit is a member named `exit`, so a file that never spells it holds none.
		if !strings.Contains(ctx.SourceFile.Text(), "exit") {
			return nil
		}
		return rule.Listeners{
			ast.KindSourceFile: func(sourceFile *ast.Node) {
				correctnessRequireBlockingStandardStreamsScanFile(ctx)
			},
		}
	},
}

// correctnessRequireBlockingStandardStreamsDeclarationFileSuffixes are where Nexus declares
// `blockStandardStreams`: its source, and the declaration file a compiled copy would ship.
var correctnessRequireBlockingStandardStreamsDeclarationFileSuffixes = []string{
	"/source/system/StandardStreams.ts",
	"/source/system/StandardStreams.d.ts",
}

// correctnessRequireBlockingStandardStreamsIndex is what the rule knows about the whole program.
type correctnessRequireBlockingStandardStreamsIndex struct {
	// imported holds every file some program file imports.
	imported map[tspath.PathKey]bool
	// canBlock holds every file whose code, or the code of a file it imports, transitively, may call
	// `blockStandardStreams`: Nexus's declaring file, every file that reaches it, and every file that
	// loads a module by a computed `import()` or `require()`, which may be anything.
	canBlock map[tspath.PathKey]bool
	// edges are each file's imports, of every kind, to files in the program.
	edges map[tspath.PathKey][]tspath.PathKey
	// files are the program's files by path.
	files map[tspath.PathKey]*ast.SourceFile
}

// correctnessRequireBlockingStandardStreamsCache holds one index per program, for the reason
// `correctness_no_import_cycle_load_time_read.go` gives: the program does not exist at registration,
// and the index is a fact about the whole run.
var correctnessRequireBlockingStandardStreamsCache struct {
	sync.Mutex
	program rule.ProgramIdentity
	index   *correctnessRequireBlockingStandardStreamsIndex
}

func correctnessRequireBlockingStandardStreamsIndexFor(program rule.Program) *correctnessRequireBlockingStandardStreamsIndex {
	correctnessRequireBlockingStandardStreamsCache.Lock()
	defer correctnessRequireBlockingStandardStreamsCache.Unlock()
	if correctnessRequireBlockingStandardStreamsCache.program == program.Identity() && correctnessRequireBlockingStandardStreamsCache.index != nil {
		return correctnessRequireBlockingStandardStreamsCache.index
	}
	index := correctnessRequireBlockingStandardStreamsBuildIndex(program)
	correctnessRequireBlockingStandardStreamsCache.program = program.Identity()
	correctnessRequireBlockingStandardStreamsCache.index = index
	return index
}

// correctnessRequireBlockingStandardStreamsFingerprint is the rule's program fingerprint (#5yfbdse): the
// files something in the program imports, by name, in a stable order.
//
// A file's verdict reads, beyond the file, whether any other file imports it, and, for the files it
// imports, transitively, their edges, whether each can reach the streams, and what their top-level code
// calls. Only the first lies outside the file's import closure. Every edge the index holds is a
// resolution from the program's own table, the table the type fingerprint's edges come from, so the
// files a file's imports reach are its closure, whose bytes and resolutions its type fingerprint holds;
// and whether one of them can reach the streams is decided by its own imports, inside that closure. So
// the imported set is the whole of it: an edit that makes a file imported or unimported moves this, and
// an edit that does neither leaves it, and the findings replay. The rule takes no options, so none reach
// it.
func correctnessRequireBlockingStandardStreamsFingerprint(program rule.Program, _ any) [sha256.Size]byte {
	index := correctnessRequireBlockingStandardStreamsIndexFor(program)
	var imported []string
	for path := range index.imported {
		imported = append(imported, rule.FingerprintPath(program, index.files[path].FileName().AsString()))
	}
	slices.Sort(imported)
	hash := sha256.New()
	for _, name := range imported {
		hash.Write([]byte(name))
		hash.Write([]byte{0})
	}
	var fingerprint [sha256.Size]byte
	copy(fingerprint[:], hash.Sum(nil))
	return fingerprint
}

// correctnessRequireBlockingStandardStreamsBuildIndex resolves every import of every source file once.
// Declaration files contribute no edges of their own (no package's types import Nexus), but may be
// the target of one, so a Nexus consumed as compiled code still seeds the closure.
func correctnessRequireBlockingStandardStreamsBuildIndex(program rule.Program) *correctnessRequireBlockingStandardStreamsIndex {
	index := &correctnessRequireBlockingStandardStreamsIndex{
		imported: map[tspath.PathKey]bool{},
		canBlock: map[tspath.PathKey]bool{},
		edges:    map[tspath.PathKey][]tspath.PathKey{},
		files:    map[tspath.PathKey]*ast.SourceFile{},
	}
	var seeds []tspath.PathKey
	for _, sourceFile := range program.SourceFiles() {
		if sourceFile == nil {
			continue
		}
		index.files[sourceFile.PathKey()] = sourceFile
		for _, suffix := range correctnessRequireBlockingStandardStreamsDeclarationFileSuffixes {
			if strings.HasSuffix(sourceFile.FileName().AsString(), suffix) {
				seeds = append(seeds, sourceFile.PathKey())
			}
		}
	}

	for _, importer := range program.SourceFiles() {
		if importer == nil || importer.IsDeclarationFile {
			continue
		}
		add := func(specifier *ast.Node) {
			if specifier == nil || !ast.IsStringLiteralLike(specifier) {
				return
			}
			resolved := program.ResolveModule(importer, specifier)
			if !resolved.IsResolved() {
				return
			}
			target := program.GetSourceFileForResolvedModule(resolved.ResolvedFileName)
			if target == nil || target == importer || index.files[target.PathKey()] == nil {
				return
			}
			index.edges[importer.PathKey()] = append(index.edges[importer.PathKey()], target.PathKey())
			index.imported[target.PathKey()] = true
		}
		for _, statement := range importer.Statements.Nodes {
			switch statement.Kind {
			case ast.KindImportDeclaration:
				declaration := statement.AsImportDeclaration()
				if declaration.ImportClause == nil || !declaration.ImportClause.IsTypeOnly() {
					add(declaration.ModuleSpecifier)
				}
			case ast.KindExportDeclaration:
				declaration := statement.AsExportDeclaration()
				if !declaration.IsTypeOnly {
					add(declaration.ModuleSpecifier)
				}
			case ast.KindImportEqualsDeclaration:
				declaration := statement.AsImportEqualsDeclaration()
				if !declaration.IsTypeOnly && declaration.ModuleReference != nil && ast.IsExternalModuleReference(declaration.ModuleReference) {
					add(declaration.ModuleReference.AsExternalModuleReference().Expression)
				}
			}
		}
		// A module loaded by a call: `import(...)` or `require(...)`. A computed specifier may load
		// anything, so the file is taken to reach the streams.
		text := importer.Text()
		if !strings.Contains(text, "import(") && !strings.Contains(text, "require(") {
			continue
		}
		var visit func(node *ast.Node) bool
		visit = func(node *ast.Node) bool {
			if node.Kind == ast.KindCallExpression && (ast.IsImportCall(node) || ast.IsRequireCall(node, false)) {
				arguments := node.AsCallExpression().Arguments
				if arguments != nil && len(arguments.Nodes) > 0 && ast.IsStringLiteralLike(arguments.Nodes[0]) {
					add(arguments.Nodes[0])
				} else {
					seeds = append(seeds, importer.PathKey())
				}
			}
			node.ForEachChild(visit)
			return false
		}
		importer.AsNode().ForEachChild(visit)
	}

	importers := map[tspath.PathKey][]tspath.PathKey{}
	for from, targets := range index.edges {
		for _, target := range targets {
			importers[target] = append(importers[target], from)
		}
	}
	queue := append([]tspath.PathKey(nil), seeds...)
	for _, seed := range seeds {
		index.canBlock[seed] = true
	}
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		for _, importer := range importers[current] {
			if !index.canBlock[importer] {
				index.canBlock[importer] = true
				queue = append(queue, importer)
			}
		}
	}
	return index
}

// correctnessRequireBlockingStandardStreamsEventKind is what one recorded event is.
type correctnessRequireBlockingStandardStreamsEventKind uint8

const (
	correctnessRequireBlockingStandardStreamsWrite correctnessRequireBlockingStandardStreamsEventKind = iota
	correctnessRequireBlockingStandardStreamsExit
	correctnessRequireBlockingStandardStreamsCatchEntered
	correctnessRequireBlockingStandardStreamsCall
	correctnessRequireBlockingStandardStreamsBlock
)

// correctnessRequireBlockingStandardStreamsEvent is one thing the graph recorded. A call carries the
// functions of this file it starts: the ones it runs before it returns, and the ones it hands on.
type correctnessRequireBlockingStandardStreamsEvent struct {
	kind              correctnessRequireBlockingStandardStreamsEventKind
	node              *ast.Node
	runs              []*ast.Node
	handsOn           []*ast.Node
	argumentsMayBlock bool
}

// correctnessRequireBlockingStandardStreamsAnalysis is one entry being read.
type correctnessRequireBlockingStandardStreamsAnalysis struct {
	ctx   rule.Context
	index *correctnessRequireBlockingStandardStreamsIndex

	// ordered is whether this file can reach the streams, so that where a block happens matters.
	ordered bool

	// functionMayBlock memoizes whether running a function may call `blockStandardStreams`: 1 yes,
	// 2 no, 3 being decided (a recursive call counts as no; the outer decision still sees every other
	// call in the cycle).
	functionMayBlock map[*ast.Node]uint8
}

func correctnessRequireBlockingStandardStreamsScanFile(ctx rule.Context) {
	sourceFile := ctx.SourceFile
	if !correctnessRequireBlockingStandardStreamsHasCandidate(ctx, sourceFile.AsNode()) {
		return
	}
	index := correctnessRequireBlockingStandardStreamsIndexFor(ctx.Program)
	reason := correctnessRequireBlockingStandardStreamsShebang
	if !strings.HasPrefix(sourceFile.Text(), "#!") {
		if index.imported[sourceFile.PathKey()] {
			return
		}
		reason = correctnessRequireBlockingStandardStreamsUnimported
	}
	analysis := &correctnessRequireBlockingStandardStreamsAnalysis{
		ctx:              ctx,
		index:            index,
		ordered:          index.canBlock[sourceFile.PathKey()],
		functionMayBlock: map[*ast.Node]uint8{},
	}
	if analysis.ordered && analysis.importsBlockAtLoad(sourceFile.PathKey()) {
		return
	}

	exits := analysis.reachedExits(sourceFile.AsNode())
	if len(exits) == 0 {
		return
	}
	sort.Slice(exits, func(left int, right int) bool { return exits[left].Pos() < exits[right].Pos() })
	count := correctnessRequireBlockingStandardStreamsOneExit
	values := map[string]string(nil)
	if len(exits) > 1 {
		count = correctnessRequireBlockingStandardStreamsSeveralExits
		values = map[string]string{"count": strconv.Itoa(len(exits))}
	}
	ctx.ReportNode(exits[0], rule.Message{
		Id:          correctnessRequireBlockingStandardStreamsId,
		Description: correctnessRequireBlockingStandardStreamsText.Render(values, reason, count),
	})
}

// correctnessRequireBlockingStandardStreamsHasCandidate says whether some function of the file, or
// its top level, holds both an exit and a write: the cheap precondition of any finding.
func correctnessRequireBlockingStandardStreamsHasCandidate(ctx rule.Context, sourceFile *ast.Node) bool {
	exits := map[*ast.Node]bool{}
	writes := map[*ast.Node]bool{}
	found := false
	var visit func(node *ast.Node) bool
	visit = func(node *ast.Node) bool {
		if found {
			return true
		}
		if node.Kind == ast.KindCallExpression {
			if correctnessNoProcessExitAfterOutputIsExit(ctx, node) {
				root := control_flow_graph.RootOf(node)
				exits[root] = true
				found = writes[root]
			} else if correctnessNoProcessExitAfterOutputIsWrite(ctx, node) {
				root := control_flow_graph.RootOf(node)
				writes[root] = true
				found = exits[root]
			}
		}
		node.ForEachChild(visit)
		return found
	}
	sourceFile.ForEachChild(visit)
	return found
}

// importsBlockAtLoad says whether a file this one imports, transitively, may block while it loads,
// before this file's own body runs. Only a file that can reach the streams can, and only through a
// call its top-level code makes.
func (analysis *correctnessRequireBlockingStandardStreamsAnalysis) importsBlockAtLoad(entry tspath.PathKey) bool {
	seen := map[tspath.PathKey]bool{entry: true}
	queue := []tspath.PathKey{entry}
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		for _, next := range analysis.index.edges[current] {
			if seen[next] {
				continue
			}
			seen[next] = true
			queue = append(queue, next)
			if !analysis.index.canBlock[next] {
				continue
			}
			file := analysis.index.files[next]
			// A compiled Nexus runs code this program cannot read.
			if file == nil || file.IsDeclarationFile || analysis.topLevelMayBlock(file.AsNode()) {
				return true
			}
		}
	}
	return false
}

// topLevelMayBlock says whether a module's load may call something that blocks: a call in its
// top-level statements, in a class's static or field initializers, or in a function it hands to a
// call there. Function and method bodies are reached only through a call, which the call's own
// verdict covers.
func (analysis *correctnessRequireBlockingStandardStreamsAnalysis) topLevelMayBlock(sourceFile *ast.Node) bool {
	blocks := false
	var visit func(node *ast.Node) bool
	visit = func(node *ast.Node) bool {
		if blocks {
			return true
		}
		if correctnessRequireBlockingStandardStreamsIsFunctionWithBody(node) {
			return false
		}
		if correctnessRequireBlockingStandardStreamsIsCallLike(node) && analysis.callMayBlock(node) {
			blocks = true
			return true
		}
		node.ForEachChild(visit)
		return blocks
	}
	sourceFile.ForEachChild(visit)
	return blocks
}

// correctnessRequireBlockingStandardStreamsIsFunctionWithBody says whether a node is a function whose
// body runs only when it is called.
func correctnessRequireBlockingStandardStreamsIsFunctionWithBody(node *ast.Node) bool {
	switch node.Kind {
	case ast.KindFunctionDeclaration, ast.KindFunctionExpression, ast.KindArrowFunction, ast.KindMethodDeclaration,
		ast.KindConstructor, ast.KindGetAccessor, ast.KindSetAccessor:
		return node.Body() != nil
	}
	return false
}

// correctnessRequireBlockingStandardStreamsIsCallLike says whether a node runs code it names: a call,
// a `new`, or a tagged template.
func correctnessRequireBlockingStandardStreamsIsCallLike(node *ast.Node) bool {
	return node.Kind == ast.KindCallExpression || node.Kind == ast.KindNewExpression || node.Kind == ast.KindTaggedTemplateExpression
}

// callMayBlock says whether running a call may call `blockStandardStreams`: through what it calls,
// or through a function it hands on.
func (analysis *correctnessRequireBlockingStandardStreamsAnalysis) callMayBlock(call *ast.Node) bool {
	var callee *ast.Node
	var arguments []*ast.Node
	switch call.Kind {
	case ast.KindCallExpression:
		if ast.IsImportCall(call) {
			return true
		}
		expression := call.AsCallExpression()
		callee = expression.Expression
		if expression.Arguments != nil {
			arguments = expression.Arguments.Nodes
		}
	case ast.KindNewExpression:
		expression := call.AsNewExpression()
		callee = expression.Expression
		if expression.Arguments != nil {
			arguments = expression.Arguments.Nodes
		}
	case ast.KindTaggedTemplateExpression:
		callee = call.AsTaggedTemplateExpression().Tag
	default:
		return false
	}
	for _, argument := range arguments {
		if analysis.handsOnMayBlock(argument) {
			return true
		}
	}
	callee = ast.SkipParentheses(callee)
	if callee.Kind == ast.KindFunctionExpression || callee.Kind == ast.KindArrowFunction {
		return analysis.mayBlockWhenRun(callee)
	}
	var name *ast.Node
	switch callee.Kind {
	case ast.KindIdentifier:
		name = callee
	case ast.KindPropertyAccessExpression:
		name = callee.AsPropertyAccessExpression().Name()
	default:
		// An element access, a call's result, or anything else the checker names no declaration for.
		return true
	}
	return analysis.symbolMayBlock(analysis.ctx.TypeChecker.GetSymbolAtLocation(name), name)
}

// handsOnMayBlock says whether an argument is a function, or names one, that may block when the
// callee runs it.
func (analysis *correctnessRequireBlockingStandardStreamsAnalysis) handsOnMayBlock(argument *ast.Node) bool {
	argument = ast.SkipParentheses(argument)
	switch argument.Kind {
	case ast.KindFunctionExpression, ast.KindArrowFunction:
		return analysis.mayBlockWhenRun(argument)
	case ast.KindIdentifier:
		symbol := analysis.ctx.TypeChecker.GetSymbolAtLocation(argument)
		if symbol != nil && symbol.Flags&ast.SymbolFlagsAlias != 0 {
			symbol = analysis.ctx.TypeChecker.GetAliasedSymbol(symbol)
		}
		if symbol == nil || symbol.Flags&ast.SymbolFlagsFunction == 0 {
			return false
		}
		return analysis.symbolMayBlock(symbol, argument)
	}
	return false
}

// symbolMayBlock says whether calling what a symbol names may block. Resolved through an import alias
// to its declaration, it may when any declaration may.
func (analysis *correctnessRequireBlockingStandardStreamsAnalysis) symbolMayBlock(symbol *ast.Symbol, name *ast.Node) bool {
	if symbol != nil && symbol.Flags&ast.SymbolFlagsAlias != 0 {
		symbol = analysis.ctx.TypeChecker.GetAliasedSymbol(symbol)
	}
	if symbol == nil || len(symbol.Declarations) == 0 {
		return true
	}
	for _, declaration := range symbol.Declarations {
		if analysis.declarationMayBlock(declaration, name) {
			return true
		}
	}
	return false
}

// declarationMayBlock says whether calling what one declaration declares may block.
func (analysis *correctnessRequireBlockingStandardStreamsAnalysis) declarationMayBlock(declaration *ast.Node, name *ast.Node) bool {
	file := ast.GetSourceFileOfNode(declaration)
	if file == nil {
		return true
	}
	if !analysis.index.canBlock[file.PathKey()] {
		// `require` is declared by `@types/node` and loads whatever it is handed.
		return file.IsDeclarationFile && name.Text() == "require"
	}
	if file.IsDeclarationFile || correctnessRequireBlockingStandardStreamsIsNexusDeclaration(declaration) {
		return true
	}
	switch declaration.Kind {
	case ast.KindFunctionDeclaration, ast.KindMethodDeclaration, ast.KindConstructor, ast.KindGetAccessor, ast.KindSetAccessor:
		if declaration.Body() == nil {
			// An overload signature, whose implementation is another declaration of the same
			// symbol and is read there. A `declare`d function, or an abstract or interface-shaped
			// member, is implemented somewhere this file cannot see.
			return declaration.Kind != ast.KindFunctionDeclaration || declaration.Flags&ast.NodeFlagsAmbient != 0
		}
		return analysis.mayBlockWhenRun(declaration)
	case ast.KindVariableDeclaration:
		if initializer := declaration.Initializer(); initializer != nil {
			initializer = ast.SkipParentheses(initializer)
			if initializer.Kind == ast.KindFunctionExpression || initializer.Kind == ast.KindArrowFunction {
				return analysis.mayBlockWhenRun(initializer)
			}
		}
	case ast.KindClassDeclaration, ast.KindClassExpression:
		return analysis.mayBlockWhenRun(declaration)
	}
	// A value that holds a function nobody here can name.
	return true
}

// correctnessRequireBlockingStandardStreamsIsNexusDeclaration says whether a declaration is Nexus's
// own `blockStandardStreams`: a function of that name in Nexus's `source/system/StandardStreams`.
func correctnessRequireBlockingStandardStreamsIsNexusDeclaration(declaration *ast.Node) bool {
	if declaration.Kind != ast.KindFunctionDeclaration || declaration.Name() == nil || declaration.Name().Text() != "blockStandardStreams" {
		return false
	}
	file := ast.GetSourceFileOfNode(declaration)
	for _, suffix := range correctnessRequireBlockingStandardStreamsDeclarationFileSuffixes {
		if file != nil && strings.HasSuffix(file.FileName().AsString(), suffix) {
			return true
		}
	}
	return false
}

// mayBlockWhenRun says whether running a function, a class's construction and initializers, or
// anything they hand on, may block: any call anywhere inside it that may.
func (analysis *correctnessRequireBlockingStandardStreamsAnalysis) mayBlockWhenRun(node *ast.Node) bool {
	switch analysis.functionMayBlock[node] {
	case 1:
		return true
	case 2, 3:
		return false
	}
	analysis.functionMayBlock[node] = 3
	blocks := false
	var visit func(child *ast.Node) bool
	visit = func(child *ast.Node) bool {
		if blocks {
			return true
		}
		if correctnessRequireBlockingStandardStreamsIsCallLike(child) && analysis.callMayBlock(child) {
			blocks = true
			return true
		}
		child.ForEachChild(visit)
		return blocks
	}
	node.ForEachChild(visit)
	if blocks {
		analysis.functionMayBlock[node] = 1
	} else {
		analysis.functionMayBlock[node] = 2
	}
	return blocks
}

// localFunction is the function of this file an expression names: a function declaration, or a
// `const` bound to a function expression or an arrow. Anything else answers nil.
func (analysis *correctnessRequireBlockingStandardStreamsAnalysis) localFunction(expression *ast.Node) *ast.Node {
	expression = ast.SkipParentheses(expression)
	switch expression.Kind {
	case ast.KindFunctionExpression, ast.KindArrowFunction:
		return expression
	case ast.KindIdentifier:
	default:
		return nil
	}
	symbol := analysis.ctx.TypeChecker.GetSymbolAtLocation(expression)
	if symbol == nil {
		return nil
	}
	for _, declaration := range symbol.Declarations {
		if ast.GetSourceFileOfNode(declaration) != analysis.ctx.SourceFile {
			continue
		}
		switch declaration.Kind {
		case ast.KindFunctionDeclaration:
			if declaration.Body() != nil {
				return declaration
			}
		case ast.KindVariableDeclaration:
			list := declaration.Parent
			initializer := declaration.Initializer()
			if list == nil || list.Kind != ast.KindVariableDeclarationList || list.Flags&ast.NodeFlagsConst == 0 || initializer == nil {
				continue
			}
			initializer = ast.SkipParentheses(initializer)
			if initializer.Kind == ast.KindFunctionExpression || initializer.Kind == ast.KindArrowFunction {
				return initializer
			}
		}
	}
	return nil
}

// startedFunctions are the functions of this file a call starts: the ones it runs before it returns
// (its callee, when that is this file's function, unless a generator, whose body waits for its
// iterator), and the ones it hands on as arguments, which may run now or on any later turn.
func (analysis *correctnessRequireBlockingStandardStreamsAnalysis) startedFunctions(call *ast.Node) (runs []*ast.Node, handsOn []*ast.Node) {
	var callee *ast.Node
	var arguments []*ast.Node
	switch call.Kind {
	case ast.KindCallExpression:
		expression := call.AsCallExpression()
		callee = expression.Expression
		if expression.Arguments != nil {
			arguments = expression.Arguments.Nodes
		}
	case ast.KindNewExpression:
		if expression := call.AsNewExpression(); expression.Arguments != nil {
			arguments = expression.Arguments.Nodes
		}
	case ast.KindTaggedTemplateExpression:
		callee = call.AsTaggedTemplateExpression().Tag
	}
	if callee != nil {
		if function := analysis.localFunction(callee); function != nil {
			if ast.GetFunctionFlags(function)&ast.FunctionFlagsGenerator != 0 {
				handsOn = append(handsOn, function)
			} else {
				runs = append(runs, function)
			}
		}
	}
	for _, argument := range arguments {
		if function := analysis.localFunction(argument); function != nil {
			handsOn = append(handsOn, function)
		}
	}
	return runs, handsOn
}

// argumentsMayBlock says whether evaluating a call's arguments may block, which they do before the
// callee's body runs.
func (analysis *correctnessRequireBlockingStandardStreamsAnalysis) argumentsMayBlock(call *ast.Node) bool {
	var arguments []*ast.Node
	switch call.Kind {
	case ast.KindCallExpression:
		if expression := call.AsCallExpression(); expression.Arguments != nil {
			arguments = expression.Arguments.Nodes
		}
	case ast.KindNewExpression:
		if expression := call.AsNewExpression(); expression.Arguments != nil {
			arguments = expression.Arguments.Nodes
		}
	}
	blocks := false
	var visit func(node *ast.Node) bool
	visit = func(node *ast.Node) bool {
		if blocks || correctnessRequireBlockingStandardStreamsIsFunctionWithBody(node) {
			return blocks
		}
		if node.Kind == ast.KindAwaitExpression || (correctnessRequireBlockingStandardStreamsIsCallLike(node) && analysis.callMayBlock(node)) {
			blocks = true
			return true
		}
		node.ForEachChild(visit)
		return blocks
	}
	for _, argument := range arguments {
		visit(argument)
	}
	return blocks
}

// reachedExits walks the process from the module's top level, function by function, and returns every
// exit a path reaches having written and not having blocked.
func (analysis *correctnessRequireBlockingStandardStreamsAnalysis) reachedExits(sourceFile *ast.Node) []*ast.Node {
	queued := map[*ast.Node]bool{sourceFile: true}
	queue := []*ast.Node{sourceFile}
	reported := map[*ast.Node]bool{}
	var exits []*ast.Node
	for len(queue) > 0 {
		root := queue[0]
		queue = queue[1:]
		found, started := analysis.analyzeRoot(root)
		for _, exit := range found {
			if !reported[exit] {
				reported[exit] = true
				exits = append(exits, exit)
			}
		}
		for _, function := range started {
			if !queued[function] {
				queued[function] = true
				queue = append(queue, function)
			}
		}
	}
	return exits
}

// analyzeRoot builds one root's graph and walks it from its entry, unblocked. It returns the exits a
// path reaches having written, and the functions of this file a path starts while still unblocked.
func (analysis *correctnessRequireBlockingStandardStreamsAnalysis) analyzeRoot(root *ast.Node) ([]*ast.Node, []*ast.Node) {
	ctx := analysis.ctx
	type event = correctnessRequireBlockingStandardStreamsEvent
	type builder = control_flow_graph.Builder[event]
	recorded := map[*ast.Node]bool{}
	unrecordable := false
	graph := control_flow_graph.Build(root, control_flow_graph.Hooks[event]{
		Expression: func(b *builder, node *ast.Node) {
			callLike := correctnessRequireBlockingStandardStreamsIsCallLike(node)
			if !callLike && node.Kind != ast.KindAwaitExpression {
				return
			}
			if correctnessNoProcessExitAfterOutputInCatchBinding(node, root) {
				// It runs before the marker the write state relies on. An exit or a block there
				// cannot be placed, so the root is given up below.
				return
			}
			if node.Kind == ast.KindAwaitExpression {
				if analysis.ordered {
					recorded[node] = true
					b.Emit(event{kind: correctnessRequireBlockingStandardStreamsBlock, node: node})
				}
				return
			}
			if node.Kind == ast.KindCallExpression && correctnessNoProcessExitAfterOutputIsExit(ctx, node) {
				recorded[node] = true
				b.Emit(event{kind: correctnessRequireBlockingStandardStreamsExit, node: node})
				return
			}
			if node.Kind == ast.KindCallExpression && correctnessNoProcessExitAfterOutputIsWrite(ctx, node) &&
				!correctnessNoProcessExitAfterOutputHoldsExit(ctx, node) {
				b.Emit(event{kind: correctnessRequireBlockingStandardStreamsWrite, node: node})
			}
			runs, handsOn := analysis.startedFunctions(node)
			if len(runs) > 0 || len(handsOn) > 0 {
				b.Emit(event{
					kind: correctnessRequireBlockingStandardStreamsCall, node: node, runs: runs, handsOn: handsOn,
					argumentsMayBlock: analysis.ordered && len(runs) > 0 && analysis.argumentsMayBlock(node),
				})
			}
			if analysis.ordered && analysis.callMayBlock(node) {
				recorded[node] = true
				b.Emit(event{kind: correctnessRequireBlockingStandardStreamsBlock, node: node})
			}
		},
		Statement: func(b *builder, node *ast.Node) {
			if node.Kind == ast.KindBlock && node.Parent != nil && node.Parent.Kind == ast.KindCatchClause {
				b.Emit(event{kind: correctnessRequireBlockingStandardStreamsCatchEntered, node: node.Parent.Parent})
			}
			if !analysis.ordered {
				return
			}
			awaits := (node.Kind == ast.KindForOfStatement && node.AsForInOrOfStatement().AwaitModifier != nil) ||
				(node.Kind == ast.KindVariableStatement && node.AsVariableStatement().DeclarationList.Flags&ast.NodeFlagsAwaitUsing == ast.NodeFlagsAwaitUsing)
			if awaits {
				b.Emit(event{kind: correctnessRequireBlockingStandardStreamsBlock, node: node})
			}
		},
	})

	// Every exit of this root, and in an ordered file every call that may block and every `await`
	// that runs while it does, must have been placed in the graph. One the graph never placed (in a
	// catch binding, or a class initializer that runs with the statement declaring it) could let a
	// path pass it unseen, so the root is given up: a missed finding, never a false one.
	var check func(node *ast.Node) bool
	check = func(node *ast.Node) bool {
		if unrecordable {
			return true
		}
		if node.Kind == ast.KindCallExpression && control_flow_graph.RootOf(node) == root &&
			correctnessNoProcessExitAfterOutputIsExit(ctx, node) && !recorded[node] {
			unrecordable = true
			return true
		}
		if analysis.ordered && !recorded[node] && correctnessRequireBlockingStandardStreamsIsInRunOf(node, root) &&
			(node.Kind == ast.KindAwaitExpression ||
				(correctnessRequireBlockingStandardStreamsIsCallLike(node) && analysis.callMayBlock(node))) {
			unrecordable = true
			return true
		}
		node.ForEachChild(check)
		return unrecordable
	}
	root.ForEachChild(check)
	if unrecordable {
		return nil, nil
	}

	walk := correctnessNoProcessExitAfterOutputWalk{root: root, chains: map[*ast.Node]string{}, tries: map[*ast.Node]int{}}
	reported := map[*ast.Node]bool{}
	started := map[*ast.Node]bool{}
	var startedOrder []*ast.Node
	start := func(function *ast.Node) {
		if !started[function] {
			started[function] = true
			startedOrder = append(startedOrder, function)
		}
	}
	type position struct {
		block *control_flow_graph.Block[event]
		state correctnessNoProcessExitAfterOutputState
	}
	visited := make([]map[string]bool, len(graph.Blocks))
	queue := []position{{block: graph.Blocks[0]}}
	visited[0] = map[string]bool{"": true}
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		state := current.state
		ended := false
		for _, recordedEvent := range current.block.Events {
			switch recordedEvent.kind {
			case correctnessRequireBlockingStandardStreamsWrite:
				state = state.with(walk.chainOf(recordedEvent.node))
			case correctnessRequireBlockingStandardStreamsCatchEntered:
				state = state.entering(walk.tryId(recordedEvent.node))
			case correctnessRequireBlockingStandardStreamsCall:
				if !recordedEvent.argumentsMayBlock {
					for _, function := range recordedEvent.runs {
						start(function)
					}
				}
				// In a file that blocks anywhere, a function handed on may run after the block.
				if !analysis.ordered {
					for _, function := range recordedEvent.handsOn {
						start(function)
					}
				}
			case correctnessRequireBlockingStandardStreamsExit:
				if len(state) > 0 {
					reported[recordedEvent.node] = true
				}
				ended = true
			case correctnessRequireBlockingStandardStreamsBlock:
				ended = true
			}
			if ended {
				break
			}
		}
		if ended {
			continue
		}
		key := state.key()
		for _, successor := range current.block.Successors {
			if successor == nil || !successor.Reachable {
				continue
			}
			index := successor.Index()
			if visited[index] == nil {
				visited[index] = map[string]bool{}
			}
			if visited[index][key] {
				continue
			}
			visited[index][key] = true
			queue = append(queue, position{block: successor, state: state})
		}
	}

	var exits []*ast.Node
	for exit := range reported {
		exits = append(exits, exit)
	}
	return exits, startedOrder
}

// correctnessRequireBlockingStandardStreamsIsInRunOf says whether node runs when root runs, rather
// than when a function inside root is later called. A class's initializers and static blocks run
// with the root that declares the class, so they count.
func correctnessRequireBlockingStandardStreamsIsInRunOf(node *ast.Node, root *ast.Node) bool {
	for current := node.Parent; current != nil && current != root; current = current.Parent {
		if correctnessRequireBlockingStandardStreamsIsFunctionWithBody(current) {
			return false
		}
	}
	return true
}
