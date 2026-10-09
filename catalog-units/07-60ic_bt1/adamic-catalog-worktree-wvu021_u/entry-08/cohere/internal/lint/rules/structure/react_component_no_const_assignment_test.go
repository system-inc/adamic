package structure

import (
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

const constAssignmentFile = "/repository/source/components/Thing.tsx"

// The tree contains zero instances of this construct, so it cannot cohere this rule at all: a clean
// run proves only that the rule does not fire on code containing none of the shape. These fixtures
// are the entire specification, which raises rather than lowers the bar on where the clean cases sit.
func TestReactComponentNoConstAssignmentFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		count      int
	}{
		{"an exported arrow component", "export const Thing = () => <div />;\n", 1},
		{"an exported function expression", "export const Thing = function () {\n    return <div />;\n};\n", 1},
		{"an unexported arrow component", "const Thing = () => <div />;\nexport default Thing;\n", 1},
		{"an unexported function expression", "const Thing = function () {\n    return <div />;\n};\nexport default Thing;\n", 1},
		// A block-bodied arrow, which is the shape the original's fix handles differently from an
		// expression body. It reports the same either way.
		{"a block-bodied arrow", "export const Thing = () => {\n    return <div />;\n};\n", 1},
		// An async component, which is one of the shapes the original's fix silently changes.
		{"an async arrow", "export const Thing = async () => <div />;\n", 1},
		// Two components on one statement. The original's fix produces two overlapping rewrites
		// here and loses one; reporting per declarator is correct and is what this pins.
		{"two components in one statement", "export const First = () => <div />,\n    Second = () => <span />;\n", 2},
		// A typed component. The annotation is on the declaration rather than the initializer, and
		// it is the other shape the original's fix drops.
		{"a typed arrow component", "export const Thing: React.FC = () => <div />;\n", 1},
		// A `let` rather than a `const`. The original visits VariableDeclaration without checking
		// the kind, so this reports too.
		{"a let-declared component", "export let Thing = () => <div />;\n", 1},
		// Nested inside a function body. This fixture was written on the silent side first, on the
		// reasoning that a component-shaped const inside a function is a callback rather than
		// something the file publishes. That reading was wrong: the original's `VariableDeclaration`
		// visitor is an ESLint selector, which fires at every depth, and its only guard skips the
		// export-wrapped case to avoid double-reporting. So a nested declaration reports.
		//
		// The tree has zero instances either way, so nothing but the original decides this, and the
		// original decides it by which selector it registered rather than by an argument. Recorded
		// because the opposite reading is the intuitive one and a later reader will have it too.
		{"an arrow inside a function body", "export function Outer() {\n    const Inner = () => <div />;\n    return <Inner />;\n}\n", 1},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, ReactComponentNoConstAssignment, constAssignmentFile, testCase.sourceText)
			expected := make([]string, testCase.count)
			for index := range expected {
				expected[index] = "noConstAssignment"
			}
			rule_testing.ExpectFindings(t, result, expected...)
		})
	}
}

func TestReactComponentNoConstAssignmentStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		fileName   string
		sourceText string
	}{
		// The shape the rule is asking for, which is what the whole tree already looks like.
		{"a function declaration", constAssignmentFile, "export function Thing() {\n    return <div />;\n}\n"},
		{"an unexported function declaration", constAssignmentFile, "function Thing() {\n    return <div />;\n}\nexport default Thing;\n"},
		// forwardRef has to be a const, because the value is what the call returns. Both spellings,
		// since the original checks the callee shape explicitly.
		{"a namespaced forwardRef", constAssignmentFile, "export const Thing = React.forwardRef(function (properties, reference) {\n    return <div ref={reference} />;\n});\n"},
		{"a bare forwardRef", constAssignmentFile, "export const Thing = forwardRef(function (properties, reference) {\n    return <div ref={reference} />;\n});\n"},
		// A capitalized const that is not a function. Each of these would report if the rule keyed
		// on the capital alone, and they are the population the rule runs over.
		{"a capitalized object constant", constAssignmentFile, "export const Colors = { primary: 'red' };\n"},
		{"a capitalized string constant", constAssignmentFile, "export const Title = 'Things';\n"},
		{"a capitalized array constant", constAssignmentFile, "export const Sizes = ['Small', 'Base'];\n"},
		{"a capitalized call result", constAssignmentFile, "export const Client = createClient();\n"},
		{"a capitalized const with no initializer", constAssignmentFile, "declare const Thing: () => JSX.Element;\n"},
		// A lowercase name is not a component by the capital convention.
		{"a lowercase arrow", constAssignmentFile, "export const buildThing = () => <div />;\n"},

		// The file gate.
		{"a plain typescript file", "/repository/source/Thing.ts", "export const Thing = () => 1;\n"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t,
				rule_testing.Run(t, ReactComponentNoConstAssignment, testCase.fileName, testCase.sourceText))
		})
	}
}
