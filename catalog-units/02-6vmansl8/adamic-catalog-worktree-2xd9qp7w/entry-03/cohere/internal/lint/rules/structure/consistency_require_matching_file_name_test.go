package structure

import (
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// The directory is fixed and the base name varies, because the base name is the subject. A fixture
// that varied the directory would be testing nothing this rule reads.
const matchingFileNameDirectory = "/repository/source/components/"

func TestReactComponentRequireMatchingFileNameFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		fileName   string
		sourceText string
	}{
		{
			"a function component named for something other than its file",
			"FinanceFloorView.tsx",
			"export function FinanceFloorViewRow() {\n    return <div />;\n}\n",
		},
		// Case is part of the name. Both directions, because a comparison folding case only on one
		// side, or only on the first letter, passes one of these.
		{
			"a file differing from its component only in the first letter's case",
			"financeFloorView.tsx",
			"export function FinanceFloorView() {\n    return <div />;\n}\n",
		},
		{
			"a file differing from its component only in an inner letter's case",
			"FinanceFloorview.tsx",
			"export function FinanceFloorView() {\n    return <div />;\n}\n",
		},
		{
			"an arrow component in a const",
			"Table.tsx",
			"export const TableRow = () => <div />;\n",
		},
		{
			"a function expression component in a const",
			"Table.tsx",
			"export const TableRow = function () {\n    return <div />;\n};\n",
		},
		// The wrapper stands for the function it wraps, so this is one component named TableRow,
		// and the file is named for neither.
		{
			"a memo wrapper absorbing the function it names",
			"Table.tsx",
			"function TableRowInner(properties: { label: string }) {\n    return <div>{properties.label}</div>;\n}\n\nexport const TableRow = React.memo(TableRowInner, function (previous, next) {\n    return previous.label === next.label;\n});\n",
		},
		// TableVirtualizedRow's shape: the memo call asserted back to the implementation's type.
		{
			"a memo wrapper under a type assertion",
			"Table.tsx",
			"export function TableRowInner<Row>(properties: { row: Row }) {\n    return <div />;\n}\n\nexport const TableRow = React.memo(TableRowInner, function (previous, next) {\n    return previous.row === next.row;\n}) as typeof TableRowInner;\n",
		},
		// Calendar's shape: a compound component built on its root.
		{
			"a compound component built with Object.assign",
			"Agenda.tsx",
			"export function CalendarRoot(properties: { day: number }) {\n    return <div />;\n}\n\nexport const Calendar = Object.assign(CalendarRoot, {\n    Size: 1,\n});\n",
		},
		// Only Object.assign is the compound idiom. Another receiver's assign is somebody else's
		// function, so it builds nothing this rule counts and the root stays the file's component.
		{
			"an assign on a receiver other than Object",
			"Calendar.tsx",
			"export function CalendarRoot(properties: { day: number }) {\n    return <div />;\n}\n\nexport const Calendar = Lodash.assign(CalendarRoot, {\n    Size: 1,\n});\n",
		},
		// Object.assign over something that is not a component builds an object, not a component,
		// so the file's one component is still the function.
		{
			"an Object.assign over a non-component beside a component",
			"Toggle.tsx",
			"const base = { a: 1 };\nexport const Settings = Object.assign(base, { b: 2 });\n\nexport function Button() {\n    return <div />;\n}\n",
		},
		{
			"a bare forwardRef wrapper around an inline function",
			"Input.tsx",
			"export const Field = forwardRef(function (properties, reference) {\n    return <input ref={reference} />;\n});\n",
		},
		{
			"a class component",
			"ErrorBoundary.tsx",
			"export class Boundary extends React.Component {\n    render() {\n        return <div />;\n    }\n}\n",
		},
		{
			"a class expression in a const",
			"ErrorBoundary.tsx",
			"export const Boundary = class extends Component {\n    render() {\n        return <div />;\n    }\n};\n",
		},
		// A default export with a name has a name to compare, whichever syntax carries it.
		{
			"a named default-exported function declaration",
			"Toggle.tsx",
			"export default function Button() {\n    return <div />;\n}\n",
		},
		{
			"a named function expression as the default export",
			"Toggle.tsx",
			"export default (function Button() {\n    return <div />;\n});\n",
		},
		// Barrel files are banned, so an index.tsx declaring a component is a component living under
		// a name that is not its own.
		{
			"an index file declaring a component",
			"index.tsx",
			"export function Button() {\n    return <div />;\n}\n",
		},
		{
			"a jsx file",
			"Toggle.jsx",
			"export function Button() {\n    return <div />;\n}\n",
		},
		// A dotted name that is not a companion suffix is compared whole, and no component can be
		// named `Button.client`.
		{
			"a dotted name that is not a test or stories file",
			"Button.client.tsx",
			"export function Button() {\n    return <div />;\n}\n",
		},
		// The three detection arms, one each. The first is IsLikelyReactComponent's parameter arm,
		// with no JSX at all. The second is JSX that only HasJsxOrReactHookCalls reaches, since
		// IsLikelyReactComponent looks at top-level returns only. The third is a hook call alone.
		{
			"a component detected only by its properties parameter",
			"Toggle.tsx",
			"export function Button(properties: { label: string }) {\n    return properties.label;\n}\n",
		},
		{
			"a component detected only by jsx nested inside an if",
			"Toggle.tsx",
			"export function Button(flag: boolean) {\n    if(flag) {\n        return <div />;\n    }\n    return null;\n}\n",
		},
		{
			"a component detected only by a hook call",
			"Toggle.tsx",
			"export function Button() {\n    const [open] = useState(false);\n    return open ? 'open' : 'closed';\n}\n",
		},
		// A hook's file holding a component. The file is named for the hook, so the component is not
		// living under its own name either.
		{
			"a component in a file named for a hook",
			"useThing.tsx",
			"export function ThingPanel() {\n    return <div />;\n}\n",
		},
		// Overload signatures each declare the name, and the signature reads as a component by its
		// parameter. Counted as two, this file would read as "several" and go silent.
		{
			"an overloaded component counted once",
			"Toggle.tsx",
			"export function Button(properties: { a: string }): JSX.Element;\nexport function Button(properties: { a: string } | { b: string }) {\n    return <div />;\n}\n",
		},
		// Things that are not components do not make a file read as several.
		{
			"a component beside a capitalized constant and a lowercase render helper",
			"Toggle.tsx",
			"const Colors = { primary: 'red' };\n\nfunction renderLabel() {\n    return <span />;\n}\n\nexport function Button() {\n    return <div>{renderLabel()}</div>;\n}\n",
		},
		// Only top-level declarations are the file's components. A component nested inside another
		// is no-multi-comp's finding, not a second candidate for this file's name.
		{
			"a component with another declared inside it",
			"Toggle.tsx",
			"export function Button() {\n    function Inner() {\n        return <span />;\n    }\n    return <div><Inner /></div>;\n}\n",
		},
		// Near misses on the Next.js grammar. The convention is a lowercase name, exact, so a
		// capitalized Page.tsx is an ordinary file, and Next's metadata digit is one digit only.
		{
			"a capitalized Page file",
			"Page.tsx",
			"export function OsPage() {\n    return <div />;\n}\n",
		},
		{
			"an icon file with two digits",
			"icon12.tsx",
			"export default function Icon() {\n    return <div />;\n}\n",
		},
		{
			"an icon file with a letter suffix",
			"iconA.tsx",
			"export default function Icon() {\n    return <div />;\n}\n",
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t,
				rule_testing.Run(t, ConsistencyRequireMatchingFileName,
					matchingFileNameDirectory+testCase.fileName, testCase.sourceText),
				"requireMatchingFileName")
		})
	}
}

func TestReactComponentRequireMatchingFileNameStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		fileName   string
		sourceText string
	}{
		// The shape the rule exists to keep, and the one nearly every component file has.
		{
			"a component in the file named for it",
			"FinanceFloorViewRow.tsx",
			"export function FinanceFloorViewRow() {\n    return <div />;\n}\n",
		},
		{
			"an arrow component in the file named for it",
			"TableRow.tsx",
			"export const TableRow = () => <div />;\n",
		},
		{
			"a jsx file named for its component",
			"Button.jsx",
			"export function Button() {\n    return <div />;\n}\n",
		},
		// The structure library's shape: an implementation function and the memo wrapper exported
		// in its name. One component, MenuItem, and the file is named for it.
		{
			"a memo wrapper named for the file, absorbing its implementation",
			"MenuItem.tsx",
			"function MenuItemComponent(properties: { label: string }) {\n    return <div>{properties.label}</div>;\n}\n\nexport const MenuItem = React.memo(MenuItemComponent, function arePropertiesEqual(previous, next) {\n    return previous.label === next.label;\n});\n",
		},
		{
			"a memo wrapper under a type assertion, named for the file",
			"TableRow.tsx",
			"export function TableRowInner<Row>(properties: { row: Row }) {\n    return <div />;\n}\n\nexport const TableRow = React.memo(TableRowInner, function (previous, next) {\n    return previous.row === next.row;\n}) as typeof TableRowInner;\n",
		},
		{
			"a compound component named for the file",
			"Calendar.tsx",
			"export function CalendarRoot(properties: { day: number }) {\n    return <div />;\n}\n\nexport const Calendar = Object.assign(CalendarRoot, {\n    Size: 1,\n});\n",
		},
		{
			"a wrapper around an inline function, named for the file",
			"MenuSearch.tsx",
			"export const MenuSearch = React.memo(\n    function MenuSearchInner(properties: { query: string }) {\n        return <input value={properties.query} />;\n    },\n);\n",
		},
		{
			"a class component named for the file",
			"ErrorBoundary.tsx",
			"export class ErrorBoundary extends React.PureComponent {\n    render() {\n        return <div />;\n    }\n}\n",
		},
		// Two or more is no-multi-comp's finding, and reporting here as well would be two findings
		// for one problem, one of them about a name the split is about to change anyway.
		{
			"two function components",
			"Table.tsx",
			"export function Table() {\n    return <TableRow />;\n}\n\nfunction TableRow() {\n    return <div />;\n}\n",
		},
		{
			"a class component beside a function component",
			"ErrorBoundary.tsx",
			"export class Boundary extends React.Component {\n    render() {\n        return <Fallback />;\n    }\n}\n\nfunction Fallback() {\n    return <div />;\n}\n",
		},
		{
			"a wrapper beside an unrelated component it does not absorb",
			"Table.tsx",
			"function TableCell() {\n    return <td />;\n}\n\nexport const TableRow = React.memo(function () {\n    return <tr />;\n});\n",
		},
		// Anonymous components are react/display-name's finding. There is no name to compare, and
		// once there is one this rule compares it.
		{
			"an anonymous default-exported function",
			"Toggle.tsx",
			"export default function () {\n    return <div />;\n}\n",
		},
		{
			"an anonymous default-exported arrow",
			"Toggle.tsx",
			"export default () => <div />;\n",
		},
		{
			"an anonymous default-exported class",
			"Toggle.tsx",
			"export default class extends React.Component {\n    render() {\n        return <div />;\n    }\n}\n",
		},
		// The anonymous one still counts toward several, so the named one is not this file's only
		// component and is not compared.
		{
			"an anonymous default beside a named component",
			"Toggle.tsx",
			"function Button() {\n    return <div />;\n}\n\nexport default function () {\n    return <Button />;\n}\n",
		},
		// The arrow spelling of the same pair. Alone, an anonymous arrow is silent whether or not it
		// is counted, so only beside a named component does this pin that it is.
		{
			"an anonymous default arrow beside a named component",
			"Toggle.tsx",
			"function Button() {\n    return <div />;\n}\n\nexport default () => <Button />;\n",
		},
		{
			"an anonymous default class beside a named component",
			"Toggle.tsx",
			"function Button() {\n    return <div />;\n}\n\nexport default class extends React.Component {\n    render() {\n        return <Button />;\n    }\n}\n",
		},
		// Nothing to name the file for.
		{
			"a file of utilities",
			"Toggle.tsx",
			"export const Colors = { primary: 'red' };\n\nexport function formatLabel(label: string) {\n    return label.trim();\n}\n",
		},
		{
			"a lowercase function returning jsx",
			"Toggle.tsx",
			"export function renderButton() {\n    return <div />;\n}\n",
		},
		{
			"a capitalized function with no jsx, no hook and no properties parameter",
			"Toggle.tsx",
			"export function Helper() {\n    return 1 + 1;\n}\n",
		},
		{
			"an uninitialized capitalized binding",
			"Toggle.tsx",
			"let Button;\n",
		},
		{
			"a barrel index that only re-exports",
			"index.tsx",
			"export { Button } from './Button';\nexport type { ButtonProperties } from './ButtonProperties';\n",
		},
		// Test and stories files are about a component rather than its home.
		{
			"a test file",
			"Button.test.tsx",
			"function Harness() {\n    return <div />;\n}\n",
		},
		{
			"a spec file",
			"Button.spec.tsx",
			"function Harness() {\n    return <div />;\n}\n",
		},
		{
			"a stories file",
			"Button.stories.tsx",
			"export function Primary() {\n    return <div />;\n}\n",
		},
		{
			"a story file",
			"Button.story.tsx",
			"export function Primary() {\n    return <div />;\n}\n",
		},
		// A .ts file is not a component file. The component is detected by a hook call rather than
		// JSX, because JSX does not parse in .ts and a fixture relying on it would be silent with or
		// without the gate.
		{
			"a typescript file holding a hook-using function",
			"Toggle.ts",
			"export function Button() {\n    const [open] = useState(false);\n    return open;\n}\n",
		},
	}

	// The Next.js grammar, every name on the list, each holding a component named otherwise. Every
	// one of these reports the moment its name leaves the list.
	for _, baseName := range []string{
		"page", "layout", "template", "loading", "error", "global-error", "not-found",
		"global-not-found", "forbidden", "unauthorized", "default", "route", "mdx-components",
		"sitemap", "robots", "manifest",
		"icon", "icon1", "apple-icon", "apple-icon2", "opengraph-image", "opengraph-image3",
		"twitter-image", "twitter-image9",
	} {
		cases = append(cases, struct {
			name       string
			fileName   string
			sourceText string
		}{
			"the next.js convention file " + baseName,
			baseName + ".tsx",
			"export default function OsPageRoute() {\n    return <div />;\n}\n",
		})
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t,
				rule_testing.Run(t, ConsistencyRequireMatchingFileName,
					matchingFileNameDirectory+testCase.fileName, testCase.sourceText))
		})
	}
}

// TestReactComponentRequireMatchingFileNameSpanAndMessage asserts where the finding points and the
// whole sentence it prints, since a single message id records neither.
//
// Three renderings: a capitalized file name, which is offered as the component's new name too; a
// lowercase one, which is not, since a lowercase component cannot be rendered as JSX; and a dashed
// one, which is not, since it cannot be a binding at all.
func TestReactComponentRequireMatchingFileNameSpanAndMessage(t *testing.T) {
	t.Parallel()

	const reasoning = "and the file is not named for it. With one component per file, the file name " +
		"is how a reader finds a component and the component name is how they know what a file " +
		"holds, and that only works as an index when the two match exactly, case included. "

	cases := []struct {
		name        string
		fileName    string
		sourceText  string
		wantSpan    string
		wantMessage string
	}{
		{
			"a capitalized file name",
			"FinanceFloorView.tsx",
			"/** The row. */\nexport function FinanceFloorViewRow() {\n    return <div />;\n}\n",
			"FinanceFloorViewRow",
			"`FinanceFloorViewRow` is the only component in `FinanceFloorView.tsx`, " + reasoning +
				"Rename the file to `FinanceFloorViewRow.tsx`, or rename the component to " +
				"`FinanceFloorView` if the file's name is the right one.",
		},
		{
			"a lowercase file name",
			"financeFloorView.tsx",
			"export const FinanceFloorView = () => <div />;\n",
			"FinanceFloorView",
			"`FinanceFloorView` is the only component in `financeFloorView.tsx`, " + reasoning +
				"Rename the file to `FinanceFloorView.tsx`.",
		},
		{
			"a dashed file name in a jsx file",
			"Finance-Floor.jsx",
			"export class FinanceFloor extends React.Component {\n    render() {\n        return <div />;\n    }\n}\n",
			"FinanceFloor",
			"`FinanceFloor` is the only component in `Finance-Floor.jsx`, " + reasoning +
				"Rename the file to `FinanceFloor.jsx`.",
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, ConsistencyRequireMatchingFileName,
				matchingFileNameDirectory+testCase.fileName, testCase.sourceText)
			rule_testing.ExpectFindings(t, result, "requireMatchingFileName")
			diagnostic := result.Diagnostics[0]
			span := testCase.sourceText[diagnostic.Range.Pos():diagnostic.Range.End()]
			if span != testCase.wantSpan {
				t.Errorf("span = %q, want %q", span, testCase.wantSpan)
			}
			if diagnostic.Message.Description != testCase.wantMessage {
				t.Errorf("message =\n%s\nwant\n%s", diagnostic.Message.Description, testCase.wantMessage)
			}
		})
	}
}
