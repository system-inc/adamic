package nexus

import (
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

func TestImportRequireNodeNamespaceFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		source  string
		wantIds []string
	}{
		{
			name:    "default import of a built-in",
			source:  "import fs from 'node:fs';\n",
			wantIds: []string{"requireNamespaceImport"},
		},
		{
			name:    "named import of a built-in",
			source:  "import { readFileSync } from 'node:fs';\n",
			wantIds: []string{"requireNamespaceImport"},
		},
		{
			name:    "namespace import with a lowercase alias",
			source:  "import * as fs from 'node:fs';\n",
			wantIds: []string{"requireCorrectAlias"},
		},
		{
			name:    "namespace import missing the node prefix",
			source:  "import * as NodeFileSystem from 'fs';\n",
			wantIds: []string{"requireNodePrefix"},
		},
		{
			// Two independent defects on one line. Both must be reported: returning after the
			// first would hide the second until the first was fixed and the gate run again.
			name:    "missing prefix and wrong alias together",
			source:  "import * as fs from 'fs';\n",
			wantIds: []string{"requireNodePrefix", "requireCorrectAlias"},
		},
		{
			name:    "underscored module name expands to PascalCase",
			source:  "import * as cp from 'node:child_process';\n",
			wantIds: []string{"requireCorrectAlias"},
		},
		{
			name:    "subpath export has its own alias",
			source:  "import * as fsp from 'node:fs/promises';\n",
			wantIds: []string{"requireCorrectAlias"},
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, ImportRequireNodeNamespace, "probe.ts", testCase.source)
			rule_testing.ExpectFindings(t, result, testCase.wantIds...)
		})
	}
}

func TestImportRequireNodeNamespaceStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		source string
	}{
		{
			name:   "canonical namespace import",
			source: "import * as NodeFileSystem from 'node:fs';\n",
		},
		{
			name:   "expanded abbreviation",
			source: "import * as NodeOperatingSystem from 'node:os';\n",
		},
		{
			name:   "underscored module, correct alias",
			source: "import * as NodeChildProcess from 'node:child_process';\n",
		},
		{
			name:   "subpath export, correct alias",
			source: "import * as NodeFileSystemPromises from 'node:fs/promises';\n",
		},
		{
			// The rule must discriminate, not merely detect. A package that happens to be named
			// like a built-in is not one, and a default import of it is ordinary code.
			name:   "a third-party package is not a built-in",
			source: "import react from 'react';\n",
		},
		{
			name:   "a relative import is not a built-in",
			source: "import { thing } from './local-module';\n",
		},
		{
			// "path-browserify" starts with "path" but is not "path".
			name:   "a package whose name begins with a built-in",
			source: "import pathBrowserify from 'path-browserify';\n",
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, ImportRequireNodeNamespace, "probe.ts", testCase.source)
			rule_testing.ExpectClean(t, result)
		})
	}
}

func TestExpectedAlias(t *testing.T) {
	t.Parallel()

	cases := map[string]string{
		"node:fs":             "NodeFileSystem",
		"fs":                  "NodeFileSystem",
		"node:path":           "NodePath",
		"node:child_process":  "NodeChildProcess",
		"node:os":             "NodeOperatingSystem",
		"node:v8":             "NodeV8",
		"node:fs/promises":    "NodeFileSystemPromises",
		"node:worker_threads": "NodeWorkerThreads",
	}
	for source, want := range cases {
		if got := expectedAlias(source); got != want {
			t.Errorf("expectedAlias(%q) = %q, want %q", source, got, want)
		}
	}
}

// The fix rewrites the module specifier, and until now nothing checked what it wrote.
//
// This rule was the first one ported and it carries an autofix, but every fixture asserted message
// ids alone. Corrupting the replacement text so the fix produced `'CORRUPTED:fs'` instead of
// `'node:fs'` left the whole suite green: a rewrite that breaks every import it touches, invisible
// to the tests. A fix is the one part of a rule that changes source, so it is the one part where an
// id assertion proves the least.
func TestTheNodePrefixFixWritesTheSpecifierItPromises(t *testing.T) {
	t.Parallel()

	result := rule_testing.Run(t, ImportRequireNodeNamespace, "app/Probe.ts",
		"import * as NodeFileSystem from 'fs';\n")

	rule_testing.ExpectFindings(t, result, "requireNodePrefix")
	rule_testing.ExpectFixedSource(t, result,
		"import * as NodeFileSystem from 'node:fs';\n")
}

// TestImportRequireNodeNamespaceNamesTheModuleAndPointsWhereTheOriginalDoes asserts each arm's whole
// message and span against the TypeScript original, ImportRequireNodeNamespaceRule.ts.
//
// Three defects, found on ahra by comparing with ESLint. The namespace message carried a hardcoded
// example, "import * as NodeFileSystem from 'node:fs'", so `import NodePath from 'node:path'` at
// `AhraOsMindLaunch.ts:17` was told to import the wrong module under the wrong alias; the original
// interpolates the module and its expected alias. The alias message dropped the expected and actual
// names the original gives ("must be 'NodeFileSystemPromises', got 'NodeFileSystem'"). And the alias
// finding pointed at the alias name where the original reports the whole `* as alias` specifier,
// column 13 against 8, while the prefix finding pointed at the specifier where the original reports
// the declaration.
func TestImportRequireNodeNamespaceNamesTheModuleAndPointsWhereTheOriginalDoes(t *testing.T) {
	t.Parallel()

	for _, testCase := range []struct {
		name     string
		source   string
		wantId   string
		wantText string
		wantSpan string
	}{
		{"a default import names its own module and alias",
			"import NodePath from 'node:path';\n",
			"requireNamespaceImport",
			"Node built-in 'node:path' must use a namespace import, so every call site says which module " +
				"it came from. Use: import * as NodePath from 'node:path'",
			"import NodePath from 'node:path';"},
		{"an alias names what it should be and what it is",
			"import * as NodeFileSystem from 'node:fs/promises';\n",
			"requireCorrectAlias",
			"Node namespace alias must be 'NodeFileSystemPromises', got 'NodeFileSystem'. The alias is the " +
				"Node-prefixed expansion of the module, so a reader forty lines down knows what they are " +
				"looking at without finding the import. Use: import * as NodeFileSystemPromises from " +
				"'node:fs/promises'",
			"* as NodeFileSystem"},
		{"a missing prefix names the module",
			"import * as NodeChildProcess from 'child_process';\n",
			"requireNodePrefix",
			"Node built-in 'child_process' must use the 'node:' prefix, which says the module is Node's " +
				"rather than a package that happens to share its name. Use: import * as NodeChildProcess " +
				"from 'node:child_process'",
			"import * as NodeChildProcess from 'child_process';"},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			result := rule_testing.Run(t, ImportRequireNodeNamespace, "probe.ts", testCase.source)
			if len(result.Diagnostics) != 1 {
				t.Fatalf("want 1 finding, got %d", len(result.Diagnostics))
			}
			diagnostic := result.Diagnostics[0]
			if diagnostic.Message.Id != testCase.wantId {
				t.Errorf("message id is %q, want %q", diagnostic.Message.Id, testCase.wantId)
			}
			if diagnostic.Message.Description != testCase.wantText {
				t.Errorf("description is\n%q\nwant\n%q", diagnostic.Message.Description, testCase.wantText)
			}
			if span := testCase.source[diagnostic.Range.Pos():diagnostic.Range.End()]; span != testCase.wantSpan {
				t.Errorf("the finding covers %q, want %q", span, testCase.wantSpan)
			}
		})
	}
}
