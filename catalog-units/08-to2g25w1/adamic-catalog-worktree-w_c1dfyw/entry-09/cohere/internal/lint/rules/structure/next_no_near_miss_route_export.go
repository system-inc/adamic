package structure

import (
	"sort"
	"strings"
	"unicode"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/ecmascript/imports"
	"github.com/system-inc/cohere/internal/lint/ecmascript/module"
	"github.com/system-inc/cohere/internal/lint/ecmascript/nextjs"
	"github.com/system-inc/cohere/internal/lint/ecmascript/text"
	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/policy"
)

// nextNoNearMissRouteExportNearMissText and nextNoNearMissRouteExportPagesRouterText are the rule's
// messages, whose wording lives in `policy/messages/next-no-near-miss-route-export.json`.
var (
	nextNoNearMissRouteExportNearMissText    = policy.MessageOf("structure/next-no-near-miss-route-export", "nearMissRouteExport")
	nextNoNearMissRouteExportPagesRouterText = policy.MessageOf("structure/next-no-near-miss-route-export", "pagesRouterExport")
)

func messageNearMissRouteExport(name string, contractName string) rule.Message {
	return rule.Message{
		Id:          nextNoNearMissRouteExportNearMissText.Id,
		Description: nextNoNearMissRouteExportNearMissText.Render(map[string]string{"name": name, "contractName": contractName}),
	}
}

func messagePagesRouterExport(name string) rule.Message {
	return rule.Message{
		Id:          nextNoNearMissRouteExportPagesRouterText.Id,
		Description: nextNoNearMissRouteExportPagesRouterText.Render(map[string]string{"name": name}),
	}
}

// pagesRouterDataFunctions are the Pages Router exports an App Router file can still be handed by
// habit. The App Router reads none of them.
var pagesRouterDataFunctions = map[string]bool{
	"getStaticProps":     true,
	"getStaticPaths":     true,
	"getServerSideProps": true,
}

// nearMissWordExpansions maps a spelled-out word to the abbreviation Next's own names use, so that
// two names differing only in how a word is spelled compare equal. Every pair is one our naming
// rules have pushed toward the long form: `generateStaticParams` became `generateStaticParameters`
// in phi web, and `maxDuration` was one suggestion away from `maximumDuration`.
var nearMissWordExpansions = map[string]string{
	"parameters":    "params",
	"parameter":     "params",
	"param":         "params",
	"properties":    "props",
	"property":      "props",
	"prop":          "props",
	"maximum":       "max",
	"configuration": "config",
}

// nearMissConfusableWords are the static-generation nouns the two routers spell differently, so one
// standing in for another is a mix-up rather than a new name: `generateStaticProps` is a Pages Router
// habit inside an App Router name.
var nearMissConfusableWords = map[string]bool{"params": true, "props": true, "paths": true}

// NextNoNearMissRouteExport flags an App Router export that misses a name Next.js reads by one step.
//
//	valid:   export async function generateStaticParams() {}
//	valid:   export const revalidate = 60
//	valid:   export function formatSlug() {}              no contract name is near it
//	invalid: export async function generateStaticParameters() {}
//	invalid: export async function generateMetaData() {}
//	invalid: export const maximumDuration = 60
//	invalid: export async function getStaticProps() {}    in an app directory
//
// The failure is silent in every direction. The file compiles, the page renders, the types pass,
// and Next simply never calls the function: no static params means no prerendering, no metadata
// means the default title. Five phi web routes exported `generateStaticParameters` from 2025-06-20,
// the commit that migrated them from the old repository, and none of them prerendered.
//
// Which names count comes from `nextjs.RouteContractExports`, read from the installed Next build
// rather than its documentation, and differs by file: a page has `generateMetadata` and no `POST`, a
// route the reverse.
//
// # What "one step" means
//
// Four tests, each the shape of a real miss rather than a general similarity:
//
//	same words, spelled out       generateStaticParameters, generateMetaData, maximumDuration
//	one router noun for another   generateStaticProps, generateStaticPaths
//	a typo                        generateStaticParam, revalidat
//	a Pages Router function       getStaticProps, getStaticPaths, getServerSideProps
//
// The first compares lowercased camelCase words after mapping spelled-out words to Next's
// abbreviations, so casing and compound splits (`MetaData`, `ViewPort`) fall out for free.
//
// The typo test is an edit distance that scales with the name, because a fixed one is wrong at both
// ends. Distance 1 from `alt` is `all`, `act` and `art`, all plausible exports of an image file, and
// distance 1 from `GET` is `SET` and `LET`. So a contract name shorter than seven characters gets no
// typo test at all, one of seven or more allows a single edit, and one of fourteen or more allows two.
//
// Only value exports count. A type export cannot be a route contract, and `export type { Metadata }`
// is the ordinary way a layout re-exports Next's own type, which a case-insensitive test would flag.
var NextNoNearMissRouteExport = rule.Rule{
	Name: "structure/next-no-near-miss-route-export",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		contract := nextjs.RouteContractExports(imports.NormalizedFileName(ctx.SourceFile))
		if len(contract) == 0 {
			return nil
		}
		contractNames := make([]string, 0, len(contract))
		for name := range contract {
			contractNames = append(contractNames, name)
		}
		// The map's order is random, and the message names one contract name, so the search runs
		// in a fixed order and a tie resolves the same way on every run.
		sort.Strings(contractNames)

		check := func(name *ast.Node, spelling string) {
			if contract[spelling] {
				return
			}
			if pagesRouterDataFunctions[spelling] {
				ctx.ReportNode(name, messagePagesRouterExport(spelling))
				return
			}
			if contractName, found := nearMissContractName(spelling, contractNames); found {
				ctx.ReportNode(name, messageNearMissRouteExport(spelling, contractName))
			}
		}

		return rule.Listeners{
			ast.KindFunctionDeclaration: func(node *ast.Node) {
				// `export default function generateStaticParameters` is the default export, which
				// Next reads by position rather than by name.
				if !module.IsExportedByName(node.Modifiers()) || node.Parent.Kind != ast.KindSourceFile {
					return
				}
				if name := node.Name(); name != nil && name.Kind == ast.KindIdentifier {
					check(name, name.Text())
				}
			},
			ast.KindVariableStatement: func(node *ast.Node) {
				if !module.IsExportedByName(node.Modifiers()) || node.Parent.Kind != ast.KindSourceFile {
					return
				}
				declarationList := node.AsVariableStatement().DeclarationList.AsVariableDeclarationList()
				for _, declaration := range declarationList.Declarations.Nodes {
					if name := declaration.Name(); name != nil && name.Kind == ast.KindIdentifier {
						check(name, name.Text())
					}
				}
			},
			// An export clause names what leaves the module, so `export { load as generateMetaData }`
			// is judged on the name after `as`, which is the one Next would look for.
			ast.KindExportDeclaration: func(node *ast.Node) {
				declaration := node.AsExportDeclaration()
				if declaration.IsTypeOnly || declaration.ExportClause == nil ||
					declaration.ExportClause.Kind != ast.KindNamedExports {
					return
				}
				for _, specifier := range declaration.ExportClause.AsNamedExports().Elements.Nodes {
					if specifier.AsExportSpecifier().IsTypeOnly {
						continue
					}
					if name := specifier.Name(); name != nil && name.Kind == ast.KindIdentifier {
						check(name, name.Text())
					}
				}
			},
		}
	},
}

// nearMissContractName returns the contract name an export misses by one step, if any. The
// candidates must be sorted; the first that matches wins.
func nearMissContractName(name string, contractNames []string) (string, bool) {
	nameWords := nearMissWords(name)
	for _, contractName := range contractNames {
		contractWords := nearMissWords(contractName)
		if strings.Join(nameWords, "") == strings.Join(contractWords, "") {
			return contractName, true
		}
		if differsByOneConfusableWord(nameWords, contractWords) {
			return contractName, true
		}
	}

	// The typo test runs only after the word tests have found nothing, so a name the words already
	// explain is never reported as the nearest typo of something else.
	for _, threshold := range []int{1, 2} {
		eligible := make([]string, 0, len(contractNames))
		for _, contractName := range contractNames {
			if typoThreshold(contractName) >= threshold {
				eligible = append(eligible, contractName)
			}
		}
		if contractName, found := text.BestMatch(name, eligible, threshold); found {
			return contractName, true
		}
	}
	return "", false
}

// typoThreshold is the greatest edit distance that still reads as a typo of this contract name.
func typoThreshold(contractName string) int {
	switch length := len(contractName); {
	case length >= 14:
		return 2
	case length >= 7:
		return 1
	}
	return 0
}

// differsByOneConfusableWord reports whether two word lists match everywhere except one position,
// where both words are static-generation nouns one router spells and the other does not.
func differsByOneConfusableWord(first []string, second []string) bool {
	if len(first) != len(second) {
		return false
	}
	differences := 0
	for index := range first {
		if first[index] == second[index] {
			continue
		}
		if !nearMissConfusableWords[first[index]] || !nearMissConfusableWords[second[index]] {
			return false
		}
		differences++
	}
	// Exactly one, though no input can tell one from more: identical lists were already caught by
	// the word test, and no contract name holds two of these nouns. A mutation sweep reports the
	// difference as surviving for that reason, and the test says what is meant.
	return differences == 1
}

// nearMissWords splits a name into lowercase words at camelCase humps, with each spelled-out word
// mapped to the abbreviation Next's names use.
//
// Humps only. Splitting on underscores and digits was written and then removed, because a mutation
// sweep showed neither could change an answer: every underscored contract name's near misses are
// already within its typo distance, and so is a trailing digit. Logic no input reaches is drift.
func nearMissWords(name string) []string {
	var words []string
	var current strings.Builder
	flush := func() {
		if current.Len() == 0 {
			return
		}
		word := strings.ToLower(current.String())
		if abbreviation, found := nearMissWordExpansions[word]; found {
			word = abbreviation
		}
		words = append(words, word)
		current.Reset()
	}
	runes := []rune(name)
	for index, character := range runes {
		if unicode.IsUpper(character) {
			// A hump starts a word, unless it continues a run of capitals (`GET`, `HTTPServer`),
			// where only the last capital before a lowercase letter starts the next word.
			previousIsUpper := index > 0 && unicode.IsUpper(runes[index-1])
			nextIsLower := index+1 < len(runes) && unicode.IsLower(runes[index+1])
			if !previousIsUpper || nextIsLower {
				flush()
			}
		}
		current.WriteRune(character)
	}
	flush()
	return words
}
