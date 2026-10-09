package core

import (
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// unusedPrivateClassMembersFile is where the fixtures pretend to live.
//
// A `.ts` name on purpose. Ten of upstream's clean cases are TypeScript, and they carry the shapes
// most likely to break a naive port: `this.#db!`, `this.#foo as any`, `readonly #select`, and a
// private method on a generic class. Under a `.js` name those ten stop parsing and pass by
// accident, which is the worst way for a fixture to be green.
const unusedPrivateClassMembersFile = "/repository/source/PrivateMembers.ts"

// The corpus is oxc's, copied rather than rewritten.
//
// Every case below is verbatim from
// `oxc/crates/oxc_linter/src/rules/eslint/no_unused_private_class_members.rs`. The extractor reports
// one Tester block holding 63 pass and 24 fail inputs, against 27 snapshot diagnostics, so three of
// the fail inputs report more than once and a fixture asserting one finding per input would be
// wrong. Recovered by name from the snapshot rather than by an in-order walk: the extra three are
// the two-member class, the `#x`/`#y` pair, and the `#b`/`#c` pair.
//
// Copied because a fixture a porter invents encodes the same belief as the port, so it passes for
// exactly the reason the code is wrong. Upstream's clean cases are the ones that catch you.
//
// # Four of the 24 fail cases are asserted clean here instead
//
// oxc reports four discarded-conditional cases that ESLint declares clean. Measured against
// `eslint@9` by running the whole corpus through it rather than reasoning about it: those four are
// the only semantic disagreement between the two implementations, and the other ten differences are
// TypeScript syntax ESLint's own parser rejects. This follows ESLint, and the four live in
// `TestNoUnusedPrivateClassMembersDivergesFromOxcOnDiscardedConditionals` with the reasoning.
func TestNoUnusedPrivateClassMembersFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		findings   int
	}{
		{"upstream fail 1", `class Foo {
                #unusedMember = 5;
            }`, 1},
		{"upstream fail 2", `class First {}
            class Second {
                #unusedMemberInSecondClass = 5;
            }`, 1},
		{"upstream fail 3", `class First {
                #unusedMemberInFirstClass = 5;
            }
            class Second {}`, 1},
		{"upstream fail 4", `class First {
                #firstUnusedMemberInSameClass = 5;
                #secondUnusedMemberInSameClass = 5;
            }`, 2},
		{"upstream fail 5", `class Foo {
                #usedOnlyInWrite = 5;
                method() {
                    this.#usedOnlyInWrite = 42;
                }
            }`, 1},
		{"upstream fail 6", `class Foo {
                #usedOnlyInWriteStatement = 5;
                method() {
                    this.#usedOnlyInWriteStatement += 42;
                }
            }`, 1},
		{"upstream fail 7", `class C {
                #usedOnlyInIncrement;

                foo() {
                    this.#usedOnlyInIncrement++;
                }
            }`, 1},
		{"upstream fail 8", `class C {
                #unusedInOuterClass;

                foo() {
                    return class {
                        #unusedInOuterClass;

                        bar() {
                            return this.#unusedInOuterClass;
                        }
                    };
                }
            }`, 1},
		{"upstream fail 9", `class C {
                #unusedOnlyInSecondNestedClass;

                foo() {
                    return class {
                        #unusedOnlyInSecondNestedClass;

                        bar() {
                            return this.#unusedOnlyInSecondNestedClass;
                        }
                    };
                }

                baz() {
                    return this.#unusedOnlyInSecondNestedClass;
                }

                bar() {
                    return class {
                        #unusedOnlyInSecondNestedClass;
                    }
                }
            }`, 1},
		{"upstream fail 10", `class Foo {
                #unusedMethod() {}
            }`, 1},
		{"upstream fail 11", `class Foo {
                #unusedMethod() {}
                #usedMethod() {
                    return 42;
                }
                publicMethod() {
                    return this.#usedMethod();
                }
            }`, 1},
		{"upstream fail 12", `class Foo {
                set #unusedSetter(value) {}
            }`, 1},
		{"upstream fail 13", `class Foo {
                #unusedForInLoop;
                method() {
                    for (this.#unusedForInLoop in bar) {

                    }
                }
            }`, 1},
		{"upstream fail 14", `class Foo {
                #unusedForOfLoop;
                method() {
                    for (this.#unusedForOfLoop of bar) {

                    }
                }
            }`, 1},
		{"upstream fail 15", `class Foo {
                #unusedInDestructuring;
                method() {
                    ({ x: this.#unusedInDestructuring } = bar);
                }
            }`, 1},
		{"upstream fail 16", `class Foo {
                #unusedInRestPattern;
                method() {
                    [...this.#unusedInRestPattern] = bar;
                }
            }`, 1},
		{"upstream fail 17", `class Foo {
                #unusedInAssignmentPattern;
                method() {
                    [this.#unusedInAssignmentPattern = 1] = bar;
                }
            }`, 1},
		{"upstream fail 18", `class Foo {
                #unusedInAssignmentPattern;
                method() {
                    [this.#unusedInAssignmentPattern] = bar;
                }
            }`, 1},
		{"upstream fail 19", `class C {
                #usedOnlyInTheSecondInnerClass;

                method(a) {
                    return class {
                        #usedOnlyInTheSecondInnerClass;

                        method2(b) {
                            foo = b.#usedOnlyInTheSecondInnerClass;
                        }

                        method3(b) {
                            foo = b.#usedOnlyInTheSecondInnerClass;
                        }
                    }
                }
            }`, 1},
		{"upstream fail 20", `class StatementLogicalAssignment { #prop; method() { this.#prop ??= 1; } }`, 1},
		// Added from reading our code, not from upstream. Every one is a use form the corpus never
		// spells out on its own, and each was written before the rule to check the rule rather than
		// after it to explain a pass.

		// A private method assigned and never called. Measured against `eslint@9`, which reports
		// it: the write-only rule applies to a method the same way it applies to a field, and
		// nothing loads the value back. oxc disagrees here, gating its read test on
		// `element.kind.is_property()` so that any mention of a method marks it used, which means
		// oxc misses this one. Reproduced ESLint's judgement rather than oxc's because a method
		// nobody ever calls is exactly what this rule exists to find.
		{"a private method assigned but never called", `class A { #m() {} n() { this.#m = 1; } }`, 1},

		// A getter and a setter of one name with neither half referenced reports once, not twice.
		// The count is the whole assertion: a rule keyed by declaration node rather than by name
		// gives two findings for one member and no upstream case pins it, because upstream's only
		// accessor-pair cases are all used.
		{"an accessor pair with neither half referenced", `class A { get #a() { return 1; } set #a(v) {} }`, 1},

		// A brand check names a member that exists, so the member is used. Inverted here: the
		// member is never brand-checked and never read, so the finding stands. This is the control
		// for the brand-check clean case, which without it could pass by the rule being blind to
		// `in` entirely rather than by reading it correctly.
		{"a member never brand checked and never read", `class A { #brand; static has(o) { return #other in o; } #other; }`, 1},

		// Static private members are the same question with a different modifier, and no upstream
		// fail case carries `static`. A rule keying on instance shape alone goes silent here.
		{"an unused static private field", `class A { static #s = 1; }`, 1},
		{"an unused static private method", `class A { static #m() {} }`, 1},

		// Written through another instance rather than through `this`. Private names are lexically
		// scoped, so `o.#foo` inside the class reaches the same member, and a write through it is
		// as dead as a write through `this`.
		{"a write through another instance", `class A { #foo; m(o: A) { o.#foo = 1; } }`, 1},

		// A member used only inside a nested class that redeclares the name. The inner `#x`
		// satisfies the inner class; the outer one is still dead. Upstream carries this shape, but
		// only in classes large enough that a rule could pass it by accident.
		{"a name shadowed by a nested class", `class A { #x; m() { return class { #x; n() { return this.#x; } }; } }`, 1},

		// A TypeScript-only declaration form. `declare #x` and `accessor #x` both parse as ordinary
		// property declarations, which the probe measured rather than the port assumed; if either
		// changed the node kind the rule would skip the member and report nothing.
		{"an unused declare field", `class A { declare #x: number; }`, 1},
		{"an unused accessor field", `class A { accessor #x = 1; }`, 1},

		// A compound assignment whose result is discarded inside a nested statement rather than at
		// the top of a method. The statement test has to look at the immediate parent, and a rule
		// that instead asked whether the method body's last statement was an expression would pass
		// the simple upstream case and fail this.
		{"a compound assignment discarded inside a block", `class A { #x = 0; m() { if (cond) { this.#x += 1; } } }`, 1},

		// Both update fixities as discarded statements. Upstream carries only the postfix form, so
		// a rule handling `x++` and forgetting `++x` passes the whole corpus. This is the shape
		// `no-ex-assign` shipped missing entirely.
		{"a prefix update discarded as a statement", `class A { #x = 0; m() { ++this.#x; } }`, 1},
		{"a prefix decrement discarded as a statement", `class A { #x = 0; m() { --this.#x; } }`, 1},
		{"a postfix decrement discarded as a statement", `class A { #x = 0; m() { this.#x--; } }`, 1},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			expected := make([]string, testCase.findings)
			for index := range expected {
				expected[index] = "noUnusedPrivateClassMember"
			}
			rule_testing.ExpectFindings(t,
				rule_testing.Run(t, NoUnusedPrivateClassMembers, unusedPrivateClassMembersFile,
					testCase.sourceText), expected...)
		})
	}
}

func TestNoUnusedPrivateClassMembersStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"upstream pass 1", `
            class Foo { #privateMember = {}; a() { return { ...this.#privateMember }; } }
        `},
		{"upstream pass 2", `
            class Test {
                #prop = undefined

                getProp() {
                    return this.#prop ??= 0
                }
            }
        `},
		{"upstream pass 3", `
            class Test {
                #prop = undefined

                getProp() {
                    return this.#prop ||= 0
                }
            }
        `},
		{"upstream pass 4", `
            class Test {
                #prop = undefined

                getProp() {
                    return this.#prop += 0
                }
            }
        `},
		{"upstream pass 5", `class Foo {}`},
		{"upstream pass 6", `class Foo {
                publicMember = 42;
            }`},
		{"upstream pass 7", `class Foo {
                #usedMember = 42;
                method() {
                    return this.#usedMember;
                }
            }`},
		{"upstream pass 8", `class Foo {
                #usedMember = 42;
                anotherMember = this.#usedMember;
            }`},
		{"upstream pass 9", `class Foo {
                #usedMember = 42;
                foo() {
                    anotherMember = this.#usedMember;
                }
            }`},
		{"upstream pass 10", `class C {
                #usedMember;

                foo() {
                    bar(this.#usedMember += 1);
                }
            }`},
		{"upstream pass 11", `class Foo {
                #usedMember = 42;
                method() {
                    return someGlobalMethod(this.#usedMember);
                }
            }`},
		{"upstream pass 12", `class C {
                #usedInOuterClass;

                foo() {
                    return class {};
                }

                bar() {
                    return this.#usedInOuterClass;
                }
            }`},
		{"upstream pass 13", `class Foo {
                #usedInForInLoop;
                method() {
                    for (const bar in this.#usedInForInLoop) {

                    }
                }
            }`},
		{"upstream pass 14", `class Foo {
                #usedInForOfLoop;
                method() {
                    for (const bar of this.#usedInForOfLoop) {

                    }
                }
            }`},
		{"upstream pass 15", `class Foo {
                #usedInAssignmentPattern;
                method() {
                    [bar = 1] = this.#usedInAssignmentPattern;
                }
            }`},
		{"upstream pass 16", `class Foo {
                #usedInArrayPattern;
                method() {
                    [bar] = this.#usedInArrayPattern;
                }
            }`},
		{"upstream pass 17", `class Foo {
                #usedInAssignmentPattern;
                method() {
                    [bar] = this.#usedInAssignmentPattern;
                }
            }`},
		{"upstream pass 18", `class C {
                #usedInObjectAssignment;

                method() {
                    ({ [this.#usedInObjectAssignment]: a } = foo);
                }
            }`},
		{"upstream pass 19", `class C {
            set #accessorWithSetterFirst(value) {
                doSomething(value);
            }
            get #accessorWithSetterFirst() {
                return something();
            }
            method() {
                this.#accessorWithSetterFirst += 1;
            }
        }`},
		{"upstream pass 20", `class Foo {
            set #accessorUsedInMemberAccess(value) {}

            method(a) {
                [this.#accessorUsedInMemberAccess] = a;
            }
        }`},
		{"upstream pass 21", `class C {
            get #accessorWithGetterFirst() {
                return something();
            }
            set #accessorWithGetterFirst(value) {
                doSomething(value);
            }
            method() {
                this.#accessorWithGetterFirst += 1;
            }
        }`},
		{"upstream pass 22", `class Foo {
                #usedMethod() {
                    return 42;
                }
                anotherMethod() {
                    return this.#usedMethod();
                }
            }`},
		{"upstream pass 23", `class C {
            set #x(value) {
                doSomething(value);
            }

            foo() {
                this.#x = 1;
            }
        }`},
		{"upstream pass 24", `type Callback<T> = () => Promise<T> | T;

         export class Issue_11039<T> {
            load: () => Promise<T>;

            constructor(callback: Callback<T>) {
                this.load = () => this.#load(callback);
            }

            async #load(callback: Callback<T>) {
                callback;
            }
         }`},
		{"upstream pass 25", `class ChildProcess extends EventEmitter { #stdioObject; #createStdioObject() {} get stdio() { return (this.#stdioObject ??= this.#createStdioObject()); } }`},
		{"upstream pass 26", "export class Foo { readonly #select = 123; override render() { return html`foo=${this.#select}`; } }"},
		{"upstream pass 27", `export class Foo { #listened = false; bar() { if (!this.#listened) return; this.#listened = false; } } `},
		{"upstream pass 28", `export class RichText { #verticalScrollContainer; init() { const verticalScrollContainer = this.#verticalScrollContainer || (this.#verticalScrollContainer = this.verticalScrollContainerGetter?.() || null); } }`},
		{"upstream pass 29", `class Foo { #a = false; on(data) { return this.#a ? [data] : data; } set setA(value) { this.#a = value; } }`},
		{"upstream pass 30", `class Foo { #a = false; on(data) { return this.#a ? [data] : data; } }`},
		{"upstream pass 31", `class WeakReference { #i = 0; inc() { return ++this.#i; }; dec() { return --this.#i; } }`},
		{"upstream pass 32", `class Foo { #d; constructor(d) { this.#d = d || kDefaultD; } get getD(): string { return this.#d!; } }`},
		{"upstream pass 33", `class F { #o; initialize(output) { this.#o = output; } text(e) { return this.#o!.text(e); } }`},
		{"upstream pass 34", `class Foo { #a; constructor(a) { this.#a = a; }; b(b?: string): this { this.#a!.setB(b); return this; } resetA() { this.#a = undefined; } }`},
		{"upstream pass 35", `let getPrivate; class C { #private; constructor(v) { this.#private = v; } static { getPrivate = klass => klass.#private; } }`},
		{"upstream pass 36", `let getPrivate; class C { #private; constructor(v) { this.#private = v; } static { getPrivate = klass => { return klass.#private; } } }`},
		{"upstream pass 37", `class C { #field = 1; static { const obj = new C(); console.log(obj.#field); } }`},
		{"upstream pass 38", `class C { #method() { return 42; } static { const obj = new C(); obj.#method(); } }`},
		{"upstream pass 39", `class C { #field = 1; static { const getField = obj => { return obj.#field; }; } }`},
		{"upstream pass 40", `export class Database<const S extends idb.DBSchema> { readonly #db: Promise<idb.IDBPDatabase<S>>; constructor(name: string, version: number, hooks: idb.OpenDBCallbacks<S>) { this.#db = idb.openDB<S>(name, version, hooks); }  async read() { let db = await this.#db; } }`},
		{"upstream pass 41", `export class A { #x; constructor(x: number) { this.#x = x; } get(y = this.#x) { return y; } }`},
		{"upstream pass 42", `class B { #value = 42; method(param = this.#value) { return param * 2; } }`},
		{"upstream pass 43", `class C { #arr = [1, 2, 3]; process(items = this.#arr) { return items.map(x => x * 2); } }`},
		{"upstream pass 44", `export class BugClass { readonly #BUG: readonly [] = []; method() { return Math.random() > 0.5 ? this.#BUG : []; } }`},
		{"upstream pass 45", `class Foo { #x; #y; method(a, b, c) { return a ? (b ? this.#x : c) : this.#y; } }`},
		{"upstream pass 46", `class Foo { #x; method() { return () => a ? this.#x : b; } }`},
		{"upstream pass 47", `class Foo { #x; method() { return a && (b ? this.#x : c); } }`},
		{"upstream pass 48", `class Foo { #x; method() { fn(a ? this.#x : b); } }`},
		{"upstream pass 49", "class Foo { #x; method() { return `${a ? this.#x : b}`; } }"},
		{"upstream pass 50", `class Foo { #x; method() { return [a ? this.#x : b]; } }`},
		{"upstream pass 51", `class Foo { #x; method() { return { key: a ? this.#x : b }; } }`},
		{"upstream pass 52", `class Foo { #x; method(val) { switch(val) { case (a ? this.#x : b): break; } } }`},
		{"upstream pass 53", `class Foo { #x; method() { throw a ? this.#x : new Error(); } }`},
		{"upstream pass 54", `class Foo { #x; method() { while (a ? this.#x : b) {} } }`},
		{"upstream pass 55", `class Bug { #flag = false; foo() { this.#flag && console.log('spam'); } }`},
		{"upstream pass 56", `class Foo { #a; #b; #c; method() { return this.#a ? this.#b : this.#c; } }`},
		{"upstream pass 57", `class ExampleFoo { #foo = 0; foo(foo) { foo = foo ?? this.#foo; return foo; } }`},
		{"upstream pass 58", `class ExampleBar { #bar = 0; bar(bar) { bar = ++this.#bar; return bar; } }`},
		{"upstream pass 59", `class Foo { #awaitedMember; async method() { await this.#awaitedMember; } }`},
		{"upstream pass 60", `class Test { #url: string; constructor(url: string) { this.#url = url; } open() { return new WebSocket(this.#url); } }`},
		{"upstream pass 61", `export class Foo { #fetch: typeof fetch; constructor() { this.#fetch = fetch; } async bar() { return (0, this.#fetch)('https://example.com'); } }`},
		{"upstream pass 62", `export class StateMachine { #state = 'idle'; step() { switch (this.#state) { case 'idle': { this.#state = 'running'; break; } case 'running': { this.#state = 'done'; break; } } } }`},
		{"upstream pass 63", `class Stopper { #promise; async stop() { this.#promise ??= this.makePromise(); await this.#promise; } makePromise() { return Promise.resolve(); } }`},
		// Added from reading our code. The use forms upstream never spells out, each one a way for
		// the rule to under-report and accuse working code of being dead.

		// The brand check. The only reference to a private name that is not a property access, so
		// a rule inspecting only `PropertyAccessExpression` nodes cannot see it and calls a live
		// member dead. Upstream has no case for it on either side.
		{"a brand check on a field", `class A { #brand; static has(o: unknown) { return #brand in o; } }`},
		{"a brand check on a method", `class A { #m() {} static has(o: unknown) { return #m in o; } }`},

		// A use inside a static block. The reference is not in any method body, so a rule that
		// walked only methods would miss it. Upstream carries five of these, which is why they are
		// listed as a group here rather than trusted to one case.
		{"a read inside a static block", `class A { #f = 1; static { const o = new A(); console.log(o.#f); } }`},
		{"a call inside a static block", `class A { #m() { return 1; } static { const o = new A(); o.#m(); } }`},

		// A use in a field initializer, which runs before any method exists.
		{"a read in a field initializer", `class A { #f = 1; g = this.#f; }`},

		// A use inside a nested arrow capturing `this`. The reference is lexically inside the class
		// but syntactically inside a function, and it is still the same member.
		{"a read inside a nested arrow", `class A { #f = 1; m() { return () => this.#f; } }`},
		{"a read inside a nested function expression", `class A { #f = 1; m(o: A) { return function () { return o.#f; }; } }`},

		// Both update fixities in value position. The mirror of the discarded pair in the Fires
		// suite: `return ++this.#x` reads the member, and a rule keying on the operator rather than
		// on whether the result is discarded reports all four.
		{"a prefix update whose value is returned", `class A { #x = 0; m() { return ++this.#x; } }`},
		{"a postfix update whose value is returned", `class A { #x = 0; m() { return this.#x++; } }`},
		{"a prefix update passed as an argument", `class A { #x = 0; m() { f(++this.#x); } }`},

		// A setter referenced only by a write, in every write form. This is the accessor rule, and
		// getting it backwards reports the ordinary way to write a private setter. The rest-pattern
		// and for-in forms are here because they are the write shapes most likely to be handled by
		// a separate arm that forgets to ask whether the member is an accessor first.
		{"a setter written through a rest pattern", `class A { set #a(v: unknown[]) {} m(b: unknown[]) { [...this.#a] = b; } }`},
		{"a setter written as a for-in target", `class A { set #a(v: string) {} m(b: object) { for (this.#a in b) {} } }`},
		{"a getter written to", `class A { get #a() { return 1; } m() { this.#a = 1; } }`},

		// A private name declared twice with mixed kinds. This is a syntax error in JavaScript, and
		// ESLint's parser rejects the file outright, so neither upstream defines behavior for it.
		// The TypeScript parser is more tolerant and hands it to us anyway, which makes it a real
		// input rather than a hypothetical: written for a surviving mutant that replaced the
		// accessor flag on the shared entry instead of accumulating it, turning `this.#a = 1` into a
		// write-only finding on code the language already refuses. Staying silent is the
		// conservative answer, since a finding here would be a second complaint about a file that
		// cannot run at all.
		{"a private name declared as both a getter and a field", `class A { get #a() { return 1; } #a = 1; m() { this.#a = 1; } }`},

		// A read through another instance rather than through `this`.
		{"a read through another instance", `class A { #foo = 1; m(o: A) { return o.#foo; } }`},

		// A static private member read from a static method.
		{"a static private field read statically", `class A { static #s = 1; static m() { return A.#s; } }`},

		// A spread of the member's value, which builds a new object and is a read. The shape is one
		// character away from `[...this.#x] = bar`, which is a write, and the only thing telling
		// them apart is whether the climb reaches an assignment operator.
		{"a spread of the member into a call", `class A { #x: unknown[] = []; m() { return f(...this.#x); } }`},
		{"a spread of the member into a new object", `class A { #x = {}; m() { return { ...this.#x }; } }`},
		{"an array literal holding the member", `class A { #x = 1; m() { return [this.#x]; } }`},
		{"an object literal holding the member", `class A { #x = 1; m() { return { k: this.#x }; } }`},

		// A destructuring-shaped literal on the *right* of an assignment. Written for a surviving
		// mutant: dropping the left-side identity test from the climb made `foo = [this.#x]` read
		// as a write, because the climb reaches an assignment operator either way and only the side
		// tells the two apart. Nothing upstream builds a value this way from a private member.
		{"an array literal on the right of an assignment", `class A { #x = 1; m(foo: unknown) { foo = [this.#x]; } }`},
		{"an object literal on the right of an assignment", `class A { #x = 1; m(foo: unknown) { foo = { k: this.#x }; } }`},
		{"a spread on the right of an assignment", `class A { #x: unknown[] = []; m(foo: unknown) { foo = [...this.#x]; } }`},
		{"a nested literal on the right of an assignment", `class A { #x = 1; m(foo: unknown) { foo = [{ k: this.#x }]; } }`},

		// A destructuring-shaped literal under a *non-assignment* binary operator. Written for a
		// second surviving mutant: dropping the operator test from the climb made `[this.#x] === bar`
		// read as a write, since the literal genuinely is the left operand and only the operator
		// says it is a comparison rather than a store.
		{"an array literal compared with a non-assignment operator", `class A { #x = 1; m(bar: unknown) { return [this.#x] === bar; } }`},
		{"an object literal compared with a non-assignment operator", `class A { #x = 1; m(bar: unknown) { return { k: this.#x } !== bar; } }`},

		// The right side of an assignment whose left side is the same member. A rule that saw the
		// member on the left and stopped would call this write-only.
		{"a self-referential assignment", `class A { #x = 1; m() { this.#x = this.#x + 1; } }`},

		// A nested destructuring target that reaches its assignment operator through several
		// levels. The climb has to pass through both literal kinds and the property assignment
		// between them, and this is a read on the *key* side, which the climb must not reach.
		{"a computed key inside a nested destructuring target", `class A { #k = "a"; m(o: never) { ({ x: { [this.#k]: y } } = o); } }`},

		// A member declared in an outer class and read from inside a nested class that does *not*
		// redeclare it. The shadowing narrowing must remove only the names the inner class
		// redeclares, not stop the walk outright.
		{"a read from a nested class that shadows a different name", `class A { #outer = 1; m() { return class { #inner = 2; n() { return this.#inner; } }; } o() { return this.#outer; } }`},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.Run(t, NoUnusedPrivateClassMembers,
				unusedPrivateClassMembersFile, testCase.sourceText))
		})
	}
}

// TestNoUnusedPrivateClassMembersDivergesFromOxcOnDiscardedConditionals pins the one place this
// rule deliberately disagrees with the reference it was ported from.
//
// All four are oxc fail cases, verbatim, asserted clean here. oxc tracks whether a conditional's
// result reaches a value context and calls the read dead when it does not; ESLint counts the read.
// Measured by running the whole corpus through `eslint@9` rather than reasoning about it: these four
// are the only semantic disagreement in 87 cases.
//
// Following ESLint, because every finding this rule produces is an accusation that code is dead.
// Reading `this.#x` in a discarded ternary still runs a getter if `#x` is one, so calling it dead is
// a claim about intent rather than about reachability, and the quieter reading is the right default
// when the two upstreams disagree on an unused-thing rule.
//
// If this is ever revisited, the decision to revisit is which of the two upstreams to follow, not
// whether the rule has a bug.
func TestNoUnusedPrivateClassMembersDivergesFromOxcOnDiscardedConditionals(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"oxc fail 21", `class Foo { #unused; method() { Math.random() > 0.5 ? this.#unused : []; } }`},
		{"oxc fail 22", `class Foo { #x; #y; method(a, b, c) { a ? (b ? this.#x : c) : this.#y; } }`},
		{"oxc fail 23", `class Foo { #x; method() { a && (b ? this.#x : c); } }`},
		{"oxc fail 24", `class Foo { #a; #b; #c; method() { this.#a ? this.#b : this.#c; } }`}}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.Run(t, NoUnusedPrivateClassMembers,
				unusedPrivateClassMembersFile, testCase.sourceText))
		})
	}
}

// TestNoUnusedPrivateClassMembersPointsAtThePrivateName asserts where every finding lands.
//
// `ExpectFindings` asserts message ids and count and nothing else, so a rule reporting the whole
// member declaration, or the class, or an off-by-one slice of the name would pass the two suites
// above completely green while pointing somewhere useless. A clone shipped exactly that defect on a
// rule with no fix at all, which is why this is asserted whether or not a repair exists.
//
// The expected span is the `#name` token including the hash, which is also what oxc underlines: its
// snapshot marks 13 columns under `#unusedMember` for a 12-character name, and ESLint reports
// `declaredNode.key.loc`, the same node. Both upstreams agree, so this is measured against them
// rather than chosen.
func TestNoUnusedPrivateClassMembersPointsAtThePrivateName(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		reported   []string
	}{
		{"a field", `class A { #unused = 1; }`, []string{"#unused"}},
		{"a method", `class A { #unusedMethod() {} }`, []string{"#unusedMethod"}},
		{"a setter", `class A { set #unusedSetter(v: number) {} }`, []string{"#unusedSetter"}},
		{"an accessor pair", `class A { get #pair() { return 1; } set #pair(v: number) {} }`, []string{"#pair"}},
		{"a static field", `class A { static #unusedStatic = 1; }`, []string{"#unusedStatic"}},
		// Two findings in one class, in declaration order. A rule reporting in map order passes the
		// count assertion and fails this one.
		{"two members in one class", `class A { #first = 1; #second = 2; }`, []string{"#first", "#second"}},
		// The finding must land on the dead member, not on the live one beside it.
		{"the dead member beside a live one", `class A { #live = 1; #dead = 2; m() { return this.#live; } }`, []string{"#dead"}},
		// A member whose name is a prefix of another. A span computed from a name search rather
		// than from the node lands on the wrong one.
		{"a name that is a prefix of a live name", `class A { #x = 1; #xy = 2; m() { return this.#xy; } }`, []string{"#x"}},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoUnusedPrivateClassMembers,
				unusedPrivateClassMembersFile, testCase.sourceText)
			if len(result.Diagnostics) != len(testCase.reported) {
				t.Fatalf("got %d findings, want %d", len(result.Diagnostics), len(testCase.reported))
			}
			for index, diagnostic := range result.Diagnostics {
				got := testCase.sourceText[diagnostic.Range.Pos():diagnostic.Range.End()]
				if got != testCase.reported[index] {
					t.Errorf("finding %d points at %q, want %q", index, got, testCase.reported[index])
				}
			}
		})
	}
}

// TestNoUnusedPrivateClassMembersProposesNoRepair asserts the rule offers nothing to apply.
//
// Upstream ESLint offers removal as a *suggestion* and oxc ships no repair at all. Deleting a member
// changes the shape of the class: a comment above it may describe it, a `declare` may be pinning a
// type another file depends on, and the author's answer is at least as often "add the reader I
// forgot" as "delete this". So neither a fix nor a suggestion is proposed here.
//
// Asserted rather than left implicit, because a later revision adding an unattended fix would be a
// rule silently deleting code while nobody is looking, and it would pass every other test in this
// file.
func TestNoUnusedPrivateClassMembersProposesNoRepair(t *testing.T) {
	t.Parallel()

	result := rule_testing.Run(t, NoUnusedPrivateClassMembers, unusedPrivateClassMembersFile,
		`class A { #unused = 1; }`)
	if len(result.Diagnostics) != 1 {
		t.Fatalf("got %d findings, want 1", len(result.Diagnostics))
	}
	if len(result.Diagnostics[0].Fixes) != 0 {
		t.Errorf("rule proposes %d fixes; removing a member changes the class shape and is a "+
			"choice a human makes", len(result.Diagnostics[0].Fixes))
	}
	if len(result.Diagnostics[0].Suggestions) != 0 {
		t.Errorf("rule proposes %d suggestions; see the comment above this test before adding one",
			len(result.Diagnostics[0].Suggestions))
	}
}

// TestNoUnusedPrivateClassMembersNamesTheMember asserts the message identifies which member is dead,
// spelled exactly as the source spells it.
//
// A class with several private members produces several identical findings otherwise, and the reader
// has only the span to tell them apart. Both upstreams interpolate the name for this reason.
//
// The assertion is on the exact rendered name rather than on `strings.Contains`, because the loose
// version shipped a defect. The name node's text already carries the leading hash, so a `'#%s'`
// format rendered `'##unusedField'`, and a containment check for `#unusedField` is satisfied by a
// string with two hashes. The dry run against the real tree caught it; this test did not. A test
// whose predicate is weaker than the property it is guarding is how that happens.
func TestNoUnusedPrivateClassMembersNamesTheMember(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		rendered   string
	}{
		{"a field", `class A { #forgotten = 1; }`, "'#forgotten' is declared"},
		{"a method", `class A { #forgottenMethod() {} }`, "'#forgottenMethod' is declared"},
		{"a static field", `class A { static #forgottenStatic = 1; }`, "'#forgottenStatic' is declared"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoUnusedPrivateClassMembers,
				unusedPrivateClassMembersFile, testCase.sourceText)
			if len(result.Diagnostics) != 1 {
				t.Fatalf("got %d findings, want 1", len(result.Diagnostics))
			}
			description := result.Diagnostics[0].Message.Description
			if !strings.HasPrefix(description, testCase.rendered) {
				t.Errorf("message begins %q, want it to begin %q",
					description[:min(len(description), len(testCase.rendered)+10)], testCase.rendered)
			}
		})
	}
}
