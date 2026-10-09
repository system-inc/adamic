package next

import (
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// Every case below is upstream's, copied byte for byte from oxc's tester block and verified
// against it mechanically rather than by reading. The file name on each case is upstream's too:
// this rule gates on the path, so the path is part of the fixture rather than scenery.
//
// The corpus is nine passing inputs and four failing ones producing five diagnostics, because one
// failing input declares two typo'd exports.
func TestNoTyposReports(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		fileName string
		source   string
		findings int
	}{
		{
			// one case change, variable form
			name:     "one case change, variable form",
			fileName: `pages/test.tsx`,
			source: `
                export default function Page() {
                return <div></div>;
                }
                export const getStaticpaths = async () => {};
                export const getStaticProps = async () => {};
            `,
			findings: 1,
		},
		{
			// two typos in one file, function form, so this reports twice
			name:     "two typos in one file, function form, so this reports twice",
			fileName: `pages/test.tsx`,
			source: `
                export default function Page() {
                    return <div></div>;
                }
                export async function getStaticPathss(){};
                export async function getStaticPropss(){};
            `,
			findings: 2,
		},
		{
			// one substitution, function form
			name:     "one substitution, function form",
			fileName: `pages/test.tsx`,
			source: `
                export default function Page() {
                    return <div></div>;
                }
                export async function getServurSideProps(){};
            `,
			findings: 1,
		},
		{
			// one substitution, variable form
			name:     "one substitution, variable form",
			fileName: `pages/test.tsx`,
			source: `
                export default function Page() {
                    return <div></div>;
                }
                export const getServurSideProps = () => {};
            `,
			findings: 1,
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoTypos, testCase.fileName, testCase.source)
			expected := make([]string, testCase.findings)
			for index := range expected {
				expected[index] = messageNoTypos.Id
			}
			rule_testing.ExpectFindings(t, result, expected...)
		})
	}
}

func TestNoTyposIsSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		fileName string
		source   string
	}{
		{
			// both correct names, variable form
			name:     "both correct names, variable form",
			fileName: `pages/test.tsx`,
			source: `
                export default function Page() {
                return <div></div>;
                }
                export const getStaticPaths = async () => {};
                export const getStaticProps = async () => {};
            `,
		},
		{
			// the correct server side name, variable form
			name:     "the correct server side name, variable form",
			fileName: `pages/test.tsx`,
			source: `
                export default function Page() {
                return <div></div>;
                }
                export const getServerSideProps = async () => {};
            `,
		},
		{
			// both correct names, function form
			name:     "both correct names, function form",
			fileName: `pages/test.tsx`,
			source: `
                export default function Page() {
                return <div></div>;
                }
                export async function getStaticPaths() {};
                export async function getStaticProps() {};`,
		},
		{
			// the correct server side name, function form
			name:     "the correct server side name, function form",
			fileName: `pages/test.tsx`,
			source: `
                export default function Page() {
                return <div></div>;
                }
                export async function getServerSideProps() {};`,
		},
		{
			// three trailing s is distance two, outside the threshold
			name:     "three trailing s is distance two, outside the threshold",
			fileName: `pages/test.tsx`,
			source: `
                export default function Page() {
                return <div></div>;
                }
                export async function getServerSidePropsss() {};
            `,
		},
		{
			// distance three, and it reads like a typo to a human
			name:     "distance three, and it reads like a typo to a human",
			fileName: `pages/test.tsx`,
			source: `
                export default function Page() {
                return <div></div>;
                }
                export async function getstatisPath() {};
            `,
		},
		{
			// a typo, but the file is not under a pages directory
			name:     "a typo, but the file is not under a pages directory",
			fileName: `test.tsx`,
			source: `
                export default function Page() {
                return <div></div>;
                }
                export const getStaticpaths = async () => {};
                export const getStaticProps = async () => {};
            `,
		},
		{
			// a typo, but the file is an api route
			name:     "a typo, but the file is an api route",
			fileName: `pages/api/test.tsx`,
			source:   `export const getStaticpaths = async () => {};`,
		},
		{
			// the windows spelling of the same api route
			name:     "the windows spelling of the same api route",
			fileName: "pages\\api\\test.tsx",
			source:   `export const getStaticpaths = async () => {};`,
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoTypos, testCase.fileName, testCase.source)
			rule_testing.ExpectClean(t, result)
		})
	}
}

// The span, which `ExpectFindings` cannot see. oxc's snapshot underlines the identifier alone, and
// ESLint's original reports the whole export statement, so a port that followed ESLint would pass
// every message-id assertion above while pointing eight columns to the left.
//
// The declaration-list case is here for the same reason at a different scale: reporting the
// statement would underline `export const a = 1, getStaticPropss = 2` entirely, which names the
// clean declarator as part of the problem.
func TestNoTyposPointsAtTheIdentifier(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		source   string
		reported []string
	}{
		{
			name:     "the variable form",
			source:   "export const getStaticpaths = async () => {};",
			reported: []string{"getStaticpaths"},
		},
		{
			name:     "the function form",
			source:   "export async function getServurSideProps(){};",
			reported: []string{"getServurSideProps"},
		},
		{
			name:     "one typo among several declarators",
			source:   "export const a = 1, getStaticPropss = 2, b = 3;",
			reported: []string{"getStaticPropss"},
		},
		{
			name:     "two typos in one file, in source order",
			source:   "export async function getStaticPathss(){};\nexport async function getStaticPropss(){};",
			reported: []string{"getStaticPathss", "getStaticPropss"},
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoTypos, "pages/test.tsx", testCase.source)
			if len(result.Diagnostics) != len(testCase.reported) {
				t.Fatalf("got %d findings, want %d", len(result.Diagnostics), len(testCase.reported))
			}
			for index, diagnostic := range result.Diagnostics {
				got := testCase.source[diagnostic.Range.Pos():diagnostic.Range.End()]
				if got != testCase.reported[index] {
					t.Fatalf("finding %d underlines %q, want %q", index, got, testCase.reported[index])
				}
			}
		})
	}
}

// The message text, asserted exactly rather than by substring. Naming the correction is the rule's
// whole value, so a message that said only "this may be a typo" would satisfy every id assertion
// while withholding the one thing the reader needs. Equality rather than `strings.Contains`, because
// a predicate weaker than the property it guards is not a guard.
func TestNoTyposNamesBothTheTypoAndTheCorrection(t *testing.T) {
	t.Parallel()

	cases := []struct {
		source string
		want   string
	}{
		{
			source: "export const getStaticpaths = async () => {};",
			want:   "`getStaticpaths` may be a typo. Did you mean `getStaticPaths`? ",
		},
		{
			source: "export async function getServurSideProps(){};",
			want:   "`getServurSideProps` may be a typo. Did you mean `getServerSideProps`? ",
		},
		{
			source: "export const getStaticPropss = 1;",
			want:   "`getStaticPropss` may be a typo. Did you mean `getStaticProps`? ",
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.want, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoTypos, "pages/test.tsx", testCase.source)
			if len(result.Diagnostics) != 1 {
				t.Fatalf("got %d findings, want 1", len(result.Diagnostics))
			}
			got := result.Diagnostics[0].Message.Description
			want := testCase.want + messageNoTypos.Description
			if got != want {
				t.Fatalf("message is\n  %q\nwant\n  %q", got, want)
			}
		})
	}
}

// Which declaration shapes carry the name, every one pinned against the release oxlint binary rather
// than inferred from the AST. Upstream matches a named export holding a function declaration or a
// variable declarator whose binding is a plain identifier, and nothing else. None of these is
// implied by the others and a port can plausibly get any of them wrong in either direction.
func TestNoTyposOnlyReadsNamedExportsOfTwoShapes(t *testing.T) {
	t.Parallel()

	reports := []struct {
		name   string
		source string
	}{
		{
			name:   "a named function export",
			source: "export function getStaticPropss() {};",
		},
		{
			name:   "a named variable export",
			source: "export const getStaticPropss = 1;",
		},
		{
			// Per declarator rather than per statement: a typo sharing a statement with clean names
			// still reports. Confirmed on the release binary, which reports at column 21 here.
			name:   "a typo among clean declarators",
			source: "export const a = 1, getStaticPropss = 2;",
		},
		{
			// `let` and `var` are the same statement kind as `const`, so nothing about the rule
			// distinguishes them. Asserted so a declaration-kind guard added later fails loudly.
			name:   "a let export",
			source: "export let getStaticPropss = 1;",
		},
	}
	for _, testCase := range reports {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoTypos, "pages/test.tsx", testCase.source)
			rule_testing.ExpectFindings(t, result, messageNoTypos.Id)
		})
	}

	silent := []struct {
		name   string
		source string
	}{
		{
			// Not exported by name, so Next never calls it by name either. `module.IsExportedByName`
			// requires `export` and refuses `default`, which is exactly this case.
			name:   "a default function export",
			source: "export default function getStaticPropss() {};",
		},
		{
			name:   "a default export of an identifier",
			source: "const getStaticPropss = () => {};\nexport default getStaticPropss;",
		},
		{
			// oxc's match arms cover VariableDeclaration and FunctionDeclaration; ESLint spells the
			// same exclusion as an explicit `case 'ClassDeclaration': break;`.
			name:   "an exported class",
			source: "export class getStaticPropss {};",
		},
		{
			// A re-export is a different node entirely and carries no declaration to walk.
			name:   "a re-export",
			source: "export { getStaticPropss } from './other';",
		},
		{
			// A local export clause, the same shape without the module specifier.
			name:   "an export clause",
			source: "const getStaticPropss = 1;\nexport { getStaticPropss };",
		},
		{
			// A destructuring pattern is not a binding identifier. Upstream matches only
			// BindingIdentifier and ESLint guards `d.id.type !== 'Identifier'`.
			name:   "a destructured export",
			source: "export const { getStaticPropss } = obj;",
		},
		{
			name:   "an array destructured export",
			source: "export const [getStaticPropss] = arr;",
		},
		{
			// Not exported at all.
			name:   "a plain local declaration",
			source: "const getStaticPropss = () => {};\nfunction getStaticPathss() {};",
		},
		{
			// A type alias is not a value Next could call, and it is neither of the two shapes.
			name:   "an exported type alias",
			source: "export type getStaticPropss = number;",
		},
	}
	for _, testCase := range silent {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoTypos, "pages/test.tsx", testCase.source)
			rule_testing.ExpectClean(t, result)
		})
	}
}

// The distance metric, which is the entire rule. Upstream's corpus exercises distances 1, 2 and 3
// but never states the boundary as a boundary, so these pin it directly: every kind of single edit
// reports, and every second edit is silent.
//
// The transposition case is the one that separates Levenshtein from Damerau. Upstream would report
// it under Damerau and does not, so a port reaching for a nicer-looking distance function changes
// the rule here and nowhere the imported corpus can see.
func TestNoTyposAppliesThresholdOne(t *testing.T) {
	t.Parallel()

	reports := []struct {
		name   string
		source string
	}{
		{"one substitution", "export const getStaticPropr = 1;"},
		{"one case change", "export const getstaticProps = 1;"},
		{"one insertion", "export const getStaticPropss = 1;"},
		{"one deletion", "export const getStaticProp = 1;"},
		{"a leading insertion", "export const xgetStaticProps = 1;"},
	}
	for _, testCase := range reports {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoTypos, "pages/test.tsx", testCase.source)
			rule_testing.ExpectFindings(t, result, messageNoTypos.Id)
		})
	}

	silent := []struct {
		name   string
		source string
	}{
		{
			// Distance 2, and the case a Damerau variant reports at 1.
			name:   "an adjacent transposition",
			source: "export const getSatticProps = 1;",
		},
		{"two substitutions", "export const getStaticPropry = 1;"},
		{"two insertions", "export const getStaticPropsss = 1;"},
		{"two deletions", "export const getStaticPro = 1;"},
		{
			// Distance 0. An exact match is correct rather than nearly correct, and `text.BestMatch`
			// declines it. A port that dropped the zero check would flag every correct spelling,
			// which is the loudest possible failure and therefore the one worth a fixture.
			name:   "the exact name",
			source: "export const getStaticProps = 1;",
		},
		{
			// Far outside the threshold, and the shape a real codebase actually contains. Measured
			// on this tree: five exported names begin `getS`, and the nearest is 13 edits away.
			name:   "an unrelated getter",
			source: "export const getServerSideNetworkService = 1;",
		},
		{
			// All lowercase is distance 4 from getStaticProps, not 1. Included because a
			// case-insensitive comparison would report it, and a case-discounting one would too.
			name:   "an all lowercase spelling",
			source: "export const getstaticprops = 1;",
		},
	}
	for _, testCase := range silent {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoTypos, "pages/test.tsx", testCase.source)
			rule_testing.ExpectClean(t, result)
		})
	}
}

// The path gate, whose four disagreements with the ESLint original were pinned against the release
// oxlint binary. Upstream's own corpus covers only three of these shapes, and the two that separate
// oxc from ESLint appear in neither corpus.
func TestNoTyposGatesOnThePagesDirectory(t *testing.T) {
	t.Parallel()

	const typo = "export const getStaticpaths = async () => {};"

	reports := []struct {
		name     string
		fileName string
	}{
		{"a page", "pages/index.tsx"},
		{"a page under a source directory", "src/pages/index.tsx"},
		{
			// Only the segment directly under `pages` exempts, so a nested api directory runs.
			// Deliberate upstream: `pages/api` is Next's API routes and a deeper one is not.
			name:     "a nested api directory",
			fileName: "pages/blog/api/x.tsx",
		},
		{
			// A segment test rather than a prefix test. ESLint exempts this, because it asks
			// whether the remaining directory startsWith `/api`. oxc reports it and so do we.
			name:     "a directory whose name begins with api",
			fileName: "pages/apiary/x.tsx",
		},
	}
	for _, testCase := range reports {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoTypos, testCase.fileName, typo)
			rule_testing.ExpectFindings(t, result, messageNoTypos.Id)
		})
	}

	silent := []struct {
		name     string
		fileName string
	}{
		{"an api route", "pages/api/user.ts"},
		{"a nested api route", "pages/api/v1/user.ts"},
		{"outside any pages directory", "components/Thing.tsx"},
		{
			// A segment test rather than a substring test. ESLint runs on this, because it splits
			// the filename on the literal text `pages` and finds it inside the longer word.
			name:     "a directory whose name ends in pages",
			fileName: "mypages/index.tsx",
		},
		{
			// The first `pages` segment decides and the second is never consulted, matching oxc's
			// walk, which returns on the component immediately after the first match.
			name:     "a pages directory nested inside an api route",
			fileName: "pages/api/pages/x.tsx",
		},
	}
	for _, testCase := range silent {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoTypos, testCase.fileName, typo)
			rule_testing.ExpectClean(t, result)
		})
	}
}
