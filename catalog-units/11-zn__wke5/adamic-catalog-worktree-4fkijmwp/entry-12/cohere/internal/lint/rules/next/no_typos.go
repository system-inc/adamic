package next

import (
	"fmt"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/ecmascript/module"
	"github.com/system-inc/cohere/internal/lint/ecmascript/nextjs"
	"github.com/system-inc/cohere/internal/lint/ecmascript/text"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// dataFetchingFunctions are the three names Next.js calls by name in the Pages Router.
//
// Order is load-bearing in principle rather than in practice: `text.BestMatch` keeps the first
// candidate on a tie, so this is oxc's array order. No name sits within distance 1 of two of these,
// so no tie is reachable at the threshold below, but reordering would still be a silent change to a
// documented behaviour.
var dataFetchingFunctions = []string{"getStaticProps", "getStaticPaths", "getServerSideProps"}

// typoThreshold is the greatest edit distance that still reads as a typo rather than a different name.
//
// One, and the value is the entire rule. Zero would mean an exact match and upstream treats that as
// correct rather than as a typo, which is why `text.BestMatch` declines it. Two would report
// `getServerSidePropsss` and `getstatisPath`, both of which upstream ships as passing cases, so a
// port that raised this "to be more helpful" would flag clean code. Upstream's own comment above the
// constant is `// 0 is the exact match`.
const typoThreshold = 1

var messageNoTypos = rule.Message{
	Id: "noTypos",
	Description: "Next.js calls its data fetching functions by exact name, so a misspelled export " +
		"is never called and the page renders with no data at all. There is no error: the function " +
		"is simply dead code the framework never looks for.",
}

// NoTypos flags an exported name one edit away from a Next.js data fetching function.
//
//	valid:   export const getStaticProps = async () => {}
//	valid:   export async function getServerSidePropsss() {}   two edits away, not a typo
//	valid:   export default function getStaticPropss() {}      not exported by name
//	invalid: export const getStaticpaths = async () => {}
//	invalid: export async function getServurSideProps() {}
//
// Ported from `@next/next/no-typos`, read against oxc's `no_typos.rs`.
//
// # The distance metric is the rule
//
// Everything else here is plumbing. The decision is `text.BestMatch(name, names, 1)`: plain
// Levenshtein, unit cost for an insertion, a deletion, or a substitution, and **no transposition**.
// Three consequences a reader should be able to see without deriving them:
//
//	getStaticpaths         distance 1 from getStaticPaths        reports    one case change
//	getStaticPropss        distance 1 from getStaticProps        reports    one insertion
//	getServerSidePropsss   distance 2 from getServerSideProps    silent     two insertions
//	getstatisPath          distance 3 from getStaticPaths        silent     reads like a typo, is not
//	getSatticProps         distance 2 from getStaticProps        silent     a transposition costs two
//
// The last two are the corpus's real content. `getstatisPath` looks like an obvious typo to a human
// and the rule does not care, and `getSatticProps` is the case a Damerau variant would report. Both
// are upstream passing cases and both are asserted in the fixtures, because they are what a port
// silently changes when it reaches for a nicer-looking distance function.
//
// Case is a full edit rather than a discount. That matters because it is the one place a plausible
// alternative disagrees: `core.GetSpellingSuggestionForStrings`, reachable through the shim and the
// closest name-match on any shelf here, charges 0.1 for a case-only substitution and 2 for any other,
// then accepts anything within 0.4 of the name's length. Measured against these three candidates it
// answers a suggestion for `getServerSidePropsss`, for `getstatisPath`, and for `getSatticProps`,
// which is three upstream passing cases reported. The reasoning is recorded on `text.BestMatch`.
//
// # Where the name has to appear
//
// A named export, and only through two shapes: a function declaration, or a variable declarator
// whose binding is a plain identifier. Everything else is silent, and each of these was pinned
// against the release binary rather than inferred, because the AST offers no reason to expect them:
//
//	export default function getStaticPropss() {}   silent   not a named export
//	export class getStaticPropss {}                silent   upstream skips ClassDeclaration outright
//	export { getStaticPropss } from './other'      silent   a re-export is a different node
//	export const { getStaticPropss } = obj         silent   a destructuring pattern, not an identifier
//	const getStaticPropss = () => {}               silent   not exported at all
//	export const a = 1, getStaticPropss = 2        REPORTS  per declarator, not per statement
//
// The last one is the one worth checking against your instinct: the rule walks every declarator in
// the list, so a typo sharing a statement with a clean name still reports, and it reports on the
// typo's own identifier rather than on the statement.
//
// TypeScript's AST has no `ExportNamedDeclaration` node, so oxc's single `ExportDeclaration` listener
// does not transfer. `export` is a modifier on the declaration itself, which is why this listens on
// the declarations and asks `module.IsExportedByName` rather than listening for an export.
//
// # Where it runs
//
// Only under a Pages Router directory, and never under its top-level `api` directory: an API route
// returns data rather than rendering a page, so it has no data fetching functions to misspell. The
// gate is answered once per file in `Run` rather than per node, which is also how upstream spells it.
//
// The gate is `nextjs.IsInPagesDirectory`, whose comment carries the two places the ESLint original
// gets this wrong and oxc fixes it.
//
// One upstream fixture deserves a note because copying its verdict without its reasoning would be
// copying an accident. oxc ships `pages\api\test.tsx` as a passing case, and on a non-Windows host it
// passes for the wrong reason: `PathBuf::components()` does not split backslashes there, so the path
// is a single component, `pages` never matches, and the file is skipped as though it were outside a
// pages directory entirely. Our harness normalizes separators before a rule sees the path, so the
// same input arrives as `pages/api/test.tsx` and is exempted as a genuine API route. Same verdict,
// and ours is the one that would still hold if the reason changed.
var NoTypos = rule.Rule{
	// No family prefix. The config writes `nextjs/no-typos` and matching strips the namespace on a
	// `/` boundary, so a prefixed name matches nothing and runs on no files while every test passes.
	Name: "@next/next/no-typos",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		if !nextjs.IsInPagesDirectory(ctx.SourceFile.FileName().AsString()) {
			return nil
		}

		checkName := func(name *ast.Node) {
			if name == nil || name.Kind != ast.KindIdentifier {
				return
			}
			suggestion, ok := text.BestMatch(name.Text(), dataFetchingFunctions, typoThreshold)
			if !ok {
				return
			}
			// The message names both strings because naming the correction is the rule's whole
			// value: "this is a typo" without "of what" leaves the reader to guess which of three
			// similar names was meant.
			message := messageNoTypos
			message.Description = fmt.Sprintf(
				"`%s` may be a typo. Did you mean `%s`? %s",
				name.Text(), suggestion, messageNoTypos.Description,
			)
			// On the identifier rather than on the statement. oxc's snapshot underlines the name
			// alone, and ESLint's original reports the whole export, so this follows oxc. The
			// narrower span is also the one that reads correctly when a statement declares several
			// names and only one is wrong.
			ctx.ReportNode(name, message)
		}

		return rule.Listeners{
			ast.KindFunctionDeclaration: func(node *ast.Node) {
				declaration := node.AsFunctionDeclaration()
				if declaration == nil || !module.IsExportedByName(node.Modifiers()) {
					return
				}
				checkName(declaration.Name())
			},
			ast.KindVariableStatement: func(node *ast.Node) {
				statement := node.AsVariableStatement()
				if statement == nil || statement.DeclarationList == nil {
					return
				}
				if !module.IsExportedByName(node.Modifiers()) {
					return
				}

				declarationList := statement.DeclarationList.AsVariableDeclarationList()
				if declarationList == nil || declarationList.Declarations == nil {
					return
				}
				for _, declaration := range declarationList.Declarations.Nodes {
					// A destructuring pattern binds names it never spells out as one identifier, and
					// upstream matches only a binding identifier. `checkName` declines any other
					// kind, so the shape check lives there rather than being repeated here.
					checkName(declaration.AsVariableDeclaration().Name())
				}
			},
		}
	},
}
