package core

import (
	"encoding/json"
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// accessorPairsFile is where the fixtures pretend to live.
const accessorPairsFile = "/repository/source/AccessorPairs.ts"

// Addressable booleans, because every option on this rule is a pointer.
//
// The pointer is not decoration: `setWithoutGet` and `enforceForClassMembers` both default to TRUE,
// so a plain bool field cannot tell "the key was absent" from "the key was written as false", and
// the zero value would silently invert the rule for anyone configuring it as a bare severity.
var (
	accessorPairsTrue  = accessorPairsBoolean(true)
	accessorPairsFalse = accessorPairsBoolean(false)
)

func accessorPairsBoolean(value bool) *bool { return &value }

// accessorPairsCase is one imported corpus row: the source, the options, and the message ids in the
// order they are reported.
type accessorPairsCase struct {
	sourceText string
	options    any
	wantIds    []string
}

// runAccessorPairs drives one case through the typed harness, routing options through the rule's own
// exported decoder rather than handing it a struct.
//
// The decoder is what applies the two true-by-default settings, so a fixture built by passing the
// struct directly would leave the single line most likely to be wrong completely untested. A nil
// options value is passed through as nil, which is what the config layer hands a rule configured as
// a bare severity.
func runAccessorPairs(t *testing.T, testCase accessorPairsCase) rule_testing.Result {
	t.Helper()
	if testCase.options == nil {
		return rule_testing.RunTyped(t, AccessorPairs, accessorPairsFile, testCase.sourceText)
	}
	encoded, err := json.Marshal(testCase.options)
	if err != nil {
		t.Fatalf("could not encode options: %v", err)
	}
	decoded, err := DecodeAccessorPairsOptions(encoded)
	if err != nil {
		t.Fatalf("could not decode options: %v", err)
	}
	return rule_testing.RunTypedWithOptions(t, AccessorPairs, accessorPairsFile,
		testCase.sourceText, decoded)
}

// The corpus is ESLint's own, extracted mechanically rather than retyped.
//
// `tests/lib/rules/accessor-pairs.js` was loaded with its RuleTester stubbed, so every case came out
// as data with its options attached, and each one was then replayed against the INSTALLED rule in
// `node_modules` to record what it actually reports. All 313 reproduced exactly, so the expectations
// below are measurements rather than transcriptions of the `errors` arrays.
//
// Two of upstream's clean cases could not be imported and are recorded in
// TestAccessorPairsGlobalShadowing instead: they turn on `languageOptions.globals` marking `Object`
// or `Reflect` as undefined, which our harness has no way to express. The mechanism they guard is
// covered there by a local binding, which is the same question asked a way this tree can ask it.
func accessorPairsFiresCases() []accessorPairsCase {
	return []accessorPairsCase{
		{"var o = { set a(value) {} };", nil, []string{"missingGetterInObjectLiteral"}},
		{"var o = { set a(value) {} };", AccessorPairsOptions{}, []string{"missingGetterInObjectLiteral"}},
		{"var o = { set a(value) {} };", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsFalse}, []string{"missingGetterInObjectLiteral"}},
		{"var o = { set a(value) {} };", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue}, []string{"missingGetterInObjectLiteral"}},
		{"var o = { get a() {} };", AccessorPairsOptions{SetWithoutGet: accessorPairsFalse, GetWithoutSet: accessorPairsTrue}, []string{"missingSetterInObjectLiteral"}},
		{"var o = { get a() {} };", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue}, []string{"missingSetterInObjectLiteral"}},
		{"var o = { get a() {} };", AccessorPairsOptions{GetWithoutSet: accessorPairsTrue}, []string{"missingSetterInObjectLiteral"}},
		{"var o = { get abc() {} };", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue}, []string{"missingSetterInObjectLiteral"}},
		{"var o = { get 'abc'() {} };", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue}, []string{"missingSetterInObjectLiteral"}},
		{"var o = { get 123() {} };", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue}, []string{"missingSetterInObjectLiteral"}},
		{"var o = { get 1e2() {} };", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue}, []string{"missingSetterInObjectLiteral"}},
		{"var o = { get ['abc']() {} };", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue}, []string{"missingSetterInObjectLiteral"}},
		{"var o = { get [`abc`]() {} };", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue}, []string{"missingSetterInObjectLiteral"}},
		{"var o = { get [123]() {} };", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue}, []string{"missingSetterInObjectLiteral"}},
		{"var o = { get [abc]() {} };", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue}, []string{"missingSetterInObjectLiteral"}},
		{"var o = { get [f(abc)]() {} };", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue}, []string{"missingSetterInObjectLiteral"}},
		{"var o = { get [a + b]() {} };", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue}, []string{"missingSetterInObjectLiteral"}},
		{"var o = { set abc(foo) {} };", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue}, []string{"missingGetterInObjectLiteral"}},
		{"var o = { set 'abc'(foo) {} };", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue}, []string{"missingGetterInObjectLiteral"}},
		{"var o = { set 123(foo) {} };", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue}, []string{"missingGetterInObjectLiteral"}},
		{"var o = { set 1e2(foo) {} };", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue}, []string{"missingGetterInObjectLiteral"}},
		{"var o = { set ['abc'](foo) {} };", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue}, []string{"missingGetterInObjectLiteral"}},
		{"var o = { set [`abc`](foo) {} };", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue}, []string{"missingGetterInObjectLiteral"}},
		{"var o = { set [123](foo) {} };", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue}, []string{"missingGetterInObjectLiteral"}},
		{"var o = { set [abc](foo) {} };", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue}, []string{"missingGetterInObjectLiteral"}},
		{"var o = { set [f(abc)](foo) {} };", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue}, []string{"missingGetterInObjectLiteral"}},
		{"var o = { set [a + b](foo) {} };", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue}, []string{"missingGetterInObjectLiteral"}},
		{"var o = { get a() {}, set b(foo) {} };", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue}, []string{"missingSetterInObjectLiteral", "missingGetterInObjectLiteral"}},
		{"var o = { set a(foo) {}, get b() {} };", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue}, []string{"missingGetterInObjectLiteral", "missingSetterInObjectLiteral"}},
		{"var o = { get 1() {}, set b(foo) {} };", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue}, []string{"missingSetterInObjectLiteral", "missingGetterInObjectLiteral"}},
		{"var o = { get a() {}, set 1(foo) {} };", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue}, []string{"missingSetterInObjectLiteral", "missingGetterInObjectLiteral"}},
		{"var o = { get a() {}, set 'a '(foo) {} };", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue}, []string{"missingSetterInObjectLiteral", "missingGetterInObjectLiteral"}},
		{"var o = { get ' a'() {}, set 'a'(foo) {} };", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue}, []string{"missingSetterInObjectLiteral", "missingGetterInObjectLiteral"}},
		{"var o = { get ''() {}, set ' '(foo) {} };", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue}, []string{"missingSetterInObjectLiteral", "missingGetterInObjectLiteral"}},
		{"var o = { get ''() {}, set null(foo) {} };", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue}, []string{"missingSetterInObjectLiteral", "missingGetterInObjectLiteral"}},
		{"var o = { get [`a`]() {}, set b(foo) {} };", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue}, []string{"missingSetterInObjectLiteral", "missingGetterInObjectLiteral"}},
		{"var o = { get [a]() {}, set [b](foo) {} };", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue}, []string{"missingSetterInObjectLiteral", "missingGetterInObjectLiteral"}},
		{"var o = { get [a]() {}, set a(foo) {} };", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue}, []string{"missingSetterInObjectLiteral", "missingGetterInObjectLiteral"}},
		{"var o = { get a() {}, set [a](foo) {} };", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue}, []string{"missingSetterInObjectLiteral", "missingGetterInObjectLiteral"}},
		{"var o = { get [a + b]() {}, set [a - b](foo) {} };", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue}, []string{"missingSetterInObjectLiteral", "missingGetterInObjectLiteral"}},
		{"var o = { get [`${0} `]() {}, set [`${0}`](foo) {} };", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue}, []string{"missingSetterInObjectLiteral", "missingGetterInObjectLiteral"}},
		{"var o = { get a() {}, get b() {} };", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue}, []string{"missingSetterInObjectLiteral", "missingSetterInObjectLiteral"}},
		{"var o = { set a(foo) {}, set b(bar) {} };", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue}, []string{"missingGetterInObjectLiteral", "missingGetterInObjectLiteral"}},
		{"var o = { get a() {}, set b(foo) {}, set c(foo) {}, get d() {} };", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue}, []string{"missingSetterInObjectLiteral", "missingGetterInObjectLiteral", "missingGetterInObjectLiteral", "missingSetterInObjectLiteral"}},
		{"var o1 = { get a() {} }, o2 = { set a(foo) {} };", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue}, []string{"missingSetterInObjectLiteral", "missingGetterInObjectLiteral"}},
		{"var o1 = { set a(foo) {} }, o2 = { get a() {} };", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue}, []string{"missingGetterInObjectLiteral", "missingSetterInObjectLiteral"}},
		{"var o = { get a() {}, get b() {}, set b(foo) {} };", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue}, []string{"missingSetterInObjectLiteral"}},
		{"var o = { get b() {}, get a() {}, set b(foo) {} };", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue}, []string{"missingSetterInObjectLiteral"}},
		{"var o = { get b() {}, set b(foo) {}, get a() {} };", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue}, []string{"missingSetterInObjectLiteral"}},
		{"var o = { set a(foo) {}, get b() {}, set b(bar) {} };", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue}, []string{"missingGetterInObjectLiteral"}},
		{"var o = { get b() {}, set a(foo) {}, set b(bar) {} };", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue}, []string{"missingGetterInObjectLiteral"}},
		{"var o = { get b() {}, set b(bar) {}, set a(foo) {} };", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue}, []string{"missingGetterInObjectLiteral"}},
		{"var o = { get v1() {}, set i1(foo) {}, get v2() {}, set v2(bar) {}, get i2() {}, set v1(baz) {} };", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue}, []string{"missingGetterInObjectLiteral", "missingSetterInObjectLiteral"}},
		{"var o = { get a() {}, get a() {} };", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue}, []string{"missingSetterInObjectLiteral", "missingSetterInObjectLiteral"}},
		{"var o = { set a(foo) {}, set a(foo) {} };", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue}, []string{"missingGetterInObjectLiteral", "missingGetterInObjectLiteral"}},
		{"var o = { a, get b() {}, c };", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue}, []string{"missingSetterInObjectLiteral"}},
		{"var o = { a, get b() {}, c, set d(foo) {} };", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue}, []string{"missingSetterInObjectLiteral", "missingGetterInObjectLiteral"}},
		{"var o = { get a() {}, a:1 };", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue}, []string{"missingSetterInObjectLiteral"}},
		{"var o = { a, get a() {} };", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue}, []string{"missingSetterInObjectLiteral"}},
		{"var o = { set a(foo) {}, a:1 };", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue}, []string{"missingGetterInObjectLiteral"}},
		{"var o = { a, set a(foo) {} };", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue}, []string{"missingGetterInObjectLiteral"}},
		{"var o = { get a() {}, ...b };", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue}, []string{"missingSetterInObjectLiteral"}},
		{"var o = { get a() {}, ...a };", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue}, []string{"missingSetterInObjectLiteral"}},
		{"var o = { set a(foo) {}, ...a };", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue}, []string{"missingGetterInObjectLiteral"}},
		{"var o = { get b() {} };", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue}, []string{"missingSetterInObjectLiteral"}},
		{"var o = {\n  set [\n a](foo) {} };", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue}, []string{"missingGetterInObjectLiteral"}},
		{"var o = {d: 1};\n Object.defineProperty(o, 'c', \n{set: function(value) {\n val = value; \n} \n});", nil, []string{"missingGetterInPropertyDescriptor"}},
		{"Reflect.defineProperty(obj, 'foo', {set: function(value) {}});", nil, []string{"missingGetterInPropertyDescriptor"}},
		{"Object.defineProperties(obj, {foo: {set: function(value) {}}});", nil, []string{"missingGetterInPropertyDescriptor"}},
		{"Object.create(null, {foo: {set: function(value) {}}});", nil, []string{"missingGetterInPropertyDescriptor"}},
		{"var o = {d: 1};\n Object?.defineProperty(o, 'c', \n{set: function(value) {\n val = value; \n} \n});", nil, []string{"missingGetterInPropertyDescriptor"}},
		{"Reflect?.defineProperty(obj, 'foo', {set: function(value) {}});", nil, []string{"missingGetterInPropertyDescriptor"}},
		{"Object?.defineProperties(obj, {foo: {set: function(value) {}}});", nil, []string{"missingGetterInPropertyDescriptor"}},
		{"Object?.create(null, {foo: {set: function(value) {}}});", nil, []string{"missingGetterInPropertyDescriptor"}},
		{"var o = {d: 1};\n (Object?.defineProperty)(o, 'c', \n{set: function(value) {\n val = value; \n} \n});", nil, []string{"missingGetterInPropertyDescriptor"}},
		{"(Reflect?.defineProperty)(obj, 'foo', {set: function(value) {}});", nil, []string{"missingGetterInPropertyDescriptor"}},
		{"(Object?.defineProperties)(obj, {foo: {set: function(value) {}}});", nil, []string{"missingGetterInPropertyDescriptor"}},
		{"(Object?.create)(null, {foo: {set: function(value) {}}});", nil, []string{"missingGetterInPropertyDescriptor"}},
		{"class A { set a(foo) {} }", nil, []string{"missingGetterInClass"}},
		{"class A { get a() {} set b(foo) {} }", AccessorPairsOptions{}, []string{"missingGetterInClass"}},
		{"class A { get a() {} }", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue}, []string{"missingSetterInClass"}},
		{"class A { set a(foo) {} }", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue}, []string{"missingGetterInClass"}},
		{"class A { static get a() {} }", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue}, []string{"missingSetterInClass"}},
		{"class A { static set a(foo) {} }", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue}, []string{"missingGetterInClass"}},
		{"A = class { get a() {} };", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue}, []string{"missingSetterInClass"}},
		{"A = class { get a() {} set b(foo) {} };", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue}, []string{"missingSetterInClass", "missingGetterInClass"}},
		{"class A { set a(value) {} }", AccessorPairsOptions{EnforceForClassMembers: accessorPairsTrue}, []string{"missingGetterInClass"}},
		{"class A { static set a(value) {} }", AccessorPairsOptions{EnforceForClassMembers: accessorPairsTrue}, []string{"missingGetterInClass"}},
		{"A = class { set a(value) {} };", AccessorPairsOptions{EnforceForClassMembers: accessorPairsTrue}, []string{"missingGetterInClass"}},
		{"(class A { static set a(value) {} });", AccessorPairsOptions{EnforceForClassMembers: accessorPairsTrue}, []string{"missingGetterInClass"}},
		{"class A { set '#a'(foo) {} }", nil, []string{"missingGetterInClass"}},
		{"class A { set #a(foo) {} }", nil, []string{"missingGetterInClass"}},
		{"class A { static set '#a'(foo) {} }", nil, []string{"missingGetterInClass"}},
		{"class A { static set #a(foo) {} }", nil, []string{"missingGetterInClass"}},
		{"class A { set a(value) {} }", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsFalse, EnforceForClassMembers: accessorPairsTrue}, []string{"missingGetterInClass"}},
		{"A = class { static set a(value) {} };", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue, EnforceForClassMembers: accessorPairsTrue}, []string{"missingGetterInClass"}},
		{"let foo = class A { get a() {} };", AccessorPairsOptions{SetWithoutGet: accessorPairsFalse, GetWithoutSet: accessorPairsTrue, EnforceForClassMembers: accessorPairsTrue}, []string{"missingSetterInClass"}},
		{"class A { static get a() {} };", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue, EnforceForClassMembers: accessorPairsTrue}, []string{"missingSetterInClass"}},
		{"(class { get a() {} });", AccessorPairsOptions{GetWithoutSet: accessorPairsTrue, EnforceForClassMembers: accessorPairsTrue}, []string{"missingSetterInClass"}},
		{"class A { get '#a'() {} };", AccessorPairsOptions{SetWithoutGet: accessorPairsFalse, GetWithoutSet: accessorPairsTrue, EnforceForClassMembers: accessorPairsTrue}, []string{"missingSetterInClass"}},
		{"class A { get #a() {} };", AccessorPairsOptions{SetWithoutGet: accessorPairsFalse, GetWithoutSet: accessorPairsTrue, EnforceForClassMembers: accessorPairsTrue}, []string{"missingSetterInClass"}},
		{"class A { static get '#a'() {} };", AccessorPairsOptions{SetWithoutGet: accessorPairsFalse, GetWithoutSet: accessorPairsTrue, EnforceForClassMembers: accessorPairsTrue}, []string{"missingSetterInClass"}},
		{"class A { static get #a() {} };", AccessorPairsOptions{SetWithoutGet: accessorPairsFalse, GetWithoutSet: accessorPairsTrue, EnforceForClassMembers: accessorPairsTrue}, []string{"missingSetterInClass"}},
		{"class A { get abc() {} }", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue, EnforceForClassMembers: accessorPairsTrue}, []string{"missingSetterInClass"}},
		{"A = class { static set 'abc'(foo) {} };", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue, EnforceForClassMembers: accessorPairsTrue}, []string{"missingGetterInClass"}},
		{"(class { get 123() {} });", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue, EnforceForClassMembers: accessorPairsTrue}, []string{"missingSetterInClass"}},
		{"class A { static get 1e2() {} }", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue, EnforceForClassMembers: accessorPairsTrue}, []string{"missingSetterInClass"}},
		{"A = class { get ['abc']() {} };", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue, EnforceForClassMembers: accessorPairsTrue}, []string{"missingSetterInClass"}},
		{"class A { set [`abc`](foo) {} }", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue, EnforceForClassMembers: accessorPairsTrue}, []string{"missingGetterInClass"}},
		{"class A { static get [123]() {} }", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue, EnforceForClassMembers: accessorPairsTrue}, []string{"missingSetterInClass"}},
		{"class A { get [abc]() {} }", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue, EnforceForClassMembers: accessorPairsTrue}, []string{"missingSetterInClass"}},
		{"class A { static get [f(abc)]() {} }", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue, EnforceForClassMembers: accessorPairsTrue}, []string{"missingSetterInClass"}},
		{"A = class { set [a + b](foo) {} };", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue, EnforceForClassMembers: accessorPairsTrue}, []string{"missingGetterInClass"}},
		{"class A { get ['constructor']() {} }", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue, EnforceForClassMembers: accessorPairsTrue}, []string{"missingSetterInClass"}},
		{"class A { get a() {} set b(foo) {} }", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue, EnforceForClassMembers: accessorPairsTrue}, []string{"missingSetterInClass", "missingGetterInClass"}},
		{"A = class { set a(foo) {} get b() {} }", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue, EnforceForClassMembers: accessorPairsTrue}, []string{"missingGetterInClass", "missingSetterInClass"}},
		{"A = class { static get a() {} static set b(foo) {} }", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue, EnforceForClassMembers: accessorPairsTrue}, []string{"missingSetterInClass", "missingGetterInClass"}},
		{"class A { get a() {} set b(foo) {} }", AccessorPairsOptions{SetWithoutGet: accessorPairsFalse, GetWithoutSet: accessorPairsTrue, EnforceForClassMembers: accessorPairsTrue}, []string{"missingSetterInClass"}},
		{"class A { get a() {} set b(foo) {} }", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsFalse, EnforceForClassMembers: accessorPairsTrue}, []string{"missingGetterInClass"}},
		{"class A { get 'a '() {} set 'a'(foo) {} }", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue, EnforceForClassMembers: accessorPairsTrue}, []string{"missingSetterInClass", "missingGetterInClass"}},
		{"class A { get 'a'() {} set 1(foo) {} }", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue, EnforceForClassMembers: accessorPairsTrue}, []string{"missingSetterInClass", "missingGetterInClass"}},
		{"class A { get 1() {} set 2(foo) {} }", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue, EnforceForClassMembers: accessorPairsTrue}, []string{"missingSetterInClass", "missingGetterInClass"}},
		{"class A { get ''() {} set null(foo) {} }", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue, EnforceForClassMembers: accessorPairsTrue}, []string{"missingSetterInClass", "missingGetterInClass"}},
		{"class A { get a() {} set [a](foo) {} }", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue, EnforceForClassMembers: accessorPairsTrue}, []string{"missingSetterInClass", "missingGetterInClass"}},
		{"class A { get [a]() {} set [b](foo) {} }", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue, EnforceForClassMembers: accessorPairsTrue}, []string{"missingSetterInClass", "missingGetterInClass"}},
		{"class A { get [a]() {} set [a++](foo) {} }", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue, EnforceForClassMembers: accessorPairsTrue}, []string{"missingSetterInClass", "missingGetterInClass"}},
		{"class A { get [a + b]() {} set [a - b](foo) {} }", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue, EnforceForClassMembers: accessorPairsTrue}, []string{"missingSetterInClass", "missingGetterInClass"}},
		{"class A { get #a() {} set '#a'(foo) {} }", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue, EnforceForClassMembers: accessorPairsTrue}, []string{"missingSetterInClass", "missingGetterInClass"}},
		{"class A { get '#a'() {} set #a(foo) {} }", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue, EnforceForClassMembers: accessorPairsTrue}, []string{"missingSetterInClass", "missingGetterInClass"}},
		{"class A { get a() {} static set a(foo) {} }", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue, EnforceForClassMembers: accessorPairsTrue}, []string{"missingSetterInClass", "missingGetterInClass"}},
		{"A = class { static get a() {} set a(foo) {} };", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue, EnforceForClassMembers: accessorPairsTrue}, []string{"missingSetterInClass", "missingGetterInClass"}},
		{"class A { set [a](foo) {} static get [a]() {} }", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue, EnforceForClassMembers: accessorPairsTrue}, []string{"missingGetterInClass", "missingSetterInClass"}},
		{"class A { static set [a](foo) {} get [a]() {} }", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue, EnforceForClassMembers: accessorPairsTrue}, []string{"missingGetterInClass", "missingSetterInClass"}},
		{"class A { get a() {} get b() {} }", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue, EnforceForClassMembers: accessorPairsTrue}, []string{"missingSetterInClass", "missingSetterInClass"}},
		{"A = class { get a() {} get [b]() {} }", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue, EnforceForClassMembers: accessorPairsTrue}, []string{"missingSetterInClass", "missingSetterInClass"}},
		{"class A { get [a]() {} get [b]() {} }", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue, EnforceForClassMembers: accessorPairsTrue}, []string{"missingSetterInClass", "missingSetterInClass"}},
		{"A = class { set a(foo) {} set b(bar) {} };", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue, EnforceForClassMembers: accessorPairsTrue}, []string{"missingGetterInClass", "missingGetterInClass"}},
		{"class A { static get a() {} static get b() {} }", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue, EnforceForClassMembers: accessorPairsTrue}, []string{"missingSetterInClass", "missingSetterInClass"}},
		{"A = class { static set a(foo) {} static set b(bar) {} }", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue, EnforceForClassMembers: accessorPairsTrue}, []string{"missingGetterInClass", "missingGetterInClass"}},
		{"class A { static get a() {} set b(foo) {} static set c(bar) {} get d() {} }", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue, EnforceForClassMembers: accessorPairsTrue}, []string{"missingSetterInClass", "missingGetterInClass", "missingGetterInClass", "missingSetterInClass"}},
		{"class A { get a() {} } class B { set a(foo) {} }", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue, EnforceForClassMembers: accessorPairsTrue}, []string{"missingSetterInClass", "missingGetterInClass"}},
		{"A = class { set a(foo) {} }, class { get a() {} };", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue, EnforceForClassMembers: accessorPairsTrue}, []string{"missingGetterInClass", "missingSetterInClass"}},
		{"A = class { get a() {} }, { set a(foo) {} }", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue, EnforceForClassMembers: accessorPairsTrue}, []string{"missingSetterInClass", "missingGetterInObjectLiteral"}},
		{"A = { get a() {} }, class { set a(foo) {} }", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue, EnforceForClassMembers: accessorPairsTrue}, []string{"missingSetterInObjectLiteral", "missingGetterInClass"}},
		{"class A { get a() {} get b() {} set b(foo) {} }", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue, EnforceForClassMembers: accessorPairsTrue}, []string{"missingSetterInClass"}},
		{"A = class { get b() {} get a() {} set b(foo) {} };", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue, EnforceForClassMembers: accessorPairsTrue}, []string{"missingSetterInClass"}},
		{"class A { set b(foo) {} get b() {} set a(bar) {} }", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue, EnforceForClassMembers: accessorPairsTrue}, []string{"missingGetterInClass"}},
		{"A = class { static get b() {} set a(foo) {} static set b(bar) {} };", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue, EnforceForClassMembers: accessorPairsTrue}, []string{"missingGetterInClass"}},
		{"class A { static set a(foo) {} get b() {} set b(bar) {} }", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue, EnforceForClassMembers: accessorPairsTrue}, []string{"missingGetterInClass"}},
		{"class A { get b() {} static get a() {} set b(bar) {} }", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue, EnforceForClassMembers: accessorPairsTrue}, []string{"missingSetterInClass"}},
		{"class A { static set b(foo) {} static get a() {} static get b() {} }", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue, EnforceForClassMembers: accessorPairsTrue}, []string{"missingSetterInClass"}},
		{"class A { get [v1](){} static set i1(foo){} static set v2(bar){} get [i2](){} static get i3(){} set [v1](baz){} static get v2(){} set i4(quux){} }", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue, EnforceForClassMembers: accessorPairsTrue}, []string{"missingGetterInClass", "missingSetterInClass", "missingSetterInClass", "missingGetterInClass"}},
		{"class A { get a() {} get a() {} }", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue, EnforceForClassMembers: accessorPairsTrue}, []string{"missingSetterInClass", "missingSetterInClass"}},
		{"A = class { set a(foo) {} set a(foo) {} };", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue, EnforceForClassMembers: accessorPairsTrue}, []string{"missingGetterInClass", "missingGetterInClass"}},
		{"A = class { static get a() {} static get a() {} };", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue, EnforceForClassMembers: accessorPairsTrue}, []string{"missingSetterInClass", "missingSetterInClass"}},
		{"class A { set a(foo) {} set a(foo) {} }", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue, EnforceForClassMembers: accessorPairsTrue}, []string{"missingGetterInClass", "missingGetterInClass"}},
		{"class A { a() {} get b() {} c() {} }", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue, EnforceForClassMembers: accessorPairsTrue}, []string{"missingSetterInClass"}},
		{"A = class { a() {} get b() {} c() {} set d(foo) {} };", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue, EnforceForClassMembers: accessorPairsTrue}, []string{"missingSetterInClass", "missingGetterInClass"}},
		{"class A { static a() {} get b() {} static c() {} }", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue, EnforceForClassMembers: accessorPairsTrue}, []string{"missingSetterInClass"}},
		{"class A { a() {} get a() {} }", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue, EnforceForClassMembers: accessorPairsTrue}, []string{"missingSetterInClass"}},
		{"A = class { static a() {} set a(foo) {} };", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue, EnforceForClassMembers: accessorPairsTrue}, []string{"missingGetterInClass"}},
		{"class A { a() {} static get b() {} c() {} }", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue, EnforceForClassMembers: accessorPairsTrue}, []string{"missingSetterInClass"}},
		{"A = class { static a() {} static set b(foo) {} static c() {} d() {} };", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue, EnforceForClassMembers: accessorPairsTrue}, []string{"missingGetterInClass"}},
		{"class A { a() {} static get a() {} a() {} }", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue, EnforceForClassMembers: accessorPairsTrue}, []string{"missingSetterInClass"}},
		{"class A { static set a(foo) {} static a() {} }", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue, EnforceForClassMembers: accessorPairsTrue}, []string{"missingGetterInClass"}},
		{"class A { get a() {} };", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue, EnforceForClassMembers: accessorPairsTrue}, []string{"missingSetterInClass"}},
		{"A = class {\n  set [\n a](foo) {} };", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue, EnforceForClassMembers: accessorPairsTrue}, []string{"missingGetterInClass"}},
		{"class A { static get b() {} };", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue, EnforceForClassMembers: accessorPairsTrue}, []string{"missingSetterInClass"}},
		{"({ set prop(value) {} });", nil, []string{"missingGetterInObjectLiteral"}},
		{"interface I { set prop(value: any): any }", AccessorPairsOptions{EnforceForTSTypes: accessorPairsTrue}, []string{"missingGetterInType"}},
		{"interface I { set prop(value: any): any, get other(): any }", AccessorPairsOptions{EnforceForTSTypes: accessorPairsTrue}, []string{"missingGetterInType"}},
		{"interface I { set prop(value: any): any, prop(): any }", AccessorPairsOptions{EnforceForTSTypes: accessorPairsTrue}, []string{"missingGetterInType"}},
		{"interface I { set [prop](value: any): any }", AccessorPairsOptions{EnforceForTSTypes: accessorPairsTrue}, []string{"missingGetterInType"}},
		{"interface I { get prop(): any } interface J { set prop(value: any): void }", AccessorPairsOptions{EnforceForTSTypes: accessorPairsTrue}, []string{"missingGetterInType"}},
		{"type T = { set prop(value: any): any }", AccessorPairsOptions{EnforceForTSTypes: accessorPairsTrue}, []string{"missingGetterInType"}},
		{"function fn(): { set prop(value: any): any }", AccessorPairsOptions{EnforceForTSTypes: accessorPairsTrue}, []string{"missingGetterInType"}},
		{"type T = { get prop(): any }", AccessorPairsOptions{EnforceForTSTypes: accessorPairsTrue, GetWithoutSet: accessorPairsTrue}, []string{"missingSetterInType"}},
	}
}

func accessorPairsSilentCases() []accessorPairsCase {
	return []accessorPairsCase{
		{"var { get: foo } = bar; ({ set: foo } = bar);", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue}, []string{}},
		{"var { set } = foo; ({ get } = foo);", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue}, []string{}},
		{"var o = { get a() {} }", nil, []string{}},
		{"var o = { get a() {} }", AccessorPairsOptions{}, []string{}},
		{"var o = {};", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue}, []string{}},
		{"var o = { a: 1 };", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue}, []string{}},
		{"var o = { a };", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue}, []string{}},
		{"var o = { a: get };", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue}, []string{}},
		{"var o = { a: set };", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue}, []string{}},
		{"var o = { get: function(){} };", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue}, []string{}},
		{"var o = { set: function(foo){} };", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue}, []string{}},
		{"var o = { get };", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue}, []string{}},
		{"var o = { set };", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue}, []string{}},
		{"var o = { [get]: function() {} };", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue}, []string{}},
		{"var o = { [set]: function(foo) {} };", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue}, []string{}},
		{"var o = { get() {} };", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue}, []string{}},
		{"var o = { set(foo) {} };", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue}, []string{}},
		{"var o = { get a() {} };", AccessorPairsOptions{SetWithoutGet: accessorPairsFalse, GetWithoutSet: accessorPairsFalse}, []string{}},
		{"var o = { get a() {} };", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsFalse}, []string{}},
		{"var o = { set a(foo) {} };", AccessorPairsOptions{SetWithoutGet: accessorPairsFalse, GetWithoutSet: accessorPairsFalse}, []string{}},
		{"var o = { set a(foo) {} };", AccessorPairsOptions{SetWithoutGet: accessorPairsFalse, GetWithoutSet: accessorPairsTrue}, []string{}},
		{"var o = { set a(foo) {} };", AccessorPairsOptions{SetWithoutGet: accessorPairsFalse}, []string{}},
		{"var o = { get a() {}, set a(foo) {} };", AccessorPairsOptions{SetWithoutGet: accessorPairsFalse, GetWithoutSet: accessorPairsTrue}, []string{}},
		{"var o = { get a() {}, set a(foo) {} };", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsFalse}, []string{}},
		{"var o = { get a() {}, set a(foo) {} };", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue}, []string{}},
		{"var o = { set a(foo) {}, get a() {} };", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue}, []string{}},
		{"var o = { get 'a'() {}, set 'a'(foo) {} };", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue}, []string{}},
		{"var o = { get a() {}, set 'a'(foo) {} };", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue}, []string{}},
		{"var o = { get ['abc']() {}, set ['abc'](foo) {} };", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue}, []string{}},
		{"var o = { get [1e2]() {}, set 100(foo) {} };", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue}, []string{}},
		{"var o = { get abc() {}, set [`abc`](foo) {} };", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue}, []string{}},
		{"var o = { get ['123']() {}, set 123(foo) {} };", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue}, []string{}},
		{"var o = { get [a]() {}, set [a](foo) {} };", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue}, []string{}},
		{"var o = { get [a]() {}, set [(a)](foo) {} };", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue}, []string{}},
		{"var o = { get [(a)]() {}, set [a](foo) {} };", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue}, []string{}},
		{"var o = { get [a]() {}, set [ a ](foo) {} };", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue}, []string{}},
		{"var o = { get [/*comment*/a/*comment*/]() {}, set [a](foo) {} };", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue}, []string{}},
		{"var o = { get [f()]() {}, set [f()](foo) {} };", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue}, []string{}},
		{"var o = { get [f(a)]() {}, set [f(a)](foo) {} };", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue}, []string{}},
		{"var o = { get [a + b]() {}, set [a + b](foo) {} };", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue}, []string{}},
		{"var o = { get [`${a}`]() {}, set [`${a}`](foo) {} };", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue}, []string{}},
		{"var o = { get a() {}, set a(foo) {}, get b() {}, set b(bar) {} };", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue}, []string{}},
		{"var o = { get a() {}, set c(foo) {}, set a(bar) {}, get b() {}, get c() {}, set b(baz) {} };", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue}, []string{}},
		{"var o = { get a() {}, set a(foo) {}, b: bar };", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue}, []string{}},
		{"var o = { get a() {}, b, set a(foo) {} };", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue}, []string{}},
		{"var o = { get a() {}, ...b, set a(foo) {} };", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue}, []string{}},
		{"var o = { get a() {}, set a(foo) {}, ...a };", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue}, []string{}},
		{"var o = { get a() {}, get a() {}, set a(foo) {}, };", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue}, []string{}},
		{"var o = { get a() {}, set a(foo) {}, get a() {} };", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue}, []string{}},
		{"var o = { get a() {}, set a(foo) {}, set a(foo) {} };", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue}, []string{}},
		{"var o = { set a(bar) {}, get a() {}, set a(foo) {} };", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue}, []string{}},
		{"var o = { get a() {}, get a() {} };", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsFalse}, []string{}},
		{"var o = { set a(foo) {}, set a(foo) {} };", AccessorPairsOptions{SetWithoutGet: accessorPairsFalse, GetWithoutSet: accessorPairsTrue}, []string{}},
		{"var o = { get a() {}, set a(foo) {}, a };", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue}, []string{}},
		{"var o = { a, get a() {}, set a(foo) {} };", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue}, []string{}},
		{"var o = { get a() {}, a:1, set a(foo) {} };", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue}, []string{}},
		{"var o = {a: 1};\n Object.defineProperty(o, 'b', \n{set: function(value) {\n val = value; \n},\n get: function() {\n return val; \n} \n});", nil, []string{}},
		{"var o = {set: function() {}}", nil, []string{}},
		{"Object.defineProperties(obj, {set: {value: function() {}}});", nil, []string{}},
		{"Object.create(null, {set: {value: function() {}}});", nil, []string{}},
		{"var o = {get: function() {}}", AccessorPairsOptions{GetWithoutSet: accessorPairsTrue}, []string{}},
		{"var o = {[set]: function() {}}", nil, []string{}},
		{"var set = 'value'; Object.defineProperty(obj, 'foo', {[set]: function(value) {}});", nil, []string{}},
		{"Object.defineProperty({ set: function(value) {} }, 'foo', { value: 1 });", nil, []string{}},
		{"Reflect.defineProperty({ get() {} }, 'foo', { value: 1 });", AccessorPairsOptions{GetWithoutSet: accessorPairsTrue}, []string{}},
		{"Object.defineProperties({ foo: { get() {} } }, { bar: { value: 1 } });", AccessorPairsOptions{GetWithoutSet: accessorPairsTrue}, []string{}},
		{"Object.create({ foo: { set(value) {} } }, { bar: { value: 1 } });", nil, []string{}},
		{"let Object; Object.defineProperty(foo, 'bar', { get() {} })", AccessorPairsOptions{GetWithoutSet: accessorPairsTrue}, []string{}},
		{"function f() { Reflect.defineProperty(foo, 'bar', { set(value) {} }); var Reflect;}", nil, []string{}},
		{"function f(Object) { Object.defineProperties(foo, { bar: { set(value) {} } }) }", nil, []string{}},
		{"if (x) { const Object = getObject(); Object.create(foo, { bar: { get() {} } }) }", AccessorPairsOptions{GetWithoutSet: accessorPairsTrue}, []string{}},
		{"class A { get a() {} }", AccessorPairsOptions{EnforceForClassMembers: accessorPairsTrue}, []string{}},
		{"class A { get #a() {} }", AccessorPairsOptions{EnforceForClassMembers: accessorPairsTrue}, []string{}},
		{"class A { set a(foo) {} }", AccessorPairsOptions{EnforceForClassMembers: accessorPairsFalse}, []string{}},
		{"class A { get a() {} set b(foo) {} static get c() {} static set d(bar) {} }", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue, EnforceForClassMembers: accessorPairsFalse}, []string{}},
		{"(class A { get a() {} set b(foo) {} static get c() {} static set d(bar) {} });", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue, EnforceForClassMembers: accessorPairsFalse}, []string{}},
		{"class A { get a() {} }", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsFalse, EnforceForClassMembers: accessorPairsTrue}, []string{}},
		{"class A { set a(foo) {} }", AccessorPairsOptions{SetWithoutGet: accessorPairsFalse, GetWithoutSet: accessorPairsTrue, EnforceForClassMembers: accessorPairsTrue}, []string{}},
		{"class A { static get a() {} }", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsFalse, EnforceForClassMembers: accessorPairsTrue}, []string{}},
		{"class A { static set a(foo) {} }", AccessorPairsOptions{SetWithoutGet: accessorPairsFalse, GetWithoutSet: accessorPairsTrue, EnforceForClassMembers: accessorPairsTrue}, []string{}},
		{"A = class { set a(foo) {} };", AccessorPairsOptions{SetWithoutGet: accessorPairsFalse, GetWithoutSet: accessorPairsTrue, EnforceForClassMembers: accessorPairsTrue}, []string{}},
		{"class A { get a() {} set b(foo) {} static get c() {} static set d(bar) {} }", AccessorPairsOptions{SetWithoutGet: accessorPairsFalse, GetWithoutSet: accessorPairsFalse, EnforceForClassMembers: accessorPairsTrue}, []string{}},
		{"class A {}", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue, EnforceForClassMembers: accessorPairsTrue}, []string{}},
		{"(class {})", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue, EnforceForClassMembers: accessorPairsTrue}, []string{}},
		{"class A { constructor () {} }", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue, EnforceForClassMembers: accessorPairsTrue}, []string{}},
		{"class A { a() {} }", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue, EnforceForClassMembers: accessorPairsTrue}, []string{}},
		{"class A { static a() {} 'b'() {} }", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue, EnforceForClassMembers: accessorPairsTrue}, []string{}},
		{"class A { [a]() {} }", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue, EnforceForClassMembers: accessorPairsTrue}, []string{}},
		{"A = class { a() {} static a() {} b() {} static c() {} }", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue, EnforceForClassMembers: accessorPairsTrue}, []string{}},
		{"class A { get a() {} set a(foo) {} }", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue, EnforceForClassMembers: accessorPairsTrue}, []string{}},
		{"class A { set a(foo) {} get a() {} }", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue, EnforceForClassMembers: accessorPairsTrue}, []string{}},
		{"class A { static get a() {} static set a(foo) {} }", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue, EnforceForClassMembers: accessorPairsTrue}, []string{}},
		{"class A { static set a(foo) {} static get a() {} }", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue, EnforceForClassMembers: accessorPairsTrue}, []string{}},
		{"(class { set a(foo) {} get a() {} });", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue, EnforceForClassMembers: accessorPairsTrue}, []string{}},
		{"class A { get 'a'() {} set ['a'](foo) {} }", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue, EnforceForClassMembers: accessorPairsTrue}, []string{}},
		{"class A { set [`a`](foo) {} get a() {} }", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue, EnforceForClassMembers: accessorPairsTrue}, []string{}},
		{"class A { get 'a'() {} set a(foo) {} }", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue, EnforceForClassMembers: accessorPairsTrue}, []string{}},
		{"A = class { static get 1e2() {} static set [100](foo) {} };", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue, EnforceForClassMembers: accessorPairsTrue}, []string{}},
		{"class A { get [a]() {} set [a](foo) {} }", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue, EnforceForClassMembers: accessorPairsTrue}, []string{}},
		{"A = class { set [(f())](foo) {} get [(f())]() {} };", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue, EnforceForClassMembers: accessorPairsTrue}, []string{}},
		{"class A { static set [f(a)](foo) {} static get [f(a)]() {} }", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue, EnforceForClassMembers: accessorPairsTrue}, []string{}},
		{"class A { get a() {} set b(foo) {} set a(bar) {} get b() {} }", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue, EnforceForClassMembers: accessorPairsTrue}, []string{}},
		{"class A { get a() {} set a(bar) {} b() {} set c(foo) {} get c() {} }", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue, EnforceForClassMembers: accessorPairsTrue}, []string{}},
		{"(class { get a() {} static set a(foo) {} set a(bar) {} static get a() {} });", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue, EnforceForClassMembers: accessorPairsTrue}, []string{}},
		{"class A { get a() {} b() {} set a(foo) {} }", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue, EnforceForClassMembers: accessorPairsTrue}, []string{}},
		{"class A { set a(foo) {} get a() {} b() {} }", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue, EnforceForClassMembers: accessorPairsTrue}, []string{}},
		{"class A { a() {} get b() {} c() {} set b(foo) {} d() {} }", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue, EnforceForClassMembers: accessorPairsTrue}, []string{}},
		{"class A { get a() {} set a(foo) {} static a() {} }", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue, EnforceForClassMembers: accessorPairsTrue}, []string{}},
		{"A = class { static get a() {} static b() {} static set a(foo) {} };", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue, EnforceForClassMembers: accessorPairsTrue}, []string{}},
		{"A = class { static set a(foo) {} static get a() {} a() {} };", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue, EnforceForClassMembers: accessorPairsTrue}, []string{}},
		{"class A { get a() {} get a() {} set a(foo) {} }", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue, EnforceForClassMembers: accessorPairsTrue}, []string{}},
		{"class A { get [a]() {} set [a](foo) {} set [a](foo) {} }", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue, EnforceForClassMembers: accessorPairsTrue}, []string{}},
		{"class A { get a() {} set 'a'(foo) {} get [`a`]() {} }", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue, EnforceForClassMembers: accessorPairsTrue}, []string{}},
		{"A = class { get a() {} set a(foo) {} a() {} }", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue, EnforceForClassMembers: accessorPairsTrue}, []string{}},
		{"A = class { a() {} get a() {} set a(foo) {} }", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue, EnforceForClassMembers: accessorPairsTrue}, []string{}},
		{"class A { static set a(foo) {} static set a(foo) {} static get a() {} }", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue, EnforceForClassMembers: accessorPairsTrue}, []string{}},
		{"class A { static get a() {} static set a(foo) {} static get a() {} }", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue, EnforceForClassMembers: accessorPairsTrue}, []string{}},
		{"class A { static set a(foo) {} static get a() {} static a() {} }", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue, EnforceForClassMembers: accessorPairsTrue}, []string{}},
		{"class A { get a() {} a() {} set a(foo) {} }", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue, EnforceForClassMembers: accessorPairsTrue}, []string{}},
		{"class A { static set a(foo) {} static a() {} static get a() {} }", AccessorPairsOptions{SetWithoutGet: accessorPairsTrue, GetWithoutSet: accessorPairsTrue, EnforceForClassMembers: accessorPairsTrue}, []string{}},
		{"interface I { get prop(): any }", nil, []string{}},
		{"type T = { set prop(value: any): void }", nil, []string{}},
		{"interface I { get prop(): any, set prop(value: any): void }", AccessorPairsOptions{EnforceForTSTypes: accessorPairsTrue}, []string{}},
		{"type T = { get prop(): any, set prop(value: any): void }", AccessorPairsOptions{EnforceForTSTypes: accessorPairsTrue}, []string{}},
		{"interface I { get prop(): any, between: true, set prop(value: any): void }", AccessorPairsOptions{EnforceForTSTypes: accessorPairsTrue}, []string{}},
		{"interface I { set prop(value: any): void, get prop(): any }", AccessorPairsOptions{EnforceForTSTypes: accessorPairsTrue}, []string{}},
		{"interface I { set prop(value: any): void, get 'prop'(): any }", AccessorPairsOptions{EnforceForTSTypes: accessorPairsTrue}, []string{}},
		{"interface I {}", AccessorPairsOptions{EnforceForTSTypes: accessorPairsTrue}, []string{}},
		{"interface I { (...args): void }", AccessorPairsOptions{EnforceForTSTypes: accessorPairsTrue}, []string{}},
		{"interface I { new(...args): unknown }", AccessorPairsOptions{EnforceForTSTypes: accessorPairsTrue}, []string{}},
		{"interface I { prop: () => any }", AccessorPairsOptions{EnforceForTSTypes: accessorPairsTrue}, []string{}},
		{"interface I { method(): any }", AccessorPairsOptions{EnforceForTSTypes: accessorPairsTrue}, []string{}},
		{"type T = { get prop(): any }", AccessorPairsOptions{EnforceForTSTypes: accessorPairsTrue}, []string{}},
	}
}

func TestAccessorPairsFires(t *testing.T) {
	t.Parallel()

	for _, testCase := range accessorPairsFiresCases() {
		t.Run(testCase.sourceText, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t, runAccessorPairs(t, testCase), testCase.wantIds...)
		})
	}
}

func TestAccessorPairsStaysSilent(t *testing.T) {
	t.Parallel()

	for _, testCase := range accessorPairsSilentCases() {
		t.Run(testCase.sourceText, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, runAccessorPairs(t, testCase))
		})
	}
}

// The three imported cases that cannot be expressed here, and what covers them instead.
//
// Upstream's corpus has three clean cases whose cleanliness comes from OUTSIDE the rule: two set
// `languageOptions.globals` to mark `Object` or `Reflect` as undefined, and one writes the same
// thing as a `/* globals Object:off */` comment directive. Our harness has no channel for either,
// so the inputs cannot be reproduced as clean and are recorded here rather than dropped or bent to
// green.
//
// Measured against the installed rule with the globals left in place, all three REPORT, and this
// asserts that agreement rather than upstream's disabled-global silence. So the divergence is in
// what the harness can say, not in what the rule decides.
//
// The mechanism those cases guard -- that the receiver must be the real global -- is covered by
// TestAccessorPairsDeclinesALocalObjectBinding below, which asks the same question with a local
// binding, a channel this tree does have. Upstream ships that case too and it is imported above.
func TestAccessorPairsGlobalShadowing(t *testing.T) {
	t.Parallel()

	cases := []accessorPairsCase{
		{"Reflect.defineProperty(foo, 'bar', { get() {} })",
			AccessorPairsOptions{GetWithoutSet: accessorPairsTrue},
			[]string{"missingSetterInPropertyDescriptor"}},
		{"/* globals Object:off */ Object.defineProperty(foo, 'bar', { set(value) {} })", nil,
			[]string{"missingGetterInPropertyDescriptor"}},
		{"Object.defineProperties(foo, { bar: { set(value) {} } })", nil,
			[]string{"missingGetterInPropertyDescriptor"}},
	}
	for _, testCase := range cases {
		t.Run(testCase.sourceText, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t, runAccessorPairs(t, testCase), testCase.wantIds...)
		})
	}
}

// A local binding named `Object` is a different object, and calling `defineProperties` on it says
// nothing about property descriptors.
//
// This is the one channel through which this tree can ask upstream's globals question, and it is the
// reason the rule declares the type checker at all: nothing syntactic separates the global `Object`
// from a parameter of the same name. Imported from upstream's own corpus and repeated here with a
// control, because it is the single case standing in for three that could not be imported.
func TestAccessorPairsDeclinesALocalObjectBinding(t *testing.T) {
	t.Parallel()

	shadowed := "function f(Object) { Object.defineProperties(foo, { bar: { set(value) {} } }) }"
	rule_testing.ExpectClean(t, rule_testing.RunTyped(t, AccessorPairs, accessorPairsFile,
		shadowed))

	// The control: the same call with no local binding reports, so the silence above is the shadow
	// rather than the shape.
	global := "Object.defineProperties(foo, { bar: { set(value) {} } })"
	rule_testing.ExpectFindings(t, rule_testing.RunTyped(t, AccessorPairs, accessorPairsFile,
		global), "missingGetterInPropertyDescriptor")
}

// The typed harness is required, and a revert to `Run` must fail loudly rather than go quiet.
//
// `GetSymbolAtLocation` on a nil checker returns nil rather than crashing, so a rule that lost its
// type information would report NOTHING on the descriptor path and every clean fixture would pass
// vacuously. This pins the direction: with no checker the descriptor finding disappears, while the
// object-literal finding, which needs no checker, survives.
func TestAccessorPairsNeedsTheTypedHarness(t *testing.T) {
	t.Parallel()

	descriptor := "Object.defineProperty(foo, 'bar', { set(value) {} })"
	rule_testing.ExpectFindings(t, rule_testing.RunTyped(t, AccessorPairs, accessorPairsFile,
		descriptor), "missingGetterInPropertyDescriptor")
	rule_testing.ExpectClean(t, rule_testing.Run(t, AccessorPairs, accessorPairsFile, descriptor))

	literal := "var o = { set a(value) {} };"
	rule_testing.ExpectFindings(t, rule_testing.Run(t, AccessorPairs, accessorPairsFile, literal),
		"missingGetterInObjectLiteral")
}

// Where the finding points, which no message-id fixture can see.
//
// The span runs from the accessor's first keyword through the end of its key, so a static member
// reports `static set a` and a quoted key reports `set 'a b'` with the quotes. Every expectation
// here was read off the INSTALLED rule's reported columns rather than off its source: upstream
// builds the span with `getFunctionHeadLoc`, whose argument is the function VALUE rather than the
// accessor, and reading that helper does not obviously predict that `static` is inside the span.
//
// A rule reporting the whole accessor including its body would pass every id fixture above.
func TestAccessorPairsPointsAtTheAccessorHead(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText string
		options    any
		wantSpans  []string
	}{
		{"var o = { set a(v) {} };", nil, []string{"set a"}},
		{"var o = { set 'a b'(v) {} };", nil, []string{"set 'a b'"}},
		{"var o = { set [x](v) {} };", nil, []string{"set [x]"}},
		{"var o = { set [ x ](v) {} };", nil, []string{"set [ x ]"}},
		{"class A { set a(v) {} }", nil, []string{"set a"}},
		{"class A { static set a(v) {} }", nil, []string{"static set a"}},
		{"class A { set #a(v) {} }", nil, []string{"set #a"}},
		{"class A { set 'a'(v) {} }", nil, []string{"set 'a'"}},
		{"class A { static get 'x y'() {} }",
			AccessorPairsOptions{GetWithoutSet: accessorPairsTrue},
			[]string{"static get 'x y'"}},
		{"var o = { get a() {} };", AccessorPairsOptions{GetWithoutSet: accessorPairsTrue},
			[]string{"get a"}},
		// Both sides of a class report, in source order, each on its own head.
		{"class A { get a() {} static set a(v) {} }",
			AccessorPairsOptions{GetWithoutSet: accessorPairsTrue,
				SetWithoutGet: accessorPairsTrue},
			[]string{"get a", "static set a"}},
		// The descriptor finding has no accessor to point at, so it points at the whole
		// descriptor object. That is a different span rule for the same rule, and it is the one
		// most likely to be silently wrong.
		{"Object.defineProperty(foo, 'bar', { set(v) {} })", nil,
			[]string{"{ set(v) {} }"}},
	}

	for _, testCase := range cases {
		t.Run(testCase.sourceText, func(t *testing.T) {
			t.Parallel()
			result := runAccessorPairs(t, accessorPairsCase{
				sourceText: testCase.sourceText, options: testCase.options})
			if len(result.Diagnostics) != len(testCase.wantSpans) {
				t.Fatalf("wanted %d findings, got %d", len(testCase.wantSpans),
					len(result.Diagnostics))
			}
			for index, want := range testCase.wantSpans {
				span := result.Diagnostics[index].Range
				got := testCase.sourceText[span.Pos():span.End()]
				if got != want {
					t.Errorf("finding %d pointed at %q, wanted %q", index, got, want)
				}
			}
		})
	}
}

// What the finding SAYS, asserted by equality on the whole rendered sentence.
//
// The accessor description is built by interpolation, so nothing above can see it: a description
// naming the wrong half, dropping `static`, or quoting a private name would satisfy every message-id
// fixture in this file. Upstream's own rendering is the specification, and these four shapes are the
// ones where its two name paths diverge.
//
// `setter '#a'` against `private setter #a` is the pair worth reading. A string key that happens to
// spell a private name renders QUOTED, while a real private name renders bare and keeps its hash.
// Upstream reaches those through two different branches, and a port collapsing them to one would
// render one of the two wrongly while every count stayed right.
func TestAccessorPairsDescribesTheAccessor(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText string
		options    any
		wantText   string
	}{
		{"var o = { set a(v) {} };", nil,
			"This declares setter 'a' with no matching getter beside it, so the property can be " +
				"written but always reads back as undefined. Declare both halves of the pair, or " +
				"drop the accessor and use a plain property."},
		{"class A { static set a(v) {} }", nil,
			"This declares static setter 'a' with no matching getter beside it, so the property " +
				"can be written but always reads back as undefined. Declare both halves of the " +
				"pair, or drop the accessor and use a plain property."},
		{"class A { set #a(v) {} }", nil,
			"This declares private setter #a with no matching getter beside it, so the property " +
				"can be written but always reads back as undefined. Declare both halves of the " +
				"pair, or drop the accessor and use a plain property."},
		{"class A { set '#a'(v) {} }", nil,
			"This declares setter '#a' with no matching getter beside it, so the property can be " +
				"written but always reads back as undefined. Declare both halves of the pair, or " +
				"drop the accessor and use a plain property."},
		// A computed key names nothing, so the description carries no name at all.
		{"var o = { set [x](v) {} };", nil,
			"This declares setter with no matching getter beside it, so the property can be " +
				"written but always reads back as undefined. Declare both halves of the pair, or " +
				"drop the accessor and use a plain property."},
		// The other half fails a different way and says so.
		{"var o = { get a() {} };", AccessorPairsOptions{GetWithoutSet: accessorPairsTrue},
			"This declares getter 'a' with no matching setter beside it, so assigning to the " +
				"property is silently ignored in sloppy mode and throws in strict mode. Declare " +
				"both halves of the pair, or drop the accessor and use a plain property."},
	}

	for _, testCase := range cases {
		t.Run(testCase.sourceText, func(t *testing.T) {
			t.Parallel()
			result := runAccessorPairs(t, accessorPairsCase{
				sourceText: testCase.sourceText, options: testCase.options})
			if len(result.Diagnostics) != 1 {
				t.Fatalf("wanted one finding, got %d", len(result.Diagnostics))
			}
			// Equality against a literal typed here, never against the rule's own constant: a
			// comparison to the constant moves with the code under mutation and asserts nothing.
			if got := result.Diagnostics[0].Message.Description; got != testCase.wantText {
				t.Errorf("rendered\n  %q\nwanted\n  %q", got, testCase.wantText)
			}
		})
	}
}

// The decoder, which is the line most likely to be wrong and has no upstream counterpart.
//
// Two of these four options default to TRUE, so the failure mode is a rule that silently stops
// reporting rather than one that errors. An absent key must keep the default; an explicit `false`
// must override it; and nil input, which is what the config layer hands a rule configured as a bare
// severity, must produce the defaults rather than the zero value.
func TestDecodeAccessorPairsOptions(t *testing.T) {
	t.Parallel()

	t.Run("nil input yields upstream's defaults", func(t *testing.T) {
		t.Parallel()
		decoded, err := DecodeAccessorPairsOptions(nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		settings := decoded.(AccessorPairsOptions).resolve()
		if settings.getWithoutSet != false || settings.setWithoutGet != true ||
			settings.enforceForClassMembers != true || settings.enforceForTSTypes != false {
			t.Fatalf("defaults came back as %+v", settings)
		}
	})

	t.Run("an explicit false overrides a true default", func(t *testing.T) {
		t.Parallel()
		decoded, err := DecodeAccessorPairsOptions([]byte(`{"setWithoutGet": false}`))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		settings := decoded.(AccessorPairsOptions).resolve()
		if settings.setWithoutGet != false {
			t.Error("an explicit false was read as the default true")
		}
		if settings.enforceForClassMembers != true {
			t.Error("an unrelated key lost its default")
		}
	})

	t.Run("the wire shape is the bare object, not upstream's array", func(t *testing.T) {
		t.Parallel()
		// cohere's config layer unwraps the `[severity, options]` tuple before dispatch, so the
		// decoder is handed `{...}` where upstream's schema writes `[{...}]`.
		if _, err := DecodeAccessorPairsOptions([]byte(`[{"setWithoutGet": false}]`)); err == nil {
			t.Error("the array spelling decoded, which means the wire shape is not what is assumed")
		}
	})
}

// A rule configured as a bare severity is handed nil options, and must still enforce the defaults.
//
// This is the failure the brief names: `options.(T)` on nil yields the zero value, which for this
// rule turns `setWithoutGet` and `enforceForClassMembers` off and leaves a rule that registers on
// every file and reports nothing. Every fixture above reaches the rule through the decoder, so none
// of them can see it.
func TestAccessorPairsWithNilOptionsUsesTheDefaults(t *testing.T) {
	t.Parallel()

	rule_testing.ExpectFindings(t, rule_testing.RunTyped(t, AccessorPairs, accessorPairsFile,
		"var o = { set a(v) {} };"), "missingGetterInObjectLiteral")
	// enforceForClassMembers defaults true, so the class arm must be live with no options at all.
	rule_testing.ExpectFindings(t, rule_testing.RunTyped(t, AccessorPairs, accessorPairsFile,
		"class A { set a(v) {} }"), "missingGetterInClass")
	// getWithoutSet defaults false, so this half must stay quiet.
	rule_testing.ExpectClean(t, rule_testing.RunTyped(t, AccessorPairs, accessorPairsFile,
		"var o = { get a() {} };"))
}

// Computed keys that differ only in SHAPE, which the imported corpus does not separate.
//
// Written for a surviving mutant: dropping the kind marker from the token rendering left every
// upstream case green, because upstream's own computed-key pairs all differ in a leaf or an
// operator too. The distinguishing shape is one where the leaves and operators agree and only the
// nesting differs, and `[[a]]` against `[a]` is the smallest of them -- an array literal holding
// the identifier renders to the same single leaf as the bare identifier.
//
// Every verdict here was measured against the installed rule before being written down, because the
// question "does upstream compare structure or text" is not answerable from `getTokens` alone: a
// token stream is flat, so upstream separates these by the BRACKET tokens it happens to include,
// while this separates them by the node kind. Different mechanisms, and the corpus does not say
// they agree, so it was measured.
func TestAccessorPairsSeparatesComputedKeysByShape(t *testing.T) {
	t.Parallel()

	both := AccessorPairsOptions{
		GetWithoutSet: accessorPairsTrue,
		SetWithoutGet: accessorPairsTrue,
	}
	cases := []accessorPairsCase{
		// An array literal wrapping the identifier is a different key from the identifier.
		{"var o = { get [[a]]() {}, set [a](v) {} };", both,
			[]string{"missingSetterInObjectLiteral", "missingGetterInObjectLiteral"}},
		// A call and its own argument likewise.
		{"var o = { get [f(a)]() {}, set [a](v) {} };", both,
			[]string{"missingSetterInObjectLiteral", "missingGetterInObjectLiteral"}},
		// Two spellings of the same property access are still different keys, because upstream
		// compares tokens rather than meaning.
		{"var o = { get [a.b]() {}, set [a[`b`]](v) {} };", both,
			[]string{"missingSetterInObjectLiteral", "missingGetterInObjectLiteral"}},
		// An object literal and an array literal holding the same identifier.
		{"var o = { get [{a}]() {}, set [[a]](v) {} };", both,
			[]string{"missingSetterInObjectLiteral", "missingGetterInObjectLiteral"}},
		// The control: identical structure pairs, so the separations above are the shape rather
		// than a rule that never pairs a computed key at all.
		{"var o = { get [a.b]() {}, set [a.b](v) {} };", both, []string{}},
	}
	for _, testCase := range cases {
		t.Run(testCase.sourceText, func(t *testing.T) {
			t.Parallel()
			result := runAccessorPairs(t, testCase)
			if len(testCase.wantIds) == 0 {
				rule_testing.ExpectClean(t, result)
				return
			}
			rule_testing.ExpectFindings(t, result, testCase.wantIds...)
		})
	}
}
