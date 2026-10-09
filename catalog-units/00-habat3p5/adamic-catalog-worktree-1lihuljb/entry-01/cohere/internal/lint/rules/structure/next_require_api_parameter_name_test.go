package structure

import (
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

const apiParameterPageFile = "/repository/app/account/page.tsx"
const apiParameterLayoutFile = "/repository/app/account/layout.tsx"

func TestNextRequireApiParameterNameFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		fileName   string
		sourceText string
		wantId     string
	}{
		{
			"parameters in generateMetadata",
			apiParameterPageFile,
			"export async function generateMetadata(argument: { parameters: unknown }) {\n    return argument;\n}\n",
			"useParamsNotParameters",
		},
		{
			"searchParameters in generateMetadata",
			apiParameterPageFile,
			"export async function generateMetadata(argument: { searchParameters: unknown }) {\n    return argument;\n}\n",
			"useSearchParamsNotSearchParameters",
		},
		// All three framework functions, since testing one and trusting the set is how a rule ends
		// up enforcing a third of what its name claims.
		{
			"parameters in generateStaticParams",
			apiParameterPageFile,
			"export async function generateStaticParams(argument: { parameters: unknown }) {\n    return argument;\n}\n",
			"useParamsNotParameters",
		},
		{
			"parameters in generateViewport",
			apiParameterPageFile,
			"export async function generateViewport(argument: { parameters: unknown }) {\n    return argument;\n}\n",
			"useParamsNotParameters",
		},
		// A layout has the same framework surface as a page.
		{
			"parameters in a layout file",
			apiParameterLayoutFile,
			"export async function generateMetadata(argument: { parameters: unknown }) {\n    return argument;\n}\n",
			"useParamsNotParameters",
		},
		// The destructured form, which is how anyone actually writes it. The binding name is not
		// judged; the type's field name is.
		{
			"a destructured argument with the wrong field name",
			apiParameterPageFile,
			"export async function generateMetadata({ parameters }: { parameters: unknown }) {\n" +
				"    return parameters;\n}\n",
			"useParamsNotParameters",
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t, rule_testing.Run(t, NextRequireApiParameterName, testCase.fileName,
				testCase.sourceText), testCase.wantId)
		})
	}
}

func TestNextRequireApiParameterNameStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		fileName   string
		sourceText string
	}{
		// The shape the rule asks for, which is the framework's spelling.
		{
			"params in generateMetadata",
			apiParameterPageFile,
			"export async function generateMetadata({ params }: { params: unknown }) {\n    return params;\n}\n",
		},
		{
			"searchParams in generateMetadata",
			apiParameterPageFile,
			"export async function generateMetadata({ searchParams }: { searchParams: unknown }) {\n" +
				"    return searchParams;\n}\n",
		},
		// The function-name gate. A page's own generate function takes whatever its author decides.
		{
			"a function Next does not call",
			apiParameterPageFile,
			"export function generateReport(argument: { parameters: unknown }) {\n    return argument;\n}\n",
		},
		// The file gate. A helper module exporting a framework-named function is not one Next calls.
		{
			"generateMetadata in an ordinary module",
			"/repository/source/api/Metadata.ts",
			"export async function generateMetadata(argument: { parameters: unknown }) {\n    return argument;\n}\n",
		},
		{
			"generateMetadata in a component file",
			"/repository/source/components/Panel.tsx",
			"export async function generateMetadata(argument: { parameters: unknown }) {\n    return argument;\n}\n",
		},
		// A named type reference is not judged: the mistake would live in that type's declaration,
		// which is elsewhere and may be shared, so a finding here points at the wrong file.
		{
			"a named type rather than an inline one",
			apiParameterPageFile,
			"interface ArgumentInterface { parameters: unknown }\n" +
				"export async function generateMetadata(argument: ArgumentInterface) {\n    return argument;\n}\n",
		},
		{
			"an untyped argument",
			apiParameterPageFile,
			"export async function generateMetadata(argument: any) {\n    return argument;\n}\n",
		},
		{
			"no arguments at all",
			apiParameterPageFile,
			"export async function generateMetadata() {\n    return {};\n}\n",
		},
		// A second parameter is the author's own, since these functions take one object from Next.
		{
			"a wrong name in a second parameter",
			apiParameterPageFile,
			"export async function generateMetadata(first: { params: unknown }, second: { parameters: unknown }) {\n" +
				"    return [first, second];\n}\n",
		},
		{
			"a page with no framework functions",
			apiParameterPageFile,
			"export default function AccountPageRoute() { return null; }\n",
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.Run(t, NextRequireApiParameterName, testCase.fileName,
				testCase.sourceText))
		})
	}
}

// The fix replaces exactly the field name, which is the whole repair.
//
// Asserted rather than assumed, because a fix is the one part of a rule that changes source: a
// wrong range here silently rewrites something else, and no fixture that only checks message ids
// would notice.
func TestNextRequireApiParameterNameFixesTheFieldName(t *testing.T) {
	t.Parallel()

	result := rule_testing.Run(t, NextRequireApiParameterName, apiParameterPageFile,
		"export async function generateMetadata(argument: { parameters: unknown }) {\n    return argument;\n}\n")

	if len(result.Diagnostics) != 1 {
		t.Fatalf("want one finding, got %d", len(result.Diagnostics))
	}
	fixes := result.Diagnostics[0].Fixes
	if len(fixes) != 1 {
		t.Fatalf("want one fix, got %d", len(fixes))
	}
	if fixes[0].Text != "params" {
		t.Fatalf("want the fix to write params, got %q", fixes[0].Text)
	}

	// The text is half the fix and the range is the other half. Writing "params" over the wrong span
	// produces the right characters in the wrong place, and every assertion above still passes.
	rule_testing.ExpectFixedSource(t, result,
		"export async function generateMetadata(argument: { params: unknown }) {\n    return argument;\n}\n")

	searchResult := rule_testing.Run(t, NextRequireApiParameterName, apiParameterPageFile,
		"export async function generateMetadata(argument: { searchParameters: unknown }) {\n    return argument;\n}\n")
	if text := searchResult.Diagnostics[0].Fixes[0].Text; text != "searchParams" {
		t.Fatalf("want the fix to write searchParams, got %q", text)
	}
	rule_testing.ExpectFixedSource(t, searchResult,
		"export async function generateMetadata(argument: { searchParams: unknown }) {\n    return argument;\n}\n")
}
