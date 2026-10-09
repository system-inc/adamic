package nexus

import (
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

func TestConsistencyNoEnumFires(t *testing.T) {
	t.Parallel()

	cases := map[string]string{
		"a plain enum":            "enum Color { Red, Green }\n",
		"an exported enum":        "export enum Color { Red, Green }\n",
		"a const enum":            "const enum Color { Red, Green }\n",
		"a string-valued enum":    "enum Color { Red = 'RED' }\n",
		"an enum inside a module": "namespace Theme { export enum Color { Red } }\n",
	}
	for name, source := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, ConsistencyNoEnum, "probe.ts", source)
			rule_testing.ExpectFindings(t, result, "noEnum")
		})
	}
}

func TestConsistencyNoEnumStaysSilent(t *testing.T) {
	t.Parallel()

	cases := map[string]string{
		"the as-const replacement": "export const ColorKind = { Red: 'Red' } as const;\n" +
			"export type ColorKindType = (typeof ColorKind)[keyof typeof ColorKind];\n",
		"a union type alias":   "type Color = 'Red' | 'Green';\n",
		"an ordinary object":   "const Color = { Red: 'Red' };\n",
		"the word in a string": "const message = 'enum Color is banned';\n",
		"an interface":         "interface Color { red: string }\n",
	}
	for name, source := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, ConsistencyNoEnum, "probe.ts", source)
			rule_testing.ExpectClean(t, result)
		})
	}
}
