package react

import (
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// forbidDomPropsFile is where the fixtures pretend to live.
//
// A .tsx extension because every case holds JSX. The rule has no suffix gate.
const forbidDomPropsFile = "/repository/source/ForbidDomProps.tsx"

// The corpus is upstream's, extracted mechanically rather than retyped.
//
// Every case in the two tables below comes from
// /tmp/lint-sources/eslint-plugin-react/tests/lib/rules/forbid-dom-props.js, read by evaluating the
// two arrays in the tester with a stubbed RuleTester and emitting Go literals from the decoded
// values, so no escape sequence was typed on the way here. Upstream carries 12 valid and 11 invalid
// cases.
//
// # The expected counts are the INSTALLED build's, not the clone's, and they differ on three cases
//
// The clone's working tree carries a `disallowedValues` feature that the installed build, version
// 7.37.5, does not have. Replaying all 23 cases against the installed build through the ESLint
// Linter API agreed on 20 and disagreed on exactly the three that exercise the feature. This port
// reproduces the installed build, because that is the artifact our gate compares against, so those
// three are carried at the installed build's verdict with the clone's verdict named beside each.

// forbidDomPropsOptions decodes a raw option body the way the config layer does.
//
// Fixtures route through the exported decoder rather than building the struct, which is what puts
// the string-or-object union and the absent-versus-empty distinction under test. An empty body is
// the nil-options case: an unconfigured rule declines everything.
func forbidDomPropsOptions(t *testing.T, raw string) any {
	t.Helper()
	decoded, err := DecodeForbidDomPropsOptions([]byte(raw))
	if err != nil {
		t.Fatalf("decoding %q: %v", raw, err)
	}
	return decoded
}

// TestForbidDomPropsFires runs every case this port reports on.
func TestForbidDomPropsFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		rawOptions string
		wantCount  int
	}{
		// the clone calls this clean because an empty disallowedValues list matches no value;
		// the installed build has no such feature and reports on the name alone.
		{"valid 8", `
        const First = (props) => (
          <div someProp="someValue" />
        );
      `, `{"forbid":[{"propName":"someProp","disallowedValues":[]}]}`, 1},
		// the clone calls this clean because the value does not match disallowedValues;
		// the installed build has no such feature and reports on the name alone.
		{"valid 10", `
        const First = (props) => (
          <div someProp="value" />
        );
      `, `{"forbid":[{"propName":"someProp","disallowedValues":["someValue"]}]}`, 1},
		{"invalid 0", `
        var First = createReactClass({
          propTypes: externalPropTypes,
          render: function() {
            return <div id="bar" />;
          }
        });
      `, `{"forbid":["id"]}`, 1},
		{"invalid 1", `
        class First extends createReactClass {
          render() {
            return <div id="bar" />;
          }
        }
      `, `{"forbid":["id"]}`, 1},
		{"invalid 2", `
        const First = (props) => (
          <div id="foo" />
        );
      `, `{"forbid":["id"]}`, 1},
		{"invalid 3", `
        const First = (props) => (
          <div className="foo" />
        );
      `, `{"forbid":[{"propName":"className","message":"Please use class instead of ClassName"}]}`, 1},
		{"invalid 4", `
        const First = (props) => (
          <span otherProp="bar" />
        );
      `, `{"forbid":[{"propName":"otherProp","disallowedFor":["span"]}]}`, 1},
		{"invalid 5", `
        const First = (props) => (
          <div someProp="someValue" />
        );
      `, `{"forbid":[{"propName":"someProp","disallowedValues":["someValue"]}]}`, 1},
		{"invalid 6", `
        const First = (props) => (
          <div className="foo">
            <div otherProp="bar" />
          </div>
        );
      `, `{"forbid":[{"propName":"className","message":"Please use class instead of ClassName"},{"propName":"otherProp","message":"Avoid using otherProp"}]}`, 2},
		{"invalid 7", `
        const First = (props) => (
          <div className="foo">
            <div otherProp="bar" />
          </div>
        );
      `, `{"forbid":[{"propName":"className"},{"propName":"otherProp","message":"Avoid using otherProp"}]}`, 2},
		{"invalid 8", `
        const First = (props) => (
          <form accept='file'>
            <input type="file" id="videoFile" accept="video/*" />
            <input type="hidden" name="fullname" />
          </form>
        );
      `, `{"forbid":[{"propName":"accept","disallowedFor":["form"],"message":"Avoid using the accept attribute on <form>"}]}`, 1},
		{"invalid 9", `
        const First = (props) => (
          <div className="foo">
            <input className="boo" />
            <span className="foobar">Foobar</span>
            <div otherProp="bar" />
          </div>
        );
      `, `{"forbid":[{"propName":"className","disallowedFor":["div","span"],"message":"Please use class instead of ClassName"},{"propName":"otherProp","message":"Avoid using otherProp"}]}`, 3},
		// the clone counts five, filtering two of the thirdProp sites by value;
		// the installed build has no such feature and reports all six by name.
		{"invalid 10", `
        const First = (props) => (
          <div className="foo">
            <input className="boo" />
            <span className="foobar">Foobar</span>
            <div otherProp="bar" />
            <p thirdProp="foo" />
            <div thirdProp="baz" />
            <p thirdProp="bar" />
            <p thirdProp="baz" />
          </div>
        );
      `, `{"forbid":[{"propName":"className","disallowedFor":["div","span"],"message":"Please use class instead of ClassName"},{"propName":"otherProp","message":"Avoid using otherProp"},{"propName":"thirdProp","disallowedFor":["p"],"disallowedValues":["bar","baz"],"message":"Do not use thirdProp with values bar and baz on p"}]}`, 6},
		// Cases upstream's corpus does not write, each measured against the installed build.

		// The value is never read, so a bare prop and an expression container both report. This is
		// the opposite of the sibling jsx-no-script-url, which decides from the value and declines
		// both shapes. A port reading the value here would silently narrow the rule.
		{"a bare prop with no value", `<div someProp />`, `{"forbid":["someProp"]}`, 1},
		{"an expression container value", `<div someProp={x} />`, `{"forbid":["someProp"]}`, 1},
		{"a numeric expression value", `<div someProp={5} />`, `{"forbid":["someProp"]}`, 1},
		// A repeated prop name keeps the LAST entry's restriction. Measured both ways; its partner
		// in the silent table uses the first entry's tag and is clean.
		{"repeated propName keeps the last restriction", `<span className="x" />`, `{"forbid":[{"propName":"className","disallowedFor":["div"]},{"propName":"className","disallowedFor":["span"]}]}`, 1},
		// An empty message string is falsy upstream, so the default text is used rather than an
		// empty finding. Measured. The message-text assertion below pins the text itself.
		{"an empty message falls back to the default text", `<div className="x" />`, `{"forbid":[{"propName":"className","message":""}]}`, 1},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			options := forbidDomPropsOptions(t, testCase.rawOptions)
			result := rule_testing.RunWithOptions(t, ForbidDomProps, forbidDomPropsFile, testCase.sourceText, options)
			wantIds := make([]string, testCase.wantCount)
			for index := range wantIds {
				wantIds[index] = "propIsForbidden"
			}
			rule_testing.ExpectFindings(t, result, wantIds...)
		})
	}
}

// TestForbidDomPropsStaysSilent runs every case this port declines.
func TestForbidDomPropsStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		rawOptions string
	}{
		{"valid 0", `
        var First = createReactClass({
          render: function() {
            return <Foo id="foo" />;
          }
        });
      `, `{"forbid":["id"]}`},
		{"valid 1", `
        var First = createReactClass({
          propTypes: externalPropTypes,
          render: function() {
            return <Foo id="bar" style={{color: "red"}} />;
          }
        });
      `, `{"forbid":["style","id"]}`},
		{"valid 2", `
        var First = createReactClass({
          propTypes: externalPropTypes,
          render: function() {
            return <this.Foo bar="baz" />;
          }
        });
      `, `{"forbid":["id"]}`},
		{"valid 3", `
        class First extends createReactClass {
          render() {
            return <this.foo id="bar" />;
          }
        }
      `, `{"forbid":["id"]}`},
		{"valid 4", `
        const First = (props) => (
          <this.Foo {...props} />
        );
      `, `{"forbid":["id"]}`},
		{"valid 5", `
        const First = (props) => (
          <fbt:param name="name">{props.name}</fbt:param>
        );
      `, `{"forbid":["id"]}`},
		{"valid 6", `
        const First = (props) => (
          <div name="foo" />
        );
      `, `{"forbid":["id"]}`},
		{"valid 7", `
        const First = (props) => (
          <div otherProp="bar" />
        );
      `, `{"forbid":[{"propName":"otherProp","disallowedFor":["span"]}]}`},
		{"valid 9", `
        const First = (props) => (
          <Foo someProp="someValue" />
        );
      `, `{"forbid":[{"propName":"someProp","disallowedValues":["someValue"]}]}`},
		{"valid 11", `
        const First = (props) => (
          <div someProp="someValue" />
        );
      `, `{"forbid":[{"propName":"someProp","disallowedValues":["someValue"],"disallowedFor":["span"]}]}`},
		// Cases upstream's corpus does not write, each measured against the installed build.

		// An empty disallowedFor list is truthy upstream and matches no tag, which is NOT the same
		// as an absent list meaning every tag. Its partner in the firing table omits the key.
		{"an empty disallowedFor matches no tag", `<div className="x" />`, `{"forbid":[{"propName":"className","disallowedFor":[]}]}`},
		// The DOM-node test asks whether the first character CHANGES under upper-casing, not
		// whether it is lowercase. An underscore upper-cases to itself, so this is clean. Measured,
		// and no corpus case writes it.
		{"a tag starting with an underscore is not a DOM node", `<_div someProp="v" />`, `{"forbid":["someProp"]}`},
		{"a capitalised tag is a component", `<Div someProp="v" />`, `{"forbid":["someProp"]}`},
		{"a member tag carries no plain name", `<a.b someProp="v" />`, `{"forbid":["someProp"]}`},
		{"a namespaced tag carries no plain name", `<svg:a someProp="v" />`, `{"forbid":["someProp"]}`},
		{"a namespaced attribute carries no plain name", `<div xlink:someProp="v" />`, `{"forbid":["someProp"]}`},
		{"a spread has no name to match", `<div {...properties} />`, `{"forbid":["someProp"]}`},
		{"no options at all", `<div className="x" />`, ``},
		{"an empty forbid list", `<div className="x" />`, `{"forbid":[]}`},
		// The partner to the repeated-name case: the FIRST entry's tag no longer applies.
		{"repeated propName drops the first restriction", `<div className="x" />`, `{"forbid":[{"propName":"className","disallowedFor":["div"]},{"propName":"className","disallowedFor":["span"]}]}`},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			options := forbidDomPropsOptions(t, testCase.rawOptions)
			result := rule_testing.RunWithOptions(t, ForbidDomProps, forbidDomPropsFile, testCase.sourceText, options)
			rule_testing.ExpectClean(t, result)
		})
	}
}

// TestForbidDomPropsReportsOnTheWholeAttribute pins WHERE the finding points.
//
// Upstream reports on the JSXAttribute, so the span covers the name and the value together rather
// than either alone. A port anchoring on the name node, or on the element, would pass every
// message-id fixture above while pointing somewhere the reader was never shown.
func TestForbidDomPropsReportsOnTheWholeAttribute(t *testing.T) {
	t.Parallel()

	const source = `<div className="foo" />`

	result := rule_testing.RunWithOptions(t, ForbidDomProps, forbidDomPropsFile, source, forbidDomPropsOptions(t, `{"forbid":["className"]}`))
	rule_testing.ExpectFindings(t, result, "propIsForbidden")

	span := result.Diagnostics[0].Range
	reported := source[span.Pos():span.End()]
	if reported != `className="foo"` {
		t.Fatalf("finding spans %q, want the whole attribute", reported)
	}
}

// TestForbidDomPropsMessageText asserts the rendered findings exactly.
//
// The default message interpolates the prop name, so an id assertion cannot see anything the format
// string does. The custom-message branch replaces the text entirely, and an EMPTY custom message
// falls back to the default rather than reporting an empty finding. All three are asserted against
// literals typed here rather than against the rule's own constants, which would move under mutation.
func TestForbidDomPropsMessageText(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		rawOptions string
		want       string
	}{
		{
			"the default message names the prop",
			`{"forbid":["className"]}`,
			`Prop "className" is forbidden on DOM Nodes`,
		},
		{
			"a custom message replaces the text",
			`{"forbid":[{"propName":"className","message":"Please use class instead of ClassName"}]}`,
			"Please use class instead of ClassName",
		},
		{
			"an empty custom message falls back to the default",
			`{"forbid":[{"propName":"className","message":""}]}`,
			`Prop "className" is forbidden on DOM Nodes`,
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunWithOptions(t, ForbidDomProps, forbidDomPropsFile, `<div className="foo" />`, forbidDomPropsOptions(t, testCase.rawOptions))
			rule_testing.ExpectFindings(t, result, "propIsForbidden")
			if got := result.Diagnostics[0].Message.Description; got != testCase.want {
				t.Fatalf("message is %q, want %q", got, testCase.want)
			}
		})
	}
}

// TestForbidDomPropsHasNoFileSuffixGate pins that every extension carrying JSX is examined.
//
// Three siblings in this package gate on `.tsx` and `.jsx`, which is oxc residue. Upstream has no
// such gate anywhere.
//
// `.ts` is absent for a PARSER reason rather than a rule one: TypeScript reads an opening angle
// bracket there as a type assertion, so no JSX attribute node exists for any rule to see. Probed
// with a control across all four extensions on a sibling rule with the same anchor.
func TestForbidDomPropsHasNoFileSuffixGate(t *testing.T) {
	t.Parallel()

	const source = `<div className="foo" />`

	for _, fileName := range []string{
		"/repository/source/ForbidDomProps.tsx",
		"/repository/source/ForbidDomProps.js",
		"/repository/source/ForbidDomProps.jsx",
	} {
		t.Run(fileName, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunWithOptions(t, ForbidDomProps, fileName, source, forbidDomPropsOptions(t, `{"forbid":["className"]}`))
			rule_testing.ExpectFindings(t, result, "propIsForbidden")
		})
	}
}

// TestDecodeForbidDomPropsOptions covers the union and the absent-versus-empty distinction.
//
// The list items are a union of a string and an object, which Go has no single shape for, so the
// decoder tries the string first. `disallowedFor` needs a pointer on the wire, because an absent
// key means every tag and an empty array means no tag, and a plain slice cannot tell them apart.
func TestDecodeForbidDomPropsOptions(t *testing.T) {
	t.Parallel()

	t.Run("an empty body forbids nothing", func(t *testing.T) {
		t.Parallel()
		decoded, err := DecodeForbidDomPropsOptions(nil)
		if err != nil {
			t.Fatalf("decoding an empty body: %v", err)
		}
		options, ok := decoded.(ForbidDomPropsOptions)
		if !ok {
			t.Fatalf("decoded to %T, want ForbidDomPropsOptions", decoded)
		}
		if len(options.Forbid) != 0 {
			t.Fatalf("forbid list is %v, want empty", options.Forbid)
		}
	})

	t.Run("the string arm", func(t *testing.T) {
		t.Parallel()
		decoded, err := DecodeForbidDomPropsOptions([]byte(`{"forbid":["id","style"]}`))
		if err != nil {
			t.Fatalf("decoding the string arm: %v", err)
		}
		options := decoded.(ForbidDomPropsOptions)
		if len(options.Forbid) != 2 {
			t.Fatalf("forbid list is %v, want two entries", options.Forbid)
		}
		for _, entry := range options.Forbid {
			if entry.DisallowedForWasGiven {
				t.Fatalf("entry %q recorded a tag restriction it was never given", entry.PropName)
			}
		}
	})

	t.Run("an absent disallowedFor means every tag", func(t *testing.T) {
		t.Parallel()
		decoded, err := DecodeForbidDomPropsOptions([]byte(`{"forbid":[{"propName":"className"}]}`))
		if err != nil {
			t.Fatalf("decoding an absent disallowedFor: %v", err)
		}
		options := decoded.(ForbidDomPropsOptions)
		if options.Forbid[0].DisallowedForWasGiven {
			t.Fatal("an absent disallowedFor was recorded as given")
		}
		if !forbidDomPropsIsForbidden(options.Forbid[0], "div") {
			t.Fatal("an absent disallowedFor declined a tag, want every tag forbidden")
		}
	})

	t.Run("an empty disallowedFor means no tag", func(t *testing.T) {
		t.Parallel()
		decoded, err := DecodeForbidDomPropsOptions([]byte(`{"forbid":[{"propName":"className","disallowedFor":[]}]}`))
		if err != nil {
			t.Fatalf("decoding an empty disallowedFor: %v", err)
		}
		options := decoded.(ForbidDomPropsOptions)
		if !options.Forbid[0].DisallowedForWasGiven {
			t.Fatal("an empty disallowedFor was not recorded as given, so it reads as absent")
		}
		if forbidDomPropsIsForbidden(options.Forbid[0], "div") {
			t.Fatal("an empty disallowedFor matched a tag, want none")
		}
	})

	t.Run("a named disallowedFor matches only those tags", func(t *testing.T) {
		t.Parallel()
		decoded, err := DecodeForbidDomPropsOptions([]byte(`{"forbid":[{"propName":"className","disallowedFor":["div","span"]}]}`))
		if err != nil {
			t.Fatalf("decoding a named disallowedFor: %v", err)
		}
		entry := decoded.(ForbidDomPropsOptions).Forbid[0]
		if !forbidDomPropsIsForbidden(entry, "div") || !forbidDomPropsIsForbidden(entry, "span") {
			t.Fatal("a named tag was declined")
		}
		if forbidDomPropsIsForbidden(entry, "input") {
			t.Fatal("an unnamed tag matched")
		}
	})
}

// TestForbidDomPropsIsDomNodeName probes the component test at the characters that separate the two
// readings.
//
// Upstream asks whether the first character CHANGES under upper-casing, not whether it is
// lowercase, and the two disagree on every character with no case. Each expectation here was
// measured by driving the installed rule on the corresponding tag.
func TestForbidDomPropsIsDomNodeName(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		want bool
	}{
		{"div", true},
		{"span", true},
		{"a", true},
		{"Div", false},
		{"Foo", false},
		// A character with no case is unchanged by upper-casing, so it is NOT a DOM node. This is
		// the whole difference between upstream's test and an is-lowercase test.
		{"_div", false},
		{"", false},
	}

	for _, testCase := range cases {
		if got := forbidDomPropsIsDomNodeName(testCase.name); got != testCase.want {
			t.Errorf("DOM-node test of %q is %v, want %v", testCase.name, got, testCase.want)
		}
	}
}
