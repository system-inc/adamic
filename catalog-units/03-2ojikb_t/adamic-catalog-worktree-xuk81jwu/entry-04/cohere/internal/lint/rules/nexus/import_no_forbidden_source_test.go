package nexus

import (
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/system-inc/cohere/internal/lint/testing"
)

const forbiddenSourceFile = "/repository/source/Thing.tsx"

func TestImportNoForbiddenSourceFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		wantId     string
		wantFix    string
	}{
		{
			"next navigation",
			"import { useRouter } from 'next/navigation';\n",
			"forbiddenNavigationImport",
			"'@structure/source/router/Navigation'",
		},
		{
			"next link",
			"import Link from 'next/link';\n",
			"forbiddenLinkImport",
			"'@structure/source/components/navigation/Link'",
		},
		{
			"next image",
			"import Image from 'next/image';\n",
			"forbiddenImageImport",
			"'@structure/source/components/images/Image'",
		},
		{
			"framer motion",
			"import { motion } from 'framer-motion';\n",
			"forbiddenMotionImport",
			"'motion/react'",
		},
		{
			"require",
			"const Link = require('next/link');\n",
			"forbiddenLinkImport",
			"'@structure/source/components/navigation/Link'",
		},
		{
			"dynamic import",
			"const Image = import('next/image');\n",
			"forbiddenImageImport",
			"'@structure/source/components/images/Image'",
		},
		{
			// A named import carries bindings the fix does not touch, on purpose. Rewriting only
			// the string is the part that is always right.
			"named imports keep their bindings",
			"import { motion, AnimatePresence } from 'framer-motion';\n",
			"forbiddenMotionImport",
			"'motion/react'",
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, ImportNoForbiddenSource, forbiddenSourceFile, testCase.sourceText)
			rule_testing.ExpectFindings(t, result, testCase.wantId)

			// A fixable rule whose fix is never checked is half-tested: the finding proves it saw
			// the problem, the fix text proves it would repair it rather than corrupt the file.
			fixes := result.Diagnostics[0].Fixes
			if len(fixes) != 1 {
				t.Fatalf("expected exactly one fix, got %d", len(fixes))
			}
			if fixes[0].Text != testCase.wantFix {
				t.Fatalf("expected fix %q, got %q", testCase.wantFix, fixes[0].Text)
			}

			// The text alone does not pin the rewrite, because the range is half of it. A correct
			// replacement over the wrong span writes the right characters into the wrong place, and
			// every assertion above still passes. Measured elsewhere in this repository: pointing a
			// deletion at the enclosing node instead of the statement removed a whole function with
			// the fix text unchanged and the suite green.
			//
			// The expectation is derived from the case rather than written out, so a new row cannot
			// forget it and cannot disagree with itself.
			rule_testing.ExpectFixedSource(t, result,
				strings.Replace(testCase.sourceText, quotedSourceOf(testCase.sourceText), testCase.wantFix, 1))
		})
	}
}

// Each wrapper the ban points to imports the forbidden source by design. Reporting it would also have
// the fix rewrite the wrapper into an import of itself, so a wrapper is exempt and produces no finding
// and no fix. A file beside a wrapper is not the wrapper, so it still reports, and its fix points at
// the wrapper.
func TestImportNoForbiddenSourceExemptsOnlyTheSanctionedWrapper(t *testing.T) {
	t.Parallel()

	wrappers := []struct {
		name       string
		fileName   string
		sourceText string
	}{
		{"the Image wrapper", "/repository/libraries/structure/source/components/images/Image.tsx",
			"import NextImage from 'next/image';\n"},
		{"the Link wrapper", "/repository/libraries/structure/source/components/navigation/Link.tsx",
			"import NextLink from 'next/link';\n"},
		{"the Navigation module", "/repository/libraries/structure/source/router/Navigation.ts",
			"import { usePathname } from 'next/navigation';\n"},
		{"Navigation's router hook", "/repository/libraries/structure/source/router/hooks/useRouter.ts",
			"import { useRouter as useNextRouter } from 'next/navigation';\n"},
		{"the Image wrapper, through a dynamic import", "/repository/source/components/images/Image.tsx",
			"const NextImage = import('next/image');\n"},
	}
	for _, wrapper := range wrappers {
		t.Run(wrapper.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.Run(t, ImportNoForbiddenSource, wrapper.fileName, wrapper.sourceText))
		})
	}

	neighbours := []struct {
		name       string
		fileName   string
		sourceText string
		wantId     string
		wantFix    string
	}{
		{"a dialog beside the Image wrapper", "/repository/libraries/structure/source/components/images/dialogs/ImageDialog.tsx",
			"import Image from 'next/image';\n", "forbiddenImageImport", "'@structure/source/components/images/Image'"},
		{"a component beside the Link wrapper", "/repository/libraries/structure/source/components/navigation/NavigationMenu.tsx",
			"import Link from 'next/link';\n", "forbiddenLinkImport", "'@structure/source/components/navigation/Link'"},
		{"framer-motion has no wrapper to exempt", "/repository/libraries/structure/source/router/hooks/useRouter.ts",
			"import { motion } from 'framer-motion';\n", "forbiddenMotionImport", "'motion/react'"},
	}
	for _, neighbour := range neighbours {
		t.Run(neighbour.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, ImportNoForbiddenSource, neighbour.fileName, neighbour.sourceText)
			rule_testing.ExpectFindings(t, result, neighbour.wantId)
			rule_testing.ExpectFixedSource(t, result,
				strings.Replace(neighbour.sourceText, quotedSourceOf(neighbour.sourceText), neighbour.wantFix, 1))
		})
	}
}

// A finding has to land on the line the author can suppress.
//
// A node's Pos() includes leading trivia, so anchoring on the enclosing declaration reports at the
// first comment above the import rather than at the import. That put real findings on line 1 of the
// Next wrapper files while the `eslint-disable-next-line` sat on line 3 covering line 4, making them
// unreachable by any suppression that could be written. A finding at the wrong location is invisible
// in a findings count and fatal to suppression, so the location is asserted rather than assumed.
func TestImportNoForbiddenSourceReportsAtTheSpecifier(t *testing.T) {
	t.Parallel()

	sourceText := "// Dependencies - Frameworks\n" +
		"// This is the only place this import is valid\n" +
		"// eslint-disable-next-line nexus/import-no-forbidden-source\n" +
		"import NextImage from 'next/image';\n"

	result := rule_testing.Run(t, ImportNoForbiddenSource, forbiddenSourceFile, sourceText)
	rule_testing.ExpectFindings(t, result, "forbiddenImageImport")

	line, _ := scanner.GetECMALineAndByteOffsetOfPosition(result.SourceFile, result.Diagnostics[0].Range.Pos())
	const importLine = 3 // zero-based, so the fourth line
	if line != importLine {
		t.Fatalf("expected the finding on the import line (%d), got line %d", importLine+1, line+1)
	}
}

func TestImportNoForbiddenSourceStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"the replacement itself", "import Link from '@structure/source/components/navigation/Link';\n"},
		{"the motion replacement", "import { motion } from 'motion/react';\n"},
		{"an unrelated next module", "import { headers } from 'next/headers';\n"},
		// A prefix match would flag these. The comparison is exact, because 'next/linkify' is a
		// different package than 'next/link' and asking a reader to rewrite it would be wrong.
		{"a longer specifier sharing a prefix", "import { thing } from 'next/link-utilities';\n"},
		{"a package whose name contains a forbidden one", "import { thing } from 'framer-motion-extras';\n"},
		{"an ordinary relative import", "import { Thing } from './Thing';\n"},
		{"a computed dynamic import", "const module = import(somePath);\n"},
		{"no imports at all", "export const Value = 1;\n"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, ImportNoForbiddenSource, forbiddenSourceFile, testCase.sourceText)
			rule_testing.ExpectClean(t, result)
		})
	}
}

// quotedSourceOf returns the quoted module specifier in a one-import fixture.
//
// The fixtures all carry exactly one quoted string, which is the specifier the fix replaces, so the
// first quoted run is it. Deliberately narrow: this is a test helper for this table and not a
// parser, and a fixture that broke the assumption would fail loudly here rather than quietly assert
// the wrong rewrite.
func quotedSourceOf(sourceText string) string {
	start := strings.IndexByte(sourceText, '\'')
	if start < 0 {
		return ""
	}
	end := strings.IndexByte(sourceText[start+1:], '\'')
	if end < 0 {
		return ""
	}
	return sourceText[start : start+end+2]
}
