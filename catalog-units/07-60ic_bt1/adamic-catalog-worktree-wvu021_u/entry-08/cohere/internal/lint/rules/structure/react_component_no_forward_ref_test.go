package structure

import (
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

const forwardRefFile = "/repository/source/components/Field.tsx"

func TestReactComponentNoForwardRefFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		wantIds    []string
	}{
		{
			"a named import",
			"import { forwardRef } from 'react';\nexport const Field = forwardRef(function Field() { return null; });\n",
			[]string{"noForwardRefImport", "noForwardRefCall"},
		},
		// The alias imports the same function under another name, so the imported name is what
		// decides. A rule testing the local name is silent on exactly the import someone writes to
		// get around a rule testing the local name.
		{
			"an aliased import",
			"import { forwardRef as wrap } from 'react';\nexport const Field = wrap(function Field() { return null; });\n",
			[]string{"noForwardRefImport"},
		},
		{
			"forwardRef among other named imports",
			"import { useState, forwardRef, useMemo } from 'react';\nexport const value = useState;\n",
			[]string{"noForwardRefImport"},
		},
		{
			"a React member call",
			"import React from 'react';\nexport const Field = React.forwardRef(function Field() { return null; });\n",
			[]string{"noForwardRefCall"},
		},
		{
			"a bare call with no import in the file",
			"declare function forwardRef(component: unknown): unknown;\nexport const Field = forwardRef(function Field() { return null; });\n",
			[]string{"noForwardRefCall"},
		},
		// The type arms, which are how the wrapper survives after the call is deleted.
		{
			"a ForwardRefExoticComponent annotation",
			"import React from 'react';\ndeclare const value: React.ForwardRefExoticComponent<{ id: string }>;\nexport const Field = value;\n",
			[]string{"noForwardRefCall"},
		},
		{
			"a ForwardedRef annotation",
			"import React from 'react';\nexport function Field(ref: React.ForwardedRef<HTMLInputElement>) { return ref; }\n",
			[]string{"noForwardRefCall"},
		},
		{
			"two calls report twice",
			"import React from 'react';\nexport const A = React.forwardRef(function A() { return null; });\n" +
				"export const B = React.forwardRef(function B() { return null; });\n",
			[]string{"noForwardRefCall", "noForwardRefCall"},
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t, rule_testing.Run(t, ReactComponentNoForwardRef, forwardRefFile,
				testCase.sourceText), testCase.wantIds...)
		})
	}
}

func TestReactComponentNoForwardRefStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		// The shape the rule asks for: ref as an ordinary property, which is what React 19 does.
		{
			"a component taking ref as a property",
			"import type { Ref } from 'react';\n" +
				"export function Field(properties: { ref?: Ref<HTMLInputElement>; id: string }) { return properties.id; }\n",
		},
		// The module gate. Identical import shape from somewhere else, so this fails if the source
		// test is dropped and the rule starts flagging any forwardRef-shaped import.
		{
			"forwardRef imported from somewhere other than react",
			"import { forwardRef } from './OurOwnHelpers.ts';\nexport const value = forwardRef;\n",
		},
		{
			"other react imports",
			"import { useState, useMemo, useRef } from 'react';\nexport const value = useState;\n",
		},
		{
			"a default react import alone",
			"import React from 'react';\nexport const value = React.useState;\n",
		},
		// A member call on something that is not React. The receiver test is what separates them.
		{
			"forwardRef on a receiver that is not React",
			"declare const Other: { forwardRef(component: unknown): unknown };\n" +
				"export const Field = Other.forwardRef(function Field() { return null; });\n",
		},
		// A qualified type whose left side is not React, which the type arm must not claim.
		{
			"a ForwardedRef on a namespace that is not React",
			"declare namespace Other { type ForwardedRef<T> = { current: T } }\n" +
				"export function Field(ref: Other.ForwardedRef<HTMLInputElement>) { return ref; }\n",
		},
		// A React type that is not one of the two named. This is the boundary a prefix test on
		// "Forward" would cross.
		{
			"an unrelated React type",
			"import React from 'react';\ndeclare const value: React.ReactNode;\nexport const Field = value;\n",
		},
		{
			"a property named forwardRef that is never called",
			"declare const config: { forwardRef: boolean };\nexport const value = config.forwardRef;\n",
		},
		{
			"a file with nothing to do with react",
			"export const value = 1;\n",
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.Run(t, ReactComponentNoForwardRef, forwardRefFile, testCase.sourceText))
		})
	}
}
