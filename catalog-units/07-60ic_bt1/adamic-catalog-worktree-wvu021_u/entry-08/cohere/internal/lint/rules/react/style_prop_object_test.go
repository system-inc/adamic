package react

import (
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// stylePropObjectFile is where the fixtures pretend to live.
//
// A .tsx extension because most cases hold JSX. The rule has no suffix gate, which a case below
// pins by writing reporting createElement source to a plain `.ts` file.
const stylePropObjectFile = "/repository/source/StylePropObject.tsx"

// The corpus is upstream's, extracted mechanically rather than retyped.
//
// Every case in the two tables below comes from
// /tmp/lint-sources/eslint-plugin-react/tests/lib/rules/style-prop-object.js, read by evaluating
// the two arrays in the tester with a stubbed RuleTester and emitting Go literals from the decoded
// values, so no escape sequence was typed on the way here. The sources were then compared byte for
// byte against the decoded corpus with a script rather than by eye. Upstream carries 27 valid and 8
// invalid cases.
//
// All 35 were replayed against the installed build, eslint-plugin-react 7.37.5, through the ESLint
// Linter API before any Go was written, and the corpus and the running rule agreed on every one.
//
// Two of upstream's cases carry byte-identical source with opposite verdicts, separated only by the
// allow list, which is what makes the option surface visible rather than assumed.

// stylePropObjectOptions decodes a raw option body the way the config layer does.
//
// An empty body is the nil-options case, which here means an EMPTY allow list and therefore full
// enforcement. That is the opposite direction from the sibling forbid rules, where an empty list
// means the rule declines everything.
func stylePropObjectOptions(t *testing.T, raw string) any {
	t.Helper()
	decoded, err := DecodeStylePropObjectOptions([]byte(raw))
	if err != nil {
		t.Fatalf("decoding %q: %v", raw, err)
	}
	return decoded
}

// TestStylePropObjectFires runs every case this port reports on.
func TestStylePropObjectFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		rawOptions string
		wantCount  int
	}{
		{"invalid 0", `<div style="color: 'red'" />`, ``, 1},
		{"invalid 1", `<Hello style="color: 'red'" />`, ``, 1},
		{"invalid 2", `<div style={true} />`, ``, 1},
		{"invalid 3", `
        const styles = 'color: "red"';
        function redDiv2() {
          return <div style={styles} />;
        }
      `, ``, 1},
		{"invalid 4", `
        const styles = 'color: "red"';
        function redDiv2() {
          return <Hello style={styles} />;
        }
      `, ``, 1},
		{"invalid 5", `
        const styles = true;
        function redDiv() {
          return <div style={styles} />;
        }
      `, ``, 1},
		{"invalid 6", `<MyComponent style="myStyle" />`, `{"allow":["MyOtherComponent"]}`, 1},
		{"invalid 7", `React.createElement(MyComponent, { style: "mySpecialStyle" })`, `{"allow":["MyOtherComponent"]}`, 1},

		// Cases upstream's corpus does not write, each measured against the installed build.

		// Our parser keeps parenthesis nodes and upstream's folds them away, so without unwrapping
		// this port would be SILENT on every shape below while the installed build reports each
		// one. No imported fixture can catch this, because upstream's corpus cannot express a node
		// its own parser deletes.
		{"a parenthesized string value", `<div style={('x')} />`, ``, 1},
		{"a doubly parenthesized string value", `<div style={(('x'))} />`, ``, 1},
		{"an identifier whose initializer is parenthesized", "const s = (('x'));\nconst q = <div style={s} />;", ``, 1},
		{"a parenthesized createElement value", `React.createElement("div", { style: ('x') });`, ``, 1},

		// The literal set is upstream's node type, which is wider than the informal reading of the
		// word. Each of these was measured; none is written in the corpus.
		{"a numeric literal", `<div style={5} />`, ``, 1},
		{"a bigint literal", `<div style={1n} />`, ``, 1},
		{"a regular expression is a literal upstream", `<div style={/x/} />`, ``, 1},
		{"a false keyword", `<div style={false} />`, ``, 1},

		// The allow list only SKIPS, so a tag with no plain name to match is still examined. This
		// is the opposite of the sibling forbid rules, where an unreadable tag ends the check.
		{"a member tag is still examined", `<a.b style="x" />`, ``, 1},
		{"a namespaced tag is still examined", `<svg:a style="x" />`, ``, 1},
		{"naming a member tag in the allow list does not exempt it", `<a.b style="x" />`, `{"allow":["a.b"]}`, 1},
		{"an empty allow list exempts nothing", `<div style="x" />`, `{"allow":[]}`, 1},

		// Upstream reads the FIRST declaration, so a name declared twice is judged by the one
		// written first. Its partner in the silent table is the same pair in the other order, and
		// the two together are what separate index zero from a loop.
		{"two declarations, the first is a string", "var s = 'x';\nvar s = {};\nconst q = <div style={s} />;", ``, 1},

		// An empty expression container reaches upstream's report branch, because the container
		// test passes and the literal test then fails on an undefined expression.
		{"an empty expression container", `<div style={} />`, ``, 1},

		// The createElement shorthand carries its own name as the value, so it takes the
		// identifier arm.
		{"a createElement shorthand property", "const style = 'x';\nReact.createElement(\"div\", { style });", ``, 1},

		// A string first argument has no name, so the allow list cannot match it.
		{"allowing div does not exempt the string element", `React.createElement("div", { style: "x" });`, `{"allow":["div"]}`, 1},

		// A member first argument is neither a name nor a match, so the element is examined.
		{"a createElement member first argument", `React.createElement(a.b, { style: 'x' });`, ``, 1},

		// More than two arguments is still more than one.
		{"a createElement call with children", `React.createElement("div", { style: 'x' }, child);`, ``, 1},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			options := stylePropObjectOptions(t, testCase.rawOptions)
			result := rule_testing.RunTypedWithOptions(t, StylePropObject, stylePropObjectFile, testCase.sourceText, options)
			wantIds := make([]string, testCase.wantCount)
			for index := range wantIds {
				wantIds[index] = "stylePropNotObject"
			}
			rule_testing.ExpectFindings(t, result, wantIds...)
		})
	}
}

// TestStylePropObjectStaysSilent runs every case this port declines.
func TestStylePropObjectStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		rawOptions string
	}{
		{"valid 0", `<div style={{ color: "red" }} />`, ``},
		{"valid 1", `<Hello style={{ color: "red" }} />`, ``},
		{"valid 2", `
        function redDiv() {
          const styles = { color: "red" };
          return <div style={styles} />;
        }
      `, ``},
		{"valid 3", `
        function redDiv() {
          const styles = { color: "red" };
          return <Hello style={styles} />;
        }
      `, ``},
		{"valid 4", `
        const styles = { color: "red" };
        function redDiv() {
          return <div style={styles} />;
        }
      `, ``},
		{"valid 5", `
        function redDiv(props) {
          return <div style={props.styles} />;
        }
      `, ``},
		{"valid 6", `
        import styles from './styles';
        function redDiv() {
          return <div style={styles} />;
        }
      `, ``},
		{"valid 7", `
        import mystyles from './styles';
        const styles = Object.assign({ color: 'red' }, mystyles);
        function redDiv() {
          return <div style={styles} />;
        }
      `, ``},
		{"valid 8", `
        const otherProps = { style: { color: "red" } };
        const { a, b, ...props } = otherProps;
        <div {...props} />
      `, ``},
		{"valid 9", `
        const styles = Object.assign({ color: 'red' }, mystyles);
        React.createElement("div", { style: styles });
      `, ``},
		{"valid 10", `<div style></div>`, ``},
		{"valid 11", `
        React.createElement(MyCustomElem, {
          [style]: true
        }, 'My custom Elem')
      `, ``},
		{"valid 12", `
        let style;
        <div style={style}></div>
      `, ``},
		{"valid 13", `
        let style = null;
        <div style={style}></div>
      `, ``},
		{"valid 14", `
        let style = undefined;
        <div style={style}></div>
      `, ``},
		{"valid 15", `<div style={undefined}></div>`, ``},
		{"valid 16", `
        const props = { style: undefined };
        <div {...props} />
      `, ``},
		{"valid 17", `
        const otherProps = { style: undefined };
        const { a, b, ...props } = otherProps;
        <div {...props} />
      `, ``},
		{"valid 18", `
        React.createElement("div", {
          style: undefined
        })
      `, ``},
		{"valid 19", `
        let style;
        React.createElement("div", {
          style
        })
      `, ``},
		{"valid 20", `<div style={null}></div>`, ``},
		{"valid 21", `
        const props = { style: null };
        <div {...props} />
      `, ``},
		{"valid 22", `
        const otherProps = { style: null };
        const { a, b, ...props } = otherProps;
        <div {...props} />
      `, ``},
		{"valid 23", `
        React.createElement("div", {
          style: null
        })
      `, ``},
		{"valid 24", `
        const MyComponent = (props) => {
          React.createElement(MyCustomElem, {
            ...props
          });
        };
      `, ``},
		{"valid 25", `<MyComponent style="myStyle" />`, `{"allow":["MyComponent"]}`},
		{"valid 26", `React.createElement(MyComponent, { style: "mySpecialStyle" })`, `{"allow":["MyComponent"]}`},

		// Cases upstream's corpus does not write, each measured against the installed build.

		// A template is a different node type from a Literal upstream, so it is clean even though
		// its value is a string.
		{"a template literal", "const q = <div style={`x`} />;", ``},
		// A unary expression wraps a literal rather than being one.
		{"a negated number", `<div style={-5} />`, ``},
		{"an array literal", `<div style={[]} />`, ``},
		// The partner to the declaration-order case above: the first declaration is the object, so
		// this is clean. A port looping over declarations would report it and would be wider than
		// upstream.
		{"two declarations, the first is an object", "var s = {};\nvar s = 'x';\nconst q = <div style={s} />;", ``},
		// A parenthesized null is still null once unwrapped, so unwrapping does not invent a
		// finding here. This is the control on the unwrap: it must not turn a clean case dirty.
		{"a parenthesized null", `<div style={(null)} />`, ``},
		{"a parenthesized undefined", `<div style={(undefined)} />`, ``},
		// A namespaced attribute carries no plain name, so it is not the style prop at all.
		{"a namespaced style attribute", `<div xlink:style="x" />`, ``},
		// A computed key is excluded even when its text is the right one.
		{"a createElement computed style key", `React.createElement("div", { ["style"]: "x" });`, ``},
		// A spread carries no key to read.
		{"createElement spread properties", `React.createElement("div", { ...p });`, ``},
		// One argument is not more than one.
		{"a createElement call with one argument", `React.createElement("div");`, ``},
		// The allow list matches an identifier first argument by name.
		{"an allowed createElement component", `React.createElement(Foo, { style: 'x' });`, `{"allow":["Foo"]}`},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			options := stylePropObjectOptions(t, testCase.rawOptions)
			result := rule_testing.RunTypedWithOptions(t, StylePropObject, stylePropObjectFile, testCase.sourceText, options)
			rule_testing.ExpectClean(t, result)
		})
	}
}

// TestStylePropObjectAnchors pins WHERE each finding points, and the two arms differ.
//
// The JSX arm reports on the whole attribute, and the createElement arm reports on the VALUE
// expression rather than on the property or the call. A port using one anchor for both would pass
// every message-id fixture above while pointing somewhere the reader was never shown in half the
// cases. The identifier arm reports on the identifier in the value position, not on its
// declaration, which is a third anchor again.
func TestStylePropObjectAnchors(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		want       string
	}{
		{"the JSX arm spans the whole attribute", `<div style="color: red" />`, `style="color: red"`},
		{"the JSX literal arm spans the whole attribute", `<div style={true} />`, `style={true}`},
		{"the createElement arm spans the value", `React.createElement("div", { style: "x" });`, `"x"`},
		{"the identifier arm spans the use, not the declaration", "const s = 'x';\nconst q = <div style={s} />;", `s`},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTypedWithOptions(t, StylePropObject, stylePropObjectFile, testCase.sourceText, stylePropObjectOptions(t, ``))
			rule_testing.ExpectFindings(t, result, "stylePropNotObject")

			// RunTyped trims the fixture and appends a newline, so the literal above and the file
			// on disk differ by any leading whitespace. Slicing the transformed source rather than
			// the literal is what keeps this assertion honest.
			source := strings.TrimSpace(testCase.sourceText) + "\n"
			span := result.Diagnostics[0].Range
			if got := source[span.Pos():span.End()]; got != testCase.want {
				t.Fatalf("finding spans %q, want %q", got, testCase.want)
			}
		})
	}
}

// TestStylePropObjectMessageText asserts the rendered finding exactly.
//
// The message carries no format verbs, so the assertion is equality against a literal typed here
// rather than against the rule's own constant, which would move with any mutation of it.
func TestStylePropObjectMessageText(t *testing.T) {
	t.Parallel()

	result := rule_testing.RunTypedWithOptions(t, StylePropObject, stylePropObjectFile, `<div style="x" />`, stylePropObjectOptions(t, ``))
	rule_testing.ExpectFindings(t, result, "stylePropNotObject")

	if got := stylePropObjectMessage.Description; got != "Style prop value must be an object" {
		t.Fatalf("message description is %q", got)
	}
	if got := stylePropObjectMessage.Id; got != "stylePropNotObject" {
		t.Fatalf("message id is %q", got)
	}
	if got := result.Diagnostics[0].Message.Description; got != "Style prop value must be an object" {
		t.Fatalf("reported message is %q", got)
	}
}

// TestStylePropObjectHasNoFileSuffixGate pins that a plain `.ts` file is examined.
//
// Three siblings in this package gate on `.tsx` and `.jsx`, which is oxc residue. Upstream has no
// such gate anywhere. The createElement arm needs no JSX, so unlike a JSX-only rule this one can
// state the point on `.ts` directly, where TypeScript would read an opening angle bracket as a type
// assertion rather than as an element.
func TestStylePropObjectHasNoFileSuffixGate(t *testing.T) {
	t.Parallel()

	const source = `React.createElement("div", { style: "x" });`

	for _, suffix := range []string{".tsx", ".ts"} {
		t.Run(suffix, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTypedWithOptions(t, StylePropObject, "/repository/source/SuffixProbe"+suffix, source, stylePropObjectOptions(t, ``))
			rule_testing.ExpectFindings(t, result, "stylePropNotObject")
		})
	}
}

// TestStylePropObjectRequiresTheTypedHarness pins that the checker is genuinely needed.
//
// The plain harness hands the rule a nil checker. Both arms guard on it, so the rule goes silent
// rather than crashing, and a silent rule looks exactly like a correct one to a StaysSilent
// fixture. This asserts the difference directly, so a later revert of the declaration fails loudly
// instead of turning every Fires case into a vacuous pass.
func TestStylePropObjectRequiresTheTypedHarness(t *testing.T) {
	t.Parallel()

	const source = `React.createElement("div", { style: "x" });`
	options := stylePropObjectOptions(t, ``)

	untyped := rule_testing.RunWithOptions(t, StylePropObject, stylePropObjectFile, source, options)
	rule_testing.ExpectClean(t, untyped)

	typed := rule_testing.RunTypedWithOptions(t, StylePropObject, stylePropObjectFile, source, options)
	rule_testing.ExpectFindings(t, typed, "stylePropNotObject")

	// The JSX arm needs the checker only for the identifier case, so a plain literal attribute
	// reports either way. Without this row the test above would read as "the rule does nothing
	// untyped", which is true of the createElement arm and false of the rule.
	jsxUntyped := rule_testing.RunWithOptions(t, StylePropObject, stylePropObjectFile, `<div style="x" />`, options)
	rule_testing.ExpectFindings(t, jsxUntyped, "stylePropNotObject")
}

// TestStylePropObjectIsNonNullLiteral probes the literal predicate directly.
//
// The set of nodes upstream's parser calls a Literal is not obvious, and the two answers a port is
// most likely to get wrong, a regular expression and a bigint, are written nowhere in the corpus.
// Each expectation was measured by driving the installed rule on the corresponding value.
func TestStylePropObjectIsNonNullLiteral(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		wantFires  bool
	}{
		{"a string", `<div style={"x"} />`, true},
		{"a number", `<div style={5} />`, true},
		{"a bigint", `<div style={1n} />`, true},
		{"a regular expression", `<div style={/x/} />`, true},
		{"true", `<div style={true} />`, true},
		{"false", `<div style={false} />`, true},
		{"null", `<div style={null} />`, false},
		{"undefined", `<div style={undefined} />`, false},
		{"a template", "const q = <div style={`x`} />;", false},
		{"a negated number", `<div style={-5} />`, false},
		{"an array", `<div style={[]} />`, false},
		{"an object", `<div style={{}} />`, false},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTypedWithOptions(t, StylePropObject, stylePropObjectFile, testCase.sourceText, stylePropObjectOptions(t, ``))
			if testCase.wantFires {
				rule_testing.ExpectFindings(t, result, "stylePropNotObject")
			} else {
				rule_testing.ExpectClean(t, result)
			}
		})
	}
}

// TestDecodeStylePropObjectOptions covers the allow list and the nil-options path.
func TestDecodeStylePropObjectOptions(t *testing.T) {
	t.Parallel()

	t.Run("an empty body allows nothing", func(t *testing.T) {
		t.Parallel()
		decoded, err := DecodeStylePropObjectOptions(nil)
		if err != nil {
			t.Fatalf("decoding an empty body: %v", err)
		}
		options, ok := decoded.(StylePropObjectOptions)
		if !ok {
			t.Fatalf("decoded to %T, want StylePropObjectOptions", decoded)
		}
		if len(options.Allow) != 0 {
			t.Fatalf("allow list is %v, want empty", options.Allow)
		}
	})

	t.Run("a named allow list", func(t *testing.T) {
		t.Parallel()
		decoded, err := DecodeStylePropObjectOptions([]byte(`{"allow":["MyComponent","Other"]}`))
		if err != nil {
			t.Fatalf("decoding a named allow list: %v", err)
		}
		options := decoded.(StylePropObjectOptions)
		if len(options.Allow) != 2 || options.Allow[0] != "MyComponent" || options.Allow[1] != "Other" {
			t.Fatalf("allow list is %v", options.Allow)
		}
	})
}
