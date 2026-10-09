package structure

import (
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// The corpus is written rather than imported, because this is a house rule and no oxc
// implementation exists. Every case below was run against the live TypeScript rule through
// ESLint's Linter API before it was written down, so each records a measured verdict rather than
// a belief about one. A rule with no imported floor is a weaker artifact than a ported one and
// the next reader should know which kind this is.
//
// The probe that produced these needed `files: ['**/*.ts', '**/*.tsx']` in its flat configuration. Without
// it ESLint matches no configuration for a TypeScript file and reports nothing, and the first run
// of 400 real files came back clean for exactly that reason. The control that caught it was the
// tree's own fixture, which is a known violation and must report.
func TestConsistencyOrganizeImportsFires(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name   string
		source string
		// reported is the exact source text the finding must span. The rule carries no fix, so
		// the span is the only thing that can catch it pointing at the wrong node.
		reported string
	}{
		{
			name:     "react sorts ahead of every other framework package",
			source:   "import a from 'next/link';\nimport React from 'react';\nimport b from 'react-dom';\n\nexport const X = 1;\n",
			reported: "import a from 'next/link';",
		},
		{
			// The react-first tiebreak, from the reporting side. `next/link` sorts before `react`
			// by string comparison, so a block written in plain alphabetical order is wrong.
			name:     "inside Frameworks, plain alphabetical order is not canonical order",
			source:   "// Dependencies - Frameworks\nimport a from 'next/link';\nimport React from 'react';\n\nexport const X = 1;\n",
			reported: "import a from 'next/link';",
		},
		{
			name:     "alphabetical inside a group",
			source:   "import b from 'zebra';\nimport a from 'alpha';\n\nexport const X = 1;\n",
			reported: "import b from 'zebra';",
		},
		{
			name:     "a single group with no header at all still needs one",
			source:   "import React from 'react';\n\nexport const X = 1;\n",
			reported: "import React from 'react';",
		},
		{
			name:     "a header naming a group that does not exist",
			source:   "// Dependencies - Bogus\nimport React from 'react';\n\nexport const X = 1;\n",
			reported: "import React from 'react';",
		},
		{
			name:     "a bare Dependencies header is not a group header",
			source:   "// Dependencies\nimport React from 'react';\n\nexport const X = 1;\n",
			reported: "import React from 'react';",
		},
		{
			name:     "a directive with no trailing comment",
			source:   "'use client';\n\nimport React from 'react';\n\nexport const X = 1;\n",
			reported: "import React from 'react';",
		},
		{
			// The other side of telling the two directives apart: a server directive carrying the
			// client comment is wrong and reports. Without this the comment table could map both
			// directives to one string and nothing would notice.
			name:     "use server carrying the client comment",
			source:   "'use server'; // Uses client-only features\n\n// Dependencies - Frameworks\nimport React from 'react';\n\nexport const X = 1;\n",
			reported: "import React from 'react';",
		},
		{
			name:     "use server takes its own comment",
			source:   "'use server';\nimport React from 'react';\nexport const X = 1;\n",
			reported: "import React from 'react';",
		},
		{
			name:     "a type-only import splits away from the value import of the same module",
			source:   "import type { T } from 'alpha';\nimport v from 'alpha';\n\nexport const X = 1;\n",
			reported: "import type { T } from 'alpha';",
		},
		{
			name:     "a statement written between two imports",
			source:   "// Dependencies - Frameworks\nimport React from 'react';\n\ninterface Thing { a: number }\n\n// Dependencies - Utilities\nimport x from '@structure/source/utilities/A';\n\nexport const X = 1;\n",
			reported: "import React from 'react';",
		},
		{
			name:     "a side-effect import sorts by its specifier like any other",
			source:   "import 'zzz-side';\nimport a from 'alpha';\n\nexport const X = 1;\n",
			reported: "import 'zzz-side';",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, ConsistencyRequireOrganizedImports, "/repository/app/Thing.tsx", testCase.source)
			rule_testing.ExpectFindings(t, result, "importsNotOrganized")

			span := result.Diagnostics[0].Range
			got := testCase.source[span.Pos():span.End()]
			if got != testCase.reported {
				t.Fatalf("finding spans %q, want %q", got, testCase.reported)
			}
		})
	}
}

func TestConsistencyOrganizeImportsStaysSilent(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		// fileName is the path the rule is handed. Empty means the default below, and it exists so
		// a case can say something about the file rather than about its contents.
		fileName string
		source   string
	}{
		{
			name:   "two groups already in canonical order",
			source: "// Dependencies - Frameworks\nimport React from 'react';\n\n// Dependencies - Utilities\nimport x from '@structure/source/utilities/A';\n\nexport const X = 1;\n",
		},
		{
			// The half that can see the react-first tiebreak. Every other Frameworks fixture is
			// unsorted anyway and reports with or without the tiebreak, so both paths reach the same
			// verdict and none of them can catch a mutation that removes it. This block is clean
			// only because react jumps the queue: by string comparison `next/link` sorts first.
			name:   "inside Frameworks, react sits above a package that sorts before it",
			source: "// Dependencies - Frameworks\nimport React from 'react';\nimport a from 'next/link';\n\nexport const X = 1;\n",
		},
		{
			// Two imports in ONE group, already ascending. Every other clean case here holds at most
			// one import per group, so the comparison inside a group is never exercised by them and
			// a mutation flipping it to descending changed no verdict at all. Measured on the
			// original: this input is clean.
			name:   "two imports in one group, already in ascending order",
			source: "// Dependencies - Third-party\nimport a from 'alpha';\nimport z from 'zebra';\n\nexport const X = 1;\n",
		},
		{
			// use server takes its OWN comment, not the client one. Nothing else here can see the
			// two directives being told apart: the reporting cases pass through a directive with no
			// comment at all, which reports whichever text the rule would have written.
			name:   "use server already carrying the server comment",
			source: "'use server'; // Uses server-only features\n\n// Dependencies - Frameworks\nimport React from 'react';\n\nexport const X = 1;\n",
		},
		{
			// Only the FIRST statement is a directive, which is what the language says too. A
			// string expression written below the imports is an ordinary statement, so it is not
			// re-emitted with a comment and its presence does not make the file unorganized.
			name:   "a directive-shaped string below the imports is not a directive",
			source: "// Dependencies - Frameworks\nimport React from 'react';\n\n'use client';\n\nexport const X = 1;\n",
		},
		{
			name:   "a directive that already carries its comment",
			source: "'use client'; // Uses client-only features\n\n// Dependencies - Frameworks\nimport React from 'react';\n\nexport const X = 1;\n",
		},
		{
			name:   "a file description block between the directive and the headers",
			source: "'use client'; // Uses client-only features\n\n/*\n * What this file does.\n */\n\n// Dependencies - Frameworks\nimport React from 'react';\n\nexport const X = 1;\n",
		},
		{
			name:   "a comment a reader wrote above an import travels with it",
			source: "// Dependencies - Frameworks\n// why we need react\nimport React from 'react';\n\nexport const X = 1;\n",
		},
		{
			name:   "a disable comment above an import is preserved when a header sits above it",
			source: "// Dependencies - Frameworks\n// eslint-disable-next-line no-restricted-imports\nimport React from 'react';\n\nexport const X = 1;\n",
		},
		{
			name:   "a file with no imports has no section to organize",
			source: "export const X = 1;\n",
		},
		{
			name:   "a preamble comment above the section stays preamble",
			source: "// Copyright 2026\n\n// Dependencies - Frameworks\nimport React from 'react';\n\nexport const X = 1;\n",
		},
		{
			name:   "no blank line before the first statement after the section",
			source: "// Dependencies - Frameworks\nimport React from 'react';\nexport const X = 1;\n",
		},
		{
			name:   "every group in canonical order, which is the ordering itself under test",
			source: "// Dependencies - Node\nimport NodePath from 'node:path';\n\n// Dependencies - Frameworks\nimport React from 'react';\n\n// Dependencies - Third-party\nimport lodash from 'lodash';\n\n// Dependencies - Nexus\nimport n from '@nexus/source/Thing';\n\n// Dependencies - Theme\nimport t from '@structure/source/theme/Theme';\n\n// Dependencies - Types\nimport type { T } from '@structure/source/Thing';\n\n// Dependencies - APIs\nimport api from '@structure/source/api/Api';\n\n// Dependencies - Hooks\nimport h from '@structure/source/hooks/useThing';\n\n// Dependencies - Components\nimport c from '@structure/source/components/Button';\n\n// Dependencies - Local Components\nimport l from './Thing';\n\n// Dependencies - Animations\nimport m from 'motion/react';\n\n// Dependencies - Assets\nimport icon from '@phosphor-icons/react';\n\n// Dependencies - Utilities\nimport u from '@structure/source/utilities/Strings';\n\nexport const X = 1;\n",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			fileName := testCase.fileName
			if fileName == "" {
				fileName = "/repository/app/Thing.tsx"
			}
			result := rule_testing.Run(t, ConsistencyRequireOrganizedImports, fileName, testCase.source)
			rule_testing.ExpectClean(t, result)
		})
	}
}

// Generated code is held to the rule like hand-written code (Kirk's ruling on generated code, #c076xbg):
// the generator writes the house form, so a finding here is a defect in the generator. The rule used to
// decline these files, and both spellings of generated output are reported now.
func TestConsistencyOrganizeImportsHoldsGeneratedFilesToTheRule(t *testing.T) {
	t.Parallel()
	for _, fileName := range []string{"/repository/app/generated/Thing.tsx", "/repository/app/Thing.generated.ts"} {
		result := rule_testing.Run(t, ConsistencyRequireOrganizedImports, fileName, "import z from 'zebra';\nimport a from 'alpha';\n\nexport const X = 1;\n")
		rule_testing.ExpectFindings(t, result, "importsNotOrganized")
	}
}

// TestConsistencyOrganizeImportsFixesTheSection pins what the fixer writes.
//
// # Why this rule ships a fixer now, having declined one
//
// The original's fixer carries three defects, each measured on the real rule and each invisible
// until applied: a disable comment above the first import is deleted, a trailing comment migrates
// to a different import, and a statement between two imports is moved below both. This port declined
// to carry them, which left every reported file to be reordered by hand.
//
// All three are addressed rather than reproduced. Two are repaired: a trailing comment now travels
// with the import it trails, and the section span was extended to cover it so the fix cannot leave a
// duplicate behind. The third cannot be repaired, because a statement written mid-section has no
// correct home once the imports around it are reordered, so those files are reported without a fix.
// A suppression above the section boundary is declined for the same reason: preserving it displaced
// is safer than deleting it and still not right.
//
// Each case below was run against the original for comparison, and the divergences are the point.
func TestConsistencyOrganizeImportsFixesTheSection(t *testing.T) {
	t.Parallel()

	// The ordinary case, which is every one of the 109 files this landed for.
	rule_testing.ExpectFixedSource(t,
		rule_testing.Run(t, ConsistencyRequireOrganizedImports, "Component.tsx",
			"import alpha from 'alpha';\nimport React from 'react';\n"),
		"// Dependencies - Frameworks\nimport React from 'react';\n\n// Dependencies - Third-party\nimport alpha from 'alpha';\n")

	/*
	 * A trailing comment stays on the import it trails. The original moves it above whichever
	 * import sorts next, because a trailing comment is syntactically the leading trivia of the
	 * following statement, so the note ends up describing a module it was never about.
	 */
	rule_testing.ExpectFixedSource(t,
		rule_testing.Run(t, ConsistencyRequireOrganizedImports, "Component.tsx",
			"import zebra from 'zebra'; // note about zebra\nimport React from 'react';\n"),
		"// Dependencies - Frameworks\nimport React from 'react';\n\n// Dependencies - Third-party\nimport zebra from 'zebra'; // note about zebra\n")

	// A comment above a non-first import travels with it and keeps suppressing what it suppressed.
	rule_testing.ExpectFixedSource(t,
		rule_testing.Run(t, ConsistencyRequireOrganizedImports, "Component.tsx",
			"import React from 'react';\n// eslint-disable-next-line no-explicit-any\nimport alpha from 'alpha';\n"),
		"// Dependencies - Frameworks\nimport React from 'react';\n\n// Dependencies - Third-party\n// eslint-disable-next-line no-explicit-any\nimport alpha from 'alpha';\n")
}

// TestConsistencyOrganizeImportsDeclinesToFixWhatItCannotMove pins the two refusals.
//
// Both still report. The finding is unchanged and a reader reorders the section by hand, which is
// what the rule did for every file before it could fix anything. What changed is only that the
// fixer says nothing rather than guessing.
func TestConsistencyOrganizeImportsDeclinesToFixWhatItCannotMove(t *testing.T) {
	t.Parallel()

	/*
	 * A statement between two imports runs before the imports below it. The rendering appends it
	 * after every one of them, which would move code across an evaluation boundary, and there is no
	 * correct alternative: the canonical form has no place to put a statement mid-section when the
	 * imports around it are being reordered.
	 */
	interleaved := rule_testing.Run(t, ConsistencyRequireOrganizedImports, "Component.tsx",
		"import alpha from 'alpha';\nconsole.info('between');\nimport React from 'react';\n")
	rule_testing.ExpectFindings(t, interleaved, "importsNotOrganized")
	if len(interleaved.Diagnostics) > 0 && len(interleaved.Diagnostics[0].Fixes) > 0 {
		t.Error("a file with an interleaved statement must not be fixed")
	}

	/*
	 * A suppression above the first import sits outside the compared section, so a fix would rewrite
	 * everything beneath it and leave it silencing a group header instead of the import it was
	 * written for. The original deletes it outright, which is worse and is the defect this refusal
	 * exists to avoid inheriting.
	 */
	suppressed := rule_testing.Run(t, ConsistencyRequireOrganizedImports, "Component.tsx",
		"// eslint-disable-next-line no-explicit-any\nimport alpha from 'alpha';\nimport React from 'react';\n")
	rule_testing.ExpectFindings(t, suppressed, "importsNotOrganized")
	if len(suppressed.Diagnostics) > 0 && len(suppressed.Diagnostics[0].Fixes) > 0 {
		t.Error("a file whose suppression precedes the section must not be fixed")
	}
}
