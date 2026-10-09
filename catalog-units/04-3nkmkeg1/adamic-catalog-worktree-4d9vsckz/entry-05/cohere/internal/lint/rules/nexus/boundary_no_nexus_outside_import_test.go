package nexus

import (
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

const nexusFile = "/repository/libraries/nexus/code-quality/Thing.ts"

func TestBoundaryNoNexusOutsideImportFires(t *testing.T) {
	t.Parallel()

	// All three import shapes, because a boundary that only guards the static form is a boundary
	// with a documented way around it.
	cases := []struct {
		name       string
		sourceText string
	}{
		{"static", "import { Thing } from '@structure/source/Thing';\n"},
		{"static default", "import Thing from '@project/ProjectSettings';\n"},
		{"static bare alias", "import '@base';\n"},
		{"dynamic", "export async function load() {\n    return await import('@structure/source/Thing');\n}\n"},
		{"require", "const thing = require('@base/source/Thing');\n"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, BoundaryNoNexusOutsideImport, nexusFile, testCase.sourceText)
			rule_testing.ExpectFindings(t, result, "forbiddenOutsideImport")
		})
	}
}

func TestBoundaryNoNexusOutsideImportStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		fileName   string
		sourceText string
	}{
		// Inside nexus, reaching sideways and downward is exactly what nexus is supposed to do.
		{"relative", nexusFile, "import { Thing } from './Thing';\n"},
		{"parent relative", nexusFile, "import { Thing } from '../Thing';\n"},
		{"nexus alias", nexusFile, "import { Thing } from '@nexus/code-quality/Thing';\n"},
		{"package", nexusFile, "import * as NodePath from 'node:path';\n"},

		// The prefix has to end at a boundary. A package genuinely named "@projections" is not
		// "@project", and flagging it would be the rule claiming a name it never meant to.
		{"lookalike alias", nexusFile, "import { Thing } from '@projections/Thing';\n"},
		{"lookalike base", nexusFile, "import { Thing } from '@basement/Thing';\n"},

		// Outside nexus, every one of these is legal, so the rule must decline the file entirely.
		{"outside nexus static", "/repository/libraries/structure/source/Thing.ts", "import { Thing } from '@project/ProjectSettings';\n"},
		{"outside nexus dynamic", "/repository/source/Thing.ts", "export async function load() {\n    return await import('@structure/source/Thing');\n}\n"},

		// A path that merely mentions nexus is not the nexus library.
		{"nexus lookalike path", "/repository/libraries/nexus-template/source/Thing.ts", "import { Thing } from '@project/ProjectSettings';\n"},

		// A computed specifier has no string to read, and the rule must not crash reaching for one.
		{"computed dynamic import", nexusFile, "export async function load(name: string) {\n    return await import(name);\n}\n"},

		// A method named require is not the CommonJS require.
		{"method named require", nexusFile, "declare const loader: { require(name: string): unknown };\nconst thing = loader.require('@base/source/Thing');\n"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, BoundaryNoNexusOutsideImport, testCase.fileName, testCase.sourceText)
			rule_testing.ExpectClean(t, result)
		})
	}
}

// TestBoundaryNoNexusOutsideImportReportsTheSpecifier pins where the finding lands.
//
// This rule reported the whole declaration until a census compared it against its two sibling
// boundary rules, which both wrap the node in `imports.SpecifierNode`. Every fixture above stayed
// green through the change, because `ExpectFindings` asserts message ids and count and nothing
// else: a rule whose defect is where it points passes a complete fixture pair while being wrong.
//
// The difference is narrower than it first looks, and the measurement is worth keeping. On a
// single-line import the two spellings agree, because `ctx.ReportNode` routes through `TokenRange`
// and strips leading trivia, so a comment above the statement does not move the finding. Measured
// on eight shapes; seven agreed.
//
// The eighth is a multiline import, which is common in real code. There the declaration starts on
// the line of the `import` keyword while the specifier sits several lines down, so the finding
// landed on line 1 and an `eslint-disable-next-line` written above the specifier could not reach
// it. A `-next-line` directive matches the line after itself, so the finding was unreachable by any
// suppression the author could write, and it read as a real finding in every count.
func TestBoundaryNoNexusOutsideImportReportsTheSpecifier(t *testing.T) {
	t.Parallel()

	for _, testCase := range []struct {
		name       string
		sourceText string
		want       string
	}{
		{
			"a single-line import reports the specifier",
			"import { Thing } from '@structure/source/Thing';",
			"'@structure/source/Thing'",
		},
		{
			// The shape that was wrong. The declaration spans four lines and the specifier is on
			// the last of them.
			"a multiline import reports the specifier rather than the declaration",
			"import {\n  Thing,\n  Other,\n} from '@structure/source/Thing';",
			"'@structure/source/Thing'",
		},
		{
			"a dynamic import reports the specifier",
			"const loaded = await import('@structure/source/Thing');",
			"'@structure/source/Thing'",
		},
		{
			"a require call reports the specifier",
			"const loaded = require('@structure/source/Thing');",
			"'@structure/source/Thing'",
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, BoundaryNoNexusOutsideImport, nexusFile, testCase.sourceText)
			if len(result.Diagnostics) != 1 {
				t.Fatalf("want exactly one finding, got %d", len(result.Diagnostics))
			}
			reported := testCase.sourceText[result.Diagnostics[0].Range.Pos():result.Diagnostics[0].Range.End()]
			if reported != testCase.want {
				t.Errorf("finding covers %q, want %q", reported, testCase.want)
			}
		})
	}
}
