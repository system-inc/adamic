// Package rename performs a program-wide symbol rename: the command-line equivalent of an editor's
// rename refactor, answered from the type graph rather than from a text search.
//
// # Why this is a verb and not a fix
//
// Every other write in this tool is reactive. A rule finds something wrong and proposes a repair,
// the blast radius is one file, and the engine's parse guard is a sufficient backstop because the
// rule already decided the edit was safe. A rename is the opposite shape: the user asserts an
// intention, nothing is wrong with the code, and the blast radius is wherever the symbol lives —
// which is a program-wide answer no rule is in a position to give. `internal/rule/rule.go` states
// the consequence at length and from experience: a rename that updates a usage and misses its
// declaration produces a file that parses and does not compile, and the engine's guard has no
// reason to fire. So this gets its own verb, its own confirmation, and its own dry-run default,
// rather than being smuggled in under `--fix` where the safety argument for autofix does not reach.
//
// # What makes it answerable here at all
//
// A reference is a checker query rather than a heuristic, because the whole type graph is resident
// in this process. `internal/unused` established the technique this builds on: walk every
// identifier, ask the checker what it resolves to, and compare symbol pointers. Measured on the
// ahra tree for `usePreviousValue`, a hook exported from the structure library: grep reports nine
// lines across five files, the checker reports four reference sites across two, and every one of
// the five differences is a comment or a markdown file. grep over-matched by five and the checker
// missed nothing.
//
// # The five things that decide whether a rename is correct
//
// Each is handled below and each is measured rather than assumed, because four of the five produce
// a program that still parses when they are got wrong, which means no downstream guard catches
// them.
//
//	OUT OF PROJECT      A declaration in node_modules or in a dependency's `.d.ts` is not the
//	                    user's to rename, and rewriting it silently breaks the package for
//	                    everything else that imports it. Refused by name.
//	SHORTHAND PUNNING   `{ err }` is `{ err: err }`. The property name is part of the object's
//	                    contract and is a DIFFERENT symbol from the binding. Measured on a fixture:
//	                    GetSymbolAtLocation on the shorthand's identifier returns a symbol with
//	                    flags 4 (Property) while GetShorthandAssignmentValueSymbol returns flags 2
//	                    (BlockScopedVariable) pointing at the declaration. They are not equal, so
//	                    renaming the binding must EXPAND the shorthand to `err: error` rather than
//	                    replace the identifier.
//	ALIASED IMPORTS     `import { err as failure }` has two halves that are two symbols. Measured:
//	                    renaming the exported `err` reports the source half at its own column and
//	                    leaves `failure` untouched, which is the correct answer in both directions.
//	STRING KEYS         `t["err"]` is invisible to the checker and always will be. This is recorded
//	                    rather than solved, and the verb REPORTS what it could not see instead of
//	                    implying completeness.
//	RANGES              A correct replacement over the wrong range writes the right characters into
//	                    the wrong place. Every edit here is anchored on the identifier token's own
//	                    span, never on an enclosing node.
package rename

import (
	"fmt"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/core"
)

// Edit is one replacement of one identifier's own span.
//
// Deliberately NOT a rule.Diagnostic. That type carries exactly one Range and has no secondary-span
// field, so a rename touching fourteen sites across six files would have to be either fourteen
// unrelated diagnostics with no way to say they are one act, or one diagnostic pointing at the
// declaration while thirteen edits happen somewhere the reader was never shown. Both readings are
// wrong in the way this project cares about: the second is the "fix applied at a location the
// reader was never shown" failure that ReportNodeWithFixes exists to prevent, one level up. A
// rename is one intention with many sites, and that shape needs a type that can say so.
type Edit struct {
	// FileName is the file this edit lands in.
	FileName string

	// Range is the identifier token's own span, never an enclosing node's.
	Range core.TextRange

	// Line and Column are 1-based, for printing. Derived once here rather than at every print site.
	Line   int
	Column int

	// Text is what gets written over Range.
	//
	// Usually the new name. For a shorthand property it is `oldName: newName`, because collapsing
	// that to the new name alone would silently rewrite the object's contract.
	Text string

	// Kind says what sort of site this is, so the dry run can be read rather than counted.
	Kind EditKind
}

// EditKind classifies a reference site. The classification is what a reader checks a dry run
// against: a rename touching an unexpected kind of site is the signal something was misidentified.
type EditKind string

const (
	// EditDeclaration is the declaration of the symbol itself.
	EditDeclaration EditKind = "declaration"

	// EditReference is an ordinary use.
	EditReference EditKind = "reference"

	// EditImportSource is the source half of `import { err as failure }`, which is the exporting
	// module's spelling and moves with the export rather than with the local binding.
	EditImportSource EditKind = "import source"

	// EditShorthandExpansion is `{ err }` becoming `{ err: error }`. Called out separately because
	// it is the one site where the replacement text is not simply the new name, and a reader
	// skimming a dry run should see that the shape changed.
	EditShorthandExpansion EditKind = "shorthand expansion"
)

// Blindness is something the rename could not see, reported so the user is never handed a number
// that implies completeness it does not have.
//
// This exists because of the failure it prevents, which is the worst one available here: a rename
// that silently misses a string-keyed use produces a codebase that compiles and is broken at
// runtime. Nothing downstream catches that — not the parse guard, not the type phase, not the
// tests, until something reads the property at runtime and gets undefined.
type Blindness struct {
	FileName string
	Line     int
	Column   int
	Text     string
	Reason   string
}

// Plan is everything one rename would do, and everything it could not see.
//
// Produced whole before a byte is written, which is what makes the dry run a real preview rather
// than a narration of work already done.
type Plan struct {
	OldName string
	NewName string

	// DeclarationFile and DeclarationLine name where the symbol being renamed is declared, so the
	// user can confirm the position resolved to what they meant.
	DeclarationFile string
	DeclarationLine int

	Edits        []Edit
	Blind        []Blindness
	FilesTouched int

	// Refusals are reasons this plan must not be applied. A non-empty Refusals means the verb
	// refuses: a rename that cannot be done correctly is not done at all, because a half-renamed
	// codebase compiles and is wrong, which is the exact failure this tool exists to eliminate.
	Refusals []string
}

// Candidate is one symbol a bare name resolved to, printed when a bare name is ambiguous.
type Candidate struct {
	Name     string
	FileName string
	Line     int
	Column   int
	Kind     string
}

// Position names a source location the way an editor does: file, 1-based line, 1-based column.
type Position struct {
	FileName string
	Line     int
	Column   int
}

// ParsePosition reads `path/to/File.ts:42:15`.
//
// Split from the right rather than the left, because a Windows path carries a drive colon and a
// left split would take `C` as the file name. Measured against nothing yet on Windows, and said
// here so the next reader knows it is reasoned rather than tested.
func ParsePosition(argument string) (Position, bool) {
	lastColon := strings.LastIndex(argument, ":")
	if lastColon <= 0 {
		return Position{}, false
	}
	secondColon := strings.LastIndex(argument[:lastColon], ":")
	if secondColon <= 0 {
		return Position{}, false
	}
	var line, column int
	if _, err := fmt.Sscanf(argument[secondColon+1:lastColon], "%d", &line); err != nil {
		return Position{}, false
	}
	if _, err := fmt.Sscanf(argument[lastColon+1:], "%d", &column); err != nil {
		return Position{}, false
	}
	if line < 1 || column < 1 {
		return Position{}, false
	}
	return Position{FileName: argument[:secondColon], Line: line, Column: column}, true
}

// IsValidIdentifier reports whether a name can be written as a plain identifier.
//
// Checked because the new name is written into source text verbatim. A name with a space or a
// hyphen in it produces a file that does not parse, which the write guard would catch, but catching
// it here names the actual mistake instead of reporting a syntax error in somebody else's file.
//
// Deliberately conservative: ASCII letters, digits, `_` and `$`, not starting with a digit.
// TypeScript accepts a much larger set through Unicode identifier classes, and accepting them here
// would mean this function has to agree with the scanner about a large table forever. Refusing a
// legal-but-exotic name is a false refusal the user can see and work around; accepting an illegal
// one writes a broken file.
func IsValidIdentifier(name string) bool {
	if name == "" {
		return false
	}
	for index := 0; index < len(name); index++ {
		character := name[index]
		isLetter := (character >= 'a' && character <= 'z') || (character >= 'A' && character <= 'Z')
		isDigit := character >= '0' && character <= '9'
		isPunctuation := character == '_' || character == '$'
		if !isLetter && !isDigit && !isPunctuation {
			return false
		}
		if index == 0 && isDigit {
			return false
		}
	}
	return !reservedWords[name]
}

// reservedWords are the names that cannot be a binding, so renaming to one produces source that
// does not parse or that parses as something else entirely.
//
// The strict-mode reserved words are included because our tree is modules, which are strict by
// construction. Contextual keywords like `type` and `as` are deliberately absent: they are legal
// identifiers and refusing them would be a false refusal.
var reservedWords = map[string]bool{
	"break": true, "case": true, "catch": true, "class": true, "const": true,
	"continue": true, "debugger": true, "default": true, "delete": true, "do": true,
	"else": true, "enum": true, "export": true, "extends": true, "false": true,
	"finally": true, "for": true, "function": true, "if": true, "import": true,
	"in": true, "instanceof": true, "new": true, "null": true, "return": true,
	"super": true, "switch": true, "this": true, "throw": true, "true": true,
	"try": true, "typeof": true, "var": true, "void": true, "while": true,
	"with": true, "implements": true, "interface": true, "let": true, "package": true,
	"private": true, "protected": true, "public": true, "static": true, "yield": true,
}
