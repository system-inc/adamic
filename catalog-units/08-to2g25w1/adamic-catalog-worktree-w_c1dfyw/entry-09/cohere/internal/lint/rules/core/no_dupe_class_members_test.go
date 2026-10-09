package core

import (
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// dupeClassMembersFile is where the fixtures pretend to live.
const dupeClassMembersFile = "/repository/source/Members.ts"

// The corpus is oxc's, copied rather than rewritten.
//
// One Tester block, 36 pass and 38 fail, with the snapshot recording 40 diagnostics from 38 inputs:
// two inputs declare the same member three times and report twice, so a fixture asserting one
// finding per input is wrong on those and right everywhere else.
func TestNoDupeClassMembersFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"two methods", "class A { foo() {} foo() {} }"},
		{"inside a class expression", "!class A { foo() {} foo() {} };"},
		{"two string keys", "class A { 'foo'() {} 'foo'() {} }"},
		// Numeric keys compare by value, so two spellings of ten are one member.
		{"ten and one-e-one", "class A { 10() {} 1e1() {} }"},
		{"two computed string keys", "class A { ['foo']() {} ['foo']() {} }"},
		{"a computed key against a plain one", "class A { static ['foo']() {} static foo() {} }"},
		{"two setters of one name", "class A { set 'foo'(value) {} set ['foo'](val) {} }"},
		{"the empty string twice", "class A { ''() {} ['']() {} }"},
		{"two template keys", "class A { [`foo`]() {} [`foo`]() {} }"},
		{"two static getters", "class A { static get [`foo`]() {} static get ['foo']() {} }"},
		// A template key and a plain identifier are the same member.
		{"an identifier against a template", "class A { foo() {} [`foo`]() {} }"},
		{"a getter against a method", "class A { get [`foo`]() {} 'foo'() {} }"},
		{"two static string keys", "class A { static 'foo'() {} static [`foo`]() {} }"},
		{"constructor as a computed key", "class A { ['constructor']() {} ['constructor']() {} }"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t,
				rule_testing.Run(t, NoDupeClassMembers, dupeClassMembersFile, testCase.sourceText),
				"noDupeClassMembers")
		})
	}
}

// The clean cases carry the whole discrimination, and four of them are about key identity.
//
// `1.0` and `'1.0'` are a number and a string, so their texts agree and they are different members.
// `0x1` and a template spelling it are the same, for the same reason in reverse. `null` is a
// keyword rather than the empty string. And a computed key naming a variable is not knowable before
// the class runs, so it collides with nothing.
//
// A rule comparing source text gets two of these right by accident and two wrong; one comparing
// cooked text gets the other pair wrong. The type has to travel with the value.
func TestNoDupeClassMembersStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"different names", "class A { foo() {} bar() {} }"},
		{"static and instance are different members", "class A { static foo() {} foo() {} }"},
		{"a getter and a setter are one member", "class A { get foo() {} set foo(value) {} }"},
		{"static method with instance accessors",
			"class A { static foo() {} get foo() {} set foo(value) {} }"},
		{"two classes with the same member", "class A { foo() { } } class B { foo() { } }"},
		{"a computed key naming a variable", "class A { [foo]() {} foo() {} }"},
		{"distinct string keys", "class A { 'foo'() {} 'bar'() {} baz() {} }"},
		{"distinct generator keys", "class A { *'foo'() {} *'bar'() {} *baz() {} }"},
		{"distinct getter keys", "class A { get 'foo'() {} get 'bar'() {} get baz() {} }"},
		{"distinct numeric keys", "class A { 1() {} 2() {} }"},
		{"distinct computed string keys", "class A { ['foo']() {} ['bar']() {} }"},
		{"distinct template keys", "class A { [`foo`]() {} [`bar`]() {} }"},
		{"distinct computed numbers", "class A { [12]() {} [123]() {} }"},
		{"a number against its string spelling", "class A { [1.0]() {} ['1.0']() {} }"},
		{"hexadecimal against its text", "class A { [0x1]() {} [`0x1`]() {} }"},
		{"null against the empty string", "class A { [null]() {} ['']() {} }"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t,
				rule_testing.Run(t, NoDupeClassMembers, dupeClassMembersFile, testCase.sourceText))
		})
	}
}

// Cases written because somebody read our code rather than upstream's.
//
// A private member is a distinct namespace, so `#foo` and `foo` coexist, and two `#foo` collide.
// Upstream tracks that with an `is_private` flag on the element and its corpus never exercises it,
// which is exactly the shape a port drops silently: the flag looks like defensive copying until a
// file declares both.
func TestNoDupeClassMembersSeparatesPrivateNames(t *testing.T) {
	t.Parallel()

	rule_testing.ExpectClean(t, rule_testing.Run(t, NoDupeClassMembers, dupeClassMembersFile,
		"class A { #foo() {} foo() {} }"))

	rule_testing.ExpectFindings(t, rule_testing.Run(t, NoDupeClassMembers, dupeClassMembersFile,
		"class A { #foo() {} #foo() {} }"), "noDupeClassMembers")
}

// TypeScript overload signatures are one member, not several.
//
// Cases written because somebody ran the binary rather than because the corpus asked, which is the
// same reason `TestNoDupeClassMembersSeparatesPrivateNames` exists. The corpus is oxc's and is
// JavaScript-shaped: a class member there always has a body, so nothing in it can reach this.
//
// A method may carry several signatures and one implementation, and only the implementation has a
// body. The signatures are erased and generate no code, so reporting them says the earlier
// declaration is dead code that reads as live, which is the reverse of what an overload is.
//
// Found on `BaseSchema.ts` in a real repository, which overloads `is` and `in` twice each: four
// reports on code `tsc --noEmit` accepts without a diagnostic.
func TestNoDupeClassMembersAllowsOverloadSignatures(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"a method with two signatures",
			"class A { is(a: string): string; is(a: number): number; is(a: unknown): unknown { return a; } }"},
		{"a getter with a signature",
			"class A { get foo(): string; get foo(): unknown { return 1; } }"},
		{"the shape BaseSchema writes",
			"class A { in<const T extends readonly unknown[]>(v: T): this; in(v: readonly unknown[]): this; in(v: readonly unknown[]): unknown { return v; } }"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t,
				rule_testing.Run(t, NoDupeClassMembers, dupeClassMembersFile, testCase.sourceText))
		})
	}
}

// A real duplicate is still reported, including one standing beside overload signatures.
//
// The half that makes the test above mean something: skipping every body-less member would let a
// genuine duplicate through, and the third case is the one that catches it, since a class can carry
// signatures and a duplicate implementation at once.
func TestNoDupeClassMembersStillReportsDuplicatesBesideOverloads(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"two implementations after a signature",
			"class A { is(a: string): string; is(a: unknown): unknown { return a; } is() {} }"},
		{"two properties, which carry no body to confuse this",
			"class A { foo = 1; foo = 2; }"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t,
				rule_testing.Run(t, NoDupeClassMembers, dupeClassMembersFile, testCase.sourceText),
				"noDupeClassMembers")
		})
	}
}
