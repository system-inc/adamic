package typescript

import (
	"strconv"
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// noUnsafeAssignmentFileFor picks the extension, because six of upstream's cases are JSX.
//
// `rule_testing` selects the script kind from the filename, so a `.tsx` case handed a `.ts` name
// parses as TypeScript, finds no JSX element, and the rule correctly reports nothing. That reads as
// a rule that cannot see rather than as a fixture named wrong.
func noUnsafeAssignmentFileFor(isJsx bool) string {
	if isJsx {
		return "/repository/source/Thing.tsx"
	}
	return "/repository/source/Thing.ts"
}

// noUnsafeAssignmentCaseName numbers a row so a failure names which one.
func noUnsafeAssignmentCaseName(index int) string {
	return "case" + strconv.Itoa(index)
}

// noUnsafeAssignmentOnDisk is what the harness actually writes for a fixture.
//
// `RunTyped` writes each file as `strings.TrimSpace(contents)+"\n"`, so a case copied from an
// upstream tester carries a leading newline the file on disk does not have, and a span sliced from
// the Go literal is one byte off.
func noUnsafeAssignmentOnDisk(sourceText string) string {
	return strings.TrimSpace(sourceText) + "\n"
}

// TestNoUnsafeAssignmentStaysSilentOnUpstreamPassCases is the imported clean corpus.
//
// All thirty-six of upstream's passing inputs, extracted from the clone's test file by parsing it
// with the TypeScript compiler rather than by reading it, so no escape sequence passed through a
// shell or a keyboard on the way here. Every one was additionally run through the installed 8.67.0
// build over a real program, which reported nothing on all thirty-six.
//
// These carry most of the rule's judgment. An inferred variable is silent because its type IS the
// value's, `unknown` is silent because absorbing an `any` is what `unknown` is for, and a generic
// whose argument matches is silent because nothing disagrees.
func TestNoUnsafeAssignmentStaysSilentOnUpstreamPassCases(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText string
		isJsx      bool
	}{
		{
			sourceText: "const x = 1;",
			isJsx:      false,
		},
		{
			sourceText: "const x: number = 1;",
			isJsx:      false,
		},
		{
			sourceText: "\nconst x = 1,\n  y = 1;\n    ",
			isJsx:      false,
		},
		{
			sourceText: "let x;",
			isJsx:      false,
		},
		{
			sourceText: "\nlet x = 1,\n  y;\n    ",
			isJsx:      false,
		},
		{
			sourceText: "function foo(a = 1) {}",
			isJsx:      false,
		},
		{
			sourceText: "\nclass Foo {\n  constructor(private a = 1) {}\n}\n    ",
			isJsx:      false,
		},
		{
			sourceText: "\nclass Foo {\n  private a = 1;\n}\n    ",
			isJsx:      false,
		},
		{
			sourceText: "\nclass Foo {\n  accessor a = 1;\n}\n    ",
			isJsx:      false,
		},
		{
			sourceText: "const x: Set<string> = new Set();",
			isJsx:      true,
		},
		{
			sourceText: "const x: Set<string> = new Set<string>();",
			isJsx:      true,
		},
		{
			sourceText: "const [x] = [1];",
			isJsx:      false,
		},
		{
			sourceText: "const [x, y] = [1, 2] as number[];",
			isJsx:      false,
		},
		{
			sourceText: "const [x, ...y] = [1, 2, 3, 4, 5];",
			isJsx:      false,
		},
		{
			sourceText: "const [x, ...y] = [1];",
			isJsx:      false,
		},
		{
			sourceText: "const [{ ...x }] = [{ x: 1 }] as [{ x: any }];",
			isJsx:      false,
		},
		{
			sourceText: "function foo(x = 1) {}",
			isJsx:      false,
		},
		{
			sourceText: "function foo([x] = [1]) {}",
			isJsx:      false,
		},
		{
			sourceText: "function foo([x, ...y] = [1, 2, 3, 4, 5]) {}",
			isJsx:      false,
		},
		{
			sourceText: "function foo([x, ...y] = [1]) {}",
			isJsx:      false,
		},
		{
			sourceText: "const x = new Set<any>();",
			isJsx:      true,
		},
		{
			sourceText: "const x = { y: 1 };",
			isJsx:      false,
		},
		{
			sourceText: "const x = { y = 1 };",
			isJsx:      false,
		},
		{
			sourceText: "const x = { y(){} };",
			isJsx:      false,
		},
		{
			sourceText: "const x: { y: number } = { y: 1 };",
			isJsx:      false,
		},
		{
			sourceText: "const x = [...[1, 2, 3]];",
			isJsx:      false,
		},
		{
			sourceText: "const [{ [`x${1}`]: x }] = [{ [`x`]: 1 }] as [{ [`x`]: any }];",
			isJsx:      false,
		},
		{
			sourceText: "\ntype T = [string, T[]];\nconst test: T = ['string', []] as T;\n    ",
			isJsx:      false,
		},
		{
			sourceText: "\ntype Props = { a: string };\ndeclare function Foo(props: Props): never;\n<Foo a={'foo'} />;\n      ",
			isJsx:      true,
		},
		{
			sourceText: "\ndeclare function Foo(props: { a: string }): never;\n<Foo a=\"foo\" />;\n      ",
			isJsx:      true,
		},
		{
			sourceText: "\ndeclare function Foo(props: { a: string }): never;\n<Foo a={} />;\n      ",
			isJsx:      true,
		},
		{
			sourceText: "const x: unknown = y as any;",
			isJsx:      false,
		},
		{
			sourceText: "const x: unknown[] = y as any[];",
			isJsx:      false,
		},
		{
			sourceText: "const x: Set<unknown> = y as Set<any>;",
			isJsx:      true,
		},
		{
			sourceText: "const x: Map<string, string> = new Map();",
			isJsx:      true,
		},
		{
			sourceText: "\ntype Foo = { bar: unknown };\nconst bar: any = 1;\nconst foo: Foo = { bar };\n    ",
			isJsx:      false,
		},
	}
	for index, testCase := range cases {
		t.Run(noUnsafeAssignmentCaseName(index), func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.RunTyped(t, NoUnsafeAssignment,
				noUnsafeAssignmentFileFor(testCase.isJsx), testCase.sourceText))
		})
	}
}

// TestNoUnsafeAssignmentFiresOnUpstreamFailCases is the imported reporting corpus.
//
// Fifty-six inputs carrying fifty-seven findings across six message ids. Three things are asserted
// per row and each catches a different defect. The ids say which of the seven arms ran, and the
// destructuring ids are the ones a naive port collapses into `anyAssignment`. The spans say where
// each finding points, which for a destructured element is the element rather than the statement.
// And the rendered text is asserted exactly, because two of the seven messages interpolate a type
// name and one interpolates two, and a message-id assertion cannot see any of that.
func TestNoUnsafeAssignmentFiresOnUpstreamFailCases(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText string
		isJsx      bool
		wantIds    []string
		wantSpans  []string
		wantTexts  []string
	}{
		{
			sourceText: "const x = 1 as any;",
			isJsx:      false,
			wantIds:    []string{"anyAssignment"},
			wantSpans:  []string{"x = 1 as any"},
			wantTexts:  []string{"Unsafe assignment of an `any` value."},
		},
		{
			sourceText: "\nconst x = 1 as any,\n  y = 1;\n      ",
			isJsx:      false,
			wantIds:    []string{"anyAssignment"},
			wantSpans:  []string{"x = 1 as any"},
			wantTexts:  []string{"Unsafe assignment of an `any` value."},
		},
		{
			sourceText: "function foo(a = 1 as any) {}",
			isJsx:      false,
			wantIds:    []string{"anyAssignment"},
			wantSpans:  []string{"a = 1 as any"},
			wantTexts:  []string{"Unsafe assignment of an `any` value."},
		},
		{
			sourceText: "\nclass Foo {\n  constructor(private a = 1 as any) {}\n}\n      ",
			isJsx:      false,
			wantIds:    []string{"anyAssignment"},
			wantSpans:  []string{"a = 1 as any"},
			wantTexts:  []string{"Unsafe assignment of an `any` value."},
		},
		{
			sourceText: "\nclass Foo {\n  private a = 1 as any;\n}\n      ",
			isJsx:      false,
			wantIds:    []string{"anyAssignment"},
			wantSpans:  []string{"private a = 1 as any;"},
			wantTexts:  []string{"Unsafe assignment of an `any` value."},
		},
		{
			sourceText: "\nclass Foo {\n  accessor a = 1 as any;\n}\n      ",
			isJsx:      false,
			wantIds:    []string{"anyAssignment"},
			wantSpans:  []string{"accessor a = 1 as any;"},
			wantTexts:  []string{"Unsafe assignment of an `any` value."},
		},
		{
			sourceText: "\nconst [x] = spooky;\n      ",
			isJsx:      false,
			wantIds:    []string{"anyAssignment"},
			wantSpans:  []string{"[x] = spooky"},
			wantTexts:  []string{"Unsafe assignment of an error typed value."},
		},
		{
			sourceText: "\nconst [[[x]]] = [spooky];\n      ",
			isJsx:      false,
			wantIds:    []string{"unsafeArrayPatternFromTuple"},
			wantSpans:  []string{"[[x]]"},
			wantTexts:  []string{"Unsafe array destructuring of a tuple element with an error typed value."},
		},
		{
			sourceText: "\nconst {\n  x: { y: z },\n} = { x: spooky };\n      ",
			isJsx:      false,
			wantIds:    []string{"unsafeObjectPattern", "anyAssignment"},
			wantSpans:  []string{"{ y: z }", "x: spooky"},
			wantTexts:  []string{"Unsafe object destructuring of a property with an error typed value.", "Unsafe assignment of an error typed value."},
		},
		{
			sourceText: "\nlet value: number;\n\nvalue = spooky;\n      ",
			isJsx:      false,
			wantIds:    []string{"anyAssignment"},
			wantSpans:  []string{"value = spooky"},
			wantTexts:  []string{"Unsafe assignment of an error typed value."},
		},
		{
			sourceText: "\nconst [x] = 1 as any;\n      ",
			isJsx:      false,
			wantIds:    []string{"anyAssignment"},
			wantSpans:  []string{"[x] = 1 as any"},
			wantTexts:  []string{"Unsafe assignment of an `any` value."},
		},
		{
			sourceText: "\nconst [x] = [] as any[];\n      ",
			isJsx:      false,
			wantIds:    []string{"unsafeArrayPattern"},
			wantSpans:  []string{"[x]"},
			wantTexts:  []string{"Unsafe array destructuring of an `any` array value."},
		},
		{
			sourceText: "const x: Set<string> = new Set<any>();",
			isJsx:      true,
			wantIds:    []string{"unsafeAssignment"},
			wantSpans:  []string{"x: Set<string> = new Set<any>()"},
			wantTexts:  []string{"Unsafe assignment of type `Set<any>` to a variable of type `Set<string>`."},
		},
		{
			sourceText: "const x: Map<string, string> = new Map<string, any>();",
			isJsx:      true,
			wantIds:    []string{"unsafeAssignment"},
			wantSpans:  []string{"x: Map<string, string> = new Map<string, any>()"},
			wantTexts:  []string{"Unsafe assignment of type `Map<string, any>` to a variable of type `Map<string, string>`."},
		},
		{
			sourceText: "const x: Set<string[]> = new Set<any[]>();",
			isJsx:      true,
			wantIds:    []string{"unsafeAssignment"},
			wantSpans:  []string{"x: Set<string[]> = new Set<any[]>()"},
			wantTexts:  []string{"Unsafe assignment of type `Set<any[]>` to a variable of type `Set<string[]>`."},
		},
		{
			sourceText: "const x: Set<Set<Set<string>>> = new Set<Set<Set<any>>>();",
			isJsx:      true,
			wantIds:    []string{"unsafeAssignment"},
			wantSpans:  []string{"x: Set<Set<Set<string>>> = new Set<Set<Set<any>>>()"},
			wantTexts:  []string{"Unsafe assignment of type `Set<Set<Set<any>>>` to a variable of type `Set<Set<Set<string>>>`."},
		},
		{
			sourceText: "const [x] = [1] as [any];",
			isJsx:      false,
			wantIds:    []string{"unsafeArrayPatternFromTuple"},
			wantSpans:  []string{"x"},
			wantTexts:  []string{"Unsafe array destructuring of a tuple element with an `any` value."},
		},
		{
			sourceText: "function foo([x] = [1] as [any]) {}",
			isJsx:      false,
			wantIds:    []string{"unsafeArrayPatternFromTuple"},
			wantSpans:  []string{"x"},
			wantTexts:  []string{"Unsafe array destructuring of a tuple element with an `any` value."},
		},
		{
			sourceText: "[x] = [1] as [any];",
			isJsx:      false,
			wantIds:    []string{"unsafeArrayPatternFromTuple"},
			wantSpans:  []string{"x"},
			wantTexts:  []string{"Unsafe array destructuring of a tuple element with an `any` value."},
		},
		{
			sourceText: "const [[[[x]]]] = [[[[1 as any]]]];",
			isJsx:      false,
			wantIds:    []string{"unsafeArrayPatternFromTuple"},
			wantSpans:  []string{"x"},
			wantTexts:  []string{"Unsafe array destructuring of a tuple element with an `any` value."},
		},
		{
			sourceText: "function foo([[[[x]]]] = [[[[1 as any]]]]) {}",
			isJsx:      false,
			wantIds:    []string{"unsafeArrayPatternFromTuple"},
			wantSpans:  []string{"x"},
			wantTexts:  []string{"Unsafe array destructuring of a tuple element with an `any` value."},
		},
		{
			sourceText: "[[[[x]]]] = [[[[1 as any]]]];",
			isJsx:      false,
			wantIds:    []string{"unsafeArrayPatternFromTuple"},
			wantSpans:  []string{"x"},
			wantTexts:  []string{"Unsafe array destructuring of a tuple element with an `any` value."},
		},
		{
			sourceText: "const [[[[x]]]] = [1 as any];",
			isJsx:      false,
			wantIds:    []string{"unsafeArrayPatternFromTuple"},
			wantSpans:  []string{"[[[x]]]"},
			wantTexts:  []string{"Unsafe array destructuring of a tuple element with an `any` value."},
		},
		{
			sourceText: "function foo([[[[x]]]] = [1 as any]) {}",
			isJsx:      false,
			wantIds:    []string{"unsafeArrayPatternFromTuple"},
			wantSpans:  []string{"[[[x]]]"},
			wantTexts:  []string{"Unsafe array destructuring of a tuple element with an `any` value."},
		},
		{
			sourceText: "const [{ x }] = [{ x: 1 }] as [{ x: any }];",
			isJsx:      false,
			wantIds:    []string{"unsafeObjectPattern"},
			wantSpans:  []string{"x"},
			wantTexts:  []string{"Unsafe object destructuring of a property with an `any` value."},
		},
		{
			sourceText: "function foo([{ x }] = [{ x: 1 }] as [{ x: any }]) {}",
			isJsx:      false,
			wantIds:    []string{"unsafeObjectPattern"},
			wantSpans:  []string{"x"},
			wantTexts:  []string{"Unsafe object destructuring of a property with an `any` value."},
		},
		{
			sourceText: "[{ x }] = [{ x: 1 }] as [{ x: any }];",
			isJsx:      false,
			wantIds:    []string{"unsafeObjectPattern"},
			wantSpans:  []string{"x"},
			wantTexts:  []string{"Unsafe object destructuring of a property with an `any` value."},
		},
		{
			sourceText: "const [{ ['x']: x }] = [{ ['x']: 1 }] as [{ ['x']: any }];",
			isJsx:      false,
			wantIds:    []string{"unsafeObjectPattern"},
			wantSpans:  []string{"x"},
			wantTexts:  []string{"Unsafe object destructuring of a property with an `any` value."},
		},
		{
			sourceText: "function foo([{ ['x']: x }] = [{ ['x']: 1 }] as [{ ['x']: any }]) {}",
			isJsx:      false,
			wantIds:    []string{"unsafeObjectPattern"},
			wantSpans:  []string{"x"},
			wantTexts:  []string{"Unsafe object destructuring of a property with an `any` value."},
		},
		{
			sourceText: "[{ ['x']: x }] = [{ ['x']: 1 }] as [{ ['x']: any }];",
			isJsx:      false,
			wantIds:    []string{"unsafeObjectPattern"},
			wantSpans:  []string{"x"},
			wantTexts:  []string{"Unsafe object destructuring of a property with an `any` value."},
		},
		{
			sourceText: "const [{ [`x`]: x }] = [{ [`x`]: 1 }] as [{ [`x`]: any }];",
			isJsx:      false,
			wantIds:    []string{"unsafeObjectPattern"},
			wantSpans:  []string{"x"},
			wantTexts:  []string{"Unsafe object destructuring of a property with an `any` value."},
		},
		{
			sourceText: "function foo([{ [`x`]: x }] = [{ [`x`]: 1 }] as [{ [`x`]: any }]) {}",
			isJsx:      false,
			wantIds:    []string{"unsafeObjectPattern"},
			wantSpans:  []string{"x"},
			wantTexts:  []string{"Unsafe object destructuring of a property with an `any` value."},
		},
		{
			sourceText: "[{ [`x`]: x }] = [{ [`x`]: 1 }] as [{ [`x`]: any }];",
			isJsx:      false,
			wantIds:    []string{"unsafeObjectPattern"},
			wantSpans:  []string{"x"},
			wantTexts:  []string{"Unsafe object destructuring of a property with an `any` value."},
		},
		{
			sourceText: "[[[[x]]]] = [1 as any];",
			isJsx:      false,
			wantIds:    []string{"unsafeAssignment"},
			wantSpans:  []string{"[[[[x]]]] = [1 as any]"},
			wantTexts:  []string{"Unsafe assignment of type `[any]` to a variable of type `[[[[any]]]]`."},
		},
		{
			sourceText: "\nconst x = [...(1 as any)];\n      ",
			isJsx:      false,
			wantIds:    []string{"unsafeArraySpread"},
			wantSpans:  []string{"...(1 as any)"},
			wantTexts:  []string{"Unsafe spread of an `any` value in an array."},
		},
		{
			sourceText: "\nconst x = [...([] as any[])];\n      ",
			isJsx:      false,
			wantIds:    []string{"unsafeArraySpread"},
			wantSpans:  []string{"...([] as any[])"},
			wantTexts:  []string{"Unsafe spread of an `any` value in an array."},
		},
		{
			sourceText: "const { x } = { x: 1 } as { x: any };",
			isJsx:      false,
			wantIds:    []string{"unsafeObjectPattern"},
			wantSpans:  []string{"x"},
			wantTexts:  []string{"Unsafe object destructuring of a property with an `any` value."},
		},
		{
			sourceText: "function foo({ x } = { x: 1 } as { x: any }) {}",
			isJsx:      false,
			wantIds:    []string{"unsafeObjectPattern"},
			wantSpans:  []string{"x"},
			wantTexts:  []string{"Unsafe object destructuring of a property with an `any` value."},
		},
		{
			sourceText: "({ x } = { x: 1 } as { x: any });",
			isJsx:      false,
			wantIds:    []string{"unsafeObjectPattern"},
			wantSpans:  []string{"x"},
			wantTexts:  []string{"Unsafe object destructuring of a property with an `any` value."},
		},
		{
			sourceText: "const { x: y } = { x: 1 } as { x: any };",
			isJsx:      false,
			wantIds:    []string{"unsafeObjectPattern"},
			wantSpans:  []string{"y"},
			wantTexts:  []string{"Unsafe object destructuring of a property with an `any` value."},
		},
		{
			sourceText: "function foo({ x: y } = { x: 1 } as { x: any }) {}",
			isJsx:      false,
			wantIds:    []string{"unsafeObjectPattern"},
			wantSpans:  []string{"y"},
			wantTexts:  []string{"Unsafe object destructuring of a property with an `any` value."},
		},
		{
			sourceText: "({ x: y } = { x: 1 } as { x: any });",
			isJsx:      false,
			wantIds:    []string{"unsafeObjectPattern"},
			wantSpans:  []string{"y"},
			wantTexts:  []string{"Unsafe object destructuring of a property with an `any` value."},
		},
		{
			sourceText: "\nconst {\n  x: { y },\n} = { x: { y: 1 } } as { x: { y: any } };\n      ",
			isJsx:      false,
			wantIds:    []string{"unsafeObjectPattern"},
			wantSpans:  []string{"y"},
			wantTexts:  []string{"Unsafe object destructuring of a property with an `any` value."},
		},
		{
			sourceText: "function foo({ x: { y } } = { x: { y: 1 } } as { x: { y: any } }) {}",
			isJsx:      false,
			wantIds:    []string{"unsafeObjectPattern"},
			wantSpans:  []string{"y"},
			wantTexts:  []string{"Unsafe object destructuring of a property with an `any` value."},
		},
		{
			sourceText: "\n({\n  x: { y },\n} = { x: { y: 1 } } as { x: { y: any } });\n      ",
			isJsx:      false,
			wantIds:    []string{"unsafeObjectPattern"},
			wantSpans:  []string{"y"},
			wantTexts:  []string{"Unsafe object destructuring of a property with an `any` value."},
		},
		{
			sourceText: "\nconst {\n  x: [y],\n} = { x: { y: 1 } } as { x: [any] };\n      ",
			isJsx:      false,
			wantIds:    []string{"unsafeArrayPatternFromTuple"},
			wantSpans:  []string{"y"},
			wantTexts:  []string{"Unsafe array destructuring of a tuple element with an `any` value."},
		},
		{
			sourceText: "function foo({ x: [y] } = { x: { y: 1 } } as { x: [any] }) {}",
			isJsx:      false,
			wantIds:    []string{"unsafeArrayPatternFromTuple"},
			wantSpans:  []string{"y"},
			wantTexts:  []string{"Unsafe array destructuring of a tuple element with an `any` value."},
		},
		{
			sourceText: "\n({\n  x: [y],\n} = { x: { y: 1 } } as { x: [any] });\n      ",
			isJsx:      false,
			wantIds:    []string{"unsafeArrayPatternFromTuple"},
			wantSpans:  []string{"y"},
			wantTexts:  []string{"Unsafe array destructuring of a tuple element with an `any` value."},
		},
		{
			sourceText: "const x = { y: 1 as any };",
			isJsx:      false,
			wantIds:    []string{"anyAssignment"},
			wantSpans:  []string{"y: 1 as any"},
			wantTexts:  []string{"Unsafe assignment of an `any` value."},
		},
		{
			sourceText: "const x = { y: { z: 1 as any } };",
			isJsx:      false,
			wantIds:    []string{"anyAssignment"},
			wantSpans:  []string{"z: 1 as any"},
			wantTexts:  []string{"Unsafe assignment of an `any` value."},
		},
		{
			sourceText: "const x: { y: Set<Set<Set<string>>> } = { y: new Set<Set<Set<any>>>() };",
			isJsx:      true,
			wantIds:    []string{"unsafeAssignment"},
			wantSpans:  []string{"y: new Set<Set<Set<any>>>()"},
			wantTexts:  []string{"Unsafe assignment of type `Set<Set<Set<any>>>` to a variable of type `Set<Set<Set<string>>>`."},
		},
		{
			sourceText: "const x = { ...(1 as any) };",
			isJsx:      false,
			wantIds:    []string{"anyAssignment"},
			wantSpans:  []string{"x = { ...(1 as any) }"},
			wantTexts:  []string{"Unsafe assignment of an `any` value."},
		},
		{
			sourceText: "\ntype Props = { a: string };\ndeclare function Foo(props: Props): never;\n<Foo a={1 as any} />;\n      ",
			isJsx:      true,
			wantIds:    []string{"anyAssignment"},
			wantSpans:  []string{"1 as any"},
			wantTexts:  []string{"Unsafe assignment of an `any` value."},
		},
		{
			sourceText: "\nfunction foo() {\n  const bar = this;\n}\n      ",
			isJsx:      false,
			wantIds:    []string{"anyAssignment"},
			wantSpans:  []string{"bar = this"},
			wantTexts:  []string{"Unsafe assignment of an `any` value."},
		},
		{
			sourceText: "\ntype T = [string, T[]];\nconst test: T = ['string', []] as any;\n      ",
			isJsx:      false,
			wantIds:    []string{"anyAssignment"},
			wantSpans:  []string{"test: T = ['string', []] as any"},
			wantTexts:  []string{"Unsafe assignment of an `any` value."},
		},
		{
			sourceText: "\ntype Foo = { bar: number };\nconst bar: any = 1;\nconst foo: Foo = { bar };\n      ",
			isJsx:      false,
			wantIds:    []string{"anyAssignment"},
			wantSpans:  []string{"bar"},
			wantTexts:  []string{"Unsafe assignment of an `any` value."},
		},
	}
	for index, testCase := range cases {
		t.Run(noUnsafeAssignmentCaseName(index), func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, NoUnsafeAssignment,
				noUnsafeAssignmentFileFor(testCase.isJsx), testCase.sourceText)
			rule_testing.ExpectFindings(t, result, testCase.wantIds...)

			onDisk := noUnsafeAssignmentOnDisk(testCase.sourceText)
			for index, wantSpan := range testCase.wantSpans {
				reported := onDisk[result.Diagnostics[index].Range.Pos():result.Diagnostics[index].Range.End()]
				if reported != wantSpan {
					t.Errorf("finding %d covers %q, want %q", index, reported, wantSpan)
				}
				if result.Diagnostics[index].Message.Description != testCase.wantTexts[index] {
					t.Errorf("finding %d reads %q, want %q", index,
						result.Diagnostics[index].Message.Description, testCase.wantTexts[index])
				}
			}
		})
	}
}

// TestNoUnsafeAssignmentDiscriminatesOnCasesUpstreamDoesNotWrite covers two guards the corpus misses.
//
// Upstream's corpus writes no rest element in a position where one would absorb an `any`, and no
// non-tuple array destructured element by element. Mutants removing either guard survived all
// ninety-two imported cases, which is what these rows exist to close.
//
// The rest guard is upstream's own comment: a rest element is not a one-to-one assignment, so it is
// skipped rather than matched against a tuple member. Without the skip, `[a, ...rest]` from
// `[string, any]` matches the rest against the tuple's second member and reports the `any` at a
// position nothing was actually assigned from.
//
// The tuple guard is what stops the element walk on a plain array: `string[]` has no per-position
// types to compare, so upstream returns before the loop. Both verdicts measured on the installed
// 8.67.0 build, each with the control that separates the guard from the shape.
func TestNoUnsafeAssignmentDiscriminatesOnCasesUpstreamDoesNotWrite(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText string
		wantIds    []string
		reason     string
	}{
		{
			sourceText: "declare const t: [string, any];\nconst [a, ...rest] = t;",
			wantIds:    nil,
			reason:     "a rest element is not a one-to-one assignment, so it is skipped",
		},
		{
			sourceText: "declare const t: [any, any];\nconst [...rest] = t;",
			wantIds:    nil,
			reason:     "and a pattern that is only a rest element reports nothing at all",
		},
		{
			sourceText: "declare const o: { a: any, b: string };\nconst { b, ...rest } = o;",
			wantIds:    nil,
			reason:     "the same holds for an object rest",
		},
		{
			sourceText: "declare const t: [string, any];\nconst [a, b] = t;",
			wantIds:    []string{"unsafeArrayPatternFromTuple"},
			reason:     "the control: the same tuple destructured by name does report",
		},
		{
			sourceText: "declare const a: string[];\nconst [x, y] = a;",
			wantIds:    nil,
			reason:     "a non-tuple array has no per-position types, so the element walk stops",
		},
		{
			sourceText: "declare const t: [any];\nconst [x] = t;",
			wantIds:    []string{"unsafeArrayPatternFromTuple"},
			reason:     "the control: a tuple of the same length does report",
		},
		{
			// The spread anchor's parent test. A spread in a CALL is `no-unsafe-argument`'s
			// question rather than this rule's, and without the test this reports there too.
			sourceText: "declare const a: any[];\ndeclare function f(...xs: string[]): void;\nf(...a);",
			wantIds:    nil,
			reason:     "a spread in a call belongs to a different rule in this family",
		},
		{
			sourceText: "declare const a: any[];\nconst b = [...a];",
			wantIds:    []string{"unsafeArraySpread"},
			reason:     "the control: the same spread in an array literal reports",
		},
		{
			// An object spread reaches a different arm and still reports, which is worth pinning
			// because the parent test above could plausibly have been written to exclude it.
			sourceText: "declare const a: any;\nconst o = { ...a };",
			wantIds:    []string{"anyAssignment"},
			reason:     "an object spread of an any reports through the assignment arm",
		},
		{
			// The assignment anchor's operator test. Upstream's selector names `=` alone, so a
			// compound assignment is a different question the rule does not ask.
			sourceText: "declare let x: string;\ndeclare const y: any;\nx += y;",
			wantIds:    nil,
			reason:     "a compound assignment is not this rule's question",
		},
		{
			sourceText: "declare let x: string;\ndeclare const y: any;\nx ||= y;",
			wantIds:    nil,
			reason:     "nor is a logical assignment",
		},
		{
			sourceText: "declare let x: string;\ndeclare const y: any;\nx = y;",
			wantIds:    []string{"anyAssignment"},
			reason:     "the control: the plain assignment does report",
		},
	}
	for index, testCase := range cases {
		t.Run(noUnsafeAssignmentCaseName(index), func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, NoUnsafeAssignment,
				noUnsafeAssignmentFileFor(false), testCase.sourceText)
			rule_testing.ExpectFindings(t, result, testCase.wantIds...)
		})
	}
}

// TestNoUnsafeAssignmentRequiresTheTypedHarness pins the checker declaration.
//
// A typed rule handed the plain harness gets a nil checker, and the guard at the top of each
// listener turns that into silence rather than a panic. Silence is the more dangerous failure: every
// clean case passes vacuously and every reporting case fails in a way that reads as a rule bug.
func TestNoUnsafeAssignmentRequiresTheTypedHarness(t *testing.T) {
	t.Parallel()

	source := "const x: string = 1 as any;"
	rule_testing.ExpectClean(t, rule_testing.Run(t, NoUnsafeAssignment,
		noUnsafeAssignmentFileFor(false), source))
	rule_testing.ExpectFindings(t, rule_testing.RunTyped(t, NoUnsafeAssignment,
		noUnsafeAssignmentFileFor(false), source), "anyAssignment")
}

// TestNoUnsafeAssignmentSurvivesShapesTheParserRecoversFrom is a crash fixture.
//
// Two hazards meet in this rule. `Elements()` and `Properties()` panic on a kind that has neither,
// and the shelf's `IsTypeAnyArrayType` indexes `getTypeArguments(t)[0]` with no length test, so an
// array type carrying no arguments panics one call in. Both are guarded, and the walk recovers per
// FILE rather than per rule, so either would cost every rule in the package every finding in that
// file while the run still printed a plausible summary.
//
// There is no finding to assert; the assertion is that the run completes. Upstream cannot reach most
// of these because its parser refuses the file outright.
func TestNoUnsafeAssignmentSurvivesShapesTheParserRecoversFrom(t *testing.T) {
	t.Parallel()

	for name, sourceText := range map[string]string{
		"emptyArrayPattern":  "declare const a: any[];\nconst [] = a;",
		"emptyObjectPattern": "declare const o: { a: any };\nconst {} = o;",
		"holeInPattern":      "declare const a: [any, any];\nconst [, x] = a;",
		"restOnly":           "declare const a: any[];\nconst [...r] = a;",
		"objectRestOnly":     "declare const o: { a: any };\nconst { ...r } = o;",
		"computedKey":        "declare const o: { a: any };\ndeclare const k: string;\nconst { [k]: v } = o;",
		"unclosedPattern":    "declare const a: any[];\nconst [x = a;",
		"noInitializer":      "const x: string;",
		"emptyArrayLiteral":  "const a = [];",
		"spreadNothing":      "const a = [...];",
		"deepNesting":        "declare const a: [[[any]]];\nconst [[[x]]] = a;",
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			rule_testing.RunTyped(t, NoUnsafeAssignment,
				noUnsafeAssignmentFileFor(false), sourceText)
		})
	}
}
