package nexus

import (
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/system-inc/cohere/internal/lint/testing"
)

func TestBoundaryNoInternalImportFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		fileName string
		source   string
		wantId   string
	}{
		{
			"aliased reach into internal",
			"/repo/libraries/structure/source/Thing.ts",
			"import { Detail } from '@structure/source/widget/internal/Detail';\n",
			"aliasedInternal",
		},
		{
			"relative reach from a sibling folder",
			"/repo/source/other/Thing.ts",
			"import { Detail } from '../widget/internal/Detail';\n",
			"outsideInternal",
		},
		{
			"require reaching in from outside",
			"/repo/source/other/Thing.ts",
			"const detail = require('../widget/internal/Detail');\n",
			"outsideInternal",
		},
		{
			"dynamic import reaching in from outside",
			"/repo/source/other/Thing.ts",
			"const detail = import('../widget/internal/Detail');\n",
			"outsideInternal",
		},
		{
			// The innermost internal owns the file, so the outer folder is still outside it.
			"nested internal is owned by the inner one",
			"/repo/source/widget/internal/Thing.ts",
			"import { Detail } from './deep/internal/Detail';\n",
			"outsideInternal",
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, BoundaryNoInternalImport, testCase.fileName, testCase.source)
			rule_testing.ExpectFindings(t, result, testCase.wantId)
		})
	}
}

func TestBoundaryNoInternalImportStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		fileName string
		source   string
	}{
		{
			"the owning folder may reach in",
			"/repo/source/widget/Thing.ts",
			"import { Detail } from './internal/Detail';\n",
		},
		{
			"a file below the owning folder may reach in",
			"/repo/source/widget/nested/Thing.ts",
			"import { Detail } from '../internal/Detail';\n",
		},
		{
			"a file already inside internal may reach its siblings",
			"/repo/source/widget/internal/Thing.ts",
			"import { Detail } from './Detail';\n",
		},
		{
			"an ordinary import naming no internal folder",
			"/repo/source/other/Thing.ts",
			"import { Thing } from '../widget/Thing';\n",
		},
		// These two are the discriminating cases. A substring match would flag both, and a rule
		// that fires on a word merely containing the one it guards is worse than no rule: the fix
		// a reader is asked to make does not exist.
		{
			"a folder whose name merely contains internal",
			"/repo/source/other/Thing.ts",
			"import { format } from '../internationalization/Format';\n",
		},
		{
			"an aliased path with internal inside a longer segment",
			"/repo/libraries/structure/source/Thing.ts",
			"import { format } from '@structure/source/internationalization/Format';\n",
		},
		{
			"a computed dynamic import has no specifier to judge",
			"/repo/source/other/Thing.ts",
			"const detail = import(somePath);\n",
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, BoundaryNoInternalImport, testCase.fileName, testCase.source)
			rule_testing.ExpectClean(t, result)
		})
	}
}

// A finding has to land on the line the author can suppress. A node's Pos() includes its leading
// trivia, so reporting the declaration anchors the finding at the first comment above the import,
// where no `eslint-disable-next-line` can reach it.
func TestBoundaryNoInternalImportReportsAtTheSpecifier(t *testing.T) {
	t.Parallel()

	sourceText := "// Dependencies\n// a second comment\nimport { Detail } from '../widget/internal/Detail';\\n"

	result := rule_testing.RunWithOptions(t, BoundaryNoInternalImport, "/repo/source/other/Thing.ts", sourceText, nil)
	if len(result.Diagnostics) == 0 {
		t.Fatalf("expected a finding, got none")
	}

	line, _ := scanner.GetECMALineAndByteOffsetOfPosition(result.SourceFile, result.Diagnostics[0].Range.Pos())
	const importLine = 2 // zero-based, so the third line
	if line != importLine {
		t.Fatalf("expected the finding on the import line (%d), got line %d", importLine+1, line+1)
	}
}
