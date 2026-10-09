package nexus

import (
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

/*
 * aliasesFromTsconfigPaths reads the aliases out of the tsconfig rather than a hand-written list.
 *
 * The fixture is shaped like a Structure project on purpose: the project's tsconfig only extends one
 * two directories down, and that one writes `paths` relative to itself. A rule resolving the targets
 * against the project's tsconfig, or against the process, would put every alias two levels above the
 * repository and suggest nothing, so the firing cases below fail on that defect rather than passing
 * around it.
 */

// structurePathsBase is the config every Structure project extends, with the three aliases they share.
const structurePathsBase = `{
	"compilerOptions": {
		"strict": true,
		"target": "ES2022",
		"lib": ["ES2022"],
		"moduleDetection": "force",
		"types": [],
		"paths": {
			"@structure/*": ["../../libraries/structure/*"],
			"@nexus/*": ["../../libraries/structure/libraries/nexus/*"],
			"@project/*": ["../../*"]
		}
	},
	"include": ["../../**/*.ts"]
}`

const structureProjectTsconfig = `{ "extends": "./libraries/structure/TypeScriptConfiguration.json" }`

const derivedImporter = "app/features/orders/Thing.ts"

// structureFixture is a Structure-shaped project around one importing file.
func structureFixture(importerSource string) map[string]string {
	return map[string]string{
		"tsconfig.json": structureProjectTsconfig,
		"libraries/structure/TypeScriptConfiguration.json": structurePathsBase,
		derivedImporter: importerSource,
	}
}

// derivedOptions turns the derivation on, rooted at the fixture's own directory.
func derivedOptions(written ...PathAlias) func(directory string) any {
	return func(directory string) any {
		return ImportRequirePathAliasOptions{
			RepositoryRoot:           directory,
			Aliases:                  written,
			AliasesFromTsconfigPaths: true,
		}
	}
}

// descriptions are the messages a run produced, for asserting what each suggests.
func descriptions(result rule_testing.Result) []string {
	texts := make([]string, 0, len(result.Diagnostics))
	for _, diagnostic := range result.Diagnostics {
		texts = append(texts, diagnostic.Message.Description)
	}
	return texts
}

func TestImportRequirePathAliasDerivesTheAliasesFromPaths(t *testing.T) {
	t.Parallel()

	result := rule_testing.RunTypedFilesWithOptionsFor(t, ImportRequirePathAlias, structureFixture(
		"import { Helper } from '../../../libraries/structure/source/Helper';\n"+
			"import { Tool } from '../../../libraries/structure/libraries/nexus/source/Tool';\n"+
			"import { Shared } from '../../shared/Shared';\n"+
			"import { Sibling } from './Sibling';\n"),
		derivedImporter, derivedOptions())
	rule_testing.ExpectFindings(t, result, "useAlias", "useAlias", "useAlias")

	texts := descriptions(result)
	for index, want := range []string{
		"'@structure/source/Helper'",
		// The nested alias wins over the one that holds it, as a written list's does.
		"'@nexus/source/Tool'",
		// The root alias, written `../../*` two directories down.
		"'@project/app/shared/Shared'",
	} {
		if !strings.Contains(texts[index], want) {
			t.Errorf("finding %d suggests the wrong alias, want %s: %s", index, want, texts[index])
		}
	}
}

// The fixer writes the derived alias, which resolves by construction.
func TestImportRequirePathAliasFixesWithADerivedAlias(t *testing.T) {
	t.Parallel()

	rule_testing.ExpectFixedSource(t,
		rule_testing.RunTypedFilesWithOptionsFor(t, ImportRequirePathAlias, structureFixture(
			"import { Helper } from '../../../libraries/structure/source/Helper';\n"),
			derivedImporter, derivedOptions()),
		"import { Helper } from '@structure/source/Helper';\n")
}

// A written alias is kept beside the derived ones, and wins a tie on the same directory.
func TestImportRequirePathAliasPrefersAWrittenAliasOnTheSameDirectory(t *testing.T) {
	t.Parallel()

	result := rule_testing.RunTypedFilesWithOptionsFor(t, ImportRequirePathAlias, structureFixture(
		"import { Helper } from '../../../libraries/structure/source/Helper';\n"),
		derivedImporter, derivedOptions(PathAlias{Directory: "libraries/structure", Alias: "@written"}))
	rule_testing.ExpectFindings(t, result, "useAlias")

	if text := descriptions(result)[0]; !strings.Contains(text, "'@written/source/Helper'") {
		t.Errorf("the derived alias beat the written one on the same directory: %s", text)
	}
}

/*
 * The derivation is asked for, never assumed.
 *
 * A config that names no aliases meant "nothing to suggest" before the option existed, and a consumer
 * outside the house may be relying on that while it writes its list. The same project and the same
 * import, with only the option unset, must stay silent, beside a firing control so the silence is the
 * option's and not the fixture's.
 */
func TestImportRequirePathAliasDerivesNothingUnlessAsked(t *testing.T) {
	t.Parallel()

	fixture := structureFixture("import { Helper } from '../../../libraries/structure/source/Helper';\n")

	rule_testing.ExpectFindings(t,
		rule_testing.RunTypedFilesWithOptionsFor(t, ImportRequirePathAlias, fixture, derivedImporter, derivedOptions()),
		"useAlias")

	rule_testing.ExpectClean(t,
		rule_testing.RunTypedFilesWithOptionsFor(t, ImportRequirePathAlias, fixture, derivedImporter,
			func(directory string) any {
				return ImportRequirePathAliasOptions{RepositoryRoot: directory}
			}))
}

/*
 * Only a wildcard mapped onto exactly one wildcard target inside the repository names a directory.
 *
 * Each left-out shape is paired with a control in the same tsconfig that does fire, `@kept`, so a
 * derivation that dropped everything would fail the control rather than pass every silent case.
 */
func TestImportRequirePathAliasLeavesOutPathsThatNameNoDirectory(t *testing.T) {
	t.Parallel()

	const base = `{
		"compilerOptions": {
			"strict": true,
			"target": "ES2022",
			"lib": ["ES2022"],
			"moduleDetection": "force",
			"types": [],
			"paths": {
				"@kept/*": ["../../kept/*"],
				"@settings": ["../../settings/Settings.ts"],
				"@fallback/*": ["../../first/*", "../../second/*"],
				"legacy-*": ["../../legacy/*"]
			}
		},
		"include": ["../../**/*.ts"]
	}`
	fixture := func(importerSource string) map[string]string {
		return map[string]string{
			"tsconfig.json": structureProjectTsconfig,
			"libraries/structure/TypeScriptConfiguration.json": base,
			derivedImporter: importerSource,
		}
	}

	result := rule_testing.RunTypedFilesWithOptionsFor(t, ImportRequirePathAlias, fixture(
		"import { Kept } from '../../../kept/Kept';\n"),
		derivedImporter, derivedOptions())
	rule_testing.ExpectFindings(t, result, "useAlias")
	if text := descriptions(result)[0]; !strings.Contains(text, "'@kept/Kept'") {
		t.Errorf("the control suggests the wrong alias: %s", text)
	}

	for _, testCase := range []struct {
		name   string
		source string
	}{
		// An exact key names one file, so nothing else lives under it.
		{"an exact key", "import { Settings } from '../../../settings/Settings';\n"},
		// Both targets of a fallback list: the alias reaches the second only while the first lacks the file.
		{"the first of a fallback list", "import { First } from '../../../first/First';\n"},
		{"the second of a fallback list", "import { Second } from '../../../second/Second';\n"},
		// A wildcard that is not a trailing `/*` maps no directory onto a prefix.
		{"a wildcard that is not a directory", "import { Legacy } from '../../../legacy/Legacy';\n"},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.RunTypedFilesWithOptionsFor(t, ImportRequirePathAlias,
				fixture(testCase.source), derivedImporter, derivedOptions()))
		})
	}
}

// A target outside the repository is not one of its directories, so it is left out at derivation.
//
// Left in, it would not merely be wrong about one directory: an outside path has no repository-relative
// spelling, the empty string is the root's, and the alias would claim every file in the tree.
func TestAliasesFromCompilerPathsLeavesOutATargetOutsideTheRepository(t *testing.T) {
	t.Parallel()

	result := rule_testing.RunTypedFilesWithOptionsFor(t, ImportRequirePathAlias, map[string]string{
		"tsconfig.json": structureProjectTsconfig,
		"libraries/structure/TypeScriptConfiguration.json": strings.Replace(structurePathsBase,
			`"@project/*": ["../../*"]`, `"@outside/*": ["../../../*"]`, 1),
		derivedImporter: "import { Shared } from '../../shared/Shared';\n",
	}, derivedImporter, func(directory string) any {
		// Rooted one level down, at app/, so `@outside` lands above the root and `@project` is gone.
		return ImportRequirePathAliasOptions{RepositoryRoot: directory + "/app", AliasesFromTsconfigPaths: true}
	})
	rule_testing.ExpectClean(t, result)
}

// Without a program there are no paths to read, and the rule declines rather than crash.
func TestImportRequirePathAliasDerivesNothingWithoutAProgram(t *testing.T) {
	t.Parallel()

	rule_testing.ExpectClean(t, rule_testing.RunWithOptions(t, ImportRequirePathAlias, aliasImporter,
		"import { Thing } from '../../foundation/Thing';\n",
		ImportRequirePathAliasOptions{RepositoryRoot: "/repository", AliasesFromTsconfigPaths: true}))
}
