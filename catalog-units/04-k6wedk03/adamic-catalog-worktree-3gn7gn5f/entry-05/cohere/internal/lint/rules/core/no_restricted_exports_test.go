package core

import (
	"encoding/json"
	"testing"

	rule_testing "github.com/system-inc/cohere/internal/lint/testing"
)

// noRestrictedExportsFile is where the fixtures pretend to live.
const noRestrictedExportsFile = "/repository/source/NoRestrictedExports.ts"

// noRestrictedExportsFlag spells an optional boolean in a fixture row.
func noRestrictedExportsFlag(value bool) *bool { return &value }

// noRestrictedExportsCase is one imported corpus row.
type noRestrictedExportsCase struct {
	sourceText string
	options    NoRestrictedExportsOptions
	wantIds    []string
}

// runNoRestrictedExports drives one case through the rule's own exported decoder.
//
// Through the decoder rather than by handing the rule a struct, because the decoder holds the one
// schema constraint that has no spelling in the type -- `default` in `restrictedNamedExports`
// beside `restrictDefaultExports` -- and the pattern compile. A fixture bypassing it would leave
// both untested and would also not be the path a real configuration takes.
func runNoRestrictedExports(t *testing.T, testCase noRestrictedExportsCase) rule_testing.Result {
	t.Helper()

	encoded, err := json.Marshal(testCase.options)
	if err != nil {
		t.Fatalf("could not encode options: %v", err)
	}
	decoded, err := DecodeNoRestrictedExportsOptions(encoded)
	if err != nil {
		t.Fatalf("could not decode options: %v", err)
	}
	return rule_testing.RunWithOptions(t, NoRestrictedExports, noRestrictedExportsFile,
		testCase.sourceText, decoded)
}

// The corpus is ESLint's own, extracted mechanically rather than retyped.
//
// `/tmp/lint-sources-fresh/eslint/tests/lib/rules/no-restricted-exports.js` was loaded with its
// RuleTester stubbed so every case came out as data with its options attached, then every case was
// replayed against the INSTALLED ESLint 10.8.1 rule through the Linter API under `sourceType:
// module` with the typescript-eslint parser and a `.ts` filename, which is the only configuration
// cohere has. The expectations below are what the installed rule ANSWERED, not what the corpus file
// annotates.
//
// The clone (10.10.0) and the installed build (10.8.1) hold byte-identical copies of this rule,
// checked with diff, so there is no version drift to reason about here.
//
// Two instrument failures were caught on the way and are worth recording, because both printed a
// clean-looking number:
//
//   - An absolute fixture path outside the flat config's base directory made the Linter answer
//     "No matching configuration found" as a MESSAGE rather than an error, so all 183 cases
//     "reported" and the control read as satisfied. The extractor now refuses any message with no
//     messageId.
//   - The first interceptor kept only one RuleTester.run call. This corpus makes one, so it was
//     harmless here, but it is the same shape that cost another porter 83% of a corpus.
//
// With both closed the corpus agrees with its own annotations exactly: 95 cases upstream calls
// valid report nothing here, 88 it calls invalid report, and zero disagree in either direction.
func noRestrictedExportsCases() []noRestrictedExportsCase {
	return []noRestrictedExportsCase{
		{"export var a;", NoRestrictedExportsOptions{}, []string{}},
		{"export function a() {}", NoRestrictedExportsOptions{}, []string{}},
		{"export class A {}", NoRestrictedExportsOptions{}, []string{}},
		{"var a; export { a };", NoRestrictedExportsOptions{}, []string{}},
		{"var b; export { b as a };", NoRestrictedExportsOptions{}, []string{}},
		{"export { a } from 'foo';", NoRestrictedExportsOptions{}, []string{}},
		{"export { b as a } from 'foo';", NoRestrictedExportsOptions{}, []string{}},
		{"export var a;", NoRestrictedExportsOptions{}, []string{}},
		{"export function a() {}", NoRestrictedExportsOptions{}, []string{}},
		{"export class A {}", NoRestrictedExportsOptions{}, []string{}},
		{"var a; export { a };", NoRestrictedExportsOptions{}, []string{}},
		{"var b; export { b as a };", NoRestrictedExportsOptions{}, []string{}},
		{"export { a } from 'foo';", NoRestrictedExportsOptions{}, []string{}},
		{"export { b as a } from 'foo';", NoRestrictedExportsOptions{}, []string{}},
		{"export var a;", NoRestrictedExportsOptions{RestrictedNamedExports: []string{}}, []string{}},
		{"export function a() {}", NoRestrictedExportsOptions{RestrictedNamedExports: []string{}}, []string{}},
		{"export class A {}", NoRestrictedExportsOptions{RestrictedNamedExports: []string{}}, []string{}},
		{"var a; export { a };", NoRestrictedExportsOptions{RestrictedNamedExports: []string{}}, []string{}},
		{"var b; export { b as a };", NoRestrictedExportsOptions{RestrictedNamedExports: []string{}}, []string{}},
		{"export { a } from 'foo';", NoRestrictedExportsOptions{RestrictedNamedExports: []string{}}, []string{}},
		{"export { b as a } from 'foo';", NoRestrictedExportsOptions{RestrictedNamedExports: []string{}}, []string{}},
		{"export var a;", NoRestrictedExportsOptions{RestrictedNamedExports: []string{"x"}}, []string{}},
		{"export let a;", NoRestrictedExportsOptions{RestrictedNamedExports: []string{"x"}}, []string{}},
		{"export const a = 1;", NoRestrictedExportsOptions{RestrictedNamedExports: []string{"x"}}, []string{}},
		{"export function a() {}", NoRestrictedExportsOptions{RestrictedNamedExports: []string{"x"}}, []string{}},
		{"export function *a() {}", NoRestrictedExportsOptions{RestrictedNamedExports: []string{"x"}}, []string{}},
		{"export async function a() {}", NoRestrictedExportsOptions{RestrictedNamedExports: []string{"x"}}, []string{}},
		{"export async function *a() {}", NoRestrictedExportsOptions{RestrictedNamedExports: []string{"x"}}, []string{}},
		{"export class A {}", NoRestrictedExportsOptions{RestrictedNamedExports: []string{"x"}}, []string{}},
		{"var a; export { a };", NoRestrictedExportsOptions{RestrictedNamedExports: []string{"x"}}, []string{}},
		{"var b; export { b as a };", NoRestrictedExportsOptions{RestrictedNamedExports: []string{"x"}}, []string{}},
		{"export { a } from 'foo';", NoRestrictedExportsOptions{RestrictedNamedExports: []string{"x"}}, []string{}},
		{"export { b as a } from 'foo';", NoRestrictedExportsOptions{RestrictedNamedExports: []string{"x"}}, []string{}},
		{"export { '' } from 'foo';", NoRestrictedExportsOptions{RestrictedNamedExports: []string{"undefined"}}, []string{}},
		{"export { '' } from 'foo';", NoRestrictedExportsOptions{RestrictedNamedExports: []string{" "}}, []string{}},
		{"export { ' ' } from 'foo';", NoRestrictedExportsOptions{RestrictedNamedExports: []string{""}}, []string{}},
		{"export { ' a', 'a ' } from 'foo';", NoRestrictedExportsOptions{RestrictedNamedExports: []string{"a"}}, []string{}},
		{"export var b = a;", NoRestrictedExportsOptions{RestrictedNamedExports: []string{"a"}}, []string{}},
		{"export let [b = a] = [];", NoRestrictedExportsOptions{RestrictedNamedExports: []string{"a"}}, []string{}},
		{"export const [b] = [a];", NoRestrictedExportsOptions{RestrictedNamedExports: []string{"a"}}, []string{}},
		{"export var { a: b } = {};", NoRestrictedExportsOptions{RestrictedNamedExports: []string{"a"}}, []string{}},
		{"export let { b = a } = {};", NoRestrictedExportsOptions{RestrictedNamedExports: []string{"a"}}, []string{}},
		{"export const { c: b = a } = {};", NoRestrictedExportsOptions{RestrictedNamedExports: []string{"a"}}, []string{}},
		{"export function b(a) {}", NoRestrictedExportsOptions{RestrictedNamedExports: []string{"a"}}, []string{}},
		{"export class A { a(){} }", NoRestrictedExportsOptions{RestrictedNamedExports: []string{"a"}}, []string{}},
		{"export class A extends B {}", NoRestrictedExportsOptions{RestrictedNamedExports: []string{"B"}}, []string{}},
		{"var a; export { a as b };", NoRestrictedExportsOptions{RestrictedNamedExports: []string{"a"}}, []string{}},
		{"var a; export { a as 'a ' };", NoRestrictedExportsOptions{RestrictedNamedExports: []string{"a"}}, []string{}},
		{"export { a as b } from 'foo';", NoRestrictedExportsOptions{RestrictedNamedExports: []string{"a"}}, []string{}},
		{"export { a as 'a ' } from 'foo';", NoRestrictedExportsOptions{RestrictedNamedExports: []string{"a"}}, []string{}},
		{"export { 'a' as 'a ' } from 'foo';", NoRestrictedExportsOptions{RestrictedNamedExports: []string{"a"}}, []string{}},
		{"export { b } from 'a';", NoRestrictedExportsOptions{RestrictedNamedExports: []string{"a"}}, []string{}},
		{"export * as b from 'a';", NoRestrictedExportsOptions{RestrictedNamedExports: []string{"a"}}, []string{}},
		{"var a;", NoRestrictedExportsOptions{RestrictedNamedExports: []string{"a"}}, []string{}},
		{"let a;", NoRestrictedExportsOptions{RestrictedNamedExports: []string{"a"}}, []string{}},
		{"const a = 1;", NoRestrictedExportsOptions{RestrictedNamedExports: []string{"a"}}, []string{}},
		{"function a() {}", NoRestrictedExportsOptions{RestrictedNamedExports: []string{"a"}}, []string{}},
		{"class A {}", NoRestrictedExportsOptions{RestrictedNamedExports: []string{"A"}}, []string{}},
		{"import a from 'foo';", NoRestrictedExportsOptions{RestrictedNamedExports: []string{"a"}}, []string{}},
		{"import { a } from 'foo';", NoRestrictedExportsOptions{RestrictedNamedExports: []string{"a"}}, []string{}},
		{"import { b as a } from 'foo';", NoRestrictedExportsOptions{RestrictedNamedExports: []string{"a"}}, []string{}},
		{"var setSomething; export { setSomething };", NoRestrictedExportsOptions{RestrictedNamedExportsPattern: "^get"}, []string{}},
		{"var foo, bar; export { foo, bar };", NoRestrictedExportsOptions{RestrictedNamedExportsPattern: "^(?!foo)(?!bar).+$"}, []string{}},
		{"var foobar; export default foobar;", NoRestrictedExportsOptions{RestrictedNamedExportsPattern: "bar$"}, []string{}},
		{"var foobar; export default foobar;", NoRestrictedExportsOptions{RestrictedNamedExportsPattern: "default"}, []string{}},
		{"export default 'default';", NoRestrictedExportsOptions{RestrictedNamedExportsPattern: "default"}, []string{}},
		{"var foobar; export { foobar as default };", NoRestrictedExportsOptions{RestrictedNamedExportsPattern: "default"}, []string{}},
		{"var foobar; export { foobar as 'default' };", NoRestrictedExportsOptions{RestrictedNamedExportsPattern: "default"}, []string{}},
		{"export { default } from 'mod';", NoRestrictedExportsOptions{RestrictedNamedExportsPattern: "default"}, []string{}},
		{"export { default as default } from 'mod';", NoRestrictedExportsOptions{RestrictedNamedExportsPattern: "default"}, []string{}},
		{"export { foobar as default } from 'mod';", NoRestrictedExportsOptions{RestrictedNamedExportsPattern: "default"}, []string{}},
		{"export * as default from 'mod';", NoRestrictedExportsOptions{RestrictedNamedExportsPattern: "default"}, []string{}},
		{"export * from 'foo';", NoRestrictedExportsOptions{RestrictedNamedExports: []string{"a"}}, []string{}},
		{"export * from 'a';", NoRestrictedExportsOptions{RestrictedNamedExports: []string{"a"}}, []string{}},
		{"export default a;", NoRestrictedExportsOptions{RestrictedNamedExports: []string{"a"}}, []string{}},
		{"export default function a() {}", NoRestrictedExportsOptions{RestrictedNamedExports: []string{"a"}}, []string{}},
		{"export default class A {}", NoRestrictedExportsOptions{RestrictedNamedExports: []string{"A"}}, []string{}},
		{"export default (function a() {});", NoRestrictedExportsOptions{RestrictedNamedExports: []string{"a"}}, []string{}},
		{"export default (class A {});", NoRestrictedExportsOptions{RestrictedNamedExports: []string{"A"}}, []string{}},
		{"export default 1;", NoRestrictedExportsOptions{RestrictedNamedExports: []string{"default"}}, []string{}},
		{"export { default as a } from 'foo';", NoRestrictedExportsOptions{RestrictedNamedExports: []string{"default"}}, []string{}},
		{"export default foo;", NoRestrictedExportsOptions{RestrictDefaultExports: &NoRestrictedExportsRestrictDefaultExports{Direct: noRestrictedExportsFlag(false)}}, []string{}},
		{"export default 42;", NoRestrictedExportsOptions{RestrictDefaultExports: &NoRestrictedExportsRestrictDefaultExports{Direct: noRestrictedExportsFlag(false)}}, []string{}},
		{"export default function foo() {}", NoRestrictedExportsOptions{RestrictDefaultExports: &NoRestrictedExportsRestrictDefaultExports{Direct: noRestrictedExportsFlag(false)}}, []string{}},
		{"const foo = 123;\nexport { foo as default };", NoRestrictedExportsOptions{RestrictDefaultExports: &NoRestrictedExportsRestrictDefaultExports{Named: noRestrictedExportsFlag(false)}}, []string{}},
		{"export { default } from 'mod';", NoRestrictedExportsOptions{RestrictDefaultExports: &NoRestrictedExportsRestrictDefaultExports{DefaultFrom: noRestrictedExportsFlag(false)}}, []string{}},
		{"export { default as default } from 'mod';", NoRestrictedExportsOptions{RestrictDefaultExports: &NoRestrictedExportsRestrictDefaultExports{DefaultFrom: noRestrictedExportsFlag(false)}}, []string{}},
		{"export { foo as default } from 'mod';", NoRestrictedExportsOptions{RestrictDefaultExports: &NoRestrictedExportsRestrictDefaultExports{DefaultFrom: noRestrictedExportsFlag(true)}}, []string{}},
		{"export { default } from 'mod';", NoRestrictedExportsOptions{RestrictDefaultExports: &NoRestrictedExportsRestrictDefaultExports{Named: noRestrictedExportsFlag(true), DefaultFrom: noRestrictedExportsFlag(false)}}, []string{}},
		{"export { 'default' } from 'mod'; ", NoRestrictedExportsOptions{RestrictDefaultExports: &NoRestrictedExportsRestrictDefaultExports{DefaultFrom: noRestrictedExportsFlag(false)}}, []string{}},
		{"export { foo as default } from 'mod';", NoRestrictedExportsOptions{RestrictDefaultExports: &NoRestrictedExportsRestrictDefaultExports{NamedFrom: noRestrictedExportsFlag(false)}}, []string{}},
		{"export { default as default } from 'mod';", NoRestrictedExportsOptions{RestrictDefaultExports: &NoRestrictedExportsRestrictDefaultExports{NamedFrom: noRestrictedExportsFlag(true)}}, []string{}},
		{"export { default as default } from 'mod';", NoRestrictedExportsOptions{RestrictDefaultExports: &NoRestrictedExportsRestrictDefaultExports{NamedFrom: noRestrictedExportsFlag(false)}}, []string{}},
		{"export { 'default' } from 'mod'; ", NoRestrictedExportsOptions{RestrictDefaultExports: &NoRestrictedExportsRestrictDefaultExports{DefaultFrom: noRestrictedExportsFlag(false), NamedFrom: noRestrictedExportsFlag(true)}}, []string{}},
		{"export * as default from 'mod';", NoRestrictedExportsOptions{RestrictDefaultExports: &NoRestrictedExportsRestrictDefaultExports{NamespaceFrom: noRestrictedExportsFlag(false)}}, []string{}},
		{"export function someFunction() {}", NoRestrictedExportsOptions{RestrictedNamedExports: []string{"someFunction"}}, []string{"restrictedNamed"}},
		{"export var a;", NoRestrictedExportsOptions{RestrictedNamedExports: []string{"a"}}, []string{"restrictedNamed"}},
		{"export var a = 1;", NoRestrictedExportsOptions{RestrictedNamedExports: []string{"a"}}, []string{"restrictedNamed"}},
		{"export let a;", NoRestrictedExportsOptions{RestrictedNamedExports: []string{"a"}}, []string{"restrictedNamed"}},
		{"export let a = 1;", NoRestrictedExportsOptions{RestrictedNamedExports: []string{"a"}}, []string{"restrictedNamed"}},
		{"export const a = 1;", NoRestrictedExportsOptions{RestrictedNamedExports: []string{"a"}}, []string{"restrictedNamed"}},
		{"export function a() {}", NoRestrictedExportsOptions{RestrictedNamedExports: []string{"a"}}, []string{"restrictedNamed"}},
		{"export function *a() {}", NoRestrictedExportsOptions{RestrictedNamedExports: []string{"a"}}, []string{"restrictedNamed"}},
		{"export async function a() {}", NoRestrictedExportsOptions{RestrictedNamedExports: []string{"a"}}, []string{"restrictedNamed"}},
		{"export async function *a() {}", NoRestrictedExportsOptions{RestrictedNamedExports: []string{"a"}}, []string{"restrictedNamed"}},
		{"export class A {}", NoRestrictedExportsOptions{RestrictedNamedExports: []string{"A"}}, []string{"restrictedNamed"}},
		{"let a; export { a };", NoRestrictedExportsOptions{RestrictedNamedExports: []string{"a"}}, []string{"restrictedNamed"}},
		{"export { a }; var a;", NoRestrictedExportsOptions{RestrictedNamedExports: []string{"a"}}, []string{"restrictedNamed"}},
		{"let b; export { b as a };", NoRestrictedExportsOptions{RestrictedNamedExports: []string{"a"}}, []string{"restrictedNamed"}},
		{"export { a } from 'foo';", NoRestrictedExportsOptions{RestrictedNamedExports: []string{"a"}}, []string{"restrictedNamed"}},
		{"export { b as a } from 'foo';", NoRestrictedExportsOptions{RestrictedNamedExports: []string{"a"}}, []string{"restrictedNamed"}},
		{"let a; export { a as 'a' };", NoRestrictedExportsOptions{RestrictedNamedExports: []string{"a"}}, []string{"restrictedNamed"}},
		{"let a; export { a as 'b' };", NoRestrictedExportsOptions{RestrictedNamedExports: []string{"b"}}, []string{"restrictedNamed"}},
		{"let a; export { a as ' b ' };", NoRestrictedExportsOptions{RestrictedNamedExports: []string{" b "}}, []string{"restrictedNamed"}},
		{"let a; export { a as '👍' };", NoRestrictedExportsOptions{RestrictedNamedExports: []string{"👍"}}, []string{"restrictedNamed"}},
		{"export { 'a' } from 'foo';", NoRestrictedExportsOptions{RestrictedNamedExports: []string{"a"}}, []string{"restrictedNamed"}},
		{"export { '' } from 'foo';", NoRestrictedExportsOptions{RestrictedNamedExports: []string{""}}, []string{"restrictedNamed"}},
		{"export { ' ' } from 'foo';", NoRestrictedExportsOptions{RestrictedNamedExports: []string{" "}}, []string{"restrictedNamed"}},
		{"export { b as 'a' } from 'foo';", NoRestrictedExportsOptions{RestrictedNamedExports: []string{"a"}}, []string{"restrictedNamed"}},
		{"export { b as '\\u0061' } from 'foo';", NoRestrictedExportsOptions{RestrictedNamedExports: []string{"a"}}, []string{"restrictedNamed"}},
		{"export * as 'a' from 'foo';", NoRestrictedExportsOptions{RestrictedNamedExports: []string{"a"}}, []string{"restrictedNamed"}},
		{"export var [a] = [];", NoRestrictedExportsOptions{RestrictedNamedExports: []string{"a"}}, []string{"restrictedNamed"}},
		{"export let { a } = {};", NoRestrictedExportsOptions{RestrictedNamedExports: []string{"a"}}, []string{"restrictedNamed"}},
		{"export const { b: a } = {};", NoRestrictedExportsOptions{RestrictedNamedExports: []string{"a"}}, []string{"restrictedNamed"}},
		{"export var [{ a }] = [];", NoRestrictedExportsOptions{RestrictedNamedExports: []string{"a"}}, []string{"restrictedNamed"}},
		{"export let { b: { c: a = d } = e } = {};", NoRestrictedExportsOptions{RestrictedNamedExports: []string{"a"}}, []string{"restrictedNamed"}},
		{"var a; export var a;", NoRestrictedExportsOptions{RestrictedNamedExports: []string{"a"}}, []string{"restrictedNamed"}},
		{"export var a; var a;", NoRestrictedExportsOptions{RestrictedNamedExports: []string{"a"}}, []string{"restrictedNamed"}},
		{"export var a = a;", NoRestrictedExportsOptions{RestrictedNamedExports: []string{"a"}}, []string{"restrictedNamed"}},
		{"export let b = a, a;", NoRestrictedExportsOptions{RestrictedNamedExports: []string{"a"}}, []string{"restrictedNamed"}},
		{"export const a = 1, b = a;", NoRestrictedExportsOptions{RestrictedNamedExports: []string{"a"}}, []string{"restrictedNamed"}},
		{"export var [a] = a;", NoRestrictedExportsOptions{RestrictedNamedExports: []string{"a"}}, []string{"restrictedNamed"}},
		{"export let { a: a } = {};", NoRestrictedExportsOptions{RestrictedNamedExports: []string{"a"}}, []string{"restrictedNamed"}},
		{"export const { a: b, b: a } = {};", NoRestrictedExportsOptions{RestrictedNamedExports: []string{"a"}}, []string{"restrictedNamed"}},
		{"export var { b: a, a: b } = {};", NoRestrictedExportsOptions{RestrictedNamedExports: []string{"a"}}, []string{"restrictedNamed"}},
		{"export let a, { a: b } = {};", NoRestrictedExportsOptions{RestrictedNamedExports: []string{"a"}}, []string{"restrictedNamed"}},
		{"export const { a: b } = {}, a = 1;", NoRestrictedExportsOptions{RestrictedNamedExports: []string{"a"}}, []string{"restrictedNamed"}},
		{"export var [a = a] = [];", NoRestrictedExportsOptions{RestrictedNamedExports: []string{"a"}}, []string{"restrictedNamed"}},
		{"export var { a: a = a } = {};", NoRestrictedExportsOptions{RestrictedNamedExports: []string{"a"}}, []string{"restrictedNamed"}},
		{"export let { a } = { a };", NoRestrictedExportsOptions{RestrictedNamedExports: []string{"a"}}, []string{"restrictedNamed"}},
		{"export function a(a) {};", NoRestrictedExportsOptions{RestrictedNamedExports: []string{"a"}}, []string{"restrictedNamed"}},
		{"export class A { A(){} };", NoRestrictedExportsOptions{RestrictedNamedExports: []string{"A"}}, []string{"restrictedNamed"}},
		{"var a; export { a as a };", NoRestrictedExportsOptions{RestrictedNamedExports: []string{"a"}}, []string{"restrictedNamed"}},
		{"let a, b; export { a as b, b as a };", NoRestrictedExportsOptions{RestrictedNamedExports: []string{"a"}}, []string{"restrictedNamed"}},
		{"const a = 1, b = 2; export { b as a, a as b };", NoRestrictedExportsOptions{RestrictedNamedExports: []string{"a"}}, []string{"restrictedNamed"}},
		{"var a; export { a as b, a };", NoRestrictedExportsOptions{RestrictedNamedExports: []string{"a"}}, []string{"restrictedNamed"}},
		{"export { a as a } from 'a';", NoRestrictedExportsOptions{RestrictedNamedExports: []string{"a"}}, []string{"restrictedNamed"}},
		{"export { a as b, b as a } from 'foo';", NoRestrictedExportsOptions{RestrictedNamedExports: []string{"a"}}, []string{"restrictedNamed"}},
		{"export { b as a, a as b } from 'foo';", NoRestrictedExportsOptions{RestrictedNamedExports: []string{"a"}}, []string{"restrictedNamed"}},
		{"export * as a from 'a';", NoRestrictedExportsOptions{RestrictedNamedExports: []string{"a"}}, []string{"restrictedNamed"}},
		{"export var a, b;", NoRestrictedExportsOptions{RestrictedNamedExports: []string{"a"}}, []string{"restrictedNamed"}},
		{"export let b, a;", NoRestrictedExportsOptions{RestrictedNamedExports: []string{"a"}}, []string{"restrictedNamed"}},
		{"export const b = 1, a = 2;", NoRestrictedExportsOptions{RestrictedNamedExports: []string{"a", "b"}}, []string{"restrictedNamed", "restrictedNamed"}},
		{"export var a, b, c;", NoRestrictedExportsOptions{RestrictedNamedExports: []string{"a", "c"}}, []string{"restrictedNamed", "restrictedNamed"}},
		{"export let { a, b, c } = {};", NoRestrictedExportsOptions{RestrictedNamedExports: []string{"b", "c"}}, []string{"restrictedNamed", "restrictedNamed"}},
		{"export const [a, b, c, d] = {};", NoRestrictedExportsOptions{RestrictedNamedExports: []string{"b", "c"}}, []string{"restrictedNamed", "restrictedNamed"}},
		{"export var { a, x: b, c, d, e: y } = {}, e, f = {};", NoRestrictedExportsOptions{RestrictedNamedExports: []string{"foo", "a", "b", "bar", "d", "e", "baz"}}, []string{"restrictedNamed", "restrictedNamed", "restrictedNamed", "restrictedNamed"}},
		{"var a, b; export { a, b };", NoRestrictedExportsOptions{RestrictedNamedExports: []string{"a"}}, []string{"restrictedNamed"}},
		{"let a, b; export { b, a };", NoRestrictedExportsOptions{RestrictedNamedExports: []string{"a"}}, []string{"restrictedNamed"}},
		{"const a = 1, b = 1; export { a, b };", NoRestrictedExportsOptions{RestrictedNamedExports: []string{"a", "b"}}, []string{"restrictedNamed", "restrictedNamed"}},
		{"export { a, b, c }; var a, b, c;", NoRestrictedExportsOptions{RestrictedNamedExports: []string{"a", "c"}}, []string{"restrictedNamed", "restrictedNamed"}},
		{"export { b as a, b } from 'foo';", NoRestrictedExportsOptions{RestrictedNamedExports: []string{"a"}}, []string{"restrictedNamed"}},
		{"export { b as a, b } from 'foo';", NoRestrictedExportsOptions{RestrictedNamedExports: []string{"b"}}, []string{"restrictedNamed"}},
		{"export { b as a, b } from 'foo';", NoRestrictedExportsOptions{RestrictedNamedExports: []string{"a", "b"}}, []string{"restrictedNamed", "restrictedNamed"}},
		{"export { a, b, c, d, x as e, f, g } from 'foo';", NoRestrictedExportsOptions{RestrictedNamedExports: []string{"foo", "b", "bar", "d", "e", "f", "baz"}}, []string{"restrictedNamed", "restrictedNamed", "restrictedNamed", "restrictedNamed"}},
		{"var getSomething; export { getSomething };", NoRestrictedExportsOptions{RestrictedNamedExportsPattern: "get*"}, []string{"restrictedNamed"}},
		{"var getSomethingFromUser; export { getSomethingFromUser };", NoRestrictedExportsOptions{RestrictedNamedExportsPattern: "User$"}, []string{"restrictedNamed"}},
		{"var foo, ab, xy; export { foo, ab, xy };", NoRestrictedExportsOptions{RestrictedNamedExportsPattern: "(b|y)$"}, []string{"restrictedNamed", "restrictedNamed"}},
		{"var foo; export { foo as ab };", NoRestrictedExportsOptions{RestrictedNamedExportsPattern: "(b|y)$"}, []string{"restrictedNamed"}},
		{"var privateUserEmail; export { privateUserEmail };", NoRestrictedExportsOptions{RestrictedNamedExportsPattern: "^privateUser"}, []string{"restrictedNamed"}},
		{"export const a = 1;", NoRestrictedExportsOptions{RestrictedNamedExportsPattern: "^(?!foo)(?!bar).+$"}, []string{"restrictedNamed"}},
		{"var a; export { a as default };", NoRestrictedExportsOptions{RestrictedNamedExports: []string{"default"}}, []string{"restrictedNamed"}},
		{"export { default } from 'foo';", NoRestrictedExportsOptions{RestrictedNamedExports: []string{"default"}}, []string{"restrictedNamed"}},
		{"export default foo;", NoRestrictedExportsOptions{RestrictDefaultExports: &NoRestrictedExportsRestrictDefaultExports{Direct: noRestrictedExportsFlag(true)}}, []string{"restrictedDefault"}},
		{"export default 42;", NoRestrictedExportsOptions{RestrictDefaultExports: &NoRestrictedExportsRestrictDefaultExports{Direct: noRestrictedExportsFlag(true)}}, []string{"restrictedDefault"}},
		{"export default function foo() {}", NoRestrictedExportsOptions{RestrictDefaultExports: &NoRestrictedExportsRestrictDefaultExports{Direct: noRestrictedExportsFlag(true)}}, []string{"restrictedDefault"}},
		{"export default foo;", NoRestrictedExportsOptions{RestrictedNamedExports: []string{"bar"}, RestrictDefaultExports: &NoRestrictedExportsRestrictDefaultExports{Direct: noRestrictedExportsFlag(true)}}, []string{"restrictedDefault"}},
		{"const foo = 123;\nexport { foo as default };", NoRestrictedExportsOptions{RestrictDefaultExports: &NoRestrictedExportsRestrictDefaultExports{Named: noRestrictedExportsFlag(true)}}, []string{"restrictedDefault"}},
		{"export { default } from 'mod';", NoRestrictedExportsOptions{RestrictDefaultExports: &NoRestrictedExportsRestrictDefaultExports{DefaultFrom: noRestrictedExportsFlag(true)}}, []string{"restrictedDefault"}},
		{"export { default as default } from 'mod';", NoRestrictedExportsOptions{RestrictDefaultExports: &NoRestrictedExportsRestrictDefaultExports{DefaultFrom: noRestrictedExportsFlag(true)}}, []string{"restrictedDefault"}},
		{"export { 'default' } from 'mod';", NoRestrictedExportsOptions{RestrictDefaultExports: &NoRestrictedExportsRestrictDefaultExports{DefaultFrom: noRestrictedExportsFlag(true)}}, []string{"restrictedDefault"}},
		{"export { foo as default } from 'mod';", NoRestrictedExportsOptions{RestrictDefaultExports: &NoRestrictedExportsRestrictDefaultExports{NamedFrom: noRestrictedExportsFlag(true)}}, []string{"restrictedDefault"}},
		{"export * as default from 'mod';", NoRestrictedExportsOptions{RestrictDefaultExports: &NoRestrictedExportsRestrictDefaultExports{NamespaceFrom: noRestrictedExportsFlag(true)}}, []string{"restrictedDefault"}},
	}
}

// TestNoRestrictedExportsMatchesUpstream replays the whole corpus.
func TestNoRestrictedExportsMatchesUpstream(t *testing.T) {
	t.Parallel()

	cases := noRestrictedExportsCases()

	// The corpus counts are asserted rather than assumed. A generator that dropped rows would leave
	// a suite that passes on whatever survived, which is the shape this project keeps finding.
	reporting := 0
	for _, testCase := range cases {
		if len(testCase.wantIds) > 0 {
			reporting++
		}
	}
	if len(cases) != 183 {
		t.Fatalf("expected 183 corpus cases, have %d", len(cases))
	}
	if reporting != 88 {
		t.Fatalf("expected 88 reporting cases, have %d", reporting)
	}

	for _, testCase := range cases {
		result := runNoRestrictedExports(t, testCase)
		rule_testing.ExpectFindings(t, result, testCase.wantIds...)
	}
}

// TestNoRestrictedExportsIsSilentWithNoConfiguration is the half a corpus of configured cases cannot
// cover.
//
// Every corpus row supplies options, so none of them can tell a rule that declines an unconfigured
// file from one that is simply never reached. This asserts the decline directly, on source that the
// rule reports on the moment a name is listed.
func TestNoRestrictedExportsIsSilentWithNoConfiguration(t *testing.T) {
	t.Parallel()

	const source = "export var a; export default b; export { c } from 'm';"

	unconfigured := rule_testing.RunWithOptions(t, NoRestrictedExports, noRestrictedExportsFile,
		source, NoRestrictedExportsOptions{})
	rule_testing.ExpectClean(t, unconfigured)

	// The control. Without it the assertion above passes for a rule that can never report at all.
	configured := runNoRestrictedExports(t, noRestrictedExportsCase{
		sourceText: source,
		options: NoRestrictedExportsOptions{
			RestrictedNamedExports: []string{"a", "c"},
			RestrictDefaultExports: &NoRestrictedExportsRestrictDefaultExports{
				Direct: noRestrictedExportsFlag(true),
			},
		},
	})
	rule_testing.ExpectFindings(t, configured, "restrictedNamed", "restrictedDefault", "restrictedNamed")
}

// TestNoRestrictedExportsRejectsContradictoryConfiguration covers the decoder's own arms.
//
// Both are cross-field or cross-library constraints that no fixture asserting message ids can
// reach, and both would otherwise fail silently: an uncompilable pattern would match nothing and
// read as a clean tree, and the contradictory pairing would report under whichever message the
// ordering happened to reach first.
func TestNoRestrictedExportsRejectsContradictoryConfiguration(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		configured string
		wantError  bool
	}{
		{"default beside the flags", `{"restrictedNamedExports":["default"],"restrictDefaultExports":{"direct":true}}`, true},
		{"default alone is fine", `{"restrictedNamedExports":["default"]}`, false},
		{"an uncompilable pattern", `{"restrictedNamedExportsPattern":"("}`, true},
		{"a lookahead pattern compiles", `{"restrictedNamedExportsPattern":"^(?!foo)(?!bar).+$"}`, false},
		{"no options at all", ``, false},
	}
	for _, testCase := range cases {
		_, err := DecodeNoRestrictedExportsOptions([]byte(testCase.configured))
		if testCase.wantError && err == nil {
			t.Errorf("%s: expected the decoder to refuse %s", testCase.name, testCase.configured)
		}
		if !testCase.wantError && err != nil {
			t.Errorf("%s: expected the decoder to accept %s, got %v", testCase.name, testCase.configured, err)
		}
	}
}

// TestNoRestrictedExportsNamedRestrictionWinsOverTheDefaultFlags pins upstream's arm ORDER, which no
// corpus case can reach.
//
// `checkExportedName` reports the named message and RETURNS, so the five default-export arms below
// it never run for a name that was already restricted. That return is invisible to every imported
// fixture: reaching it needs `default` listed in `restrictedNamedExports` alongside a
// `restrictDefaultExports` flag, and upstream's schema forbids exactly that pairing, which is why
// the decoder here refuses it too and why upstream's corpus has no such case.
//
// It is still reachable, because a caller can build the options struct directly, and without the
// return that configuration reports TWICE on one export. Found by a mutation that deleted the return
// and survived every one of the 183 corpus rows.
func TestNoRestrictedExportsNamedRestrictionWinsOverTheDefaultFlags(t *testing.T) {
	t.Parallel()

	// Deliberately not through the decoder, which refuses this pairing. The point is the rule's own
	// arm order, not the configuration surface.
	options := NoRestrictedExportsOptions{
		RestrictedNamedExports: []string{"default"},
		RestrictDefaultExports: &NoRestrictedExportsRestrictDefaultExports{
			Named:         noRestrictedExportsFlag(true),
			DefaultFrom:   noRestrictedExportsFlag(true),
			NamedFrom:     noRestrictedExportsFlag(true),
			NamespaceFrom: noRestrictedExportsFlag(true),
		},
	}
	for _, source := range []string{
		"var a; export { a as default };",
		"export { default } from 'mod';",
		"export { foo as default } from 'mod';",
		"export * as default from 'mod';",
	} {
		result := rule_testing.RunWithOptions(t, NoRestrictedExports, noRestrictedExportsFile,
			source, options)
		rule_testing.ExpectFindings(t, result, "restrictedNamed")
	}

	// The control: with the name NOT restricted, the same sources reach the default arms and report
	// under the other message. Without this the assertions above would also pass for a rule that
	// had simply stopped reading the flags at all.
	flagsOnly := NoRestrictedExportsOptions{RestrictDefaultExports: options.RestrictDefaultExports}
	for _, source := range []string{
		"var a; export { a as default };",
		"export { default } from 'mod';",
		"export { foo as default } from 'mod';",
		"export * as default from 'mod';",
	} {
		result := rule_testing.RunWithOptions(t, NoRestrictedExports, noRestrictedExportsFile,
			source, flagsOnly)
		rule_testing.ExpectFindings(t, result, "restrictedDefault")
	}
}

// TestNoRestrictedExportsIgnoresExportEquals covers a construct upstream's corpus cannot contain.
//
// `export = foo` is TypeScript's own module-export syntax and shares `KindExportAssignment` with
// `export default foo`, separated only by the `IsExportEquals` field. Upstream parses JavaScript for
// this corpus and has no such syntax, so no imported fixture can distinguish a rule that reads that
// field from one that ignores it, and a mutation deleting the guard survived all 183 rows.
//
// Treating `export =` as a default export would be wrong on its own terms as well as unfaithful:
// it publishes the module's whole value under a name the importer supplies via `import x = require`,
// which is not the `export default` construct `restrictDefaultExports.direct` names.
func TestNoRestrictedExportsIgnoresExportEquals(t *testing.T) {
	t.Parallel()

	options := NoRestrictedExportsOptions{
		RestrictDefaultExports: &NoRestrictedExportsRestrictDefaultExports{
			Direct: noRestrictedExportsFlag(true),
		},
	}

	silent := rule_testing.RunWithOptions(t, NoRestrictedExports, noRestrictedExportsFile,
		"const foo = 1;\nexport = foo;", options)
	rule_testing.ExpectClean(t, silent)

	// The control. The same option on the same file with `export default` instead reports, so the
	// silence above is the guard rather than a rule that never fires under this configuration.
	reporting := rule_testing.RunWithOptions(t, NoRestrictedExports, noRestrictedExportsFile,
		"const foo = 1;\nexport default foo;", options)
	rule_testing.ExpectFindings(t, reporting, "restrictedDefault")
}
