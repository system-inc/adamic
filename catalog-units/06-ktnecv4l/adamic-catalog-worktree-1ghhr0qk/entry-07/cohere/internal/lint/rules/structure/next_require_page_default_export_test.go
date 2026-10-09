package structure

import (
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

const pageFile = "/repository/app/account/settings/page.tsx"
const layoutFile = "/repository/app/account/settings/layout.tsx"

func TestNextRequirePageDefaultExportFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		wantIds    []string
	}{
		{
			"no default export at all",
			"export function SettingsPageRoute() {\n    return null;\n}\n",
			[]string{"pageRequireDefaultExport"},
		},
		{
			"a file with only named exports",
			"export const value = 1;\nexport function helper() { return 1; }\n",
			[]string{"pageRequireDefaultExport"},
		},
		{
			"an empty page file",
			"",
			[]string{"pageRequireDefaultExport"},
		},
		// `export = x` is the CommonJS form and is not a default export, so a page using it still
		// has no route. Distinguishing the two is what keeps this a missing-export finding rather
		// than a name complaint about an export Next will not read. Found by treating export
		// equals as a default and watching the suite stay green.
		{
			"an export equals rather than a default export",
			"declare function SettingsPageRoute(): null;\nexport = SettingsPageRoute;\n",
			[]string{"pageRequireDefaultExport"},
		},
		// The separate-statement form. Two findings, because the name is wrong as well as the
		// placement, and they are two different edits.
		{
			"a separate default export statement",
			"function Settings() {\n    return null;\n}\nexport default Settings;\n",
			[]string{"pageDefaultExportInline", "pageDefaultExportNameSuffix"},
		},
		{
			"a separate default export of a correctly named function",
			"function SettingsPageRoute() {\n    return null;\n}\nexport default SettingsPageRoute;\n",
			[]string{"pageDefaultExportInline"},
		},
		{
			"an inline default export with the wrong name",
			"export default function Settings() {\n    return null;\n}\n",
			[]string{"pageDefaultExportNameSuffix"},
		},
		// A name holding PageRoute as a substring rather than a suffix is still wrong.
		{
			"a name with PageRoute in the middle",
			"export default function SettingsPageRouteView() {\n    return null;\n}\n",
			[]string{"pageDefaultExportNameSuffix"},
		},
		{
			"a default exported class with the wrong name",
			"export default class Settings {\n    render() { return null; }\n}\n",
			[]string{"pageDefaultExportNameSuffix"},
		},
		// The specifier spelling of the separate-statement form. The function is declared in this
		// file, so it could carry the export itself, and the name is this file's to choose.
		{
			"a local function exported as default through a specifier",
			"function Settings() {\n    return null;\n}\nexport { Settings as default };\n",
			[]string{"pageDefaultExportInline", "pageDefaultExportNameSuffix"},
		},
		{
			"a correctly named local function exported as default through a specifier",
			"function SettingsPageRoute() {\n    return null;\n}\nexport { SettingsPageRoute as default };\n",
			[]string{"pageDefaultExportInline"},
		},
		// A type-only re-export carries no value, so Next still finds no page. Counting it would
		// turn the specifier support into a way to silence a missing route.
		{
			"a type-only default re-export",
			"export type { SettingsPage as default } from './SettingsPage';\n",
			[]string{"pageRequireDefaultExport"},
		},
		{
			"a type-only default specifier",
			"export { type SettingsPage as default } from './SettingsPage';\n",
			[]string{"pageRequireDefaultExport"},
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t, rule_testing.Run(t, NextRequirePageDefaultExport, pageFile,
				testCase.sourceText), testCase.wantIds...)
		})
	}
}

func TestNextRequirePageDefaultExportStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		fileName   string
		sourceText string
	}{
		// The shape the rule asks for.
		{
			"an inline default export ending in PageRoute",
			pageFile,
			"export default function SettingsPageRoute() {\n    return null;\n}\n",
		},
		{
			"a correctly named page alongside other exports",
			pageFile,
			"export const metadata = { title: 'Settings' };\n" +
				"export default function SettingsPageRoute() {\n    return null;\n}\n",
		},
		{
			"a default exported class ending in PageRoute",
			pageFile,
			"export default class SettingsPageRoute {\n    render() { return null; }\n}\n",
		},
		// The file gate. A layout has its own contract that this rule would misdescribe, so it is
		// not held to the page convention.
		{
			"a layout file with no default export",
			layoutFile,
			"export function Settings() {\n    return null;\n}\n",
		},
		{
			"an ordinary component file",
			"/repository/source/components/Settings.tsx",
			"export function Settings() {\n    return null;\n}\n",
		},
		// A page.ts is not a React file, so it is not a page in the sense this rule means.
		{
			"a page.ts rather than page.tsx",
			"/repository/app/account/page.ts",
			"export function Settings() {\n    return null;\n}\n",
		},
		// A named export is not the route's default, and treating it as one would report every
		// helper a page file exports.
		{
			"a named export that is not default",
			pageFile,
			"export function SettingsPageRoute() { return null; }\n" +
				"export default function OtherPageRoute() { return null; }\n",
		},
		// Re-exported defaults, which Next resolves exactly as it resolves a declared one. Both are
		// real pages, verbatim: www-phi-health has 78 page files in these two shapes and the rule
		// reported every one of them as having no default export.
		{
			"a named page re-exported as default (www-phi-health app/ops/analytics/sessions/page.tsx)",
			pageFile,
			"// Dependencies - Frameworks\nimport type { Metadata } from 'next';\n\n" +
				"// Next.js Metadata\nexport function generateMetadata(): Metadata {\n    return {\n" +
				"        title: 'Sessions • Analytics • Ops',\n    };\n}\n\n" +
				"// Shim the default export from Structure\n" +
				"export { AnalyticsSessionsPage as default } from '@structure/source/modules/engagement/ops/sessions/AnalyticsSessionsPage';\n",
		},
		{
			"another page's default re-exported (www-phi-health app/(main-layout)/research/page.tsx)",
			pageFile,
			"// Next.js Metadata\nexport function generateMetadata() {\n    return {\n        title: 'Research',\n    };\n}\n\n" +
				"export { default } from './subscribe/page';\n",
		},
		// An imported binding exported as default. Its declaration is in another module, so it cannot
		// be moved onto a declaration here and its name is not this file's to choose. Reporting
		// either would demand an edit this file cannot make.
		{
			"a default import exported as default",
			pageFile,
			"import SettingsPage from '@structure/source/modules/account/SettingsPage';\n\nexport default SettingsPage;\n",
		},
		{
			"a named import exported as default",
			pageFile,
			"import { SettingsPage } from '@structure/source/modules/account/SettingsPage';\n\nexport default SettingsPage;\n",
		},
		{
			"a named import exported as default through a specifier",
			pageFile,
			"import { SettingsPage } from '@structure/source/modules/account/SettingsPage';\n\nexport { SettingsPage as default };\n",
		},
		// A wrapped component has no declaration to carry the export and no name of its own, so
		// neither repair applies. ESLint's original reports neither; this rule used to report the
		// inline form, asking for `export default function` on a call.
		{
			"a default export of a call",
			pageFile,
			"import { memo } from 'react';\nfunction SettingsPageRoute() {\n    return null;\n}\nexport default memo(SettingsPageRoute);\n",
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.Run(t, NextRequirePageDefaultExport, testCase.fileName,
				testCase.sourceText))
		})
	}
}
