package parser

import (
	"path/filepath"
	"strings"
	"testing"
)

func wholeCases() []string {
	cases := []string{
		`export default await(x); export = await(x); export default interface I { x: number; } export default async function*() { yield await(x); }`,
		`declare module "x" with { type: "json" } { export interface I {} } declare module "y"; interface I<in T, out U> { f(x: T): U; }`,
		``,
		"function f() { return\nx\n}\nlet a = 1\nlet b = 2\nthrow a;",
		`class C { static; readonly; abstract; declare; override; async; accessor; get; set; get<T>() {} set<T>() {} }`,
		`@dec class C { @dec [key]() {} @(N[key]) p: number; @dec(N[key]) m() {} }`,
		`type T = any.X<string>; type T = this is I; type T = asserts x is X; type T = infer U extends V ? A : B;`,
		`var x; let y = 1; const z = 2; let {a: b = 1, ...c} = d; let [e,,...f] = g;`,
		`function f<T extends U = V>(x?: T): T; async function* g(x = 1) { yield x; await h(); return x; }`,
		`if (x) if (y) a(); else b(); else c(); while (x) { break; } do { continue; } while (x);`,
		`for (;;) ; for (let i = 0; i < 3; i++) f(i); for (const x of xs) f(x); for (var k in o) f(k); async function f() { for await (const x of xs) g(x); }`,
		`switch (x) { case 1: f(); break; default: g(); case 2: h(); } try { f(); } catch (e: any) { throw e; } finally { g(); } try {} catch {} outer: while (x) { continue outer; break outer; } debugger;`,
		`with (x) y();`,
		`export const x = 1; export function f() {}; export default function() {}; export default class {}; export abstract class A {}; declare function f(x: number): void;`,
		`class C<T> extends B<T> implements I<T>, J { ; static { f(); } public readonly x?: T; private y!: number; #z = 1; constructor(public p: T) {} get a(): T { return this.p; } set a(x: T) {} async *m<U>(x: U): Promise<U> { return x; } [name]() {} accessor value = 1; }`,
		`abstract class C { abstract m(): void; [key: string]: number; static [key: number]: string; }`,
		`interface I<T> extends A<T>, N.B { readonly x?: T; m<U>(x: U): T; (x: T): T; new(x: T): I<T>; [key: string]: T; get value(): T; set value(v: T); }`,
		`enum E { A, B = 1, C = f(), "D" = "d", } export const enum F { A = 1 }`,
		`namespace N.A.B { export const x = 1; } module M { export interface I {} } declare module "x" { export = M; } declare global { interface X {} }`,
		`import "x"; import X from "x"; import * as X from "x"; import { a, b as c, type D, type E as F, "x" as x } from "x"; import X, {y} from "x"; import X, * as Y from "x";`,
		`import type X from "x"; import type * as X from "x"; import type {X} from "x"; import defer * as X from "x"; import type from "x"; import type from from "x";`,
		`import X = require("x"); import X = N.Y; import type X = require("x"); export import Y = X; export import type Z = N.X;`,
		`export {}; export {x, y as z, type T, type U as V}; export {x} from "x"; export * from "x"; export * as N from "x"; export type {T} from "x"; export type * from "x"; export type * as N from "x"; export as namespace Library; export = x; export default x + y;`,
		`import {type, type as as, type as X, type as} from "x"; export {type, type as as, type as X, type as} from "x";`,
		`import X from "x" with {type: "json", "mode": "x",}; export * from "x" with {type: "json"};`,
		`@dec export class C { @dec p: number; @dec m(@dec p: number) {} } export @dec class D {} @dec(x) @N.dec class E { @dec accessor x = 1; }`,
		`using resource = acquire(); await using resource = acquire(); for (using r of resources) f(r);`,
		`type T = any; type T = unknown; type T = number; type T = bigint; type T = object; type T = boolean; type T = string; type T = symbol; type T = void; type T = undefined; type T = never; type T = intrinsic;`,
		`type A = N.T<string>; type B = (x: any) => x is string; type C = (x: any) => asserts x; type D = (x: any) => asserts x is string; type E = new<T>(x: T) => T; type F = abstract new() => I;`,
		`type T = typeof N.x<string>; type T = { readonly x?: string; m(): void }; type T = T[]; type T = [T, U?, ...V[]]; type T = [x: T, y?: U, ...z: V[]];`,
		`type T = A | B & C; type T = A extends infer U extends B ? U : never; type T = (A); type T = this; type T = keyof T; type T = readonly T[]; type T = unique symbol; type T = T[K];`,
		"type T = { -readonly [K in keyof T as `get${K & string}`]+?: T[K] }; type T = `a${string}b${number}`;",
		`type T = "x" | 1 | -1 | true | false | null | 1n | -1n; type T = typeof import("x", {with: {type: "json"}}).N.T<string>; type T = import("x");`,
		`type T = *; type T = ?T; type T = !T; type T = T!; type T = T?;`,
	}
	// Shift every position across UTF-8/UTF-16 boundaries and exercise CRLF trivia.
	original := append([]string(nil), cases...)
	for _, source := range original {
		cases = append(cases, "/* 💡 π */\r\n"+strings.ReplaceAll(source, "\n", "\r\n"))
	}
	return cases

}
// Not parallel: native.runtimeBuilds and the runtime archive cache at os.UserCacheDir()/adamic/runtime via native.Build.
func TestWholeGeneratedAgrees(t *testing.T) {
	cases := wholeCases()
	manifest := wholeManifest(t, cases)
	oracle := goOracle(t)
	want := execute(t, "", oracle, "--manifest", manifest, "--whole")
	directory, err := filepath.Abs(".")
	if err != nil {
		t.Fatal(err)
	}
	got := wholeNode(t, directory, manifest, false)
	if diff := difference(got.output, want.output); diff != "" {
		t.Fatalf("Node: %s", diff)
	}
	got = execute(t, "", buildPort(t, directory, true), "--manifest", manifest, "--whole")
	if diff := difference(got.output, want.output); diff != "" {
		t.Fatalf("native: %s", diff)
	}
	t.Logf("%d generated whole files, %d identical bytes", len(cases), len(want.output))
}

// Not parallel: native.runtimeBuilds and the runtime archive cache at os.UserCacheDir()/adamic/runtime via native.Build.
func TestObsoleteImportAttributesAgrees(t *testing.T) {
	manifest := wholeManifest(t, []string{`import X from "x" assert {type: "json"}; export * from "x" assert {type: "json"}; type T = import("x", {assert: {type: "json"}});`})
	oracle := goOracle(t)
	want := execute(t, "", oracle, "--manifest", manifest, "--whole", "--allow-obsolete-assert")
	directory, err := filepath.Abs(".")
	if err != nil {
		t.Fatal(err)
	}
	for _, side := range []struct {
		name string
		got  execution
	}{{"Node", wholeNode(t, directory, manifest, false)}, {"native", execute(t, "", buildPort(t, directory, true), "--manifest", manifest, "--whole")}} {
		if diff := difference(side.got.output, want.output); diff != "" {
			t.Fatalf("%s: %s", side.name, diff)
		}
	}
	t.Logf("obsolete assert forms: Go required to emit only diagnostic 2880; %d identical tree bytes", len(want.output))
}
