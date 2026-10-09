// Whether source text can spell what a React Compiler pass recognises, answered before any lowering.
//
// Both gates here answer "maybe" or a proof of "no" from the text, and fall back to the syntax tree
// only when the text holds a backslash, because an escape can spell a name without its letters
// appearing in a row while every recogniser reads the cooked `node.Text()`.
package high_level_intermediate_representation

import (
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	utilsreact "github.com/system-inc/cohere/internal/lint/ecmascript/react"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// MayNameManualMemoization reports whether a function could contain a `useMemo` or `useCallback`
// call that `DropManualMemoization` would recognise. False is a proof that the erasure finds nothing
// in it; true is only a maybe. text is the function's own source range.
//
// Two callers decide on it whether to lower at all: `ForFunctionWithoutManualMemoization`, which
// shares the memo-intact lowering when the erasure could change nothing, and
// `react/preserve-manual-memoization`, whose findings all live inside memo blocks. Both are wrong
// in the same silent direction if this answers false for a function the erasure would have
// changed, so they share one answer rather than each keeping a copy that can drift.
//
// # Why false proves the erasure finds nothing
//
// `recogniseManualMemoCall` matches a temporary holding either a `LoadGlobal` whose `Name` is
// `useMemo` or `useCallback`, or a `Primitive` string of one of those names, the `React.useMemo`
// property or a computed `React['useMemo']`. Both come from `node.Text()` of an identifier,
// property name, or string literal inside the lowered function, so the name is spelled in that
// function's source range. An import alias does not escape this: `LoadGlobal.Name` is the local
// spelling, so `import { useMemo as memo }` is not recognised by the pipeline either.
//
// # Why a backslash falls back to the syntax tree
//
// `node.Text()` is the cooked value, and an escape can spell the name without its letters appearing
// in a row: `useMemo` as an identifier, `'use\x4Demo'` or a line continuation in a string.
// So a range with a backslash and no plain spelling is answered by reading the cooked text of every
// identifier and string-like literal under the function, which are exactly the nodes the two
// recognised instructions are lowered from. A substring search alone missed these: measured before
// this was shared, five escaped spellings of `useCallback` wrapping a setter left
// `react-hooks/set-state-in-effect` silent where the plain spelling reported, because the cache
// handed it the memo-intact graph. The walk costs a fraction of lowering, and it matters: measured on
// ahra, about 600 of 3,785 files carry a backslash, and lowering every function in them on a "maybe"
// was half of what preserve-manual-memoization still cost after the plain-text check.
func MayNameManualMemoization(functionNode *ast.Node, text string) bool {
	if SpellsManualMemoization(text) {
		return true
	}
	if strings.IndexByte(text, '\\') < 0 {
		return false
	}
	return cooksToName(functionNode, func(name string) bool { return name == "useMemo" || name == "useCallback" })
}

// SpellsManualMemoization reports whether text spells either name plainly.
func SpellsManualMemoization(text string) bool {
	return strings.Contains(text, "useMemo") || strings.Contains(text, "useCallback")
}

// MayHoldComponentOrHook reports whether this file could hold a function React Compiler would
// compile. False is a proof that none of the seven React Compiler rules that lower this file's
// functions can report in it; true is only a maybe. Computed once per file and shared.
//
// # Why false proves silence for all seven
//
// Each of them reports only inside a component or a hook, and every gate they use asks for evidence
// from the body, not only a name. `IsComponentOrHookLike` (immutability, purity, set-state-in-render
// and the two effect rules) wants JSX or a call whose callee is an identifier or a namespace member
// spelled `use[A-Z0-9]...`. refs wants a hook or component kind and then JSX, a `LoadGlobal` named
// `use[A-Z]...`, or a method call whose property string is one. static-components wants JSX. So a file
// with no possible JSX and no cooked identifier or string spelling a hook name holds no unit.
//
// JSX is possible only in a file parsed with the JSX variant, and only where a `<` appears. A hook
// name is looked for as text first, `use` followed by an uppercase letter or a digit anywhere, which
// is wider than any of the gates; a file holding a backslash and no such text is answered by the
// cooked text of its identifiers and string-like literals, the nodes every hook test reads.
//
// # Why this waited for the effect rules to gate
//
// set-state-in-effect and no-deriving-state-in-effects classify an effect hook by its type, so a
// value typed `typeof useEffect` reaches them from a file that spells no hook name. While they judged
// every function, this gate would have been inexact for them, and they lowered every function anyway,
// so gating the other five saved almost nothing. They gate on `IsComponentOrHookLike` now, as upstream
// does, and this is exact for all seven.
func MayHoldComponentOrHook(ctx rule.Context) bool {
	if ctx.SourceFile == nil {
		return true
	}
	return rule.Cached(ctx.FileCache, "hir.MayHoldComponentOrHook", func() bool {
		return mayHoldComponentOrHook(ctx.SourceFile)
	})
}

// mayHoldComponentOrHook is MayHoldComponentOrHook without the per-file cache.
func mayHoldComponentOrHook(sourceFile *ast.SourceFile) bool {
	text := sourceFile.Text()
	if sourceFile.LanguageVariant == core.LanguageVariantJSX && strings.IndexByte(text, '<') >= 0 {
		return true
	}
	if spellsHookName(text) {
		return true
	}
	if strings.IndexByte(text, '\\') < 0 {
		return false
	}
	return cooksToName(sourceFile.AsNode(), utilsreact.IsCompilerHookName)
}

// spellsHookName reports whether text holds `use` followed by an uppercase letter or a digit.
func spellsHookName(text string) bool {
	for offset := strings.Index(text, "use"); offset >= 0; {
		next := offset + 3
		if next < len(text) {
			if character := text[next]; (character >= 'A' && character <= 'Z') || (character >= '0' && character <= '9') {
				return true
			}
		}
		following := strings.Index(text[next:], "use")
		if following < 0 {
			return false
		}
		offset = next + following
	}
	return false
}

// cooksToName reports whether any identifier or string-like literal under node has cooked text the
// predicate accepts.
//
// These are the node kinds whose `Text()` lowering copies into a `LoadGlobal.Name` (an identifier
// read) or a `Primitive` string (a string literal, a template without substitutions, or the property
// name of a method call), and the kinds the syntactic hook tests read. Every other kind lowers to an
// instruction no recogniser matches, so a kind missing here would be a missed finding and a kind
// added here only costs a comparison.
func cooksToName(node *ast.Node, accepts func(name string) bool) bool {
	found := false
	var visit func(child *ast.Node) bool
	visit = func(child *ast.Node) bool {
		switch child.Kind {
		case ast.KindIdentifier, ast.KindPrivateIdentifier, ast.KindStringLiteral,
			ast.KindNoSubstitutionTemplateLiteral:
			if accepts(child.Text()) {
				found = true
				return true
			}
		}
		return child.ForEachChild(visit)
	}
	visit(node)
	return found
}
