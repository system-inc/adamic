package nexus

import (
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

var pathAliasOptions = ImportRequirePathAliasOptions{
	RepositoryRoot: "/repository",
	Aliases: []PathAlias{
		{Directory: "libraries/base/source", Alias: "@base/source"},
		{Directory: "libraries/base", Alias: "@base"},
		{Directory: "source", Alias: "@project/source"},
	},
	StrictRoots: []string{"libraries/base"},
}

const aliasImporter = "/repository/source/features/orders/Thing.ts"

func TestImportRequirePathAliasFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		fileName   string
		sourceText string
		wantId     string
	}{
		{"climbs two levels", aliasImporter, "import { Thing } from '../../foundation/Thing';\n", "useAlias"},
		{"climbs three levels", aliasImporter, "import { Thing } from '../../../source/foundation/Thing';\n", "useAlias"},
		{"dynamic import climbs", aliasImporter, "const thing = import('../../foundation/Thing');\n", "useAlias"},
		{"require climbs", aliasImporter, "const thing = require('../../foundation/Thing');\n", "useAlias"},
		// In a strict root even a sibling must be aliased.
		{"sibling inside a strict root", "/repository/libraries/base/source/Thing.ts", "import { Other } from './Other';\n", "useAliasInStrictRoot"},
		{"parent inside a strict root", "/repository/libraries/base/source/deep/Thing.ts", "import { Other } from '../Other';\n", "useAliasInStrictRoot"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunWithOptions(t, ImportRequirePathAlias, testCase.fileName, testCase.sourceText, pathAliasOptions)
			rule_testing.ExpectFindings(t, result, testCase.wantId)
		})
	}
}

func TestImportRequirePathAliasStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		fileName   string
		sourceText string
		options    any
	}{
		// One level stays relative on purpose: those are facts the reader already has.
		{"sibling import", aliasImporter, "import { Sibling } from './Sibling';\n", pathAliasOptions},
		{"parent import", aliasImporter, "import { Parent } from '../Parent';\n", pathAliasOptions},
		{"already aliased", aliasImporter, "import { Thing } from '@base/source/Thing';\n", pathAliasOptions},
		{"a package import", aliasImporter, "import * as NodePath from 'node:path';\n", pathAliasOptions},
		// Reporting here would demand exactly what boundary-no-internal-import rejects.
		{"an internal path must stay relative", aliasImporter, "import { Detail } from '../../internal/Detail';\n", pathAliasOptions},
		// A path leaving the repository is not ours to name.
		{"climbs out of the repository", "/repository/Thing.ts", "import { Thing } from '../../elsewhere/Thing';\n", pathAliasOptions},
		// No alias covers the destination, so there is nothing to suggest.
		{"no alias covers it", aliasImporter, "import { Thing } from '../../../other/Thing';\n", pathAliasOptions},
		{"a computed dynamic import", aliasImporter, "const thing = import(somePath);\n", pathAliasOptions},
		// A misconfigured rule declines rather than guessing.
		{"no options at all", aliasImporter, "import { Thing } from '../../foundation/Thing';\n", nil},
		// These two carry a real options value with one half missing. Both stay silent even without
		// the explicit guard, because aliasForPath finds nothing with an empty alias list and an
		// empty root makes every path fall outside the repository. The guard is deliberate
		// redundancy rather than the only thing holding these up, and saying so is more useful
		// than implying a mutation would catch it.
		{"no aliases configured", aliasImporter, "import { Thing } from '../../foundation/Thing';\n", ImportRequirePathAliasOptions{RepositoryRoot: "/repository"}},
		{"no repository root", aliasImporter, "import { Thing } from '../../foundation/Thing';\n", ImportRequirePathAliasOptions{Aliases: pathAliasOptions.Aliases}},
		{"file outside the repository", "/elsewhere/Thing.ts", "import { Thing } from '../../foundation/Thing';\n", pathAliasOptions},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunWithOptions(t, ImportRequirePathAlias, testCase.fileName, testCase.sourceText, testCase.options)
			rule_testing.ExpectClean(t, result)
		})
	}
}

// Longest directory wins, so a nested root is not shadowed by its parent.
//
// The alias strings here are deliberately shaped so the two orders produce different text. With the
// natural spellings they do not: "@base" + "/source/foundation/Thing" and "@base/source" +
// "/foundation/Thing" are the same string, so a fixture using those cannot tell a correct sort from
// a reversed one. That fixture existed first and a reversed-sort mutation sailed straight past it.
func TestImportRequirePathAliasPrefersTheLongestRoot(t *testing.T) {
	t.Parallel()

	options := ImportRequirePathAliasOptions{
		RepositoryRoot: "/repository",
		Aliases: []PathAlias{
			{Directory: "libraries/base", Alias: "@base"},
			{Directory: "libraries/base/source", Alias: "@basesource"},
		},
	}
	result := rule_testing.RunWithOptions(t, ImportRequirePathAlias,
		"/repository/libraries/base/deep/nested/Thing.ts",
		"import { Thing } from '../../source/foundation/Thing';\n", options)
	if len(result.Diagnostics) == 0 {
		t.Fatalf("expected a finding, got none")
	}
	description := result.Diagnostics[0].Message.Description
	if !strings.Contains(description, "@basesource/foundation/Thing") {
		t.Fatalf("expected the longest matching alias, got: %s", description)
	}
}

// The suggestion has to be a path someone can paste, so it is asserted rather than assumed.
func TestImportRequirePathAliasSuggestsTheAliasedPath(t *testing.T) {
	t.Parallel()

	result := rule_testing.RunWithOptions(t, ImportRequirePathAlias, aliasImporter,
		"import { Thing } from '../../foundation/Thing';\n", pathAliasOptions)
	rule_testing.ExpectFindings(t, result, "useAlias")

	description := result.Diagnostics[0].Message.Description
	if !strings.Contains(description, "@project/source/foundation/Thing") {
		t.Fatalf("expected the aliased suggestion, got: %s", description)
	}
}

// TestImportRequirePathAliasAcceptsTheRepositoryRoot pins the spelling a tsconfig uses.
//
// # This was a silent defect in both implementations
//
// A tsconfig maps the repository root as `"@project/*": ["../../*"]`, so a reader configuring this
// rule writes `{Directory: ".", Alias: "@project"}` -- the only spelling that reads like what it
// means. The matcher compares against repository-relative paths like `app/Foo.ts`, which equal
// neither `.` nor anything starting with `./`, so that entry matched nothing at all. Every deep
// relative import under the root went unreported, and a tree with 238 of them read clean.
//
// Found by counting `../../` imports by hand and getting 230 where the rule reported 0, not by any
// test: a suite whose fixtures all name real subdirectories cannot see this, because the bug is in
// the one directory nobody writes as a subdirectory.
func TestImportRequirePathAliasAcceptsTheRepositoryRoot(t *testing.T) {
	t.Parallel()

	rootOptions := ImportRequirePathAliasOptions{
		RepositoryRoot: "/repository",
		Aliases: []PathAlias{
			{Directory: "libraries/base", Alias: "@base"},
			{Directory: ".", Alias: "@project"},
		},
	}

	result := rule_testing.RunWithOptions(t, ImportRequirePathAlias, "/repository/app/features/orders/Thing.ts",
		"import { Helper } from '../../shared/Helper';\n", rootOptions)
	rule_testing.ExpectFindings(t, result, "useAlias")

	description := result.Diagnostics[0].Message.Description
	if !strings.Contains(description, "@project/app/shared/Helper") {
		t.Fatalf("expected the root-aliased suggestion, got: %s", description)
	}
}

// TestImportRequirePathAliasNormalizesConfiguredDirectories accepts the spellings a reader may write.
//
// `./app` and `app/` name the same directory as `app`, and a reader has no reason to expect them to
// differ. Normalising once at configuration read is also what keeps the alias matcher and the
// strict-root matcher from drifting, since both now ask the same function.
func TestImportRequirePathAliasNormalizesConfiguredDirectories(t *testing.T) {
	t.Parallel()

	for _, spelling := range []string{"app", "./app", "app/"} {
		t.Run(spelling, func(t *testing.T) {
			t.Parallel()
			options := ImportRequirePathAliasOptions{
				RepositoryRoot: "/repository",
				Aliases:        []PathAlias{{Directory: spelling, Alias: "@app"}},
			}
			result := rule_testing.RunWithOptions(t, ImportRequirePathAlias, "/repository/app/features/orders/Thing.ts",
				"import { Helper } from '../../shared/Helper';\n", options)
			rule_testing.ExpectFindings(t, result, "useAlias")

			description := result.Diagnostics[0].Message.Description
			if !strings.Contains(description, "@app/shared/Helper") {
				t.Fatalf("expected the normalised suggestion, got: %s", description)
			}
		})
	}
}

// TestImportRequirePathAliasRootDoesNotShadowASubdirectory pins the sort.
//
// The root contains everything, so an unsorted matcher that reached it first would alias every path
// as `@project` and no path as `@base`. Aliases are ordered longest-directory-first and the root
// normalises to the empty string, which is the shortest, so it is consulted last by construction.
func TestImportRequirePathAliasRootDoesNotShadowASubdirectory(t *testing.T) {
	t.Parallel()

	rootOptions := ImportRequirePathAliasOptions{
		RepositoryRoot: "/repository",
		Aliases: []PathAlias{
			{Directory: ".", Alias: "@project"},
			{Directory: "libraries/base", Alias: "@base"},
		},
	}

	result := rule_testing.RunWithOptions(t, ImportRequirePathAlias, "/repository/app/features/orders/Thing.ts",
		"import { Thing } from '../../../libraries/base/source/Thing';\n", rootOptions)
	rule_testing.ExpectFindings(t, result, "useAlias")

	description := result.Diagnostics[0].Message.Description
	if !strings.Contains(description, "@base/source/Thing") {
		t.Fatalf("expected the longer root to win, got: %s", description)
	}
}

// TestImportRequirePathAliasFixesTheSpecifier pins what the fixer writes.
//
// The replacement covers the literal's whole token range, quotes included, so the fix has to supply
// its own. The quote character is read off the source rather than hardcoded: this codebase writes
// single quotes and Prettier would restore them either way, but a fixer that silently changed a byte
// nobody asked about is a fixer people stop trusting.
func TestImportRequirePathAliasFixesTheSpecifier(t *testing.T) {
	t.Parallel()

	rule_testing.ExpectFixedSource(t,
		rule_testing.RunWithOptions(t, ImportRequirePathAlias, aliasImporter,
			"import { Thing } from '../../foundation/Thing';\n", pathAliasOptions),
		"import { Thing } from '@project/source/foundation/Thing';\n")

	// A double-quoted specifier keeps its quotes.
	rule_testing.ExpectFixedSource(t,
		rule_testing.RunWithOptions(t, ImportRequirePathAlias, aliasImporter,
			"import { Thing } from \"../../foundation/Thing\";\n", pathAliasOptions),
		"import { Thing } from \"@project/source/foundation/Thing\";\n")

	// The root alias, which is the spelling that matched nothing before the directory fix.
	rootOptions := ImportRequirePathAliasOptions{
		RepositoryRoot: "/repository",
		Aliases:        []PathAlias{{Directory: ".", Alias: "@project"}},
	}
	rule_testing.ExpectFixedSource(t,
		rule_testing.RunWithOptions(t, ImportRequirePathAlias, "/repository/app/features/orders/Thing.ts",
			"import { Helper } from '../../shared/Helper';\n", rootOptions),
		"import { Helper } from '@project/app/shared/Helper';\n")
}
