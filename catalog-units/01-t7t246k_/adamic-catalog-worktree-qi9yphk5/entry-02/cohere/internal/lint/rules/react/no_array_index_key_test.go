package react

import (
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// noArrayIndexKeyFile is where the fixtures pretend to live.
//
// A .tsx extension because most cases hold JSX. The rule has no suffix gate; see the note on the
// suffix test below for why only one extension is expressible under the typed harness.
const noArrayIndexKeyFile = "/repository/source/Keys.tsx"

// The corpus is upstream's, extracted mechanically rather than retyped.
//
// Every case in the two tables below comes from
// /tmp/lint-sources/eslint-plugin-react/tests/lib/rules/no-array-index-key.js. The invalid array is
// built with `[].concat(...)` rather than as a plain literal, which a brace-matching extractor
// walks straight past and silently reports as empty; it was read by stubbing the tester and
// capturing the object it was handed, so both `concat` and the template literals inside it are
// handled by the JavaScript engine rather than by a parser written here.
//
// Upstream carries 30 valid and 41 invalid cases, each invalid one naming exactly one messageId,
// which is 41 findings from 41 inputs.
//
// Four cases contain a backtick and so cannot be a Go raw string. Those are emitted as interpreted
// strings whose escaping was derived from the decoded bytes rather than typed, which is the one
// place this file departs from raw strings and the reason is mechanical.

// TestNoArrayIndexKeyFires runs the forty-one failing cases from upstream.
func TestNoArrayIndexKeyFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		wantIds    []string
	}{
		{"invalid 0 foo.map((bar, i) => <Foo key={i} />)", `foo.map((bar, i) => <Foo key={i} />)`, []string{"noArrayIndex"}},
		{"invalid 1 [{}, {}].map((bar, i) => <Foo key={i} />)", `[{}, {}].map((bar, i) => <Foo key={i} />)`, []string{"noArrayIndex"}},
		{"invalid 2 foo.map((bar, anything) => <Foo key={anything} />)", `foo.map((bar, anything) => <Foo key={anything} />)`, []string{"noArrayIndex"}},
		{"invalid 3 foo.map((bar, i) => <Foo key={`foo-${i}`} />)", "foo.map((bar, i) => <Foo key={`foo-${i}`} />)", []string{"noArrayIndex"}},
		{"invalid 4 foo.map((bar, i) => <Foo key={'foo-' + i} />)", `foo.map((bar, i) => <Foo key={'foo-' + i} />)`, []string{"noArrayIndex"}},
		{"invalid 5 foo.map((bar, i) => <Foo key={'foo-' + i + '-bar'} />)", `foo.map((bar, i) => <Foo key={'foo-' + i + '-bar'} />)`, []string{"noArrayIndex"}},
		{"invalid 6 foo.map((baz, i) => React.cloneElement(someChild, { ...someChi", `foo.map((baz, i) => React.cloneElement(someChild, { ...someChild.props, key: i }))`, []string{"noArrayIndex"}},
		{"invalid 7 import { cloneElement } from 'react'; foo.map((baz, i) => clon", `
        import { cloneElement } from 'react';

        foo.map((baz, i) => cloneElement(someChild, { ...someChild.props, key: i }))
      `, []string{"noArrayIndex"}},
		{"invalid 8 foo.map((item, i) => { return React.cloneElement(someChild, { ", `
        foo.map((item, i) => {
          return React.cloneElement(someChild, {
            key: i
          })
        })
      `, []string{"noArrayIndex"}},
		{"invalid 9 import { cloneElement } from 'react'; foo.map((item, i) => { r", `
        import { cloneElement } from 'react';

        foo.map((item, i) => {
          return cloneElement(someChild, {
            key: i
          })
        })
      `, []string{"noArrayIndex"}},
		{"invalid 10 foo.forEach((bar, i) => { baz.push(<Foo key={i} />); })", `foo.forEach((bar, i) => { baz.push(<Foo key={i} />); })`, []string{"noArrayIndex"}},
		{"invalid 11 foo.filter((bar, i) => { baz.push(<Foo key={i} />); })", `foo.filter((bar, i) => { baz.push(<Foo key={i} />); })`, []string{"noArrayIndex"}},
		{"invalid 12 foo.some((bar, i) => { baz.push(<Foo key={i} />); })", `foo.some((bar, i) => { baz.push(<Foo key={i} />); })`, []string{"noArrayIndex"}},
		{"invalid 13 foo.every((bar, i) => { baz.push(<Foo key={i} />); })", `foo.every((bar, i) => { baz.push(<Foo key={i} />); })`, []string{"noArrayIndex"}},
		{"invalid 14 foo.find((bar, i) => { baz.push(<Foo key={i} />); })", `foo.find((bar, i) => { baz.push(<Foo key={i} />); })`, []string{"noArrayIndex"}},
		{"invalid 15 foo.findIndex((bar, i) => { baz.push(<Foo key={i} />); })", `foo.findIndex((bar, i) => { baz.push(<Foo key={i} />); })`, []string{"noArrayIndex"}},
		{"invalid 16 foo.reduce((a, b, i) => a.concat(<Foo key={i} />), [])", `foo.reduce((a, b, i) => a.concat(<Foo key={i} />), [])`, []string{"noArrayIndex"}},
		{"invalid 17 foo.flatMap((a, i) => <Foo key={i} />)", `foo.flatMap((a, i) => <Foo key={i} />)`, []string{"noArrayIndex"}},
		{"invalid 18 foo.reduceRight((a, b, i) => a.concat(<Foo key={i} />), [])", `foo.reduceRight((a, b, i) => a.concat(<Foo key={i} />), [])`, []string{"noArrayIndex"}},
		{"invalid 19 foo.map((bar, i) => React.createElement('Foo', { key: i }))", `foo.map((bar, i) => React.createElement('Foo', { key: i }))`, []string{"noArrayIndex"}},
		{"invalid 20 foo.map((bar, i) => React.createElement('Foo', { key: `foo-${i", "foo.map((bar, i) => React.createElement('Foo', { key: `foo-${i}` }))", []string{"noArrayIndex"}},
		{"invalid 21 foo.map((bar, i) => React.createElement('Foo', { key: 'foo-' +", `foo.map((bar, i) => React.createElement('Foo', { key: 'foo-' + i }))`, []string{"noArrayIndex"}},
		{"invalid 22 foo.map((bar, i) => React.createElement('Foo', { key: 'foo-' +", `foo.map((bar, i) => React.createElement('Foo', { key: 'foo-' + i + '-bar' }))`, []string{"noArrayIndex"}},
		{"invalid 23 foo.forEach((bar, i) => { baz.push(React.createElement('Foo', ", `foo.forEach((bar, i) => { baz.push(React.createElement('Foo', { key: i })); })`, []string{"noArrayIndex"}},
		{"invalid 24 foo.filter((bar, i) => { baz.push(React.createElement('Foo', {", `foo.filter((bar, i) => { baz.push(React.createElement('Foo', { key: i })); })`, []string{"noArrayIndex"}},
		{"invalid 25 foo.some((bar, i) => { baz.push(React.createElement('Foo', { k", `foo.some((bar, i) => { baz.push(React.createElement('Foo', { key: i })); })`, []string{"noArrayIndex"}},
		{"invalid 26 foo.every((bar, i) => { baz.push(React.createElement('Foo', { ", `foo.every((bar, i) => { baz.push(React.createElement('Foo', { key: i })); })`, []string{"noArrayIndex"}},
		{"invalid 27 foo.find((bar, i) => { baz.push(React.createElement('Foo', { k", `foo.find((bar, i) => { baz.push(React.createElement('Foo', { key: i })); })`, []string{"noArrayIndex"}},
		{"invalid 28 foo.findIndex((bar, i) => { baz.push(React.createElement('Foo'", `foo.findIndex((bar, i) => { baz.push(React.createElement('Foo', { key: i })); })`, []string{"noArrayIndex"}},
		{"invalid 29 Children.map(this.props.children, (child, index) => { return R", `
        Children.map(this.props.children, (child, index) => {
          return React.cloneElement(child, { key: index });
        })
      `, []string{"noArrayIndex"}},
		{"invalid 30 import { cloneElement } from 'react'; Children.map(this.props.", `
        import { cloneElement } from 'react';

        Children.map(this.props.children, (child, index) => {
          return cloneElement(child, { key: index });
        })
      `, []string{"noArrayIndex"}},
		{"invalid 31 React.Children.map(this.props.children, (child, index) => { re", `
        React.Children.map(this.props.children, (child, index) => {
          return React.cloneElement(child, { key: index });
        })
      `, []string{"noArrayIndex"}},
		{"invalid 32 import { cloneElement } from 'react'; React.Children.map(this.", `
        import { cloneElement } from 'react';

        React.Children.map(this.props.children, (child, index) => {
          return cloneElement(child, { key: index });
        })
      `, []string{"noArrayIndex"}},
		{"invalid 33 Children.forEach(this.props.children, (child, index) => { retu", `
        Children.forEach(this.props.children, (child, index) => {
          return React.cloneElement(child, { key: index });
        })
      `, []string{"noArrayIndex"}},
		{"invalid 34 import { cloneElement } from 'react'; Children.forEach(this.pr", `
        import { cloneElement } from 'react';

        Children.forEach(this.props.children, (child, index) => {
          return cloneElement(child, { key: index });
        })
      `, []string{"noArrayIndex"}},
		{"invalid 35 React.Children.forEach(this.props.children, (child, index) => ", `
        React.Children.forEach(this.props.children, (child, index) => {
          return React.cloneElement(child, { key: index });
        })
      `, []string{"noArrayIndex"}},
		{"invalid 36 import { cloneElement } from 'react'; React.Children.forEach(t", `
        import { cloneElement } from 'react';

        React.Children.forEach(this.props.children, (child, index) => {
          return cloneElement(child, { key: index });
        })
      `, []string{"noArrayIndex"}},
		{"invalid 37 foo?.map((child, i) => <Foo key={i} />)", `foo?.map((child, i) => <Foo key={i} />)`, []string{"noArrayIndex"}},
		{"invalid 38 foo.map((bar, index) => ( <Element key={index.toString()} bar=", `
        foo.map((bar, index) => (
          <Element key={index.toString()} bar={bar} />
        ))
      `, []string{"noArrayIndex"}},
		{"invalid 39 foo.map((bar, index) => ( <Element key={String(index)} bar={ba", `
        foo.map((bar, index) => (
          <Element key={String(index)} bar={bar} />
        ))
      `, []string{"noArrayIndex"}},
		{"invalid 40 foo.map((bar, index) => ( <Element key={index} bar={bar} /> ))", `
        foo.map((bar, index) => (
          <Element key={index} bar={bar} />
        ))
      `, []string{"noArrayIndex"}},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, NoArrayIndexKey, noArrayIndexKeyFile, testCase.sourceText)
			rule_testing.ExpectFindings(t, result, testCase.wantIds...)
		})
	}
}

// TestNoArrayIndexKeyStaysSilent runs the thirty passing cases from upstream.
//
// These are the false positives upstream already thought about, and several are near misses rather
// than obvious passes: a key built from the item rather than the index, an index used as a
// subscript rather than as the key, and a callback with too few parameters to have an index at all.
func TestNoArrayIndexKeyStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"valid 0 <Foo key='foo' />;", `<Foo key="foo" />;`},
		{"valid 1 <Foo key={i} />;", `<Foo key={i} />;`},
		{"valid 2 <Foo key />;", `<Foo key />;`},
		{"valid 3 <Foo key={`foo-${i}`} />;", "<Foo key={`foo-${i}`} />;"},
		{"valid 4 <Foo key={'foo-' + i} />;", `<Foo key={'foo-' + i} />;`},
		{"valid 5 foo.bar((baz, i) => <Foo key={i} />)", `foo.bar((baz, i) => <Foo key={i} />)`},
		{"valid 6 foo.bar((bar, i) => <Foo key={`foo-${i}`} />)", "foo.bar((bar, i) => <Foo key={`foo-${i}`} />)"},
		{"valid 7 foo.bar((bar, i) => <Foo key={'foo-' + i} />)", `foo.bar((bar, i) => <Foo key={'foo-' + i} />)`},
		{"valid 8 foo.map((baz) => <Foo key={baz.id} />)", `foo.map((baz) => <Foo key={baz.id} />)`},
		{"valid 9 foo.map((baz, i) => <Foo key={baz.id} />)", `foo.map((baz, i) => <Foo key={baz.id} />)`},
		{"valid 10 foo.map((baz, i) => <Foo key={'foo' + baz.id} />)", `foo.map((baz, i) => <Foo key={'foo' + baz.id} />)`},
		{"valid 11 foo.map((baz, i) => React.cloneElement(someChild, { ...someChi", `foo.map((baz, i) => React.cloneElement(someChild, { ...someChild.props }))`},
		{"valid 12 foo.map((baz, i) => cloneElement(someChild, { ...someChild.pro", `foo.map((baz, i) => cloneElement(someChild, { ...someChild.props }))`},
		{"valid 13 foo.map((item, i) => { return React.cloneElement(someChild, { ", `
        foo.map((item, i) => {
          return React.cloneElement(someChild, {
            key: item.id
          })
        })
      `},
		{"valid 14 foo.map((item, i) => { return cloneElement(someChild, { key: i", `
        foo.map((item, i) => {
          return cloneElement(someChild, {
            key: item.id
          })
        })
      `},
		{"valid 15 foo.map((baz, i) => <Foo key />)", `foo.map((baz, i) => <Foo key />)`},
		{"valid 16 foo.reduce((a, b) => a.concat(<Foo key={b.id} />), [])", `foo.reduce((a, b) => a.concat(<Foo key={b.id} />), [])`},
		{"valid 17 foo.map((bar, i) => <Foo key={i.baz.toString()} />)", `foo.map((bar, i) => <Foo key={i.baz.toString()} />)`},
		{"valid 18 foo.map((bar, i) => <Foo key={i.toString} />)", `foo.map((bar, i) => <Foo key={i.toString} />)`},
		{"valid 19 foo.map((bar, i) => <Foo key={String()} />)", `foo.map((bar, i) => <Foo key={String()} />)`},
		{"valid 20 foo.map((bar, i) => <Foo key={String(baz)} />)", `foo.map((bar, i) => <Foo key={String(baz)} />)`},
		{"valid 21 foo.flatMap((a) => <Foo key={a} />)", `foo.flatMap((a) => <Foo key={a} />)`},
		{"valid 22 foo.reduce((a, b, i) => a.concat(<Foo key={b.id} />), [])", `foo.reduce((a, b, i) => a.concat(<Foo key={b.id} />), [])`},
		{"valid 23 foo.reduceRight((a, b) => a.concat(<Foo key={b.id} />), [])", `foo.reduceRight((a, b) => a.concat(<Foo key={b.id} />), [])`},
		{"valid 24 foo.reduceRight((a, b, i) => a.concat(<Foo key={b.id} />), [])", `foo.reduceRight((a, b, i) => a.concat(<Foo key={b.id} />), [])`},
		{"valid 25 React.Children.map(this.props.children, (child, index, arr) =>", `
        React.Children.map(this.props.children, (child, index, arr) => {
          return React.cloneElement(child, { key: child.id });
        })
      `},
		{"valid 26 React.Children.map(this.props.children, (child, index, arr) =>", `
        React.Children.map(this.props.children, (child, index, arr) => {
          return cloneElement(child, { key: child.id });
        })
      `},
		{"valid 27 Children.forEach(this.props.children, (child, index, arr) => {", `
        Children.forEach(this.props.children, (child, index, arr) => {
          return React.cloneElement(child, { key: child.id });
        })
      `},
		{"valid 28 Children.forEach(this.props.children, (child, index, arr) => {", `
        Children.forEach(this.props.children, (child, index, arr) => {
          return cloneElement(child, { key: child.id });
        })
      `},
		{"valid 29 foo?.map(child => <Foo key={child.i} />)", `foo?.map(child => <Foo key={child.i} />)`},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, NoArrayIndexKey, noArrayIndexKeyFile, testCase.sourceText)
			rule_testing.ExpectClean(t, result)
		})
	}
}

// TestNoArrayIndexKeyRequiresTheTypedHarness pins that the rule declares the checker.
//
// Only the imported-cloneElement path needs it, and that path is a small share of the rule, so a
// later revert of the declaration would leave most fixtures green while silently losing that arm.
// The listener guards on a nil checker, so the plain harness makes the whole rule silent rather
// than crashing, which is the quieter and more dangerous failure.
func TestNoArrayIndexKeyRequiresTheTypedHarness(t *testing.T) {
	t.Parallel()

	source := "foo.map((bar, i) => <Foo key={i} />)\n"

	typed := rule_testing.RunTyped(t, NoArrayIndexKey, noArrayIndexKeyFile, source)
	rule_testing.ExpectFindings(t, typed, "noArrayIndex")

	untyped := rule_testing.Run(t, NoArrayIndexKey, noArrayIndexKeyFile, source)
	rule_testing.ExpectClean(t, untyped)
}

// TestNoArrayIndexKeyHasNoFileSuffixGate pins that no filename decides anything.
//
// Three rules in this package carried a `.tsx`/`.jsx` gate as oxc residue and it was removed;
// upstream gates nothing. Measured on the installed build under .tsx, .jsx and .js: all three
// report.
//
// Only the .tsx row is expressible here, and the reason is the harness rather than the rule.
// `rule_testing.RunTyped` writes a tsconfig whose include list is `["**/*.ts", "**/*.tsx"]`
// (internal/rule_testing/program.go:30), so a .jsx or .js fixture is not in the program at all and
// the run fails before any rule is offered a file. Weakening the rule to make such a fixture pass
// would turn a fact about the harness into a fact about the rule.
func TestNoArrayIndexKeyHasNoFileSuffixGate(t *testing.T) {
	t.Parallel()

	source := "foo.map((bar, i) => <Foo key={i} />)\n"
	result := rule_testing.RunTyped(t, NoArrayIndexKey, "/repository/source/Suffix.tsx", source)
	rule_testing.ExpectFindings(t, result, "noArrayIndex")
}

// TestNoArrayIndexKeyIteratorMethods pins the ten methods and the position each one's index sits at.
//
// The two folding methods are the reason the position is a table rather than a constant: their
// index is the third parameter, and a port using position one everywhere is silent exactly where
// upstream is silent, so only the two-parameter fold separates them. The corpus does test both
// folds, and the near-miss rows below are not in it.
func TestNoArrayIndexKeyIteratorMethods(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		wantIds    []string
	}{
		{"map", "foo.map((bar, i) => <Foo key={i} />)\n", []string{"noArrayIndex"}},
		{"forEach", "foo.forEach((bar, i) => <Foo key={i} />)\n", []string{"noArrayIndex"}},
		{"filter", "foo.filter((bar, i) => <Foo key={i} />)\n", []string{"noArrayIndex"}},
		{"every", "foo.every((bar, i) => <Foo key={i} />)\n", []string{"noArrayIndex"}},
		{"some", "foo.some((bar, i) => <Foo key={i} />)\n", []string{"noArrayIndex"}},
		{"find", "foo.find((bar, i) => <Foo key={i} />)\n", []string{"noArrayIndex"}},
		{"findIndex", "foo.findIndex((bar, i) => <Foo key={i} />)\n", []string{"noArrayIndex"}},
		{"flatMap", "foo.flatMap((bar, i) => <Foo key={i} />)\n", []string{"noArrayIndex"}},

		// The folds take the accumulator first, so their index is at position two.
		{"reduce at position two", "foo.reduce((acc, bar, i) => <Foo key={i} />)\n", []string{"noArrayIndex"}},
		{"reduceRight at position two", "foo.reduceRight((acc, bar, i) => <Foo key={i} />)\n", []string{"noArrayIndex"}},

		// The row that separates a table from a constant: a two-parameter fold's second parameter
		// is the ITEM, not an index, so nothing is pushed.
		{"reduce with only two parameters", "foo.reduce((acc, bar) => <Foo key={bar} />)\n", nil},

		// Not in the set at all.
		{"a method that is not an iterator", "foo.each((bar, i) => <Foo key={i} />)\n", nil},
		{"sort", "foo.sort((bar, i) => <Foo key={i} />)\n", nil},

		// Too few parameters to have an index.
		{"map with one parameter", "foo.map((bar) => <Foo key={bar} />)\n", nil},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, NoArrayIndexKey, noArrayIndexKeyFile, testCase.sourceText)
			rule_testing.ExpectFindings(t, result, testCase.wantIds...)
		})
	}
}

// TestNoArrayIndexKeyCalleeShapes pins the shapes a callee can take, which is where a port panics
// rather than where it mis-decides.
//
// `ast.Node.Text()` panics on a property access and on a call expression, and the walk recovers per
// FILE rather than per rule, so one such node costs every rule in the package every finding in that
// file while the run still prints a plausible summary. Each row below reaches the callee inspection
// with a different node kind, and two of them would panic if the kind were not checked before the
// text was read.
//
// No `ExpectFindings` fixture can see a panic, so these are as much a crash test as a behaviour
// test, and the two answers differ: the computed form is silent upstream and the call-as-callee
// form reports, so a guard that declined both would be wrong in one direction.
func TestNoArrayIndexKeyCalleeShapes(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		wantIds    []string
	}{
		{"a plain member callee", "foo.map((bar, i) => <Foo key={i} />)\n", []string{"noArrayIndex"}},
		{"a deep member callee", "a.b.c.map((bar, i) => <Foo key={i} />)\n", []string{"noArrayIndex"}},
		{"a call as the receiver", "getFoo().map((bar, i) => <Foo key={i} />)\n", []string{"noArrayIndex"}},
		{"an optional call", "foo?.map((bar, i) => <Foo key={i} />)\n", []string{"noArrayIndex"}},
		{"an optional member receiver", "foo.bar?.map((baz, i) => <Foo key={i} />)\n", []string{"noArrayIndex"}},

		// A computed access has no identifier property, so upstream reads `callee.property.type`
		// as something other than Identifier and declines. Measured silent.
		{"a computed callee", "foo[\"map\"]((bar, i) => <Foo key={i} />)\n", nil},

		// A bare identifier callee is not a member expression at all.
		{"a bare identifier callee", "map((bar, i) => <Foo key={i} />)\n", nil},

		// A callback passed by reference rather than written inline.
		{"a callback reference", "foo.map(cb)\n", nil},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, NoArrayIndexKey, noArrayIndexKeyFile, testCase.sourceText)
			rule_testing.ExpectFindings(t, result, testCase.wantIds...)
		})
	}
}

// TestNoArrayIndexKeyParameterShapes pins what counts as an index parameter name.
//
// Upstream reads `params[n].name`, which is undefined for anything that is not a plain identifier,
// so a destructuring pattern and a default value are both silent. The default row is the surprising
// one: `(bar, i = 0)` reads as an index to a human and is an assignment pattern to the parser.
// Both measured against the installed build.
func TestNoArrayIndexKeyParameterShapes(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		wantIds    []string
	}{
		{"a plain identifier parameter", "foo.map((bar, i) => <Foo key={i} />)\n", []string{"noArrayIndex"}},
		{"a destructured index parameter", "foo.map((bar, {i}) => <Foo key={i} />)\n", nil},
		{"an index parameter with a default", "foo.map((bar, i = 0) => <Foo key={i} />)\n", nil},
		{"a rest parameter", "foo.map((...args) => <Foo key={args} />)\n", nil},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, NoArrayIndexKey, noArrayIndexKeyFile, testCase.sourceText)
			rule_testing.ExpectFindings(t, result, testCase.wantIds...)
		})
	}
}

// TestNoArrayIndexKeyKeyExpressionShapes pins the five shapes upstream accepts, and the near misses.
//
// The near misses are the load-bearing half. Upstream hardcodes `String` and `toString` rather than
// reasoning about conversion, so `Number(i)` and `i.toFixed()` are clean even though both convert an
// index to a key; and an element access is not one of the five shapes even though the index is
// plainly inside it. A port that generalized any of these would report on code upstream leaves
// alone, and nothing in the corpus writes `Number(i)` or `i.toFixed()`.
func TestNoArrayIndexKeyKeyExpressionShapes(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		key     string
		wantIds []string
	}{
		{"the index itself", "{i}", []string{"noArrayIndex"}},
		{"a template naming the index", "{`x-${i}`}", []string{"noArrayIndex"}},
		{"a template not naming it", "{`x-${bar}`}", nil},
		{"a binary expression", "{'x-' + i}", []string{"noArrayIndex"}},
		{"a nested binary expression", "{'x-' + i + '-y'}", []string{"noArrayIndex"}},

		// A LOGICAL operator is a different node type upstream and is not dispatched to the binary
		// arm, so all three of these are silent there while being `KindBinaryExpression` here.
		// Found on the real tree rather than in a fixture: the dry run produced four findings
		// ESLint does not, every one of them `key={item.id ?? index}`, which is an ordinary thing
		// to write and which no imported case covers. All measured silent on the installed build.
		{"nullish coalescing with the index on the right", "{bar.id ?? i}", nil},
		{"logical or with the index on the right", "{bar.id || i}", nil},
		{"logical and with the index on the right", "{bar.id && i}", nil},

		// And the subtree of a logical operator is not walked either, so an arithmetic expression
		// nested inside one stays silent. Without that the collector would find the index through
		// the logical node it was meant to decline.
		{"an arithmetic expression nested in a logical one", "{bar.id ?? ('x-' + i)}", nil},
		{"toString on the index", "{i.toString()}", []string{"noArrayIndex"}},
		{"String of the index", "{String(i)}", []string{"noArrayIndex"}},

		// The near misses.
		{"Number of the index", "{Number(i)}", nil},
		{"toFixed on the index", "{i.toFixed()}", nil},
		{"toString on something else", "{bar.toString()}", nil},
		{"the index as a subscript", "{bar[i]}", nil},
		{"a string literal key", "\"x\"", nil},
		{"a shorthand key", "", nil},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			attribute := "key"
			if testCase.key != "" {
				attribute = "key=" + testCase.key
			}
			source := "foo.map((bar, i) => <Foo " + attribute + " />)\n"
			result := rule_testing.RunTyped(t, NoArrayIndexKey, noArrayIndexKeyFile, source)
			rule_testing.ExpectFindings(t, result, testCase.wantIds...)
		})
	}
}

// TestNoArrayIndexKeyStackBehaviour pins that the index names form a stack rather than one name.
//
// Nothing in the corpus nests one iterator inside another, so a port holding a single name passes
// every imported case and gets the first row here wrong. The last two rows are the other half of
// the stack: a name is in scope only inside its own callback, so the same identifier is clean
// before and after.
func TestNoArrayIndexKeyStackBehaviour(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		wantIds    []string
	}{
		{
			"the outer index used inside a nested callback",
			"foo.map((a, i) => bar.map((b, j) => <Foo key={i} />))\n",
			[]string{"noArrayIndex"},
		},
		{
			"the inner index used inside a nested callback",
			"foo.map((a, i) => bar.map((b, j) => <Foo key={j} />))\n",
			[]string{"noArrayIndex"},
		},
		{
			"both indexes used in one nested callback",
			"foo.map((a, i) => bar.map((b, j) => <Foo key={`${i}-${j}`} />))\n",
			[]string{"noArrayIndex", "noArrayIndex"},
		},
		{
			"a same-named identifier outside any callback",
			"const i = 1;\nconst element = <Foo key={i} />;\n",
			nil,
		},
		{
			"a same-named identifier after the callback has closed",
			"foo.map((bar, i) => null);\nconst element = <Foo key={i} />;\n",
			nil,
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, NoArrayIndexKey, noArrayIndexKeyFile, testCase.sourceText)
			rule_testing.ExpectFindings(t, result, testCase.wantIds...)
		})
	}
}

// TestNoArrayIndexKeyReactChildren pins the argument shift for the React children iterator.
//
// `Children.map(collection, callback)` takes the callback SECOND, so a port reading argument zero
// finds the collection and pushes nothing. Upstream detects this by the receiver's name, and the
// last row shows how narrow that test is: any other receiver keeps the ordinary argument order, so
// the same two-argument shape is silent.
func TestNoArrayIndexKeyReactChildren(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		wantIds    []string
	}{
		{"Children.map", "Children.map(c, (child, i) => <Foo key={i} />)\n", []string{"noArrayIndex"}},
		{"React.Children.map", "React.Children.map(c, (child, i) => <Foo key={i} />)\n", []string{"noArrayIndex"}},
		{"React.Children.forEach", "React.Children.forEach(c, (child, i) => <Foo key={i} />)\n", []string{"noArrayIndex"}},

		// The callback in argument zero is not read for a Children call.
		{"Children.map with the callback first", "Children.map((child, i) => <Foo key={i} />)\n", nil},

		// Any other receiver keeps the ordinary order, so argument zero is the collection.
		{"another receiver with two arguments", "Other.map(c, (child, i) => <Foo key={i} />)\n", nil},

		// Upstream restricts the argument shift to `map` and `forEach`, which is NARROWER than the
		// ten-method set the caller has already matched. So a Children receiver with any other
		// iterator method keeps the ordinary order, reads argument zero, finds the collection
		// rather than a callback, and pushes nothing. A mutation removing that restriction survived
		// every other fixture here, because the corpus only ever writes Children with map or
		// forEach. All four rows measured silent on the installed build.
		{"Children.filter with two arguments", "Children.filter(c, (child, i) => <Foo key={i} />)\n", nil},
		{"Children.some with two arguments", "Children.some(c, (child, i) => <Foo key={i} />)\n", nil},
		{"Children.reduce with three arguments", "Children.reduce(c, (acc, child, i) => <Foo key={i} />)\n", nil},
		{"React.Children.filter with two arguments", "React.Children.filter(c, (child, i) => <Foo key={i} />)\n", nil},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, NoArrayIndexKey, noArrayIndexKeyFile, testCase.sourceText)
			rule_testing.ExpectFindings(t, result, testCase.wantIds...)
		})
	}
}

// TestNoArrayIndexKeyCreateElement pins the props-object path and the import resolution under it.
//
// The imported form is the only part of this rule that needs the type checker: upstream resolves
// the bare name through its own variable index and requires the import's source to be exactly
// `react`. The last row is what that requirement buys, and nothing in the corpus writes it.
func TestNoArrayIndexKeyCreateElement(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		wantIds    []string
	}{
		{
			"React.createElement with a key",
			"foo.map((bar, i) => React.createElement(\"div\", { key: i }))\n",
			[]string{"noArrayIndex"},
		},
		{
			"React.cloneElement with a key",
			"foo.map((bar, i) => React.cloneElement(c, { key: i }))\n",
			[]string{"noArrayIndex"},
		},
		{
			"a spread before the key",
			"foo.map((bar, i) => React.createElement(\"div\", { ...p, key: i }))\n",
			[]string{"noArrayIndex"},
		},
		{
			"a spread and no key",
			"foo.map((bar, i) => React.createElement(\"div\", { ...p }))\n",
			nil,
		},
		{
			"a computed key property",
			"foo.map((bar, i) => React.createElement(\"div\", { [\"key\"]: i }))\n",
			nil,
		},
		{
			"a single argument",
			"foo.map((bar, i) => React.createElement(\"div\"))\n",
			nil,
		},
		{
			"props that are not an object literal",
			"foo.map((bar, i) => React.createElement(\"div\", p))\n",
			nil,
		},
		{
			"the same call outside any iterator",
			"React.createElement(\"div\", { key: i })\n",
			nil,
		},
		{
			"cloneElement imported from react",
			"import { cloneElement } from 'react';\n" +
				"foo.map((bar, i) => cloneElement(c, { key: i }))\n",
			[]string{"noArrayIndex"},
		},
		{
			// The row the checker is declared for. Same identifier, different module.
			"cloneElement imported from somewhere else",
			"import { cloneElement } from 'other';\n" +
				"foo.map((bar, i) => cloneElement(c, { key: i }))\n",
			nil,
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, NoArrayIndexKey, noArrayIndexKeyFile, testCase.sourceText)
			rule_testing.ExpectFindings(t, result, testCase.wantIds...)
		})
	}
}

// TestNoArrayIndexKeySpans asserts where each finding points.
//
// Four of the five shapes report the whole key expression and `String(i)` reports the ARGUMENT,
// which is upstream passing a different node at that one site. A message-id fixture cannot see the
// difference and the corpus asserts none of it.
//
// `RunTyped` writes `strings.TrimSpace(source) + "\n"` to disk, so the expectation is sliced from
// the same transform rather than from the Go literal, which would be one byte off.
func TestNoArrayIndexKeySpans(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		wantText   string
	}{
		{"the index itself", "foo.map((bar, i) => <Foo key={i} />)\n", "i"},
		{"a template underlines the whole template", "foo.map((bar, i) => <Foo key={`x-${i}`} />)\n", "`x-${i}`"},
		{"a binary expression underlines the whole expression", "foo.map((bar, i) => <Foo key={'x-' + i} />)\n", "'x-' + i"},
		{"toString underlines the whole call", "foo.map((bar, i) => <Foo key={i.toString()} />)\n", "i.toString()"},

		// The one site that differs: the argument rather than the call.
		{"String underlines its argument only", "foo.map((bar, i) => <Foo key={String(i)} />)\n", "i"},

		// A createElement key reports the value expression the same way.
		{
			"a createElement key underlines the value",
			"foo.map((bar, i) => React.createElement(\"div\", { key: i }))\n",
			"i",
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, NoArrayIndexKey, noArrayIndexKeyFile, testCase.sourceText)
			if len(result.Diagnostics) != 1 {
				t.Fatalf("wanted exactly one finding, got %d", len(result.Diagnostics))
			}
			written := strings.TrimSpace(testCase.sourceText) + "\n"
			diagnostic := result.Diagnostics[0]
			got := written[diagnostic.Range.Pos():diagnostic.Range.End()]
			if got != testCase.wantText {
				t.Errorf("span text = %q, wanted %q", got, testCase.wantText)
			}
		})
	}
}

// TestNoArrayIndexKeyMessage asserts the message by identity.
//
// A rule.Message is {Id, Description} with no interpolation, so there is nothing to render.
// Asserted against a literal typed here rather than against the rule's own constant, which would
// move on both sides under mutation and could not fail.
func TestNoArrayIndexKeyMessage(t *testing.T) {
	t.Parallel()

	if messageNoArrayIndexKey.Id != "noArrayIndex" {
		t.Errorf("message id = %q", messageNoArrayIndexKey.Id)
	}
	if len(messageNoArrayIndexKey.Description) < 80 {
		t.Errorf("description is too short to say why: %q", messageNoArrayIndexKey.Description)
	}
}
