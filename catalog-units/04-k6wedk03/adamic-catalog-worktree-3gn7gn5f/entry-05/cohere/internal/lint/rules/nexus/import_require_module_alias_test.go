package nexus

import (
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/system-inc/cohere/internal/lint/testing"
)

const moduleAliasFile = "/repository/source/Thing.tsx"

func TestImportRequireModuleAliasFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		wantIds    []string
	}{
		{
			"react bound to the wrong name",
			"import Reakt from 'react';\n",
			[]string{"requireAliasName"},
		},
		{
			"react as a namespace when the default form is required",
			"import * as React from 'react';\n",
			[]string{"requireDefaultStyle"},
		},
		{
			// Style is reported rather than name, because a namespace import bound to the right
			// name is still the wrong shape and renaming it would fix nothing.
			"react as a namespace under a wrong name still reports the style",
			"import * as Reakt from 'react';\n",
			[]string{"requireDefaultStyle"},
		},
		{
			"typescript bound to the wrong name",
			"import * as ts from 'typescript';\n",
			[]string{"requireAliasName"},
		},
		{
			"typescript as a default when the namespace form is required",
			"import TypeScript from 'typescript';\n",
			[]string{"requireNamespaceStyle"},
		},
		{
			// The default binding and the named ones live in one clause. The named half is not this
			// rule's question, so exactly one finding is correct here; two would mean the named
			// filter had come off.
			"a wrong default name beside named imports reports only the default",
			"import Reakt, { useState } from 'react';\n",
			[]string{"requireAliasName"},
		},
		{
			// A type-only import binds a name in exactly the same way, so it is pinned the same way.
			"a type-only import is pinned too",
			"import type Reakt from 'react';\n",
			[]string{"requireAliasName"},
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, ImportRequireModuleAlias, moduleAliasFile, testCase.sourceText)
			rule_testing.ExpectFindings(t, result, testCase.wantIds...)
		})
	}
}

// A finding has to land on the line the author can suppress.
//
// A node's Pos() includes leading trivia, so anchoring on the enclosing declaration reports at the
// first comment above the import rather than at the import itself. A `-next-line` directive can only
// match the line after itself, so such a finding is unreachable by any suppression that could be
// written while still reading as a real finding in every count. This rule reports the binding, which
// is both on the import line and the text a reader has to change.
func TestImportRequireModuleAliasReportsAtTheBinding(t *testing.T) {
	t.Parallel()

	sourceText := "// Dependencies - React\n" +
		"// The binding below is deliberately misnamed\n" +
		"// eslint-disable-next-line nexus/import-require-module-alias\n" +
		"import Reakt from 'react';\n"

	result := rule_testing.Run(t, ImportRequireModuleAlias, moduleAliasFile, sourceText)
	rule_testing.ExpectFindings(t, result, "requireAliasName")

	line, _ := scanner.GetECMALineAndByteOffsetOfPosition(result.SourceFile, result.Diagnostics[0].Range.Pos())
	const importLine = 3 // zero-based, so the fourth line
	if line != importLine {
		t.Fatalf("expected the finding on the import line (%d), got line %d", importLine+1, line+1)
	}
}

// The rule reports and never rewrites. See the doc comment for why: every fix the original offers
// renames a binding without following its references, and the corrupted output still parses, so the
// edit engine's parse guard cannot catch it.
func TestImportRequireModuleAliasProposesNoFix(t *testing.T) {
	t.Parallel()

	result := rule_testing.Run(t, ImportRequireModuleAlias, moduleAliasFile, "import Reakt from 'react';\n")
	rule_testing.ExpectFindings(t, result, "requireAliasName")

	if len(result.Diagnostics[0].Fixes) != 0 {
		t.Fatalf("expected no fixes, got %d", len(result.Diagnostics[0].Fixes))
	}
	if len(result.Diagnostics[0].Suggestions) != 0 {
		t.Fatalf("expected no suggestions, got %d", len(result.Diagnostics[0].Suggestions))
	}
}

func TestImportRequireModuleAliasStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"react under its pinned name and form", "import React from 'react';\n"},
		{"a type-only react import under its pinned name", "import type React from 'react';\n"},
		{"typescript under its pinned name and form", "import * as TypeScript from 'typescript';\n"},
		{"the pinned default beside named imports", "import React, { useState } from 'react';\n"},
		// Named imports are a different question and this rule has no opinion until a default or
		// namespace binding exists. Without this exemption every destructured React import in the
		// tree would report.
		{"named imports alone from a pinned package", "import { useState, useEffect } from 'react';\n"},
		{"a named type import alone from a pinned package", "import type { ReactNode } from 'react';\n"},
		// No clause at all, so no binding to pin.
		{"a side effect import of a pinned package", "import 'react';\n"},
		// An unconfigured package is not this rule's business, whatever it is bound to.
		{"an unpinned package bound to anything", "import whatever from 'lodash';\n"},
		{"an unpinned package as a namespace", "import * as whatever from 'lodash';\n"},
		// A prefix or substring match would flag these. The map lookup is exact, because
		// 'react-dom' is a different package than 'react'.
		{"a package whose name extends a pinned one", "import ReactDom from 'react-dom';\n"},
		{"a scoped package containing a pinned name", "import { thing } from '@types/react';\n"},
		{"an ordinary relative import", "import { Thing } from './Thing';\n"},
		{"no imports at all", "export const Value = 1;\n"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, ImportRequireModuleAlias, moduleAliasFile, testCase.sourceText)
			rule_testing.ExpectClean(t, result)
		})
	}
}

// A JSX element name is a plain KindIdentifier in typescript-go, where the TypeScript parser gives
// it a distinct node type. A rule that walks identifiers therefore inherits an exemption the
// original got from its AST for free. This rule keys on ImportDeclaration and so never sees them,
// and this fixture is what keeps that true if the listener set ever widens.
func TestImportRequireModuleAliasIgnoresJsxAndOtherIdentifiers(t *testing.T) {
	t.Parallel()

	sourceText := "import React from 'react';\n" +
		"const Reakt = 1;\n" +
		"export function Component() {\n" +
		"    return <Reakt.Thing ts={Reakt} />;\n" +
		"}\n"

	result := rule_testing.Run(t, ImportRequireModuleAlias, moduleAliasFile, sourceText)
	rule_testing.ExpectClean(t, result)
}

func TestImportRequireModuleAliasReadsOptions(t *testing.T) {
	t.Parallel()

	options := ImportRequireModuleAliasOptions{
		Modules: map[string]ModuleAlias{
			// A package the defaults say nothing about.
			"lodash": {Name: "Lodash"},
			// A default re-stated, which must win over the built-in entry.
			"react": {Name: "React", Style: ImportStyleNamespace},
		},
	}

	t.Run("an added package is pinned", func(t *testing.T) {
		t.Parallel()
		result := rule_testing.RunWithOptions(
			t, ImportRequireModuleAlias, moduleAliasFile, "import lodash from 'lodash';\n", options,
		)
		rule_testing.ExpectFindings(t, result, "requireAliasName")
	})

	t.Run("an added package under its pinned name is clean", func(t *testing.T) {
		t.Parallel()
		result := rule_testing.RunWithOptions(
			t, ImportRequireModuleAlias, moduleAliasFile, "import Lodash from 'lodash';\n", options,
		)
		rule_testing.ExpectClean(t, result)
	})

	// An omitted style means Default, so the shorthand pins the form as well as the name. If the
	// style were left genuinely unset this case would pass silently, which is the drift the default
	// exists to stop.
	t.Run("an added package with no style still requires the default form", func(t *testing.T) {
		t.Parallel()
		result := rule_testing.RunWithOptions(
			t, ImportRequireModuleAlias, moduleAliasFile, "import * as Lodash from 'lodash';\n", options,
		)
		rule_testing.ExpectFindings(t, result, "requireDefaultStyle")
	})

	// The override direction is the half a merge in the wrong order would break: react now wants the
	// namespace form, so the built-in default entry must lose.
	t.Run("a re-stated default is overridden", func(t *testing.T) {
		t.Parallel()
		result := rule_testing.RunWithOptions(
			t, ImportRequireModuleAlias, moduleAliasFile, "import React from 'react';\n", options,
		)
		rule_testing.ExpectFindings(t, result, "requireNamespaceStyle")
	})

	t.Run("the overridden form is clean", func(t *testing.T) {
		t.Parallel()
		result := rule_testing.RunWithOptions(
			t, ImportRequireModuleAlias, moduleAliasFile, "import * as React from 'react';\n", options,
		)
		rule_testing.ExpectClean(t, result)
	})

	// The rule guards the built-in packages with or without configuration, so unlike a rule that
	// declines every file without its options, it is not Required in the registry. This is the
	// fixture that says so.
	t.Run("the built-in defaults apply with no options at all", func(t *testing.T) {
		t.Parallel()
		result := rule_testing.Run(t, ImportRequireModuleAlias, moduleAliasFile, "import ts from 'typescript';\n")
		rule_testing.ExpectFindings(t, result, "requireNamespaceStyle")
	})
}
