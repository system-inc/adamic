package core

import (
	"encoding/json"
	"testing"

	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/internal/lint/testing"
)

// noRestrictedPropertiesFile is where the fixtures pretend to live.
const noRestrictedPropertiesFile = "/repository/source/NoRestrictedProperties.ts"

// noRestrictedPropertiesCase is one imported corpus row.
type noRestrictedPropertiesCase struct {
	sourceText string
	options    any
	wantIds    []string
}

// runNoRestrictedProperties drives one case through the rule's own exported decoder.
//
// Through the decoder rather than by building the struct, because the decoder is where the three
// schema constraints upstream expresses declaratively are actually checked. A fixture handing the
// rule a struct would leave all three untested.
func runNoRestrictedProperties(
	t *testing.T,
	testCase noRestrictedPropertiesCase,
) rule_testing.Result {
	t.Helper()
	if testCase.options == nil {
		return rule_testing.Run(t, NoRestrictedProperties, noRestrictedPropertiesFile,
			testCase.sourceText)
	}
	// Written the way upstream's config writes it: each restriction its own element.
	encoded, err := json.Marshal(testCase.options.(NoRestrictedPropertiesOptions).Restrictions)
	if err != nil {
		t.Fatalf("could not encode options: %v", err)
	}
	decoded, err := DecodeNoRestrictedPropertiesOptions(encoded)
	if err != nil {
		t.Fatalf("could not decode options: %v", err)
	}
	return rule_testing.RunWithOptions(t, NoRestrictedProperties, noRestrictedPropertiesFile,
		testCase.sourceText, decoded)
}

// The corpus is ESLint's own, extracted mechanically rather than retyped.
//
// `tests/lib/rules/no-restricted-properties.js` was loaded with its RuleTester stubbed so every case
// came out as data with its restrictions attached, then replayed against the INSTALLED rule to
// record what it reports. All 89 reproduced.
//
// Upstream's options are a variadic list of restriction objects, one per element. The struct here
// carries them as a slice and the runner writes that slice back as upstream's list, so each entry
// keeps upstream's own field names and crosses the decoder on the shape the config layer delivers.
func noRestrictedPropertiesFiresCases() []noRestrictedPropertiesCase {
	return []noRestrictedPropertiesCase{
		{"someObject.disallowedProperty", NoRestrictedPropertiesOptions{Restrictions: []NoRestrictedPropertiesRestriction{{Object: "someObject", Property: "disallowedProperty"}}}, []string{"restrictedObjectProperty"}},
		{"someObject.disallowedProperty", NoRestrictedPropertiesOptions{Restrictions: []NoRestrictedPropertiesRestriction{{Object: "someObject", Property: "disallowedProperty", Message: "Please use someObject.allowedProperty instead."}}}, []string{"restrictedObjectProperty"}},
		{"someObject.disallowedProperty; anotherObject.anotherDisallowedProperty()", NoRestrictedPropertiesOptions{Restrictions: []NoRestrictedPropertiesRestriction{{Object: "someObject", Property: "disallowedProperty"}, {Object: "anotherObject", Property: "anotherDisallowedProperty"}}}, []string{"restrictedObjectProperty", "restrictedObjectProperty"}},
		{"foo.__proto__", NoRestrictedPropertiesOptions{Restrictions: []NoRestrictedPropertiesRestriction{{Property: "__proto__", Message: "Please use Object.getPrototypeOf instead."}}}, []string{"restrictedProperty"}},
		{"foo['__proto__']", NoRestrictedPropertiesOptions{Restrictions: []NoRestrictedPropertiesRestriction{{Property: "__proto__", Message: "Please use Object.getPrototypeOf instead."}}}, []string{"restrictedProperty"}},
		{"foo.bar.baz;", NoRestrictedPropertiesOptions{Restrictions: []NoRestrictedPropertiesRestriction{{Object: "foo"}}}, []string{"restrictedObjectProperty"}},
		{"foo.bar();", NoRestrictedPropertiesOptions{Restrictions: []NoRestrictedPropertiesRestriction{{Object: "foo"}}}, []string{"restrictedObjectProperty"}},
		{"foo.bar.baz();", NoRestrictedPropertiesOptions{Restrictions: []NoRestrictedPropertiesRestriction{{Object: "foo"}}}, []string{"restrictedObjectProperty"}},
		{"foo.bar.baz;", NoRestrictedPropertiesOptions{Restrictions: []NoRestrictedPropertiesRestriction{{Property: "bar"}}}, []string{"restrictedProperty"}},
		{"foo.bar();", NoRestrictedPropertiesOptions{Restrictions: []NoRestrictedPropertiesRestriction{{Property: "bar"}}}, []string{"restrictedProperty"}},
		{"foo.bar.baz();", NoRestrictedPropertiesOptions{Restrictions: []NoRestrictedPropertiesRestriction{{Property: "bar"}}}, []string{"restrictedProperty"}},
		{"foo[/(?<zero>0)/]", NoRestrictedPropertiesOptions{Restrictions: []NoRestrictedPropertiesRestriction{{Property: "/(?<zero>0)/"}}}, []string{"restrictedProperty"}},
		{"require.call({}, 'foo')", NoRestrictedPropertiesOptions{Restrictions: []NoRestrictedPropertiesRestriction{{Object: "require", Message: "Please call require() directly."}}}, []string{"restrictedObjectProperty"}},
		{"require['resolve']", NoRestrictedPropertiesOptions{Restrictions: []NoRestrictedPropertiesRestriction{{Object: "require"}}}, []string{"restrictedObjectProperty"}},
		{"let {bar} = foo;", NoRestrictedPropertiesOptions{Restrictions: []NoRestrictedPropertiesRestriction{{Object: "foo", Property: "bar"}}}, []string{"restrictedObjectProperty"}},
		{"let {bar: baz} = foo;", NoRestrictedPropertiesOptions{Restrictions: []NoRestrictedPropertiesRestriction{{Object: "foo", Property: "bar"}}}, []string{"restrictedObjectProperty"}},
		{"let {'bar': baz} = foo;", NoRestrictedPropertiesOptions{Restrictions: []NoRestrictedPropertiesRestriction{{Object: "foo", Property: "bar"}}}, []string{"restrictedObjectProperty"}},
		{"let {bar: {baz: qux}} = foo;", NoRestrictedPropertiesOptions{Restrictions: []NoRestrictedPropertiesRestriction{{Object: "foo", Property: "bar"}}}, []string{"restrictedObjectProperty"}},
		{"let {bar} = foo;", NoRestrictedPropertiesOptions{Restrictions: []NoRestrictedPropertiesRestriction{{Object: "foo"}}}, []string{"restrictedObjectProperty"}},
		{"let {bar: baz} = foo;", NoRestrictedPropertiesOptions{Restrictions: []NoRestrictedPropertiesRestriction{{Object: "foo"}}}, []string{"restrictedObjectProperty"}},
		{"let {bar} = foo;", NoRestrictedPropertiesOptions{Restrictions: []NoRestrictedPropertiesRestriction{{Property: "bar"}}}, []string{"restrictedProperty"}},
		{"let bar; ({bar} = foo);", NoRestrictedPropertiesOptions{Restrictions: []NoRestrictedPropertiesRestriction{{Object: "foo", Property: "bar"}}}, []string{"restrictedObjectProperty"}},
		{"let bar; ({bar: baz = 1} = foo);", NoRestrictedPropertiesOptions{Restrictions: []NoRestrictedPropertiesRestriction{{Object: "foo", Property: "bar"}}}, []string{"restrictedObjectProperty"}},
		{"function qux({bar} = foo) {}", NoRestrictedPropertiesOptions{Restrictions: []NoRestrictedPropertiesRestriction{{Object: "foo", Property: "bar"}}}, []string{"restrictedObjectProperty"}},
		{"function qux({bar: baz} = foo) {}", NoRestrictedPropertiesOptions{Restrictions: []NoRestrictedPropertiesRestriction{{Object: "foo", Property: "bar"}}}, []string{"restrictedObjectProperty"}},
		{"var {['foo']: qux, bar} = baz", NoRestrictedPropertiesOptions{Restrictions: []NoRestrictedPropertiesRestriction{{Object: "baz", Property: "foo"}}}, []string{"restrictedObjectProperty"}},
		{"obj['#foo']", NoRestrictedPropertiesOptions{Restrictions: []NoRestrictedPropertiesRestriction{{Property: "#foo"}}}, []string{"restrictedProperty"}},
		{"const { bar: { bad } = {} } = foo;", NoRestrictedPropertiesOptions{Restrictions: []NoRestrictedPropertiesRestriction{{Property: "bad"}}}, []string{"restrictedProperty"}},
		{"const { bar: { bad } } = foo;", NoRestrictedPropertiesOptions{Restrictions: []NoRestrictedPropertiesRestriction{{Property: "bad"}}}, []string{"restrictedProperty"}},
		{"const { bad } = foo();", NoRestrictedPropertiesOptions{Restrictions: []NoRestrictedPropertiesRestriction{{Property: "bad"}}}, []string{"restrictedProperty"}},
		{"({ bad } = foo());", NoRestrictedPropertiesOptions{Restrictions: []NoRestrictedPropertiesRestriction{{Property: "bad"}}}, []string{"restrictedProperty"}},
		{"({ bar: { bad } } = foo);", NoRestrictedPropertiesOptions{Restrictions: []NoRestrictedPropertiesRestriction{{Property: "bad"}}}, []string{"restrictedProperty"}},
		{"({ bar: { bad } = {} } = foo);", NoRestrictedPropertiesOptions{Restrictions: []NoRestrictedPropertiesRestriction{{Property: "bad"}}}, []string{"restrictedProperty"}},
		{"({ bad }) => {};", NoRestrictedPropertiesOptions{Restrictions: []NoRestrictedPropertiesRestriction{{Property: "bad"}}}, []string{"restrictedProperty"}},
		{"({ bad } = {}) => {};", NoRestrictedPropertiesOptions{Restrictions: []NoRestrictedPropertiesRestriction{{Property: "bad"}}}, []string{"restrictedProperty"}},
		{"({ bad: bar }) => {};", NoRestrictedPropertiesOptions{Restrictions: []NoRestrictedPropertiesRestriction{{Property: "bad"}}}, []string{"restrictedProperty"}},
		{"({ bar: { bad } = {} }) => {};", NoRestrictedPropertiesOptions{Restrictions: []NoRestrictedPropertiesRestriction{{Property: "bad"}}}, []string{"restrictedProperty"}},
		{"[{ bad }] = foo;", NoRestrictedPropertiesOptions{Restrictions: []NoRestrictedPropertiesRestriction{{Property: "bad"}}}, []string{"restrictedProperty"}},
		{"const [{ bad }] = foo;", NoRestrictedPropertiesOptions{Restrictions: []NoRestrictedPropertiesRestriction{{Property: "bad"}}}, []string{"restrictedProperty"}},
		{"someObject.disallowedProperty", NoRestrictedPropertiesOptions{Restrictions: []NoRestrictedPropertiesRestriction{{Property: "disallowedProperty", AllowObjects: []string{"anotherObject"}}}}, []string{"restrictedProperty"}},
		{"someObject.disallowedProperty", NoRestrictedPropertiesOptions{Restrictions: []NoRestrictedPropertiesRestriction{{Property: "disallowedProperty", AllowObjects: []string{"anotherObject"}, Message: "Please use someObject.allowedProperty instead."}}}, []string{"restrictedProperty"}},
		{"someObject.disallowedProperty; anotherObject.anotherDisallowedProperty()", NoRestrictedPropertiesOptions{Restrictions: []NoRestrictedPropertiesRestriction{{Property: "disallowedProperty", AllowObjects: []string{"anotherObject"}}, {Property: "anotherDisallowedProperty", AllowObjects: []string{"someObject"}}}}, []string{"restrictedProperty", "restrictedProperty"}},
		{"someObject.disallowedProperty", NoRestrictedPropertiesOptions{Restrictions: []NoRestrictedPropertiesRestriction{{Object: "someObject", AllowProperties: []string{"allowedProperty"}}}}, []string{"restrictedObjectProperty"}},
		{"someObject.disallowedProperty", NoRestrictedPropertiesOptions{Restrictions: []NoRestrictedPropertiesRestriction{{Object: "someObject", AllowProperties: []string{"allowedProperty"}, Message: "Please use someObject.allowedProperty instead."}}}, []string{"restrictedObjectProperty"}},
		{"someObject.disallowedProperty; anotherObject.anotherDisallowedProperty()", NoRestrictedPropertiesOptions{Restrictions: []NoRestrictedPropertiesRestriction{{Object: "someObject", AllowProperties: []string{"anotherDisallowedProperty"}}, {Object: "anotherObject", AllowProperties: []string{"disallowedProperty"}}}}, []string{"restrictedObjectProperty", "restrictedObjectProperty"}},
	}
}

func noRestrictedPropertiesSilentCases() []noRestrictedPropertiesCase {
	return []noRestrictedPropertiesCase{
		{"someObject.someProperty", NoRestrictedPropertiesOptions{Restrictions: []NoRestrictedPropertiesRestriction{{Object: "someObject", Property: "disallowedProperty"}}}, nil},
		{"anotherObject.disallowedProperty", NoRestrictedPropertiesOptions{Restrictions: []NoRestrictedPropertiesRestriction{{Object: "someObject", Property: "disallowedProperty"}}}, nil},
		{"someObject.someProperty()", NoRestrictedPropertiesOptions{Restrictions: []NoRestrictedPropertiesRestriction{{Object: "someObject", Property: "disallowedProperty"}}}, nil},
		{"anotherObject.disallowedProperty()", NoRestrictedPropertiesOptions{Restrictions: []NoRestrictedPropertiesRestriction{{Object: "someObject", Property: "disallowedProperty"}}}, nil},
		{"anotherObject.disallowedProperty()", NoRestrictedPropertiesOptions{Restrictions: []NoRestrictedPropertiesRestriction{{Object: "someObject", Property: "disallowedProperty", Message: "Please use someObject.allowedProperty instead."}}}, nil},
		{"anotherObject['disallowedProperty']()", NoRestrictedPropertiesOptions{Restrictions: []NoRestrictedPropertiesRestriction{{Object: "someObject", Property: "disallowedProperty"}}}, nil},
		{"obj.toString", NoRestrictedPropertiesOptions{Restrictions: []NoRestrictedPropertiesRestriction{{Object: "obj", Property: "__proto__"}}}, nil},
		{"toString.toString", NoRestrictedPropertiesOptions{Restrictions: []NoRestrictedPropertiesRestriction{{Object: "obj", Property: "foo"}}}, nil},
		{"obj.toString", NoRestrictedPropertiesOptions{Restrictions: []NoRestrictedPropertiesRestriction{{Object: "obj", Property: "foo"}}}, nil},
		{"foo.bar", NoRestrictedPropertiesOptions{Restrictions: []NoRestrictedPropertiesRestriction{{Property: "baz"}}}, nil},
		{"foo.bar", NoRestrictedPropertiesOptions{Restrictions: []NoRestrictedPropertiesRestriction{{Object: "baz"}}}, nil},
		{"foo()", NoRestrictedPropertiesOptions{Restrictions: []NoRestrictedPropertiesRestriction{{Object: "foo"}}}, nil},
		{"foo;", NoRestrictedPropertiesOptions{Restrictions: []NoRestrictedPropertiesRestriction{{Object: "foo"}}}, nil},
		{"foo[/(?<zero>0)/]", NoRestrictedPropertiesOptions{Restrictions: []NoRestrictedPropertiesRestriction{{Property: "null"}}}, nil},
		{"let bar = foo;", NoRestrictedPropertiesOptions{Restrictions: []NoRestrictedPropertiesRestriction{{Object: "foo", Property: "bar"}}}, nil},
		{"let {baz: bar} = foo;", NoRestrictedPropertiesOptions{Restrictions: []NoRestrictedPropertiesRestriction{{Object: "foo", Property: "bar"}}}, nil},
		{"let {unrelated} = foo;", NoRestrictedPropertiesOptions{Restrictions: []NoRestrictedPropertiesRestriction{{Object: "foo", Property: "bar"}}}, nil},
		{"let {baz: {bar: qux}} = foo;", NoRestrictedPropertiesOptions{Restrictions: []NoRestrictedPropertiesRestriction{{Object: "foo", Property: "bar"}}}, nil},
		{"let {bar} = foo.baz;", NoRestrictedPropertiesOptions{Restrictions: []NoRestrictedPropertiesRestriction{{Object: "foo", Property: "bar"}}}, nil},
		{"let {baz: bar} = foo;", NoRestrictedPropertiesOptions{Restrictions: []NoRestrictedPropertiesRestriction{{Property: "bar"}}}, nil},
		{"let baz; ({baz: bar} = foo)", NoRestrictedPropertiesOptions{Restrictions: []NoRestrictedPropertiesRestriction{{Object: "foo", Property: "bar"}}}, nil},
		{"let bar;", NoRestrictedPropertiesOptions{Restrictions: []NoRestrictedPropertiesRestriction{{Object: "foo", Property: "bar"}}}, nil},
		{"let bar; ([bar = 5] = foo);", NoRestrictedPropertiesOptions{Restrictions: []NoRestrictedPropertiesRestriction{{Object: "foo", Property: "1"}}}, nil},
		{"function qux({baz: bar} = foo) {}", NoRestrictedPropertiesOptions{Restrictions: []NoRestrictedPropertiesRestriction{{Object: "foo", Property: "bar"}}}, nil},
		{"let [bar, baz] = foo;", NoRestrictedPropertiesOptions{Restrictions: []NoRestrictedPropertiesRestriction{{Object: "foo", Property: "1"}}}, nil},
		{"let [, bar] = foo;", NoRestrictedPropertiesOptions{Restrictions: []NoRestrictedPropertiesRestriction{{Object: "foo", Property: "0"}}}, nil},
		{"let [, bar = 5] = foo;", NoRestrictedPropertiesOptions{Restrictions: []NoRestrictedPropertiesRestriction{{Object: "foo", Property: "1"}}}, nil},
		{"let bar; ([bar = 5] = foo);", NoRestrictedPropertiesOptions{Restrictions: []NoRestrictedPropertiesRestriction{{Object: "foo", Property: "0"}}}, nil},
		{"function qux([bar] = foo) {}", NoRestrictedPropertiesOptions{Restrictions: []NoRestrictedPropertiesRestriction{{Object: "foo", Property: "0"}}}, nil},
		{"function qux([, bar] = foo) {}", NoRestrictedPropertiesOptions{Restrictions: []NoRestrictedPropertiesRestriction{{Object: "foo", Property: "0"}}}, nil},
		{"function qux([, bar] = foo) {}", NoRestrictedPropertiesOptions{Restrictions: []NoRestrictedPropertiesRestriction{{Object: "foo", Property: "1"}}}, nil},
		{"class C { #foo; foo() { this.#foo; } }", NoRestrictedPropertiesOptions{Restrictions: []NoRestrictedPropertiesRestriction{{Property: "#foo"}}}, nil},
		{"someObject.disallowedProperty", NoRestrictedPropertiesOptions{Restrictions: []NoRestrictedPropertiesRestriction{{Property: "disallowedProperty", AllowObjects: []string{"someObject"}}}}, nil},
		{"someObject.disallowedProperty; anotherObject.disallowedProperty();", NoRestrictedPropertiesOptions{Restrictions: []NoRestrictedPropertiesRestriction{{Property: "disallowedProperty", AllowObjects: []string{"someObject", "anotherObject"}}}}, nil},
		{"someObject.disallowedProperty()", NoRestrictedPropertiesOptions{Restrictions: []NoRestrictedPropertiesRestriction{{Property: "disallowedProperty", AllowObjects: []string{"someObject"}}}}, nil},
		{"someObject['disallowedProperty']()", NoRestrictedPropertiesOptions{Restrictions: []NoRestrictedPropertiesRestriction{{Property: "disallowedProperty", AllowObjects: []string{"someObject"}}}}, nil},
		{"let {bar} = foo;", NoRestrictedPropertiesOptions{Restrictions: []NoRestrictedPropertiesRestriction{{Property: "bar", AllowObjects: []string{"foo"}}}}, nil},
		{"let {baz: bar} = foo;", NoRestrictedPropertiesOptions{Restrictions: []NoRestrictedPropertiesRestriction{{Property: "baz", AllowObjects: []string{"foo"}}}}, nil},
		{"someObject.disallowedProperty", NoRestrictedPropertiesOptions{Restrictions: []NoRestrictedPropertiesRestriction{{Object: "someObject", AllowProperties: []string{"disallowedProperty"}}}}, nil},
		{"someObject.disallowedProperty; someObject.anotherDisallowedProperty();", NoRestrictedPropertiesOptions{Restrictions: []NoRestrictedPropertiesRestriction{{Object: "someObject", AllowProperties: []string{"disallowedProperty", "anotherDisallowedProperty"}}}}, nil},
		{"someObject.disallowedProperty()", NoRestrictedPropertiesOptions{Restrictions: []NoRestrictedPropertiesRestriction{{Object: "someObject", AllowProperties: []string{"disallowedProperty"}}}}, nil},
		{"someObject['disallowedProperty']()", NoRestrictedPropertiesOptions{Restrictions: []NoRestrictedPropertiesRestriction{{Object: "someObject", AllowProperties: []string{"disallowedProperty"}}}}, nil},
		{"let {bar} = foo;", NoRestrictedPropertiesOptions{Restrictions: []NoRestrictedPropertiesRestriction{{Object: "foo", AllowProperties: []string{"bar"}}}}, nil},
		{"let {baz: bar} = foo;", NoRestrictedPropertiesOptions{Restrictions: []NoRestrictedPropertiesRestriction{{Object: "foo", AllowProperties: []string{"baz"}}}}, nil},
	}
}

func TestNoRestrictedPropertiesFires(t *testing.T) {
	t.Parallel()

	for _, testCase := range noRestrictedPropertiesFiresCases() {
		t.Run(testCase.sourceText, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t, runNoRestrictedProperties(t, testCase),
				testCase.wantIds...)
		})
	}
}

func TestNoRestrictedPropertiesStaysSilent(t *testing.T) {
	t.Parallel()

	for _, testCase := range noRestrictedPropertiesSilentCases() {
		t.Run(testCase.sourceText, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, runNoRestrictedProperties(t, testCase))
		})
	}
}

// With no restrictions configured, the rule registers nothing at all.
//
// This is the property that makes the rule's violation count on any codebase structurally zero
// rather than evidence of a clean tree, and it is why this port ships REGISTERED BUT NOT ENABLED:
// enabling it means choosing restrictions, which is a decision about this codebase rather than a
// porting question.
//
// Verified against the installed build the same way: `foo.bar; baz.qux;` with no options reports
// nothing, on source that names exactly the shape a restriction would ban.
func TestNoRestrictedPropertiesEnforcesNothingUnconfigured(t *testing.T) {
	t.Parallel()

	source := "foo.bar; baz.qux; var { bar } = foo;"
	rule_testing.ExpectClean(t, rule_testing.Run(t, NoRestrictedProperties,
		noRestrictedPropertiesFile, source))

	// The control: the same source under a restriction reports, so the silence above is the empty
	// configuration rather than a rule that cannot fire.
	decoded, err := DecodeNoRestrictedPropertiesOptions(
		[]byte(`[{"object":"foo","property":"bar"}]`))
	if err != nil {
		t.Fatalf("could not decode options: %v", err)
	}
	rule_testing.ExpectFindings(t, rule_testing.RunWithOptions(t, NoRestrictedProperties,
		noRestrictedPropertiesFile, source, decoded),
		"restrictedObjectProperty", "restrictedObjectProperty")
}

// What the finding SAYS, asserted by equality on the whole rendered sentence.
//
// Every message on this rule is built by interpolation from up to four pieces: the object name, the
// property name, an allowance list, and the project's own message. Nothing above can see any of
// them -- a rule naming the wrong half of the pair, joining an allowance list with the wrong
// separator, or dropping the configured message would satisfy every count assertion in this file.
//
// Backticks replace upstream's single quotes around identifiers, matching how every other message in
// this tree renders code. The rest of each sentence is upstream's wording, including the ordering
// of the two optional suffixes: the allowance first, then the project message.
func TestNoRestrictedPropertiesRendersTheWholeMessage(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		options    NoRestrictedPropertiesOptions
		wantId     string
		wantText   string
	}{
		{
			name:       "the specific pair, with nothing optional",
			sourceText: "someObject.disallowedProperty",
			options: NoRestrictedPropertiesOptions{
				Restrictions: []NoRestrictedPropertiesRestriction{
					{Object: "someObject", Property: "disallowedProperty"},
				}},
			wantId:   "restrictedObjectProperty",
			wantText: "`someObject.disallowedProperty` is restricted from being used.",
		},
		{
			name:       "the specific pair, carrying the project's message",
			sourceText: "someObject.disallowedProperty",
			options: NoRestrictedPropertiesOptions{
				Restrictions: []NoRestrictedPropertiesRestriction{
					{Object: "someObject", Property: "disallowedProperty",
						Message: "Please use someObject.allowedProperty instead."},
				}},
			wantId: "restrictedObjectProperty",
			wantText: "`someObject.disallowedProperty` is restricted from being used. " +
				"Please use someObject.allowedProperty instead.",
		},
		{
			name:       "an object-only restriction naming its allowance",
			sourceText: "someObject.disallowedProperty",
			options: NoRestrictedPropertiesOptions{
				Restrictions: []NoRestrictedPropertiesRestriction{
					{Object: "someObject", AllowProperties: []string{"allowedProperty"}},
				}},
			wantId: "restrictedObjectProperty",
			wantText: "`someObject.disallowedProperty` is restricted from being used. " +
				"Only these properties are allowed: allowedProperty.",
		},
		{
			name:       "a property-only restriction naming its allowance",
			sourceText: "someObject.disallowedProperty",
			options: NoRestrictedPropertiesOptions{
				Restrictions: []NoRestrictedPropertiesRestriction{
					{Property: "disallowedProperty", AllowObjects: []string{"anotherObject"}},
				}},
			wantId: "restrictedProperty",
			wantText: "`disallowedProperty` is restricted from being used. Property " +
				"`disallowedProperty` is only allowed on these objects: anotherObject.",
		},
		{
			name:       "both optional pieces at once, in upstream's order",
			sourceText: "someObject.disallowedProperty",
			options: NoRestrictedPropertiesOptions{
				Restrictions: []NoRestrictedPropertiesRestriction{
					{Property: "disallowedProperty", AllowObjects: []string{"anotherObject"},
						Message: "Please use someObject.allowedProperty instead."},
				}},
			wantId: "restrictedProperty",
			wantText: "`disallowedProperty` is restricted from being used. Property " +
				"`disallowedProperty` is only allowed on these objects: anotherObject. " +
				"Please use someObject.allowedProperty instead.",
		},
		{
			name:       "an allowance list with more than one entry, to pin the separator",
			sourceText: "someObject.disallowedProperty",
			options: NoRestrictedPropertiesOptions{
				Restrictions: []NoRestrictedPropertiesRestriction{
					{Object: "someObject", AllowProperties: []string{"first", "second", "third"}},
				}},
			wantId: "restrictedObjectProperty",
			wantText: "`someObject.disallowedProperty` is restricted from being used. " +
				"Only these properties are allowed: first, second, third.",
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := runNoRestrictedProperties(t, noRestrictedPropertiesCase{
				sourceText: testCase.sourceText, options: testCase.options})
			if len(result.Diagnostics) != 1 {
				t.Fatalf("wanted one finding, got %d", len(result.Diagnostics))
			}
			// Equality against a literal typed here, never against the rule's own constant: a
			// comparison to the constant moves with the code under mutation and asserts nothing.
			if got := result.Diagnostics[0].Message.Description; got != testCase.wantText {
				t.Errorf("rendered\n  %q\nwanted\n  %q", got, testCase.wantText)
			}
			if got := result.Diagnostics[0].Message.Id; got != testCase.wantId {
				t.Errorf("message id was %q, wanted %q", got, testCase.wantId)
			}
		})
	}
}

// Where the finding points, which no message-id fixture can see.
//
// A member access reports on the WHOLE access rather than on the property, and a destructuring
// pattern reports on the whole pattern rather than on the property being read -- which is why one
// pattern restricting two properties reports twice on the same span. Both are upstream's `node`.
func TestNoRestrictedPropertiesPointsAtTheAccess(t *testing.T) {
	t.Parallel()

	restrictBar := NoRestrictedPropertiesOptions{
		Restrictions: []NoRestrictedPropertiesRestriction{{Property: "bar"}},
	}
	cases := []struct {
		sourceText string
		options    NoRestrictedPropertiesOptions
		wantSpans  []string
	}{
		{"foo.bar;", restrictBar, []string{"foo.bar"}},
		{"foo['bar'];", restrictBar, []string{"foo['bar']"}},
		{"foo.baz.bar;", restrictBar, []string{"foo.baz.bar"}},
		// The pattern, not the property inside it.
		{"var { bar } = foo;", restrictBar, []string{"{ bar }"}},
		{"({ bar } = foo);", restrictBar, []string{"{ bar }"}},
		// Two restricted properties in one pattern report twice on the identical span.
		{"var { bar, bar: other } = foo;", restrictBar, []string{
			"{ bar, bar: other }", "{ bar, bar: other }"}},
	}
	for _, testCase := range cases {
		t.Run(testCase.sourceText, func(t *testing.T) {
			t.Parallel()
			result := runNoRestrictedProperties(t, noRestrictedPropertiesCase{
				sourceText: testCase.sourceText, options: testCase.options})
			if len(result.Diagnostics) != len(testCase.wantSpans) {
				t.Fatalf("wanted %d findings, got %d", len(testCase.wantSpans),
					len(result.Diagnostics))
			}
			for index, want := range testCase.wantSpans {
				span := result.Diagnostics[index].Range
				if got := testCase.sourceText[span.Pos():span.End()]; got != want {
					t.Errorf("finding %d pointed at %q, wanted %q", index, got, want)
				}
			}
		})
	}
}

// Destructuring written as an ASSIGNMENT, which our parser gives a different node kind.
//
// ESTree calls `var {bar} = foo` and `({bar} = foo)` the same `ObjectPattern`, so upstream needs one
// listener. Ours parses the second as an object LITERAL whose members are property assignments, so
// the same judgment needs a second entry point over a different shape entirely.
//
// Six of upstream's imported cases reported nothing until that arm existed, and a seventh -- the
// nested pattern -- until the walk accepted an enclosing object literal. A port stopping at the
// binding pattern ships with a whole syntactic form silently exempt, and only the corpus says so.
//
// The last two rows are the controls that separate "is a pattern" from "is any object literal".
func TestNoRestrictedPropertiesReadsAssignmentDestructuring(t *testing.T) {
	t.Parallel()

	restrictBad := NoRestrictedPropertiesOptions{
		Restrictions: []NoRestrictedPropertiesRestriction{{Property: "bad"}},
	}
	restrictFooBar := NoRestrictedPropertiesOptions{
		Restrictions: []NoRestrictedPropertiesRestriction{{Object: "foo", Property: "bar"}},
	}
	cases := []noRestrictedPropertiesCase{
		{"let bar; ({bar} = foo);", restrictFooBar, []string{"restrictedObjectProperty"}},
		{"({ bad } = foo());", restrictBad, []string{"restrictedProperty"}},
		{"({ bar: { bad } } = foo);", restrictBad, []string{"restrictedProperty"}},
		{"[{ bad }] = foo;", restrictBad, []string{"restrictedProperty"}},
		// An ordinary object literal in a call argument is not a pattern and must stay clean.
		{"foo({ bad: 1 });", restrictBad, nil},
		// Nor is one on the RIGHT of an assignment.
		{"x = { bad: 1 };", restrictBad, nil},
	}
	for _, testCase := range cases {
		t.Run(testCase.sourceText, func(t *testing.T) {
			t.Parallel()
			result := runNoRestrictedProperties(t, testCase)
			if len(testCase.wantIds) == 0 {
				rule_testing.ExpectClean(t, result)
				return
			}
			rule_testing.ExpectFindings(t, result, testCase.wantIds...)
		})
	}
}

// The property-name predicate, which is NOT the shelf's `property.AccessedName`.
//
// The shelf helper is the closest thing on any shelf and disagrees with upstream on four of the
// eight spellings this rule can meet. That was measured with a probe rather than read off its doc
// comment, which is accurate about what its code does and says nothing about whether it matches the
// implementation being ported:
//
//	`this.#foo`   upstream CLEAN, shelf answers "#foo" and would report
//	`foo[/re/]`   upstream reports "/re/", shelf answers nothing
//	`foo[null]`   upstream reports "null", shelf answers nothing
//	`foo[1n]`     upstream reports "1", shelf answers nothing
//
// Three of those cost findings and one adds a false one. Widening the shelf helper would change four
// other rules, so this rule carries its own predicate; these rows are what pin the difference.
func TestNoRestrictedPropertiesReadsUpstreamsPropertyNames(t *testing.T) {
	t.Parallel()

	restrict := func(property string) NoRestrictedPropertiesOptions {
		return NoRestrictedPropertiesOptions{
			Restrictions: []NoRestrictedPropertiesRestriction{{Property: property}},
		}
	}
	cases := []noRestrictedPropertiesCase{
		// A string key that spells a private name reports; the private name itself does not.
		{"foo['#bar'];", restrict("#bar"), []string{"restrictedProperty"}},
		{"class C { #foo = 1; m() { return this.#foo; } }", restrict("#foo"), nil},
		// A regular expression subscript stringifies to its own source, slashes included.
		{"foo[/(?<zero>0)/];", restrict("/(?<zero>0)/"), []string{"restrictedProperty"}},
		// `null`, `true` and a bigint all stringify.
		{"foo[null];", restrict("null"), []string{"restrictedProperty"}},
		{"foo[true];", restrict("true"), []string{"restrictedProperty"}},
		{"foo[1n];", restrict("1"), []string{"restrictedProperty"}},
		// A numeric and a template subscript, which the shelf and upstream agree on.
		{"foo[1];", restrict("1"), []string{"restrictedProperty"}},
		{"foo[`bar`];", restrict("bar"), []string{"restrictedProperty"}},
		// A variable subscript names whatever it holds, so nothing is knowable before it runs.
		{"foo[bar];", restrict("bar"), nil},
	}
	for _, testCase := range cases {
		t.Run(testCase.sourceText, func(t *testing.T) {
			t.Parallel()
			result := runNoRestrictedProperties(t, testCase)
			if len(testCase.wantIds) == 0 {
				rule_testing.ExpectClean(t, result)
				return
			}
			rule_testing.ExpectFindings(t, result, testCase.wantIds...)
		})
	}
}

// The decoder, which enforces three constraints upstream expresses in its schema.
//
// We have no schema layer, so an entry naming neither an object nor a property would sit in the
// config matching nothing forever, and the two self-contradictory pairings would be accepted
// silently. None of these is reachable from a fixture that builds the options struct directly.
func TestDecodeNoRestrictedPropertiesOptions(t *testing.T) {
	t.Parallel()

	t.Run("nil input yields an empty list", func(t *testing.T) {
		t.Parallel()
		decoded, err := DecodeNoRestrictedPropertiesOptions(nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(decoded.(NoRestrictedPropertiesOptions).Restrictions) != 0 {
			t.Error("nil input produced restrictions")
		}
	})

	t.Run("an entry naming neither an object nor a property is rejected", func(t *testing.T) {
		t.Parallel()
		if _, err := DecodeNoRestrictedPropertiesOptions(
			[]byte(`[{"message":"nope"}]`)); err == nil {
			t.Error("an entry that can never match anything decoded")
		}
	})

	t.Run("object paired with allowObjects is rejected", func(t *testing.T) {
		t.Parallel()
		if _, err := DecodeNoRestrictedPropertiesOptions(
			[]byte(`[{"object":"foo","allowObjects":["bar"]}]`)); err == nil {
			t.Error("a self-contradictory pairing decoded")
		}
	})

	t.Run("property paired with allowProperties is rejected", func(t *testing.T) {
		t.Parallel()
		if _, err := DecodeNoRestrictedPropertiesOptions(
			[]byte(`[{"property":"foo","allowProperties":["bar"]}]`)); err == nil {
			t.Error("a self-contradictory pairing decoded")
		}
	})

	t.Run("every element is a restriction, upstream's variadic spelling", func(t *testing.T) {
		t.Parallel()
		decoded, err := DecodeNoRestrictedPropertiesOptions(
			[]byte(`[{"object":"foo","property":"bar"},{"property":"__proto__","message":"No."}]`))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		restrictions := decoded.(NoRestrictedPropertiesOptions).Restrictions
		if len(restrictions) != 2 || restrictions[1].Property != "__proto__" || restrictions[1].Message != "No." {
			t.Errorf("the second element was not read as a restriction: %+v", restrictions)
		}
	})

	t.Run("shapes the config layer does not deliver, or upstream refuses, are rejected", func(t *testing.T) {
		t.Parallel()
		for _, raw := range []string{
			// The wrapper this decoder used to read.
			`{"restrictions":[{"object":"foo"}]}`,
			// A restriction carrying a key upstream's schema does not declare.
			`[{"object":"foo","mesage":"x"}]`,
			// An element that is not a restriction object.
			`["foo"]`,
		} {
			if decoded, err := DecodeNoRestrictedPropertiesOptions([]byte(raw)); err == nil {
				t.Errorf("%s decoded to %+v; it must be refused", raw, decoded)
			}
		}
	})
}

// The empty-list guard registers NO listener, which is a cost property no findings test can see.
//
// Written for a surviving mutant. Removing the guard changes no verdict -- an index built from an
// empty list matches nothing, so both paths reach the same silence -- and `ExpectFindings` asserts
// findings, so nothing in this file could tell the two apart.
//
// It is not equivalent, though, and the distinguishing observable was measured rather than argued.
// In a probe tree with the rule enabled and no restrictions configured, the guard is the difference
// between the timing report showing `0 files, 0 registrations` and showing `1 file, 2
// registrations`: without it every member access and every object literal in the tree is walked to
// reach a lookup that can never match.
//
// So the assertion is on the listener set rather than on any finding.
func TestNoRestrictedPropertiesRegistersNothingUnconfigured(t *testing.T) {
	t.Parallel()

	listenersFor := func(options any) rule.Listeners {
		return NoRestrictedProperties.Run(rule.Context{}, options)
	}

	if got := len(listenersFor(nil)); got != 0 {
		t.Errorf("nil options registered %d listeners, wanted none: an unconfigured rule that "+
			"walks the tree costs every file for a lookup that cannot match", got)
	}
	if got := len(listenersFor(NoRestrictedPropertiesOptions{})); got != 0 {
		t.Errorf("an empty restriction list registered %d listeners, wanted none", got)
	}

	// The control: one restriction registers the full set, so the zero above is the guard rather
	// than a rule that never registers anything at all.
	configured := NoRestrictedPropertiesOptions{
		Restrictions: []NoRestrictedPropertiesRestriction{{Property: "bar"}},
	}
	if got := len(listenersFor(configured)); got == 0 {
		t.Error("a configured restriction registered no listeners")
	}
}

// Two shapes upstream's corpus does not write, each found by a surviving mutant.
//
//	a parenthesized receiver   `(foo).bar` still has the receiver name `foo`. Espree gives
//	                          parentheses no node, so upstream's `node.object.name` reads the
//	                          identifier directly and the skip here is fidelity rather than a
//	                          widening -- the same correction `no-eq-null` documents.
//	a rest element            `{...rest}` reads no single property, so its binding name must NOT
//	                          be compared against a restriction. Without the guard, `rest` is read
//	                          as a property name and a project restricting a property called
//	                          `rest` would report on every rest destructuring in the tree.
//
// Both verdicts were measured against the installed rule before the rows were written.
func TestNoRestrictedPropertiesSkipsParenthesesAndRestElements(t *testing.T) {
	t.Parallel()

	restrictFooBar := NoRestrictedPropertiesOptions{
		Restrictions: []NoRestrictedPropertiesRestriction{{Object: "foo", Property: "bar"}},
	}
	restrictRest := NoRestrictedPropertiesOptions{
		Restrictions: []NoRestrictedPropertiesRestriction{{Property: "rest"}},
	}
	restrictFoo := NoRestrictedPropertiesOptions{
		Restrictions: []NoRestrictedPropertiesRestriction{{Object: "foo"}},
	}
	cases := []struct {
		name       string
		sourceText string
		options    NoRestrictedPropertiesOptions
		wantIds    []string
	}{
		{"a parenthesized receiver", "(foo).bar;", restrictFooBar,
			[]string{"restrictedObjectProperty"}},
		{"a doubly parenthesized receiver", "((foo)).bar;", restrictFooBar,
			[]string{"restrictedObjectProperty"}},
		// The control: an unparenthesized receiver, so the two rows above are the skip rather than
		// a rule matching every receiver.
		{"the unparenthesized control", "foo.bar;", restrictFooBar,
			[]string{"restrictedObjectProperty"}},
		{"a different receiver stays clean", "bar.bar;", restrictFooBar, nil},
		// A rest element under each of the three restriction kinds.
		{"a rest element against a property restriction", "var {...rest} = foo;", restrictRest,
			nil},
		{"a rest element against an object restriction", "var {...rest} = foo;", restrictFoo, nil},
		{"a rest element in an assignment pattern", "({...rest} = foo);", restrictRest, nil},
		// The control: the same binding name written as an ordinary property DOES report.
		{"the same name as a real property", "var {rest} = foo;", restrictRest,
			[]string{"restrictedProperty"}},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := runNoRestrictedProperties(t, noRestrictedPropertiesCase{
				sourceText: testCase.sourceText, options: testCase.options})
			if len(testCase.wantIds) == 0 {
				rule_testing.ExpectClean(t, result)
				return
			}
			rule_testing.ExpectFindings(t, result, testCase.wantIds...)
		})
	}
}
