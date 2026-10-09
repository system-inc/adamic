package structure

import (
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

const reactImportFile = "/repository/source/components/Field.tsx"

func TestReactImportNoDestructuringFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		wantIds    []string
	}{
		// One named import with no default produces both import findings: the specifier is wrong,
		// and there is no namespace to reach through.
		{
			"a named import with no default",
			"import { useState } from 'react';\nexport const value = 1;\n",
			[]string{"noNamedImport", "noDestructuringFromReact"},
		},
		// With a React default present the namespace exists, so only the specifier is at fault.
		{
			"a named import alongside the React default",
			"import React, { useState } from 'react';\nexport const value = React.version;\n",
			[]string{"noNamedImport"},
		},
		{
			"three named imports report three times",
			"import React, { useState, useMemo, useRef } from 'react';\nexport const value = React.version;\n",
			[]string{"noNamedImport", "noNamedImport", "noNamedImport"},
		},
		// The use arms. An import plus a use is two findings, which is the point: they are two
		// edits and a reader needs to see both.
		{
			"a destructured hook that is then called",
			"import React, { useState } from 'react';\nexport function Field() { return useState(0); }\n",
			[]string{"noNamedImport", "noCallWithoutPrefix"},
		},
		{
			"a destructured type that is then used",
			"import React, { type Ref } from 'react';\nexport function Field(reference: Ref<HTMLElement>) { return reference; }\n",
			[]string{"noNamedImport", "noTypeReferenceWithoutPrefix"},
		},
		{
			"a hook called twice reports at each call",
			"import React, { useState } from 'react';\n" +
				"export function Field() { const a = useState(0); const b = useState(1); return [a, b]; }\n",
			[]string{"noNamedImport", "noCallWithoutPrefix", "noCallWithoutPrefix"},
		},
		// The alias is what a later use is written as, so the tracking set has to hold the local
		// name rather than the imported one.
		{
			"an aliased hook is tracked by its local name",
			"import React, { useState as useValue } from 'react';\nexport function Field() { return useValue(0); }\n",
			[]string{"noNamedImport", "noCallWithoutPrefix"},
		},
		{
			"a type-only import declaration",
			"import type { Ref } from 'react';\nexport function Field(reference: Ref<HTMLElement>) { return reference; }\n",
			[]string{"noNamedImport", "noDestructuringFromReact", "noTypeReferenceWithoutPrefix"},
		},
		// A default import that is not named React does not provide the namespace the rule asks
		// for, so the destructuring finding still stands.
		// The call arm is hooks only, so a destructured non-hook produces the import finding and
		// nothing more. Belongs here rather than in the silent table: the import is a real
		// violation, and the claim being tested is that the call arm adds nothing to it.
		{
			"a destructured non-hook that is called",
			"import React, { createElement } from 'react';\nexport function Field() { return createElement('div'); }\n",
			[]string{"noNamedImport"},
		},
		{
			"a default import not named React",
			"import Reakt, { useState } from 'react';\nexport const value = Reakt.version;\n",
			[]string{"noNamedImport", "noDestructuringFromReact"},
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t, rule_testing.Run(t, ImportRequireReactNamespace, reactImportFile,
				testCase.sourceText), testCase.wantIds...)
		})
	}
}

func TestReactImportNoDestructuringStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		// The shape the rule asks for.
		{
			"the React default alone, used through the namespace",
			"import React from 'react';\nexport function Field() { return React.useState(0); }\n",
		},
		// A namespace import provides the namespace and destructures nothing.
		{
			"a namespace import",
			"import * as React from 'react';\nexport function Field() { return React.useState(0); }\n",
		},
		// The module gate. Identical shape from elsewhere, so this fails if the source test goes.
		{
			"a named import from somewhere other than react",
			"import { useState } from './OurHooks.ts';\nexport function Field() { return useState(0); }\n",
		},
		{
			"a named import from react-dom",
			"import { createPortal } from 'react-dom';\nexport const value = createPortal;\n",
		},
		// A bare hook call with no react import is somebody's own function.
		{
			"a hook-shaped call that was never imported from react",
			"declare function useThing(): number;\nexport function Field() { return useThing(); }\n",
		},
		{
			"a type reference that was never imported from react",
			"type Ref<T> = { current: T };\nexport function Field(reference: Ref<HTMLElement>) { return reference; }\n",
		},
		// A qualified type name is already namespaced, which is the thing being asked for.
		{
			"a qualified type reference",
			"import React from 'react';\nexport function Field(reference: React.Ref<HTMLElement>) { return reference; }\n",
		},
		{
			"a file with no react at all",
			"export const value = 1;\n",
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.Run(t, ImportRequireReactNamespace, reactImportFile, testCase.sourceText))
		})
	}
}

// The use arms depend on the import having been seen, and this pins the direction.
//
// Imports sit at the top of a file in practice, so the ordering is not something anyone hits. It is
// tested because the rule reads a set that an earlier listener fills, which makes the behavior
// depend on traversal order rather than on any single node, and a change to how listeners are
// dispatched would move it silently.
func TestReactImportNoDestructuringDependsOnImportOrder(t *testing.T) {
	t.Parallel()

	// The ordinary case: the import is above the use, and both are reported.
	rule_testing.ExpectFindings(t, rule_testing.Run(t, ImportRequireReactNamespace, reactImportFile,
		"import React, { useState } from 'react';\nexport function Field() { return useState(0); }\n"),
		"noNamedImport", "noCallWithoutPrefix")

	// The import below the use. The specifier is still reported, since that listener does not
	// depend on anything earlier, but the call is not, because the set was empty when it was
	// visited. Recorded as the rule's behavior rather than defended as correct: the shape is not
	// one anyone writes, and matching the original matters more than improving on it here.
	rule_testing.ExpectFindings(t, rule_testing.Run(t, ImportRequireReactNamespace, reactImportFile,
		"export function Field() { return useState(0); }\nimport React, { useState } from 'react';\n"),
		"noNamedImport")
}
