package react

import (
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// The corpus is upstream's thirty cases, extracted by stubbing its RuleTester and capturing the
// case objects the test file hands it, so no case was retyped. All thirty were replayed against the
// installed build before being written down and every one reproduced exactly, with no substitution.
//
// Not one case carries an `output` field, which is the corpus confirming what the rule file says:
// upstream deprecated this rule and commented its fixer out, so there is no repair to port.
//
// Options are carried as the raw payload the config layer delivers and routed through this rule's
// own decoder rather than built as a struct.
const sortDefaultPropsFile = "/repository/source/SortDefaultProps.tsx"

// sortDefaultPropsDecode routes a fixture's raw options through the exported decoder.
//
// A case with no `options` is spelled `""` in the rows above, which is the JSON empty string rather
// than an empty payload, so it is mapped to the empty payload the config layer would actually hand
// a rule configured as a bare "error". Keeping the two distinguishable in the rows makes the
// generated table readable; collapsing them here keeps the decoder receiving what it will see.
func sortDefaultPropsDecode(t *testing.T, raw string) any {
	t.Helper()
	if raw == `""` {
		raw = ""
	}
	decoded, err := DecodeSortDefaultPropsOptions([]byte(raw))
	if err != nil {
		t.Fatalf("decoding %s returned %v", raw, err)
	}
	return decoded
}

func TestSortDefaultPropsStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct{ name, sourceText, rawOptions string }{
		{"upstream valid 0", "\n        var First = createReactClass({\n          render: function() {\n            return <div />;\n          }\n        });\n      ", ""},
		{"upstream valid 1", "\n        var First = createReactClass({\n          propTypes: {\n            A: PropTypes.any,\n            Z: PropTypes.string,\n            a: PropTypes.any,\n            z: PropTypes.string\n          },\n          getDefaultProps: function() {\n            return {\n              A: \"A\",\n              Z: \"Z\",\n              a: \"a\",\n              z: \"z\"\n            };\n          },\n          render: function() {\n            return <div />;\n          }\n        });\n      ", ""},
		{"upstream valid 2", "\n        var First = createReactClass({\n          propTypes: {\n            a: PropTypes.any,\n            A: PropTypes.any,\n            z: PropTypes.string,\n            Z: PropTypes.string\n          },\n          getDefaultProps: function() {\n            return {\n              a: \"a\",\n              A: \"A\",\n              z: \"z\",\n              Z: \"Z\"\n            };\n          },\n          render: function() {\n            return <div />;\n          }\n        });\n      ", "{\"ignoreCase\": true}"},
		{"upstream valid 3", "\n        var First = createReactClass({\n          propTypes: {\n            a: PropTypes.any,\n            z: PropTypes.string\n          },\n          getDefaultProps: function() {\n            return {\n              a: \"a\",\n              z: \"z\"\n            };\n          },\n          render: function() {\n            return <div />;\n          }\n        });\n        var Second = createReactClass({\n          propTypes: {\n            AA: PropTypes.any,\n            ZZ: PropTypes.string\n          },\n          getDefaultProps: function() {\n            return {\n              AA: \"AA\",\n              ZZ: \"ZZ\"\n            };\n          },\n          render: function() {\n            return <div />;\n          }\n        });\n      ", ""},
		{"upstream valid 4", "\n        class First extends React.Component {\n          render() {\n            return <div />;\n          }\n        }\n        First.propTypes = {\n          a: PropTypes.string,\n          z: PropTypes.string\n        };\n        First.propTypes.justforcheck = PropTypes.string;\n        First.defaultProps = {\n          a: a,\n          z: z\n        };\n        First.defaultProps.justforcheck = \"justforcheck\";\n      ", ""},
		{"upstream valid 5", "\n        class First extends React.Component {\n          render() {\n            return <div />;\n          }\n        }\n        First.propTypes = {\n          a: PropTypes.any,\n          A: PropTypes.any,\n          z: PropTypes.string,\n          Z: PropTypes.string\n        };\n        First.defaultProps = {\n          a: \"a\",\n          A: \"A\",\n          z: \"z\",\n          Z: \"Z\"\n        };\n      ", "{\"ignoreCase\": true}"},
		{"upstream valid 6", "\n        class Component extends React.Component {\n          static propTypes = {\n            a: PropTypes.any,\n            b: PropTypes.any,\n            c: PropTypes.any\n          };\n          static defaultProps = {\n            a: \"a\",\n            b: \"b\",\n            c: \"c\"\n          };\n          render() {\n            return <div />;\n          }\n        }\n      ", ""},
		{"upstream valid 7", "\n        class Hello extends React.Component {\n          render() {\n            return <div>Hello</div>;\n          }\n        }\n        Hello.propTypes = {\n          \"aria-controls\": PropTypes.string\n        };\n        Hello.defaultProps = {\n          \"aria-controls\": \"aria-controls\"\n        };\n      ", "{\"ignoreCase\": true}"},
		{"upstream valid 8", "\n        var Hello = createReactClass({\n          render: function() {\n            let { a, ...b } = obj;\n            let c = { ...d };\n            return <div />;\n          }\n        });\n      ", ""},
		{"upstream valid 9", "\n        var First = createReactClass({\n          propTypes: {\n            barRequired: PropTypes.func.isRequired,\n            onBar: PropTypes.func,\n            z: PropTypes.any\n          },\n          getDefaultProps: function() {\n            return {\n              barRequired: \"barRequired\",\n              onBar: \"onBar\",\n              z: \"z\"\n            };\n          },\n          render: function() {\n            return <div />;\n          }\n        });\n      ", ""},
		{"upstream valid 10", "\n        export default class ClassWithSpreadInPropTypes extends BaseClass {\n          static propTypes = {\n            b: PropTypes.string,\n            ...c.propTypes,\n            a: PropTypes.string\n          }\n          static defaultProps = {\n            b: \"b\",\n            ...c.defaultProps,\n            a: \"a\"\n          }\n        }\n      ", ""},
		{"upstream valid 11", "\n        export default class ClassWithSpreadInPropTypes extends BaseClass {\n          static propTypes = {\n            a: PropTypes.string,\n            b: PropTypes.string,\n            c: PropTypes.string,\n            d: PropTypes.string,\n            e: PropTypes.string,\n            f: PropTypes.string\n          }\n          static defaultProps = {\n            a: \"a\",\n            b: \"b\",\n            ...c.defaultProps,\n            e: \"e\",\n            f: \"f\",\n            ...d.defaultProps\n          }\n        }\n      ", ""},
		{"upstream valid 12", "\n        const defaults = {\n          b: \"b\"\n        };\n        const types = {\n          a: PropTypes.string,\n          b: PropTypes.string,\n          c: PropTypes.string\n        };\n        function StatelessComponentWithSpreadInPropTypes({ a, b, c }) {\n          return <div>{a}{b}{c}</div>;\n        }\n        StatelessComponentWithSpreadInPropTypes.propTypes = types;\n        StatelessComponentWithSpreadInPropTypes.defaultProps = {\n          c: \"c\",\n          ...defaults,\n          a: \"a\"\n        };\n      ", ""},
		{"upstream valid 13", "\n        const propTypes = require('./externalPropTypes')\n        const defaultProps = require('./externalDefaultProps')\n        const TextFieldLabel = (props) => {\n          return <div />;\n        };\n        TextFieldLabel.propTypes = propTypes;\n        TextFieldLabel.defaultProps = defaultProps;\n      ", ""},
		{"upstream valid 14", "\n        const First = (props) => <div />;\n        export const propTypes = {\n            a: PropTypes.any,\n            z: PropTypes.string,\n        };\n        export const defaultProps = {\n            a: \"a\",\n            z: \"z\",\n        };\n        First.propTypes = propTypes;\n        First.defaultProps = defaultProps;\n      ", ""},
		{"upstream valid 15", "\n        const defaults = {\n          b: \"b\"\n        };\n        const First = (props) => <div />;\n        export const propTypes = {\n            a: PropTypes.string,\n            b: PropTypes.string,\n            z: PropTypes.string,\n        };\n        export const defaultProps = {\n            ...defaults,\n            a: \"a\",\n            z: \"z\",\n        };\n        First.propTypes = propTypes;\n        First.defaultProps = defaultProps;\n      ", ""},
		{"upstream valid 16", "\n        class First extends React.Component {\n          render() {\n            return <div />;\n          }\n        }\n\n        First.defaultProps = {\n            a: PropTypes.any,\n            onBar: PropTypes.func,\n            onFoo: PropTypes.func,\n            z: PropTypes.string,\n        };\n      ", ""},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTypedWithOptions(t, SortDefaultProps, sortDefaultPropsFile,
				testCase.sourceText, sortDefaultPropsDecode(t, testCase.rawOptions))
			rule_testing.ExpectClean(t, result)
		})
	}
}

func TestSortDefaultPropsFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name, sourceText, rawOptions string
		messageIds                   []string
	}{
		{"upstream invalid 0", "\n        class Component extends React.Component {\n          static propTypes = {\n            a: PropTypes.any,\n            b: PropTypes.any,\n            c: PropTypes.any\n          };\n          static defaultProps = {\n            a: \"a\",\n            c: \"c\",\n            b: \"b\"\n          };\n          render() {\n            return <div />;\n          }\n        }\n      ", "", []string{"propsNotSorted"}},
		{"upstream invalid 1", "\n        class Component extends React.Component {\n          static propTypes = {\n            a: PropTypes.any,\n            b: PropTypes.any,\n            c: PropTypes.any\n          };\n          static defaultProps = {\n            c: \"c\",\n            b: \"b\",\n            a: \"a\"\n          };\n          render() {\n            return <div />;\n          }\n        }\n      ", "", []string{"propsNotSorted", "propsNotSorted"}},
		{"upstream invalid 2", "\n        class Component extends React.Component {\n          static propTypes = {\n            a: PropTypes.any,\n            b: PropTypes.any\n          };\n          static defaultProps = {\n            Z: \"Z\",\n            a: \"a\",\n          };\n          render() {\n            return <div />;\n          }\n        }\n      ", "{\"ignoreCase\": true}", []string{"propsNotSorted"}},
		{"upstream invalid 3", "\n        class Component extends React.Component {\n          static propTypes = {\n            a: PropTypes.any,\n            z: PropTypes.any\n          };\n          static defaultProps = {\n            a: \"a\",\n            Z: \"Z\",\n          };\n          render() {\n            return <div />;\n          }\n        }\n      ", "", []string{"propsNotSorted"}},
		{"upstream invalid 4", "\n        class Hello extends React.Component {\n          render() {\n            return <div>Hello</div>;\n          }\n        }\n        Hello.propTypes = {\n          \"a\": PropTypes.string,\n          \"b\": PropTypes.string\n        };\n        Hello.defaultProps = {\n          \"b\": \"b\",\n          \"a\": \"a\"\n        };\n      ", "", []string{"propsNotSorted"}},
		{"upstream invalid 5", "\n        class Hello extends React.Component {\n          render() {\n            return <div>Hello</div>;\n          }\n        }\n        Hello.propTypes = {\n          \"a\": PropTypes.string,\n          \"b\": PropTypes.string,\n          \"c\": PropTypes.string\n        };\n        Hello.defaultProps = {\n          \"c\": \"c\",\n          \"b\": \"b\",\n          \"a\": \"a\"\n        };\n      ", "", []string{"propsNotSorted", "propsNotSorted"}},
		{"upstream invalid 6", "\n        class Hello extends React.Component {\n          render() {\n            return <div>Hello</div>;\n          }\n        }\n        Hello.propTypes = {\n          \"a\": PropTypes.string,\n          \"B\": PropTypes.string,\n        };\n        Hello.defaultProps = {\n          \"a\": \"a\",\n          \"B\": \"B\",\n        };\n      ", "", []string{"propsNotSorted"}},
		{"upstream invalid 7", "\n        class Hello extends React.Component {\n          render() {\n            return <div>Hello</div>;\n          }\n        }\n        Hello.propTypes = {\n          \"a\": PropTypes.string,\n          \"B\": PropTypes.string,\n        };\n        Hello.defaultProps = {\n          \"B\": \"B\",\n          \"a\": \"a\",\n        };\n      ", "{\"ignoreCase\": true}", []string{"propsNotSorted"}},
		{"upstream invalid 8", "\n        const First = (props) => <div />;\n        const propTypes = {\n          z: PropTypes.string,\n          a: PropTypes.any,\n        };\n        const defaultProps = {\n          z: \"z\",\n          a: \"a\",\n        };\n        First.propTypes = propTypes;\n        First.defaultProps = defaultProps;\n      ", "", []string{"propsNotSorted"}},
		{"upstream invalid 9", "\n        export default class ClassWithSpreadInPropTypes extends BaseClass {\n          static propTypes = {\n            b: PropTypes.string,\n            ...c.propTypes,\n            a: PropTypes.string\n          }\n          static defaultProps = {\n            b: \"b\",\n            a: \"a\",\n            ...c.defaultProps\n          }\n        }\n      ", "", []string{"propsNotSorted"}},
		{"upstream invalid 10", "\n        export default class ClassWithSpreadInPropTypes extends BaseClass {\n          static propTypes = {\n            a: PropTypes.string,\n            b: PropTypes.string,\n            c: PropTypes.string,\n            d: PropTypes.string,\n            e: PropTypes.string,\n            f: PropTypes.string\n          }\n          static defaultProps = {\n            b: \"b\",\n            a: \"a\",\n            ...c.defaultProps,\n            f: \"f\",\n            e: \"e\",\n            ...d.defaultProps\n          }\n        }\n      ", "", []string{"propsNotSorted", "propsNotSorted"}},
		{"upstream invalid 11", "\n        const defaults = {\n          b: \"b\"\n        };\n        const types = {\n          a: PropTypes.string,\n          b: PropTypes.string,\n          c: PropTypes.string\n        };\n        function StatelessComponentWithSpreadInPropTypes({ a, b, c }) {\n          return <div>{a}{b}{c}</div>;\n        }\n        StatelessComponentWithSpreadInPropTypes.propTypes = types;\n        StatelessComponentWithSpreadInPropTypes.defaultProps = {\n          c: \"c\",\n          a: \"a\",\n          ...defaults,\n        };\n      ", "", []string{"propsNotSorted"}},
		{"upstream invalid 12", "\n        class First extends React.Component {\n          render() {\n            return <div />;\n          }\n        }\n\n        First.defaultProps = {\n            a: PropTypes.any,\n            z: PropTypes.string,\n            onFoo: PropTypes.func,\n            onBar: PropTypes.func,\n        };\n      ", "", []string{"propsNotSorted", "propsNotSorted"}},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTypedWithOptions(t, SortDefaultProps, sortDefaultPropsFile,
				testCase.sourceText, sortDefaultPropsDecode(t, testCase.rawOptions))
			rule_testing.ExpectFindings(t, result, testCase.messageIds...)
		})
	}
}

// TestSortDefaultPropsRawKeyTextComparison pins the comparison being on SOURCE TEXT.
//
// Upstream's `getKey` calls `getText` on the key node, so quotes the author wrote are part of the
// string being compared. The consequence is an ordering nobody would design, and all three rows
// were measured against the installed build on 2026-08-27:
//
//	{a: 1, "b": 2}    REPORTS, `"b"` begins with a quote which sorts before `a`
//	{"a": 1, b: 2}    silent, the same rule running the other way
//	{[x]: 1, a: 2}    REPORTS, the computed key renders as `x`
//
// Upstream's corpus writes no mixed-quoting object, so nothing imported can distinguish a port
// comparing raw text from one comparing resolved names. Reproduced rather than corrected, because
// comparing resolved names would change the verdict on every quoted key in the tree.
func TestSortDefaultPropsRawKeyTextComparison(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name, sourceText string
		findings         int
	}{
		{
			"a quoted key sorts by its quote, so a then quoted b reports",
			"class C extends React.Component { static defaultProps = { a: 1, \"b\": 2 }; }",
			1,
		},
		{
			"quoted a then bare b is silent for the same reason",
			"class C extends React.Component { static defaultProps = { \"a\": 1, b: 2 }; }",
			0,
		},
		{
			// Four computed-key rows, because this is where the two trees disagree about what the
			// key node IS. ESTree's `key` for `{[x]: 1}` is the inner expression `x`; our parser
			// wraps it in a ComputedPropertyName carrying the brackets. Since `[` sorts between
			// uppercase and lowercase, comparing the wrapper flips the verdict on two of these four
			// and leaves the other two unchanged, which is exactly the shape that makes a wrong
			// port look right. All four measured against the installed build.
			"a computed key renders as its expression text, not the brackets",
			"class C extends React.Component { static defaultProps = { [x]: 1, a: 2 }; }",
			1,
		},
		{
			"a computed key before an uppercase key reports either way",
			"class C extends React.Component { static defaultProps = { [x]: 1, A: 2 }; }",
			1,
		},
		{
			"a computed key after a lowercase key is silent either way",
			"class C extends React.Component { static defaultProps = { b: 1, [x]: 2 }; }",
			0,
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTypedWithOptions(t, SortDefaultProps, sortDefaultPropsFile,
				testCase.sourceText, sortDefaultPropsDecode(t, `""`))
			if testCase.findings == 0 {
				rule_testing.ExpectClean(t, result)
				return
			}
			rule_testing.ExpectFindings(t, result, "propsNotSorted")
		})
	}
}

// TestSortDefaultPropsSpreadRestartsTheRun pins the spread behaviour.
//
// Upstream's reduce returns the element AFTER the spread rather than the spread itself, so that
// element becomes the accumulator without being compared to anything. A spread therefore splits the
// object into independently sorted runs rather than being skipped over. Measured: `{b, ...x, a}` is
// clean while both `{...x, b, a}` and `{b, a, ...x}` report.
func TestSortDefaultPropsSpreadRestartsTheRun(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name, sourceText string
		findings         int
	}{
		{
			"a spread between two out of order keys clears the comparison",
			"class C extends React.Component { static defaultProps = { b: 1, ...x, a: 2 }; }",
			0,
		},
		{
			"a leading spread does not exempt what follows",
			"class C extends React.Component { static defaultProps = { ...x, b: 1, a: 2 }; }",
			1,
		},
		{
			"a trailing spread does not exempt what precedes",
			"class C extends React.Component { static defaultProps = { b: 1, a: 2, ...x }; }",
			1,
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTypedWithOptions(t, SortDefaultProps, sortDefaultPropsFile,
				testCase.sourceText, sortDefaultPropsDecode(t, `""`))
			if testCase.findings == 0 {
				rule_testing.ExpectClean(t, result)
				return
			}
			rule_testing.ExpectFindings(t, result, "propsNotSorted")
		})
	}
}

// TestSortDefaultPropsAccumulatorDoesNotAdvanceOnAFinding pins how many findings one object yields.
//
// When a property is out of order upstream returns the PREVIOUS accumulator rather than advancing
// past the offender, so `{c, b, a}` reports twice, both measured against the same `c`. A port that
// resynced would report once and pass every imported case, because upstream's longest failing
// object has a single inversion.
func TestSortDefaultPropsAccumulatorDoesNotAdvanceOnAFinding(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name, sourceText string
		findings         int
	}{
		{
			// This row does NOT distinguish the two designs and is kept for contrast: a descending
			// run reports twice whether the accumulator holds or advances, which is why a mutation
			// making it advance survived until the rows below existed.
			"a fully descending run reports on each step either way",
			"class C extends React.Component { static defaultProps = { c: 1, b: 2, a: 3 }; }",
			2,
		},
		{
			// This one does. Holding at `c` makes both `a` and `b` report; advancing past `a`
			// would leave `b` in order and report once. Upstream reports twice, measured.
			"holding the accumulator makes every later key report against the same one",
			"class C extends React.Component { static defaultProps = { c: 1, a: 2, b: 3 }; }",
			2,
		},
		{
			"a longer run makes the difference wider",
			"class C extends React.Component { static defaultProps = { z: 1, a: 2, b: 3, c: 4 }; }",
			3,
		},
		{
			"a single inversion followed by an ordered key reports once",
			"class C extends React.Component { static defaultProps = { b: 1, a: 2, c: 3 }; }",
			1,
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTypedWithOptions(t, SortDefaultProps, sortDefaultPropsFile,
				testCase.sourceText, sortDefaultPropsDecode(t, `""`))
			ids := make([]string, testCase.findings)
			for index := range ids {
				ids[index] = "propsNotSorted"
			}
			rule_testing.ExpectFindings(t, result, ids...)
		})
	}
}

// TestSortDefaultPropsIgnoreCase pins the option, using the pair that inverts across it.
//
// `{B: 1, a: 2}` is SORTED by default, because uppercase letters sort before lowercase, and
// UNSORTED once case is folded. Both measured. A pair that inverts is the only kind that can tell a
// working option from one that is read and discarded.
func TestSortDefaultPropsIgnoreCase(t *testing.T) {
	t.Parallel()

	sensitive := rule_testing.RunTypedWithOptions(t, SortDefaultProps, sortDefaultPropsFile,
		"class C extends React.Component { static defaultProps = { B: 1, a: 2 }; }",
		sortDefaultPropsDecode(t, `""`))
	rule_testing.ExpectClean(t, sensitive)

	folded := rule_testing.RunTypedWithOptions(t, SortDefaultProps, sortDefaultPropsFile,
		"class C extends React.Component { static defaultProps = { B: 1, a: 2 }; }",
		sortDefaultPropsDecode(t, `{"ignoreCase": true}`))
	rule_testing.ExpectFindings(t, folded, "propsNotSorted")
}

// TestSortDefaultPropsWhichObjectsAreReached pins the two routes and three shapes that are not.
//
// The `getDefaultProps` METHOD row is the one worth reading twice: the name is one of the two this
// rule accepts, and the object is right there in the return, and it is still silent. Upstream looks
// at the property's VALUE rather than at a return statement, so a method never reaches the check.
// Measured.
func TestSortDefaultPropsWhichObjectsAreReached(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name, sourceText string
		findings         int
	}{
		{
			"a class field holding the object is reached",
			"class C extends React.Component { static defaultProps = { b: 1, a: 2 }; }",
			1,
		},
		{
			"an assignment after the class is reached",
			"class C extends React.Component {}\nC.defaultProps = { b: 1, a: 2 };",
			1,
		},
		{
			"an identifier naming the object is resolved and reached",
			"const props = { b: 1, a: 2 };\nclass C extends React.Component {}\nC.defaultProps = props;",
			1,
		},
		{
			"a getDefaultProps METHOD returning the object is NOT reached",
			"var C = createReactClass({ getDefaultProps: function(){ return { b: 1, a: 2 }; } });",
			0,
		},
		{
			"a plain object property named defaultProps is not an assignment",
			"const o = { defaultProps: { b: 1, a: 2 } };",
			0,
		},
		{
			"an empty object has nothing to compare",
			"class C extends React.Component { static defaultProps = {}; }",
			0,
		},
		{
			"a single property is compared against itself and cannot fail",
			"class C extends React.Component { static defaultProps = { a: 1 }; }",
			0,
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTypedWithOptions(t, SortDefaultProps, sortDefaultPropsFile,
				testCase.sourceText, sortDefaultPropsDecode(t, `""`))
			if testCase.findings == 0 {
				rule_testing.ExpectClean(t, result)
				return
			}
			rule_testing.ExpectFindings(t, result, "propsNotSorted")
		})
	}
}

// TestSortDefaultPropsSpanAndMessage asserts where the finding points and what it says.
//
// The finding anchors on the offending PROPERTY rather than on the object or the declaration, so a
// reader is shown the key that is in the wrong place. No message here interpolates, but the text is
// asserted against a literal typed here rather than against the rule's own constant, which would
// move both sides together under mutation.
func TestSortDefaultPropsSpanAndMessage(t *testing.T) {
	t.Parallel()

	source := "class C extends React.Component { static defaultProps = { b: 1, a: 2 }; }"
	result := rule_testing.RunTypedWithOptions(t, SortDefaultProps, sortDefaultPropsFile, source,
		sortDefaultPropsDecode(t, `""`))
	if len(result.Diagnostics) != 1 {
		t.Fatalf("wanted one finding, got %d", len(result.Diagnostics))
	}
	onDisk := strings.TrimSpace(source) + "\n"
	reported := onDisk[result.Diagnostics[0].Range.Pos():result.Diagnostics[0].Range.End()]
	if reported != "a: 2" {
		t.Fatalf("finding points at %q, wanted %q", reported, "a: 2")
	}
	want := "Default prop declarations should be sorted alphabetically. A sorted list is one a " +
		"reader can scan for a name instead of reading end to end, and it keeps two people adding " +
		"a default in the same week from colliding in the same spot."
	if result.Diagnostics[0].Message.Description != want {
		t.Fatalf("message is %q", result.Diagnostics[0].Message.Description)
	}
	if result.Diagnostics[0].Message.Id != "propsNotSorted" {
		t.Fatalf("id is %q", result.Diagnostics[0].Message.Id)
	}
	// Upstream commented its fixer out along with `fixable: 'code'`, so there is no repair to port.
	if len(result.Diagnostics[0].Fixes) != 0 || len(result.Diagnostics[0].Suggestions) != 0 {
		t.Fatal("upstream ships no repair for this rule and neither does this port")
	}
}

// TestDecodeSortDefaultPropsOptions routes configuration through the exported decoder.
//
// The nil row is the one that matters: a rule configured as a bare "error" is handed empty input,
// and this option defaults to FALSE so the zero value happens to be right. That is worth pinning
// rather than relying on, because the sibling rule in this package needed a pointer decoder for
// exactly the opposite reason.
func TestDecodeSortDefaultPropsOptions(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name, raw string
		want      bool
	}{
		{"absent configuration is case sensitive", "", false},
		{"an empty object is case sensitive", "{}", false},
		{"an explicit false is case sensitive", `{"ignoreCase": false}`, false},
		{"an explicit true folds case", `{"ignoreCase": true}`, true},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			decoded, err := DecodeSortDefaultPropsOptions([]byte(testCase.raw))
			if err != nil {
				t.Fatalf("decode returned %v", err)
			}
			options, isOptions := decoded.(SortDefaultPropsOptions)
			if !isOptions {
				t.Fatalf("decode returned %T", decoded)
			}
			if options.IgnoreCase != testCase.want {
				t.Fatalf("ignoreCase is %v, wanted %v", options.IgnoreCase, testCase.want)
			}
		})
	}
}

// TestSortDefaultPropsNilOptionsUsesTheDefault bypasses the decoder entirely.
func TestSortDefaultPropsNilOptionsUsesTheDefault(t *testing.T) {
	t.Parallel()

	result := rule_testing.RunTyped(t, SortDefaultProps, sortDefaultPropsFile,
		"class C extends React.Component { static defaultProps = { b: 1, a: 2 }; }")
	rule_testing.ExpectFindings(t, result, "propsNotSorted")
}

// TestSortDefaultPropsRequiresTheTypedHarness asserts the checker is genuinely needed.
//
// Only the identifier route needs resolution, so the typed harness is asserted through that route
// specifically: `GetSymbolAtLocation` on a nil checker returns nil rather than crashing, which
// would make this rule quietly narrower rather than obviously broken.
func TestSortDefaultPropsRequiresTheTypedHarness(t *testing.T) {
	t.Parallel()

	if !SortDefaultProps.NeedsTypeChecker {
		t.Fatal("the identifier route resolves through the checker and must declare it")
	}
	source := "const props = { b: 1, a: 2 };\nclass C extends React.Component {}\nC.defaultProps = props;"
	untyped := rule_testing.RunWithOptions(t, SortDefaultProps, sortDefaultPropsFile, source,
		DefaultSortDefaultPropsOptions())
	rule_testing.ExpectClean(t, untyped)

	typed := rule_testing.RunTypedWithOptions(t, SortDefaultProps, sortDefaultPropsFile, source,
		DefaultSortDefaultPropsOptions())
	rule_testing.ExpectFindings(t, typed, "propsNotSorted")
}

// TestSortDefaultPropsLeadingTriviaIsNotPartOfTheKey guards a large false positive class.
//
// `Pos()` includes leading trivia here, so slicing a key node raw picks up the newline, indentation
// and any comment written before it. Whitespace and `/` sort below every letter, so an untrimmed
// key compares as less than almost anything and a correctly SORTED object reports.
//
// Both rows are measured silent against the installed build, and both report without the trim. The
// multiline row is the one that matters in practice: a defaults object written across lines is the
// normal shape, so this guard is the difference between a quiet rule and one that flags every
// sorted multiline object in the tree.
//
// Upstream cannot see this because its `getText(node)` returns a node's own text with trivia
// already excluded; the hazard is entirely ours.
func TestSortDefaultPropsLeadingTriviaIsNotPartOfTheKey(t *testing.T) {
	t.Parallel()

	cases := []struct{ name, sourceText string }{
		{
			"a sorted object written across lines is silent",
			"class C extends React.Component { static defaultProps = {\n  a: 1,\n  b: 2\n}; }",
		},
		{
			"a sorted object with a comment before a key is silent",
			"class C extends React.Component { static defaultProps = { a: 1, /* c */ b: 2 }; }",
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTypedWithOptions(t, SortDefaultProps, sortDefaultPropsFile,
				testCase.sourceText, sortDefaultPropsDecode(t, `""`))
			rule_testing.ExpectClean(t, result)
		})
	}
}

// TestSortDefaultPropsGetDefaultPropsSpelling pins the second accepted name.
//
// Upstream accepts `getDefaultProps` alongside `defaultProps` in both arms, and a mutation dropping
// it survived because upstream's corpus writes the name only as a METHOD, which is silent for a
// different reason. As a field or an assignment it reports, both measured.
func TestSortDefaultPropsGetDefaultPropsSpelling(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name, sourceText string
		findings         int
	}{
		{
			"getDefaultProps as a class field is judged",
			"class C extends React.Component { static getDefaultProps = { b: 1, a: 2 }; }",
			1,
		},
		{
			"getDefaultProps as an assignment is judged",
			"class C extends React.Component {}\nC.getDefaultProps = { b: 1, a: 2 };",
			1,
		},
		{
			"getDefaultProps as an object property is not an assignment and is silent",
			"var C = createReactClass({ getDefaultProps: { b: 1, a: 2 } });",
			0,
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTypedWithOptions(t, SortDefaultProps, sortDefaultPropsFile,
				testCase.sourceText, sortDefaultPropsDecode(t, `""`))
			if testCase.findings == 0 {
				rule_testing.ExpectClean(t, result)
				return
			}
			rule_testing.ExpectFindings(t, result, "propsNotSorted")
		})
	}
}
