package structure

import (
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

const (
	nearMissPageFile    = "/repository/app/blog/[slug]/page.tsx"
	nearMissRouteFile   = "/repository/app/api/report/route.ts"
	nearMissImageFile   = "/repository/app/opengraph-image.tsx"
	nearMissPlainFile   = "/repository/source/Thing.ts"
	nearMissSitemapFile = "/repository/app/sitemap.ts"
)

// TestNextNoNearMissRouteExportFiresOnThePhiWebSites runs the five real sites by their real paths.
//
// These are the files that lost prerendering, each line as the tree has it. Three are async and two
// are not, and both shapes are here because the listener sees a function declaration either way.
func TestNextNoNearMissRouteExportFiresOnThePhiWebSites(t *testing.T) {
	t.Parallel()

	sites := []struct {
		filePath string
		line     string
	}{
		{"/www-phi-health/app/(main-layout)/blog/[slugWithId]/page.tsx", "export async function generateStaticParameters() {"},
		{"/www-phi-health/app/(main-layout)/podcasts/frequency/[slugWithId]/page.tsx", "export async function generateStaticParameters() {"},
		{"/www-phi-health/app/(main-layout)/(shop)/shop/(products)/[productIdentifier]/page.tsx", "export async function generateStaticParameters() {"},
		{"/www-phi-health/app/(main-layout)/(shop)/shop/(products)/axis/[[...variant]]/page.tsx", "export function generateStaticParameters() {"},
		{"/www-phi-health/app/(main-layout)/stack/[[...variant]]/page.tsx", "export function generateStaticParameters() {"},
	}
	for _, site := range sites {
		t.Run(site.filePath, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NextNoNearMissRouteExport, site.filePath, site.line+"\n    return [];\n}")
			rule_testing.ExpectFindings(t, result, "nearMissRouteExport")
			if !strings.Contains(result.Diagnostics[0].Message.Description, "Rename it to `generateStaticParams`.") {
				t.Errorf("the message does not name the contract export: %q", result.Diagnostics[0].Message.Description)
			}
		})
	}
}

// TestNextNoNearMissRouteExportFires covers each of the four ways an export misses by one step.
func TestNextNoNearMissRouteExportFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		filePath   string
		sourceText string
		wantId     string
		wantNamed  string
	}{
		// Same words once spelled-out words map back and casing is ignored.
		{"metadata written as two words", nearMissPageFile, "export async function generateMetaData() { return {}; }", "nearMissRouteExport", "generateMetadata"},
		{"the abbreviation rule's old advice", nearMissRouteFile, "export const maximumDuration = 60;", "nearMissRouteExport", "maxDuration"},
		{"viewport written as two words", nearMissPageFile, "export function generateViewPort() { return {}; }", "nearMissRouteExport", "generateViewport"},
		{"a capitalised data export", nearMissPageFile, "export const Revalidate = 60;", "nearMissRouteExport", "revalidate"},
		{"dynamic params spelled out", nearMissPageFile, "export const dynamicParameters = false;", "nearMissRouteExport", "dynamicParams"},
		{"config spelled out", nearMissRouteFile, "export const configuration = {};", "nearMissRouteExport", "config"},
		{"a lowercase http method", nearMissRouteFile, "export async function post() { return new Response(''); }", "nearMissRouteExport", "POST"},
		// One router's static-generation noun inside the other router's name.
		{"props for params", nearMissPageFile, "export async function generateStaticProps() { return []; }", "nearMissRouteExport", "generateStaticParams"},
		{"paths for params", nearMissPageFile, "export async function generateStaticPaths() { return []; }", "nearMissRouteExport", "generateStaticParams"},
		// A run of capitals is one word, so the noun swap still lines up word for word.
		{"a shouted router noun", nearMissPageFile, "export async function generateStaticPROPS() { return []; }", "nearMissRouteExport", "generateStaticParams"},
		// Typos, scaled to the name's length.
		{"one edit on a long name", nearMissPageFile, "export async function generateStaticParam() { return []; }", "nearMissRouteExport", "generateStaticParams"},
		{"two edits on a long name", nearMissImageFile, "export function generateImageMetadaat() { return []; }", "nearMissRouteExport", "generateImageMetadata"},
		{"one edit on a mid-length name", nearMissPageFile, "export const revalidat = 60;", "nearMissRouteExport", "revalidate"},
		// No underscore split: the typo distance reaches it, which is why the split was removed.
		{"an underscored name in camelCase", nearMissPageFile, "export const unstableInstant = false;", "nearMissRouteExport", "unstable_instant"},
		{"a sitemap typo", nearMissSitemapFile, "export function generateSitemap() { return []; }", "nearMissRouteExport", "generateSitemaps"},
		// Pages Router data functions, which the App Router never calls.
		{"getStaticProps in an app page", nearMissPageFile, "export async function getStaticProps() { return { props: {} }; }", "pagesRouterExport", ""},
		{"getServerSideProps in an app page", nearMissPageFile, "export const getServerSideProps = async () => ({ props: {} });", "pagesRouterExport", ""},
		// An export clause is judged on the name it exports, not the local one.
		{"an export clause renaming to the near miss", nearMissPageFile, "async function load() { return {}; }\nexport { load as generateMetaData };", "nearMissRouteExport", "generateMetadata"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NextNoNearMissRouteExport, testCase.filePath, testCase.sourceText)
			rule_testing.ExpectFindings(t, result, testCase.wantId)
			if testCase.wantNamed != "" && !strings.Contains(result.Diagnostics[0].Message.Description, "Rename it to `"+testCase.wantNamed+"`.") {
				t.Errorf("want the message to name %q: %q", testCase.wantNamed, result.Diagnostics[0].Message.Description)
			}
		})
	}
}

// TestNextNoNearMissRouteExportStaysSilent covers every contract name and the neighbours that must
// not trip it.
//
// The short-name rows are why the typo distance scales: `all` is one edit from `alt` and `SET` one
// from `GET`, and both are plausible exports Next simply does not read.
func TestNextNoNearMissRouteExportStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		filePath   string
		sourceText string
	}{
		{"every page contract name", nearMissPageFile, `
export const config = {};
export async function generateStaticParams() { return []; }
export const unstable_instant = false;
export const unstable_dynamicStaleTime = 30;
export const revalidate = 60;
export const dynamic = 'force-static';
export const dynamicParams = false;
export const fetchCache = 'default-cache';
export const preferredRegion = 'auto';
export const runtime = 'nodejs';
export const maxDuration = 30;
export const metadata = {};
export async function generateMetadata() { return {}; }
export const viewport = {};
export function generateViewport() { return {}; }
export default function Page() { return null; }`},
		{"every route handler", nearMissRouteFile, `
export async function GET() { return new Response(''); }
export async function HEAD() { return new Response(''); }
export const OPTIONS = async () => new Response('');
export const POST = async () => new Response('');
export async function PUT() { return new Response(''); }
export async function DELETE() { return new Response(''); }
export async function PATCH() { return new Response(''); }`},
		{"every image metadata name", nearMissImageFile, `
export const alt = 'About';
export const size = { width: 1200, height: 630 };
export const contentType = 'image/png';
export function generateImageMetadata() { return []; }`},
		{"a short name one edit from alt", nearMissImageFile, "export const all = 1;"},
		{"a short name one edit from GET", nearMissRouteFile, "export const SET = 1;"},
		{"an ordinary helper", nearMissPageFile, "export function formatSlug(slug: string) { return slug; }"},
		{"a type re-export of Next's own Metadata", nearMissPageFile, "export type { Metadata } from 'next';"},
		{"a type-only specifier", nearMissPageFile, "import type { Viewport } from 'next';\nexport { type Viewport };"},
		{"a default export is read by position", nearMissPageFile, "export default function generateStaticParameters() { return []; }"},
		{"an unexported function", nearMissPageFile, "function generateStaticParameters() { return []; }\ngenerateStaticParameters();"},
		// `export` inside a namespace exports from the namespace, not the module Next reads.
		{"a function exported from a namespace", nearMissPageFile, "export namespace Helpers {\n    export function generateStaticParameters() { return []; }\n}"},
		{"a constant exported from a namespace", nearMissPageFile, "export namespace Helpers {\n    export const maximumDuration = 60;\n}"},
		{"a nested declaration is no export", nearMissPageFile, "export function load() {\n    const generateMetaData = 1;\n    return generateMetaData;\n}"},
		// Contracts differ by file, and so do near misses: a route has no metadata to miss.
		{"metadata in a route is not near anything", nearMissRouteFile, "export async function generateMetaData() { return {}; }"},
		{"the near miss outside a route file", nearMissPlainFile, "export async function generateStaticParameters() { return []; }"},
		{"a pages router function outside the app directory", "/repository/pages/index.tsx", "export async function getStaticProps() { return { props: {} }; }"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NextNoNearMissRouteExport, testCase.filePath, testCase.sourceText)
			rule_testing.ExpectClean(t, result)
		})
	}
}

// TestNextNoNearMissRouteExportAnchorsOnTheName asserts the finding sits on the exported name, which
// is the token an author renames.
func TestNextNoNearMissRouteExportAnchorsOnTheName(t *testing.T) {
	t.Parallel()

	source := "export async function generateStaticParameters() { return []; }"
	result := rule_testing.Run(t, NextNoNearMissRouteExport, nearMissPageFile, source)
	if len(result.Diagnostics) != 1 {
		t.Fatalf("want 1 finding, got %d", len(result.Diagnostics))
	}
	onDisk := strings.TrimSpace(source) + "\n"
	diagnostic := result.Diagnostics[0]
	if got := strings.TrimSpace(onDisk[diagnostic.Range.Pos():diagnostic.Range.End()]); got != "generateStaticParameters" {
		t.Errorf("the finding spans %q, want the exported name", got)
	}
}
