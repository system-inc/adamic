// Package rule defines what a rule is and how it speaks.
//
// Every rule in cohere walks the same AST that the type checker already built, in the same process
// and the same address space. That is the whole architecture in one sentence, and it is why rule
// number five hundred costs what rule number one hundred costs. Measured on a 3,416-file codebase,
// 107 rules living inside the linter parse 24.7 MB and run in 0.19s; 53 rules of identical shape
// living behind a JavaScript boundary cost 1.81s, entirely for the crossing.
//
// The shape here is adapted from typescript-eslint/tsgolint (MIT), which solved it well against
// this same checker.
package rule

import (
	"crypto/sha256"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
)

// Message is what a rule tells a reader when it fires.
//
// Id is stable and machine-facing: it survives rewording, so suppressions and metrics keep
// resolving. Description is the sentence a person reads, and it carries the reasoning rather than
// just the verdict — a rule that only says what is wrong gets disabled the first time it is
// inconvenient, while one that says why gets fixed.
type Message struct {
	Id          string
	Description string
}

// Fix is a proposed replacement of a range of source text.
//
// A rule proposes; it never applies. The edit engine owns what actually lands, because it is the
// only component that can see every proposal at once and reason about the cases a single rule
// cannot: two fixes overlapping the same bytes, one fix shifting the offsets every later fix was
// computed against, and a rewrite that parses as something other than what the author meant.
//
// An empty Text is a deletion.
type Fix struct {
	Range core.TextRange
	Text  string
}

// Suggestion is a fix that must not be applied automatically.
//
// The distinction is intent, not confidence. A fix preserves what the code means; a suggestion
// changes it, and a human has to agree. Awaiting a non-promise is a fix — removing the await cannot
// alter behavior. Adding a missing switch case is a suggestion, because only the author knows what
// belongs in the body.
type Suggestion struct {
	Message Message
	Fixes   []Fix
}

// Diagnostic is one finding: a range, what is wrong there, and optionally how to repair it.
type Diagnostic struct {
	RuleName    string
	Range       core.TextRange
	Message     Message
	SourceFile  *ast.SourceFile
	Fixes       []Fix
	Suggestions []Suggestion
}

// Location is where a finding is, as every view of it names the place: the file's name, and the 1-based
// line and byte column of the range's start. Empty and zero for a finding with no file. cohere's own
// printers and the in-process runner (rule_runner) both read it, so the two name one place one way.
func (diagnostic Diagnostic) Location() (fileName string, line int, column int) {
	if diagnostic.SourceFile == nil {
		return "", 0, 0
	}
	line, character := scanner.GetECMALineAndByteOffsetOfPosition(diagnostic.SourceFile, diagnostic.Range.Pos())
	return diagnostic.SourceFile.FileName().AsString(), line + 1, character + 1
}

// Context is what a rule is handed for one file.
//
// TypeChecker is present for every rule rather than gated behind a separate tier, and that is the
// point of building in the checker's own language: asking a type question is a function call rather
// than a subprocess. The tool this replaces spent 2.09 seconds per run rebuilding a 9,530-file
// program in a separate process to answer four such questions.
//
// TypeChecker is nil only when the program could not be built, which the caller reports as a
// type-phase failure before any rule runs.
type Context struct {
	SourceFile  *ast.SourceFile
	Program     Program
	TypeChecker *checker.Checker

	// Report emits a finding. Prefer the helpers below, which spare a rule from restating how to
	// turn a node into a range.
	//
	// Report and RecordNote are good only during Run and the listeners Run returns, for this file. The
	// walk makes each rule's pair once per worker and points it at whichever file it is dispatching,
	// so a rule that keeps them past its file reports into another file, or panics between files.
	Report func(Diagnostic)

	// FileCache holds work that is expensive per file and identical for every rule that wants it.
	//
	// One walk serves every rule, which is the whole architecture, but that only covers work the
	// walk itself does. A rule that derives something from the file outside the walk pays for it
	// alone, and three rules deriving the same thing pay three times. That is not hypothetical:
	// `cohere --timing` measured three comment rules at 1,777ms combined, each visiting exactly one
	// node per file, because each rescanned the same comment trivia independently. Flat node counts
	// with unequal times is the signature.
	//
	// Nil is legal and means no caching, so a harness that builds a Context by hand keeps working
	// and simply recomputes.
	FileCache *FileCache

	// RecordNote counts one note under a key. Rules call Note, which tolerates a nil RecordNote, so a
	// harness that builds a Context by hand needs none.
	RecordNote func(key string)
}

// FileCache memoizes per-file derived work across the rules that share it.
//
// Deliberately untyped and keyed by string rather than holding named fields. This package must not
// know what a comment scan is, or a scope table, or whatever the next expensive shared derivation
// turns out to be. The rule packages own those; this owns only the fact that they are worth
// computing once.
//
// Not safe for concurrent use, and it does not need to be: one cache is created per file and every
// rule for that file runs on the one goroutine that owns it. Sharing a cache across files or
// workers would be a bug, which is why nothing here has a mutex to make that look safe.
type FileCache struct {
	entries map[string]any

	// fills counts how many times each derivation was computed, keyed the same way the entries are.
	// Once per key is the cache working.
	fills map[string]int

	// aroundFill, when set, runs each fill, so the caller can say whose work the fill is.
	//
	// This exists because per-rule timing lies about shared work. Whichever rule asks for a
	// derivation first pays for it, and since files are walked in parallel the identity of that
	// rule varies per file. Measured on three comment rules sharing one scan: 171ms, 132ms, and
	// 1.0ms for identical work, where the 1.0ms rule was simply the one that asked last. Under
	// --timing the walk sets this to bill the fill to the derivation rather than to the rule that
	// asked, which is the only reading that stays true as the order changes.
	aroundFill func(key string, fill func())
}

// NewFileCache returns a cache for one file.
func NewFileCache() *FileCache {
	return &FileCache{
		entries: map[string]any{},
		fills:   map[string]int{},
	}
}

// Fills reports how many times each derivation was computed in this file, by key.
//
// Empty for a cache that was never filled, which is the common case for a file no comment rule
// looked at.
func (c *FileCache) Fills() map[string]int {
	if c == nil {
		return nil
	}
	return c.fills
}

// SetAroundFill makes every later fill run inside around. Nil restores plain fills.
func (c *FileCache) SetAroundFill(around func(key string, fill func())) {
	if c == nil {
		return
	}
	c.aroundFill = around
}

// Cached returns the value stored under key, computing it once on the first ask.
//
// A nil cache computes every time rather than failing, so a rule reads the same whether or not the
// caller supplied one. That keeps the fast path an optimization rather than a requirement.
func Cached[Value any](cache *FileCache, key string, compute func() Value) Value {
	if cache == nil || cache.entries == nil {
		return compute()
	}
	if existing, isCached := cache.entries[key]; isCached {
		if typed, isTyped := existing.(Value); isTyped {
			return typed
		}
		// Two callers used one key for different types. Recomputing is the safe answer, and it is
		// silent on purpose: the alternative is a rule package crashing a lint run over a cache.
		return compute()
	}
	var computed Value
	if cache.aroundFill != nil {
		cache.aroundFill(key, func() { computed = compute() })
	} else {
		computed = compute()
	}
	cache.fills[key]++

	cache.entries[key] = computed
	return computed
}

// Listeners maps an AST node kind to the function a rule wants called when the walk reaches it.
//
// One walk serves every rule. A rule that asked for its own pass over the tree would multiply the
// only genuinely unavoidable cost in the tool by the number of rules, which is the arrangement this
// design exists to refuse.
type Listeners map[ast.Kind]func(node *ast.Node)

// TypeReach is what a type-aware rule can see of imported files. See Rule.TypeReach.
type TypeReach string

const (
	// TypeReachContents keys a type-aware rule's findings on the bytes of its file's whole import closure.
	TypeReachContents TypeReach = ""

	// TypeReachShapes keys them on the closure's shapes, for a rule that never reads into an imported
	// function's body.
	TypeReachShapes TypeReach = "Shapes"
)

// NoListenerKind is why a rule may register no listener on a file it was offered and still have done
// its job there.
//
// Coverage cannot tell three rules apart by their numbers: offered files, listened to none, reported
// nothing. One answers from the file in Run and needs no walk, one declines every file it is not about
// and met none, and one is left inert by its options and checks nothing anywhere. The first two are
// working and the third is dead, so the rule says which it is (#j69gvka). A rule that declares nothing
// is expected to listen, and registering no listener anywhere is then reported as checking nothing.
type NoListenerKind string

const (
	// NoListenerUndeclared is a rule that listens on the files it is about, so no listener anywhere
	// means it checked nothing.
	NoListenerUndeclared NoListenerKind = ""

	// NoListenerAnswersInRun is a rule that does its whole job in Run and never registers a listener:
	// it reads what it needs off the file, reports, and returns nil.
	NoListenerAnswersInRun NoListenerKind = "AnswersInRun"

	// NoListenerDeclinesIrrelevantFiles is a rule that registers listeners only on files that can hold
	// what it checks (by path, by language, or by what the file imports), so a tree with no such file
	// gives it none.
	NoListenerDeclinesIrrelevantFiles NoListenerKind = "DeclinesIrrelevantFiles"
)

// Rule is a name, and a function that returns what it wants to listen to.
//
// Run is called once per file, so a rule may allocate per-file state in its closure and read
// ctx.SourceFile while deciding whether to listen at all. Returning nil listeners is how a rule
// declines a file cheaply — the discipline that matters most for speed, since the cheapest rule is
// one that looks at a filename and stops.
type Rule struct {
	Name string
	Run  func(ctx Context, options any) Listeners

	// NoListener declares why this rule may register no listener on a file it was offered and still
	// have done its job. See NoListenerKind. Left undeclared, a rule offered files that registers no
	// listener on any of them is reported by default as having checked nothing.
	NoListener NoListenerKind

	// NeedsTypeChecker declares that this rule reads ctx.TypeChecker.
	//
	// The checker is handed out under an exclusive per-file lock, so acquiring it serializes the
	// whole walk of that file against every other file's walk. Declaring the need lets the walk skip
	// the acquisition entirely on files where no applicable rule wants types.
	//
	// Measured before this field existed: the exclusive lock cost the lint phase about 50%, 366-417ms
	// against 557-575ms on the same tree, paid on every file for rules that never asked a type
	// question.
	//
	// # How much the skip is still worth, as a dated measurement rather than a standing claim
	//
	// **2026-08-25: 44 of 216 registered rules declare this, roughly a fifth.** So the acquisition
	// is skipped on a file only when none of that fifth applies to it, which is a real saving and is
	// nothing like the blanket one this comment used to describe.
	//
	// The share is what the claim rests on rather than either count, and
	// `TestCheckerDeclarationShareStillSupportsTheConclusion` asserts it as a band for that reason.
	// A porting wave moves the total on every commit without touching what this paragraph concludes;
	// a shift in the ratio is what would make it false.
	//
	// It previously read "every file in the current catalog: of 112 rule files exactly one reads the
	// checker, and it is the tsgolint adapter, which registers no rules yet." That was true when it
	// was written and false the moment the adapter registered, and nothing announced the change. It
	// is quoted here rather than deleted because it is the evidence for the rule underneath: **a rule
	// count in a comment decays silently, so write it with a date and treat it as a measurement that
	// expires.**
	//
	// A rule that leaves this false and then reads ctx.TypeChecker gets nil, which is the same thing
	// it gets when the program fails to build. That is deliberate: a rule silently reading a checker
	// it did not declare would reintroduce the data race this lock exists to prevent, and a nil
	// dereference is a loud failure where a race is a quiet one.
	//
	// # Why there is no shared references-to-a-binding helper
	//
	// Several rules want to know whether an identifier refers to a particular binding, and the
	// obvious move is a shelf helper that answers it. We decided against one, and the cost above is
	// the reason: a helper whose cost is a 50% serialization of the file walk is not a helper, it is
	// a decision, and putting it on the shelf hides the decision behind an import. A helper also has
	// to be correct for every caller, so it would have to be scope-aware, so it would have to take
	// the checker, so every caller would pay whether or not its own rule could be shadowed.
	//
	// So the question is answered per rule, at the cheapest tier that is correct for that rule:
	//
	//	name matching     enough where shadowing is impossible. core.NoExAssign resolves its
	//	                  binding this way, because a catch clause gives it a subtree to stop at.
	//	                  That is a claim about the tier and not a warrant for the rule: the same
	//	                  rule shipped for weeks missing `e++` entirely, which is a write-detection
	//	                  gap rather than a scope one, and neither corpus tested the shape.
	//	symbol identity   where shadowing is possible. Resolve the binding once with
	//	                  GetSymbolAtLocation, resolve each candidate, compare symbols. Declare
	//	                  NeedsTypeChecker so files with no such rule skip the acquisition.
	//
	// Note which question that answers. Find-all-references answers "where is this used", and these
	// rules already know where to look, since they walk one file. They need "is this the same
	// binding", which is a symbol comparison. Reaching for a reverse index here would be answering a
	// forward question with the wrong instrument.
	//
	// Symbol comparison is unmeasured on our tree. The first rule that needs it times it with
	// --timing against planted violations, and that number decides whether the rules after it follow
	// or stay on name matching with the shadowing hazard written at their own site.
	NeedsTypeChecker bool

	// ProgramReads declares what this rule reads through ctx.Program beyond the file it was handed.
	// See ProgramRead for the kinds and what covers each.
	//
	// The declaration exists for the findings cache, and the failure it prevents is the one this tool
	// exists to catch. A cache keyed on one file serves a stale result when something the rule also
	// read has changed and the key has not: zero findings, forever, indistinguishable from a clean
	// tree. So it cannot live in a convention, and ctx.Program enforces it: a method outside the
	// declaration panics with the rule's name, and the walk names the panic on every run.
	//
	// It replaced ReadsProgram, one flag for every kind, which kept 45 rules out of the cache although
	// most read only compiler options, the default library, or their own file's module resolution,
	// all of which the key already covers. Declaring more than a rule reads costs a cache miss;
	// declaring less is refused at the read. A rule reading a helper's program methods declares what
	// the helper reads.
	ProgramReads ProgramRead

	// ProgramFingerprint, on a rule that declares ReadsOtherFiles, is a hash of exactly the program-wide data the
	// rule's verdict on one file depends on beyond that file: the theme interfaces it indexes, the import graph's
	// edges, the one file it reads. With it, the findings cache replays the rule on a file whose bytes are
	// unchanged while this hash is too (#kdee854). It is called for each set of options the run's
	// configuration gives the rule, with a Program viewed under the rule's own ProgramReads and the options as
	// Run receives them, since an option can choose which data the rule reads: no-html-link-for-pages' pagesDir
	// names the directory its routes come from (#s9k38p3). The options themselves need no hashing: every
	// config file is in the findings cache's key. A walk calls it once for each set of options it gives the
	// rule, on whichever worker asks first, while other rules' fingerprints may be computed on other workers,
	// so it must be safe for concurrent use.
	//
	// It is a claim of completeness, and it decides soundness: anything the verdict reads outside the file and
	// outside this hash replays a stale finding when it changes. So it is proven both ways for each rule, a
	// planted change to the indexed data moving it and an edit elsewhere leaving it alone. A rule that declares
	// ReadsOtherFiles without one is walked on every file, as before: the default is always the sound one.
	ProgramFingerprint func(program Program, options any) [sha256.Size]byte

	// TypeReach declares what a type-aware rule can see of the files its file imports, and so what its
	// cached findings are keyed on. Contents, the default, keys them on the bytes of every file in the
	// import closure. Shapes keys them on each imported file's shape: its declaration output and its text
	// with function bodies cut out (program.SignatureEntry), so an edit inside a body of a file imported
	// by half the tree re-runs this rule on that one file rather than on half the tree.
	//
	// Shapes is a claim that the rule never reads into an imported function's body, which a shape does
	// not cover. The claim is checked rather than trusted: a scan of every rule and the helpers it
	// reaches refuses Shapes on any rule that can reach both an imported declaration and a body
	// (TestRulesClaimShapesOnlyWhereTheScanAllowsIt). Claiming wrongly replays a stale finding; not
	// claiming costs a re-run.
	TypeReach TypeReach

	// ResolvesReactValueTypes declares that this rule identifies a React value by asking the checker
	// for its TYPE, so a file where the hook call resolves to `any` costs it every finding it would
	// otherwise make, silently.
	//
	// This is narrower than NeedsTypeChecker and the difference is the whole reason the field exists.
	// Measured on 2026-08-24: of the forty-one rules declaring NeedsTypeChecker, only the two that
	// call GetTypeAtLocation and read the resulting type go blind when `useState` resolves to `any`.
	// The rest ask GetSymbolAtLocation, which resolves a BINDING rather than a type, and a binding to
	// `declare function useState<T>(initial: T): any` resolves exactly as well as one to the real
	// declaration. Probed both ways on the same input: under the ambient shim the symbol comes back
	// named `useState` with one declaration, while the call's type comes back `any` with TypeFlagsAny.
	//
	// So a message naming every NeedsTypeChecker rule as blinded would be confidently wrong about
	// thirty-nine of them, and a message naming a hardcoded list would be confidently wrong the week a
	// fifth rule ships. `structure/react-hook-no-any-type` reads this flag off the live catalog instead,
	// which is why the flag is a declaration on the rule rather than a list somewhere else: the rule
	// that goes blind is the only thing that knows it does.
	//
	// Under-declaring costs a reader the knowledge that suppressing the tripwire disables this rule
	// too. Over-declaring names a rule that would have survived. Neither is silent, which is what
	// makes this field cheaper to get wrong than the two above it.
	ResolvesReactValueTypes bool
}

// ReportNode is the common case: this node is wrong, here is why.
//
// The finding is anchored on the node's own text rather than on node.Loc, and that is the whole
// difference between a usable finding and an unsuppressable one.
//
// Loc.Pos() sits before leading trivia, so a node preceded by a comment reports at the comment. A
// real case from the tree this gates: in modules/mcp/McpApi.ts the binding `params` is on line 242
// with its `eslint-disable-next-line` on 241, correctly covering it. Anchored on Loc the rule
// reported 240, swallowing both the comment and the indentation. **A `-next-line` directive only
// matches the line after itself, so that finding could not be suppressed by anything the author was
// able to write.**
//
// It is worse than a cosmetic offset because it is invisible in every count. A finding at the wrong
// line still reads as a real finding, costs nothing observable, and surfaces only as a suppression
// that mysteriously does not work. This root cause produced five distinct symptoms in one night: a
// fix that ate whitespace, an exemption that failed to fire, a line span measured from the wrong
// end, three findings reported at the wrong node, and this.
//
// So it is fixed here rather than at 30 call sites across 16 rule files. Same argument as the fix
// builders taking a Context: a rule should not be able to get this wrong by writing the obvious
// thing.
//
// A rule that genuinely wants the trivia included says so with ReportRange.
// Note records one occurrence of a fact the rule wants counted rather than reported: an exemption it
// applied, say, so that a tag silencing writes shows up as a number someone reads instead of as nothing.
// A note is not a finding and fails nothing. The walk counts notes per file, per rule and per key.
func (c Context) Note(key string) {
	if c.RecordNote != nil {
		c.RecordNote(key)
	}
}

// SkippedNotePrefix opens the key of every note Skip records, which is how --coverage tells a skip apart
// from the facts a rule counts while it judges.
const SkippedNotePrefix = "skipped: "

// Skip records that the rule declined this file because a compiler option or a missing precondition left
// it unable to judge, and says which. Call it once, then return no listeners. A rule that declines this
// way otherwise reads exactly like a rule that looked and found nothing: no-useless-default-assignment
// declined every file in ahra and www while every run read green (#6ar414z, #pa7k7zv). It is not for a
// file the rule is simply not about, a `.d.ts` or a file outside its directory, which is its scope.
func (c Context) Skip(reason string) {
	c.Note(SkippedNotePrefix + reason)
}

// Cover is another check that reports what a rule declined to, named so a run can tell whether it ran.
type Cover string

// CoverTypeCheck is the types phase: the compiler reporting the same thing itself, under an option the
// rule found on.
const CoverTypeCheck Cover = "types"

// CoveredSkipNotePrefix opens the key of every note SkipCovered records: the cover, then the reason.
const CoveredSkipNotePrefix = "skipped, covered by "

// SkipCovered records that the rule declined this file because another check reports the same thing, and
// names that check. A skip whose cover ran this run left nothing unchecked, so the footer does not mark it;
// one whose cover did not run (`--lint`, or a types phase that bailed) is a gap like any other skip. A rule
// declining for any other reason calls Skip: choosing between the two is how a skip says which it is.
func (c Context) SkipCovered(cover Cover, reason string) {
	c.Note(CoveredSkipNotePrefix + string(cover) + ": " + reason)
}

func (c Context) ReportNode(node *ast.Node, message Message) {
	c.Report(Diagnostic{
		Range:      TokenRange(c.SourceFile, node),
		Message:    message,
		SourceFile: c.SourceFile,
	})
}

// ReportNodeWithFixes reports a node and proposes repairs the edit engine may apply unattended.
//
// Anchored on the node's own text, for the reason spelled out on ReportNode. This one matters twice
// over: a finding reported at the wrong line while carrying a fix means the fix is applied at a
// location the reader was never shown.
func (c Context) ReportNodeWithFixes(node *ast.Node, message Message, fixes ...Fix) {
	c.Report(Diagnostic{
		Range:      TokenRange(c.SourceFile, node),
		Message:    message,
		SourceFile: c.SourceFile,
		Fixes:      fixes,
	})
}

// ReportNodeWithSuggestions reports a node and offers repairs that need a human to choose them.
func (c Context) ReportNodeWithSuggestions(node *ast.Node, message Message, suggestions ...Suggestion) {
	c.Report(Diagnostic{
		Range:       TokenRange(c.SourceFile, node),
		Message:     message,
		SourceFile:  c.SourceFile,
		Suggestions: suggestions,
	})
}

// ReportRange reports a span that is not exactly one node — a portion of a string literal, or the
// gap between two tokens.
func (c Context) ReportRange(textRange core.TextRange, message Message) {
	c.Report(Diagnostic{
		Range:      textRange,
		Message:    message,
		SourceFile: c.SourceFile,
	})
}

// ReportRangeWithSuggestions reports a span that is not one node and offers repairs a human chooses.
//
// The node helpers came in three shapes and the range helpers in one, so a rule reporting a sub-range
// of a string literal with suggestions had no helper and had to hand-build a Diagnostic. That works
// and it is the wrong thing to make a rule author do: `Report` exists for the cases the helpers do not
// cover, and every hand-rolled Diagnostic is a place the Range can be built from `Loc` instead of
// `TokenRange` without anything downstream noticing.
//
// The shape was already settled elsewhere. The tsgolint adapter has exactly this method, because an
// adapted rule needed it first. Native rules simply did not have it, which is the kind of asymmetry
// that is invisible until someone writes the rule that falls in the gap.
func (c Context) ReportRangeWithSuggestions(textRange core.TextRange, message Message, suggestions ...Suggestion) {
	c.Report(Diagnostic{
		Range:       textRange,
		Message:     message,
		SourceFile:  c.SourceFile,
		Suggestions: suggestions,
	})
}

// ReportRangeWithFixes reports a span that is not one node and proposes repairs the engine applies
// unattended.
//
// The last corner of the same asymmetry ReportRangeWithSuggestions describes. The node helpers came
// in three shapes and the range helpers had reached two, so a rule whose finding is a computed span
// and whose repair is safe to apply had no helper and would have hand-built a Diagnostic.
//
// No trimming happens here, and that is the point rather than an omission. A range helper exists
// precisely because the caller computed a span the AST does not name, so a helper that adjusted it
// would be second-guessing the only party that knows what it means.
//
// The fixture harness has looked for this method name since before it existed: fixture_pair_test
// matches on `ReportRangeWithFixes` alongside `ReportNodeWithFixes` when deciding whether a rule
// proposes a fix. A rule reporting a computed span with a fix through a hand-built Diagnostic would
// have read to that guard as proposing nothing.
func (c Context) ReportRangeWithFixes(textRange core.TextRange, message Message, fixes ...Fix) {
	c.Report(Diagnostic{
		Range:      textRange,
		Message:    message,
		SourceFile: c.SourceFile,
		Fixes:      fixes,
	})
}

// TokenRange is a node's own text, without the trivia that precedes it.
//
// This distinction is the single sharpest edge in the fix API, and getting it wrong produces damage
// that no downstream check can catch.
//
// `node.Loc.Pos()` is the position *before* leading trivia, not the start of the node's own text.
// For `import * as X from 'fs'`, the module specifier's Loc spans `" 'fs'"` — the space is inside
// the range. A fix built from Loc therefore replaces the whitespace too:
//
//	import * as X from 'fs'   ->   import * as X from'node:fs'
//
// and `from /* pinned */ 'fs'` loses the comment permanently.
//
// The reason this is worse than cosmetic: **the corrupted output still parses.** The edit engine
// refuses a rewrite that breaks syntax, and that guard never fires here, so the damage sails
// through with every downstream check green. A range wider than the rule intended is damage that
// validation downstream is structurally unable to detect. The parse guard is necessary and not
// sufficient.
//
// Trimming needs the SourceFile, because finding where a token actually starts means scanning
// forward past the trivia. That is why the helpers below hang off Context: a rule always has one,
// so the safe form is also the convenient one, and the shape that eats trivia is not reachable by
// accident.
//
// Found by @system_cohere_lint_fix, running real rules through the engine on three trivia shapes
// rather than reasoning about the ranges. Upstream tsgolint gets this right via
// `utils.TrimNodeTextRange`; our port dropped the step, most likely because the signature had no
// SourceFile to trim with.
//
// This is also why vendoring tsgolint's `rule/` package is declined while its `utils/` is taken.
// Its builders are correct and do trim, but they take the SourceFile as a parameter rather than a
// receiver (`RuleFixReplace(file, node, text)`), so the call that forgets it is available. Ours is
// not. Exposing both would give a rule author two reachable spellings of one act, one of them
// unsafe, and two helpers that almost agree drift permanently.
//
// The first token's start is read with scanner.GetTokenPosOfNode, the compiler's own trivia skip. It used to
// come from scanner.GetRangeOfTokenAtPosition, which builds a scanner on every call: 252 MB and 1.5M objects of
// a cold ahra run, for a position a trivia skip gives directly (#9xfg09f). The text is read inside the compiler
// rather than with sourceFile.Text() here, since the shape-keyed rules' guard cannot tell this file from an
// imported declaration's (TestRulesClaimShapesOnlyWhereTheScanAllowsIt).
//
// GetTokenPosOfNode is a plain trivia skip, the scanner's answer, except for four shapes it reads on purpose
// another way: a missing node, whose position it returns unskipped; JSDoc and JSX text, where it stops at a
// comment; and a node inside JSDoc. Those keep the scanner. TestTokenRangeStartsWhereTheScannerDoes holds the
// two to the scanner's start on every node of the fixtures and of the four consumers, and found exactly those
// shapes when the scanner was dropped for every node (14,522 of 6,024,171).
func TokenRange(sourceFile *ast.SourceFile, node *ast.Node) core.TextRange {
	if sourceFile == nil || node == nil {
		return node.Loc
	}
	if ast.NodeIsMissing(node) || ast.IsJSDocNode(node) || node.Kind == ast.KindJsxText || node.Flags&ast.NodeFlagsJSDoc != 0 {
		return scanner.GetRangeOfTokenAtPosition(sourceFile, node.Pos()).WithEnd(node.End())
	}
	return core.NewTextRange(scanner.GetTokenPosOfNode(node, sourceFile, false), node.End())
}

// NodeText is a node's own source text, from its first token to its end, without the leading trivia
// node.Pos() includes: what a reader sees as the node, and what a message or a fix quotes.
//
// One home for the slice of TokenRange that three rules each wrote privately (no-implicit-coercion,
// prefer-exponentiation-operator, valid-typeof), so the next rule reaches for this instead of a fourth copy.
//
// A method on the context, reading the context's own file, rather than a function taking a file: the
// type-reach guard counts text read out of any other file as descending into an imported body, so the
// only spelling offered is the one that cannot (internal/guard/type_reach_test.go).
func (c Context) NodeText(node *ast.Node) string {
	span := TokenRange(c.SourceFile, node)
	return c.SourceFile.Text()[span.Pos():span.End()]
}

// The fix helpers come in two families, and picking the wrong one is the most common way a correct
// rule produces a wrong edit.
//
// The node forms below (ReplaceNode, RemoveNode, InsertBefore, InsertAfter) are methods on Context
// because they trim to the token and trimming needs the SourceFile. Reach for these by default: a
// rule that says "this node" means the node, not the comment above it. The range forms
// (ReplaceRange, RemoveRange) are package-level and trim nothing, because a caller that computed its
// own span already said exactly what it meant. That split is why the safe form is also the
// convenient one, and why the trivia-eating shape is unreachable rather than merely discouraged.
//
// Two properties of the engine that consumes these are worth knowing before writing a fix.
//
// A Fix must preserve meaning, because it is applied unattended. When the correct edit requires
// choosing between alternatives only the author can rank, that is a Suggestion instead, and the
// engine never applies suggestions. core.NoCaseDeclarations is the shipped example: wrapping a case
// clause in braces changes what the code means, so it reports a Suggestion even though the edit is
// mechanical and unambiguous. Mechanical is not the test; meaning-preserving is.
//
// Fixes reported together are not applied together. ProposalsFrom flattens a diagnostic's fixes into
// independent proposals, so overlap resolution may admit one and refuse the other. A pair that only
// means something jointly needs grouping in the engine first; reporting them side by side does not
// buy atomicity. core.NoCaseDeclarations shows this too: its InsertBefore and InsertAfter are an
// opening and a closing brace, and half of that pair is broken code. It is safe today only because
// it is a Suggestion and never reaches the engine. See ProposalsFrom in internal/fix.
//
// The engine's refusal guard checks that the rewritten file parses, and parsing is necessary rather
// than sufficient. Anything sayable in valid syntax about a name that no longer exists slips
// through it: a rename that updates a usage and misses its declaration produces a file that parses
// and does not compile, and the guard has no reason to fire.
//
// That is not hypothetical. structure/react-component-require-properties-type-suffix renames a
// component's properties type at both the usage and the declaration, using a map that one listener
// fills and the other reads. Source order decides which listener runs first, so on the conventional
// declaration-first ordering the declaration is visited against an empty map and skipped: the usage
// is renamed, the declaration is not, and the file then references a type that does not exist. It
// parses. Measured: one diagnostic on that ordering against two on the other, and the rewritten
// source fails to type-check with "cannot find name".
//
// So a fix that renames anything carries an obligation the engine cannot discharge for it: every
// reference to the old name must move in the same pass, and a fixture has to assert it on both
// orderings, since either one alone passes. The failure also repairs the evidence of itself, because
// the rule's own fix silences the finding that would have reported the rest of the work.

// ReplaceNode proposes replacing a node's own text, leaving the trivia before it untouched.
func (c Context) ReplaceNode(node *ast.Node, text string) Fix {
	return Fix{Range: TokenRange(c.SourceFile, node), Text: text}
}

// RemoveNode proposes deleting a node's own text.
//
// Leading trivia is deliberately left behind rather than swept up with the node. Removing a node
// and removing the blank line above it are different intentions, and a helper that guessed would
// be wrong half the time. A rule wanting the surrounding whitespace gone should say so with
// ReplaceRange over a range it computed itself.
func (c Context) RemoveNode(node *ast.Node) Fix {
	return Fix{Range: TokenRange(c.SourceFile, node), Text: ""}
}

// InsertBefore proposes inserting text immediately before a node's own text.
//
// Before the token rather than before its trivia, which is almost always what a rule means: adding
// a modifier to a declaration should land next to the declaration, not above the comment that
// documents it.
func (c Context) InsertBefore(node *ast.Node, text string) Fix {
	tokenRange := TokenRange(c.SourceFile, node)
	return Fix{Range: tokenRange.WithEnd(tokenRange.Pos()), Text: text}
}

// InsertAfter proposes inserting text immediately after a node, touching no existing bytes.
//
// No trimming needed: a node's End is already past its own text, and trailing trivia belongs to
// whatever comes next.
func (c Context) InsertAfter(node *ast.Node, text string) Fix {
	return Fix{Range: node.Loc.WithPos(node.End()), Text: text}
}

// ReplaceRange proposes replacing an arbitrary span.
//
// The escape hatch for a rule that computed its own range: a portion of a string literal, or the
// gap between two tokens. Nothing is trimmed, because the caller already said exactly what it
// meant.
func ReplaceRange(textRange core.TextRange, text string) Fix {
	return Fix{Range: textRange, Text: text}
}

// RemoveRange proposes deleting an arbitrary span.
func RemoveRange(textRange core.TextRange) Fix {
	return Fix{Range: textRange, Text: ""}
}
