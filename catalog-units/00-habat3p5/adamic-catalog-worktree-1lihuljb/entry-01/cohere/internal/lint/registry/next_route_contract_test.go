package registry

import (
	"fmt"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/system-inc/cohere/internal/lint/ecmascript/nextjs"
	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/internal/lint/testing"
)

// nextRouteContractFixtures are App Router files exporting every name Next.js reads from them, in
// the shapes route files are actually written in: function declarations, async functions, arrow
// constants, and plain data constants.
//
// The paths matter as much as the text. `nextjs.RouteContractExports` decides from the path, so each
// file sits under an `app` segment with the stem Next looks for.
var nextRouteContractFixtures = map[string]string{
	"/repository/app/blog/[slug]/page.tsx": `
export const config = {};
export async function generateStaticParams() {
    return [];
}
export const unstable_instant = false;
export const unstable_dynamicStaleTime = 30;
export const revalidate = 60;
export const dynamic = 'force-static';
export const dynamicParams = false;
export const fetchCache = 'default-cache';
export const preferredRegion = 'auto';
export const runtime = 'nodejs';
export const maxDuration = 30;
export const metadata = { title: 'Title' };
export async function generateMetadata() {
    return { title: 'Title' };
}
export const viewport = { themeColor: 'black' };
export function generateViewport() {
    return {};
}
export default function Page() {
    return null;
}
`,
	"/repository/app/api/report/route.ts": `
export const revalidate = false;
export const dynamic = 'force-dynamic';
export const maxDuration = 60;
export const runtime = 'nodejs';
export async function GET() {
    return new Response('');
}
export async function HEAD() {
    return new Response('');
}
export const OPTIONS = async () => new Response('');
export const POST = async () => new Response('');
export async function PUT() {
    return new Response('');
}
export async function DELETE() {
    return new Response('');
}
export async function PATCH() {
    return new Response('');
}
`,
	"/repository/app/opengraph-image.tsx": `
export const alt = 'About';
export const size = { width: 1200, height: 630 };
export const contentType = 'image/png';
export const runtime = 'nodejs';
export function generateImageMetadata() {
    return [];
}
export default function Image() {
    return null;
}
`,
	"/repository/app/sitemap.ts": `
export const revalidate = 3600;
export function generateSitemaps() {
    return [];
}
export default function sitemap() {
    return [];
}
`,
}

// TestNoRuleRewritesANextRouteContractExport runs every registered rule over route files that export
// every contract name, and fails on any fix that touches one.
//
// Next.js reads these names off the module with no import anywhere, so a rename, a recasing or a
// dropped `export` compiles, renders, and silently stops working: five phi web routes exported
// `generateStaticParameters` for fifteen months and never prerendered. A per-rule fixture can only
// protect the rules somebody remembered to write one for. This walks the registry instead, so a rule
// added next year with a rename fixer is held to the same line on its first run.
//
// Two levels of evidence, because a fix is the dangerous half and a finding is how the bad rename
// got made by hand: any FIX whose range touches a contract name fails, and so does any finding a
// naming rule anchors on one, since its message is advice to rename a name that must not change.
//
// Every rule runs with nil options, which is the strict case: a consumer that passes no framework
// list still gets the floor `nextjs.IsRouteContractExport` provides.
func TestNoRuleRewritesANextRouteContractExport(t *testing.T) {
	t.Parallel()

	rules := All()
	if len(rules) < 100 {
		// A registry that loaded nothing would pass this test for the wrong reason.
		t.Fatalf("the registry holds %d rules, so this test could not see the catalog", len(rules))
	}

	for filePath, sourceText := range nextRouteContractFixtures {
		contract := nextjs.RouteContractExports(filePath)
		if len(contract) == 0 {
			t.Fatalf("%s is not a route file by nextjs.RouteContractExports, so its fixture proves nothing", filePath)
		}

		t.Run(filePath, func(t *testing.T) {
			t.Parallel()

			sawContractNames := 0
			for _, subject := range rules {
				result := rule_testing.RunTyped(t, subject, filePath, sourceText)
				contractRanges := contractExportNameRanges(result.SourceFile, contract)
				if sawContractNames == 0 {
					sawContractNames = len(contractRanges)
				}

				for _, diagnostic := range result.Diagnostics {
					for _, fix := range diagnostic.Fixes {
						if name, touches := touchedContractName(fix.Range, contractRanges); touches {
							t.Errorf("%s: a fix rewrites the Next contract export %q (%q over %s): %s",
								subject.Name, name, fix.Text, describeRange(result.SourceFile, fix.Range), diagnostic.Message.Description)
						}
					}
					if !isNamingRule(subject.Name) {
						continue
					}
					if name, touches := touchedContractName(diagnostic.Range, contractRanges); touches {
						t.Errorf("%s: reports on the Next contract export %q, advising a change Next would silently ignore: %s",
							subject.Name, name, diagnostic.Message.Description)
					}
				}
			}

			// Every contract name in the fixture must have been found, or the ranges the assertions
			// compare against were empty and every rule passed by seeing nothing.
			if want := countFixtureContractNames(sourceText, contract); sawContractNames != want {
				t.Fatalf("found %d contract export names in the parsed file, the fixture writes %d", sawContractNames, want)
			}
		})
	}
}

// isNamingRule names the rules whose findings are advice to rename, recase or un-export a
// declaration. Their FINDINGS on a contract name are defects, not just their fixes.
func isNamingRule(name string) bool {
	for _, prefix := range []string{
		"nexus/consistency-no-abbreviated-identifier",
		"nexus/consistency-no-ambiguous-identifier",
		"nexus/consistency-require-constant-casing",
		"nexus/consistency-no-screaming-snake-case",
		"nexus/consistency-no-shouting",
		"nexus/consistency-require-type-suffix",
		"camelcase",
		"id-length",
		"id-match",
		"@typescript-eslint/naming-convention",
	} {
		if name == prefix {
			return true
		}
	}
	return false
}

// contractExportNameRanges collects the token range of every top-level exported declaration name
// that Next reads from this file.
func contractExportNameRanges(sourceFile *ast.SourceFile, contract map[string]bool) map[string]core.TextRange {
	ranges := map[string]core.TextRange{}
	if sourceFile == nil {
		return ranges
	}
	record := func(name *ast.Node) {
		if name != nil && name.Kind == ast.KindIdentifier && contract[name.Text()] {
			ranges[name.Text()] = rule.TokenRange(sourceFile, name)
		}
	}
	for _, statement := range sourceFile.Statements.Nodes {
		if !hasExportModifier(statement) {
			continue
		}
		switch statement.Kind {
		case ast.KindFunctionDeclaration:
			record(statement.Name())
		case ast.KindVariableStatement:
			declarationList := statement.AsVariableStatement().DeclarationList.AsVariableDeclarationList()
			for _, declaration := range declarationList.Declarations.Nodes {
				record(declaration.Name())
			}
		}
	}
	return ranges
}

func hasExportModifier(statement *ast.Node) bool {
	modifiers := statement.Modifiers()
	if modifiers == nil {
		return false
	}
	for _, modifier := range modifiers.Nodes {
		if modifier.Kind == ast.KindExportKeyword {
			return true
		}
	}
	return false
}

// touchedContractName reports the contract name a range overlaps, if any. Touching counts, not just
// containing: a fix spanning the whole declaration rewrites the name inside it as surely as one
// spanning the name alone.
func touchedContractName(subject core.TextRange, contractRanges map[string]core.TextRange) (string, bool) {
	for name, nameRange := range contractRanges {
		if subject.Pos() < nameRange.End() && nameRange.Pos() < subject.End() {
			return name, true
		}
	}
	return "", false
}

func describeRange(sourceFile *ast.SourceFile, textRange core.TextRange) string {
	text := sourceFile.Text()
	if textRange.Pos() < 0 || textRange.End() > len(text) || textRange.Pos() > textRange.End() {
		return fmt.Sprintf("[%d,%d)", textRange.Pos(), textRange.End())
	}
	return fmt.Sprintf("%q", text[textRange.Pos():textRange.End()])
}

// countFixtureContractNames counts the contract names the fixture text declares, independently of
// the parse, so the parsed count has something to be checked against.
func countFixtureContractNames(sourceText string, contract map[string]bool) int {
	count := 0
	for _, line := range strings.Split(sourceText, "\n") {
		for _, prefix := range []string{"export const ", "export async function ", "export function "} {
			if !strings.HasPrefix(line, prefix) {
				continue
			}
			rest := strings.TrimPrefix(line, prefix)
			end := strings.IndexAny(rest, " (=")
			if end > 0 && contract[rest[:end]] {
				count++
			}
		}
	}
	return count
}
