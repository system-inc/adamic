package nexus

import (
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// The require-atomic-updates corpus, replayed through the rule that replaces it.
//
// Every case is taken verbatim: the 63 of upstream's own corpus as `core/require_atomic_updates_corpus_test.go`
// imports it, then the 21 cases cohere's port of that rule added while settling its divergences from
// ESLint (`TestRequireAtomicUpdatesEscapeTable`, `...ResolutionDivergence`, `...ProcessCaseWithDeclaration`,
// `...FinallyReportsOnce`, `...CatchDoesNotInheritTryRefresh` and `...RestoreInFinallyIsJudged`).
//
// `atomic` is what the old rule reports: upstream's own expectation for the corpus, and cohere's port
// measured on 2026-10-01 for the added cases. `lostUpdate` is this rule. Where they differ the reason is
// stated, and every difference runs one way: the old rule reports and this one is silent because the
// write is not computed from the value it overwrites. The cases where both report are the corpus's
// real read-modify-writes, so the replay pins both directions at once.
//
// Two upstream-clean cases (allowProperties) carry that option upstream; this rule takes none and is
// silent on them either way.
var concurrencyNoLostUpdateAtomicCorpus = []struct {
	source     string
	atomic     int
	lostUpdate int
	reason     string
}{
	{
		source:     `let foo; async function x() { foo += bar; }`,
		atomic:     0,
		lostUpdate: 0,
		reason:     "",
	},
	{
		source:     `let foo; async function x() { foo = foo + bar; }`,
		atomic:     0,
		lostUpdate: 0,
		reason:     "",
	},
	{
		source:     `let foo; async function x() { foo = await bar + foo; }`,
		atomic:     0,
		lostUpdate: 0,
		reason:     "",
	},
	{
		source:     `async function x() { let foo; foo += await bar; }`,
		atomic:     0,
		lostUpdate: 0,
		reason:     "",
	},
	{
		source:     `let foo; async function x() { foo = (await result)(foo); }`,
		atomic:     0,
		lostUpdate: 0,
		reason:     "",
	},
	{
		source:     `let foo; async function x() { foo = bar(await something, foo) }`,
		atomic:     0,
		lostUpdate: 0,
		reason:     "",
	},
	{
		source:     `function* x() { let foo; foo += yield bar; }`,
		atomic:     0,
		lostUpdate: 0,
		reason:     "",
	},
	{
		source:     `const foo = {}; async function x() { foo.bar = await baz; }`,
		atomic:     0,
		lostUpdate: 0,
		reason:     "",
	},
	{
		source:     `const foo = []; async function x() { foo[x] += 1;  }`,
		atomic:     0,
		lostUpdate: 0,
		reason:     "",
	},
	{
		source:     `let foo; function* x() { foo = bar + foo; }`,
		atomic:     0,
		lostUpdate: 0,
		reason:     "",
	},
	{
		source:     `async function x() { let foo; bar(() => baz += 1); foo += await amount; }`,
		atomic:     0,
		lostUpdate: 0,
		reason:     "",
	},
	{
		source:     `let foo; async function x() { foo = condition ? foo : await bar; }`,
		atomic:     0,
		lostUpdate: 0,
		reason:     "",
	},
	{
		source:     `async function x() { let foo; bar(() => { let foo; blah(foo); }); foo += await result; }`,
		atomic:     0,
		lostUpdate: 0,
		reason:     "",
	},
	{
		source:     `let foo; async function x() { foo = foo + 1; await bar; }`,
		atomic:     0,
		lostUpdate: 0,
		reason:     "",
	},
	{
		source:     `async function x() { foo += await bar; }`,
		atomic:     0,
		lostUpdate: 0,
		reason:     "",
	},
	{
		source: `
            async function foo() {
                if (1);
                if (2);
                if (3);
                if (4);
                if (5);
                if (6);
                if (7);
                if (8);
                if (9);
                if (10);
                if (11);
                if (12);
                if (13);
                if (14);
                if (15);
                if (16);
                if (17);
                if (18);
                if (19);
                if (20);
            }
        `,
		atomic:     0,
		lostUpdate: 0,
		reason:     "",
	},
	{
		source: `
            async function foo() {
                return [
                    1 ? a : b,
                    2 ? a : b,
                    3 ? a : b,
                    4 ? a : b,
                    5 ? a : b,
                    6 ? a : b,
                    7 ? a : b,
                    8 ? a : b,
                    9 ? a : b,
                    10 ? a : b,
                    11 ? a : b,
                    12 ? a : b,
                    13 ? a : b,
                    14 ? a : b,
                    15 ? a : b,
                    16 ? a : b,
                    17 ? a : b,
                    18 ? a : b,
                    19 ? a : b,
                    20 ? a : b
                ];
            }
        `,
		atomic:     0,
		lostUpdate: 0,
		reason:     "",
	},
	{
		source: `
            async function f() {
                let records
                records = await a.records
                g(() => { records })
            }
        `,
		atomic:     0,
		lostUpdate: 0,
		reason:     "",
	},
	{
		source: `
            async function f() {
                try {
                    this.foo = doSomething();
                } catch (e) {
                    this.foo = null;
                    await doElse();
                }
            }
        `,
		atomic:     0,
		lostUpdate: 0,
		reason:     "",
	},
	{
		source: `
            async function f(foo) {
                let bar = await get(foo.id);
                bar.prop = foo.prop;
            }
        `,
		atomic:     0,
		lostUpdate: 0,
		reason:     "",
	},
	{
		source: `
            async function f(foo) {
                let bar = await get(foo.id);
                foo = bar.prop;
            }
        `,
		atomic:     0,
		lostUpdate: 0,
		reason:     "",
	},
	{
		source: `
            async function f() {
                let foo = {}
                let bar = await get(foo.id);
                foo.prop = bar.prop;
            }
        `,
		atomic:     0,
		lostUpdate: 0,
		reason:     "",
	},
	{
		source: `
            let count = 0
            let queue = []
            async function A(...args) {
                count += 1
                await new Promise(resolve=>resolve())
                count -= 1
                return
            }
        `,
		atomic:     0,
		lostUpdate: 0,
		reason:     "",
	},
	{
		source: `
            async function foo(e) {
            }

            async function run() {
              const input = [];
              const props = [];

              for(const entry of input) {
                const prop = props.find(a => a.id === entry.id) || null;
                await foo(entry);
              }

              for(const entry of input) {
                const prop = props.find(a => a.id === entry.id) || null;
              }

              for(const entry2 of input) {
                const prop = props.find(a => a.id === entry2.id) || null;
              }
            }
        `,
		atomic:     0,
		lostUpdate: 0,
		reason:     "",
	},
	{
		source: `
            async function run() {
              {
                let entry;
                await entry;
              }
              {
                let entry;
                () => entry;

                entry = 1;
              }
            }
        `,
		atomic:     0,
		lostUpdate: 0,
		reason:     "",
	},
	{
		source: `
            async function run() {
                await a;
                b = 1;
            }
        `,
		atomic:     0,
		lostUpdate: 0,
		reason:     "",
	},
	{
		source: `
                async function a(foo) {
                    if (foo.bar) {
                        foo.bar = await something;
                    }
                }
            `,
		atomic:     0,
		lostUpdate: 0,
		reason:     "",
	},
	{
		source: `
                function* g(foo) {
                    baz = foo.bar;
                    yield something;
                    foo.bar = 1;
                }
            `,
		atomic:     0,
		lostUpdate: 0,
		reason:     "",
	},
	{
		source:     `let foo; async function x() { foo += await amount; }`,
		atomic:     1,
		lostUpdate: 1,
		reason:     "",
	},
	{
		source:     `if (1); let foo; async function x() { foo += await amount; }`,
		atomic:     1,
		lostUpdate: 1,
		reason:     "",
	},
	{
		source:     `let foo; async function x() { while (condition) { foo += await amount; } }`,
		atomic:     1,
		lostUpdate: 1,
		reason:     "",
	},
	{
		source:     `let foo; async function x() { foo = foo + await amount; }`,
		atomic:     1,
		lostUpdate: 1,
		reason:     "",
	},
	{
		source:     `let foo; async function x() { foo = foo + (bar ? baz : await amount); }`,
		atomic:     1,
		lostUpdate: 1,
		reason:     "",
	},
	{
		source:     `let foo; async function x() { foo = foo + (bar ? await amount : baz); }`,
		atomic:     1,
		lostUpdate: 1,
		reason:     "",
	},
	{
		source:     `let foo; async function x() { foo = condition ? foo + await amount : somethingElse; }`,
		atomic:     1,
		lostUpdate: 1,
		reason:     "",
	},
	{
		source:     `let foo; async function x() { foo = (condition ? foo : await bar) + await bar; }`,
		atomic:     1,
		lostUpdate: 1,
		reason:     "",
	},
	{
		source:     `let foo; async function x() { foo += bar + await amount; }`,
		atomic:     1,
		lostUpdate: 1,
		reason:     "",
	},
	{
		source:     `async function x() { let foo; bar(() => foo); foo += await amount; }`,
		atomic:     1,
		lostUpdate: 0,
		reason:     "a closure only reads foo, so nothing else can write it during the suspension",
	},
	{
		source:     `let foo; function* x() { foo += yield baz }`,
		atomic:     1,
		lostUpdate: 1,
		reason:     "",
	},
	{
		source:     `let foo; async function x() { foo = bar(foo, await something) }`,
		atomic:     1,
		lostUpdate: 1,
		reason:     "",
	},
	{
		source:     `const foo = {}; async function x() { foo.bar += await baz }`,
		atomic:     1,
		lostUpdate: 1,
		reason:     "",
	},
	{
		source:     `const foo = []; async function x() { foo[bar].baz += await result;  }`,
		atomic:     1,
		lostUpdate: 1,
		reason:     "",
	},
	{
		source:     `const foo = {}; class C { #bar; async wrap() { foo.#bar += await baz } }`,
		atomic:     1,
		lostUpdate: 1,
		reason:     "",
	},
	{
		source:     `let foo; async function* x() { foo = (yield foo) + await bar; }`,
		atomic:     1,
		lostUpdate: 0,
		reason:     "a yield's value is what the caller sends, so foo is not computed from foo",
	},
	{
		source:     `let foo; async function x() { foo = foo + await result(foo); }`,
		atomic:     1,
		lostUpdate: 1,
		reason:     "",
	},
	{
		source:     `let foo; async function x() { foo = await result(foo, await somethingElse); }`,
		atomic:     1,
		lostUpdate: 0,
		reason:     "the value written is an awaited call's result, the callee's answer rather than a computation made here",
	},
	{
		source:     `function* x() { let foo; yield async function y() { foo += await bar; } }`,
		atomic:     1,
		lostUpdate: 1,
		reason:     "",
	},
	{
		source:     `let foo; async function* x() { foo = await foo + (yield bar); }`,
		atomic:     1,
		lostUpdate: 1,
		reason:     "",
	},
	{
		source:     `let foo; async function x() { foo = bar + await foo; }`,
		atomic:     1,
		lostUpdate: 1,
		reason:     "",
	},
	{
		source:     `let foo = {}; async function x() { foo[bar].baz = await (foo.bar += await foo[bar].baz) }`,
		atomic:     2,
		lostUpdate: 1,
		reason:     "the outer write is a plain `=` through a computed key, which is not judged; the inner `+=` is",
	},
	{
		source:     `let foo = ''; async function x() { foo += await bar; }`,
		atomic:     1,
		lostUpdate: 1,
		reason:     "",
	},
	{
		source:     `let foo = 0; async function x() { foo = (a ? b : foo) + await bar; if (baz); }`,
		atomic:     1,
		lostUpdate: 1,
		reason:     "",
	},
	{
		source:     `let foo = 0; async function x() { foo = (a ? b ? c ? d ? foo : e : f : g : h) + await bar; if (baz); }`,
		atomic:     1,
		lostUpdate: 1,
		reason:     "",
	},
	{
		source: `
                async function f(foo) {
                    let buz = await get(foo.id);
                    foo.bar = buz.bar;
                }
            `,
		atomic:     1,
		lostUpdate: 0,
		reason:     "a plain write of a value fetched after the read, not computed from foo.bar",
	},
	{
		source: `
                async () => {
                    opts.spec = process.stdin;
                    try {
                        const { exit_code } = await run(opts);
                        process.exitCode = exit_code;
                    } catch (e) {
                        process.exitCode = 1;
                    }
              };
            `,
		atomic:     2,
		lostUpdate: 0,
		reason:     "plain writes of an exit code",
	},
	{
		source: `
                async function a(foo) {
                    if (foo.bar) {
                        foo.bar = await something;
                    }
                }
            `,
		atomic:     1,
		lostUpdate: 0,
		reason:     "check-then-act: the value written is not computed from the guard read",
	},
	{
		source: `
                function* g(foo) {
                    baz = foo.bar;
                    yield something;
                    foo.bar = 1;
                }
            `,
		atomic:     1,
		lostUpdate: 0,
		reason:     "a constant",
	},
	{
		source: `
                async function a(foo) {
                    if (foo.bar) {
                        foo.bar = await something;
                    }
                }
            `,
		atomic:     1,
		lostUpdate: 0,
		reason:     "check-then-act: the value written is not computed from the guard read",
	},
	{
		source: `
                function* g(foo) {
                    baz = foo.bar;
                    yield something;
                    foo.bar = 1;
                }
            `,
		atomic:     1,
		lostUpdate: 0,
		reason:     "a constant",
	},
	{
		source: `
                async function a(foo) {
                    if (foo.bar) {
                        foo.bar = await something;
                    }
                }
            `,
		atomic:     1,
		lostUpdate: 0,
		reason:     "check-then-act: the value written is not computed from the guard read",
	},
	{
		source: `
                function* g(foo) {
                    baz = foo.bar;
                    yield something;
                    foo.bar = 1;
                }
            `,
		atomic:     1,
		lostUpdate: 0,
		reason:     "a constant",
	},
	{
		source: `
                let foo;
                async function a() {
                    if (foo) {
                        foo = await something;
                    }
                }
            `,
		atomic:     1,
		lostUpdate: 0,
		reason:     "check-then-act: the value written is not computed from the guard read",
	},
	{
		source: `
                let foo;
                function* g() {
                    baz = foo;
                    yield something;
                    foo = 1;
                }
            `,
		atomic:     1,
		lostUpdate: 0,
		reason:     "a constant",
	},
	{
		source:     `async function x() { let foo; foo += await bar; }`,
		atomic:     0,
		lostUpdate: 0,
		reason:     "",
	},
	{
		source:     `async function x() { let foo; bar(() => foo); foo += await amount; }`,
		atomic:     1,
		lostUpdate: 0,
		reason:     "a closure only reads foo, so nothing else can write it during the suspension",
	},
	{
		source:     `async function x() { let foo; bar(() => baz += 1); foo += await amount; }`,
		atomic:     0,
		lostUpdate: 0,
		reason:     "",
	},
	{
		source:     `let foo; async function x() { foo += await amount; }`,
		atomic:     1,
		lostUpdate: 1,
		reason:     "",
	},
	{
		source:     `async function f(foo) { foo = await bar; }`,
		atomic:     0,
		lostUpdate: 0,
		reason:     "",
	},
	{
		source:     `async function f(foo) { let b = await get(foo.id); foo.bar = b.bar; }`,
		atomic:     1,
		lostUpdate: 0,
		reason:     "a plain write of a value fetched after the read",
	},
	{
		source:     `async function f() { let foo = {}; let bar = await get(foo.id); foo.prop = bar.prop; }`,
		atomic:     0,
		lostUpdate: 0,
		reason:     "",
	},
	{
		source:     `let foo; function x() { foo = foo + amount; }`,
		atomic:     0,
		lostUpdate: 0,
		reason:     "",
	},
	{
		source:     `let holder: any; async function f() { const q = holder.a; try { const r = await run(); holder.b = r; } catch (e) { holder.b = 1; } }`,
		atomic:     2,
		lostUpdate: 0,
		reason:     "plain writes of a fetched value and a constant",
	},
	{
		source:     `async function f() { const q = neverDeclaredAnywhere.a; await run(); neverDeclaredAnywhere.b = 1; }`,
		atomic:     0,
		lostUpdate: 0,
		reason:     "",
	},
	{
		source: `declare const process: any;
declare function run(options: any): Promise<any>;
async function main() {
    const opts: any = {};
    opts.spec = process.stdin;
    try {
        const { exit_code } = await run(opts);
        process.exitCode = exit_code;
    } catch (e) {
        process.exitCode = 1;
    }
}`,
		atomic:     2,
		lostUpdate: 0,
		reason:     "plain writes of an exit code",
	},
	{
		source:     `function o() { let g = false; async function f() { if (g) return; g = true; try { await s(); } finally { g = false; } } }`,
		atomic:     1,
		lostUpdate: 0,
		reason:     "a guard flag reset to a constant",
	},
	{
		source:     `function o() { let g = false; async function f() { g = true; try { await s(); } finally { g = false; } } }`,
		atomic:     0,
		lostUpdate: 0,
		reason:     "",
	},
	{
		source:     `function o() { let g = false; async function f() { if (g) return; g = true; try { await s(); } finally { } } }`,
		atomic:     0,
		lostUpdate: 0,
		reason:     "",
	},
	{
		source:     `function o() { let g = false; async function f() { if (g) return; await s(); g = false; } }`,
		atomic:     1,
		lostUpdate: 0,
		reason:     "a guard flag reset to a constant",
	},
	{
		source: `declare function a(): Promise<void>;
declare function b(): Promise<number>;
declare function use(value: unknown): void;
async function f(entry: any) { if (entry.status !== 1) return;
    try { await a(); entry.position = await b(); } catch (error) { entry.error = 1; } }`,
		atomic:     2,
		lostUpdate: 0,
		reason:     "plain writes of a fetched value and a constant",
	},
	{
		source: `declare function a(): Promise<void>;
declare function b(): Promise<number>;
declare function use(value: unknown): void;
async function f(entry: any) { if (entry.status !== 1) return;
    try { await a(); entry.position = await b(); use(entry.position); } catch (error) { entry.error = 1; } }`,
		atomic:     2,
		lostUpdate: 0,
		reason:     "plain writes of a fetched value and a constant",
	},
	{
		source: `declare function a(): Promise<void>;
declare function b(): Promise<number>;
declare function use(value: unknown): void;
async function f(entry: any) {
    try { await a(); } catch (error) { entry.error = 1; } }`,
		atomic:     0,
		lostUpdate: 0,
		reason:     "",
	},
	{
		source: `declare function callback(): Promise<void>;
const holder = { slot: 0 };
async function withRestore() {
    const saved = holder.slot;
    try { await callback(); }
    finally { holder.slot = saved; }
}`,
		atomic:     1,
		lostUpdate: 0,
		reason:     "a restore of exactly the saved value",
	},
	{
		source: `declare function callback(): Promise<void>;
declare const ambientHolder: { slot: number };
async function withRestore() {
    const saved = ambientHolder.slot;
    try { await callback(); }
    finally { ambientHolder.slot = saved; }
}`,
		atomic:     1,
		lostUpdate: 0,
		reason:     "a restore of exactly the saved value",
	},
	{
		source: `declare function callback(): Promise<void>;
const holder = { slot: 0 };
async function withRestore() {
    try { await callback(); }
    finally { holder.slot = 0; }
}`,
		atomic:     0,
		lostUpdate: 0,
		reason:     "",
	},
}

func TestConcurrencyNoLostUpdateReplaysTheAtomicUpdatesCorpus(t *testing.T) {
	t.Parallel()

	if len(concurrencyNoLostUpdateAtomicCorpus) != 84 {
		t.Fatalf("the replay holds %d cases, want 84", len(concurrencyNoLostUpdateAtomicCorpus))
	}
	reporting := 0
	for index, testCase := range concurrencyNoLostUpdateAtomicCorpus {
		if testCase.lostUpdate > 0 {
			reporting++
		}
		if testCase.lostUpdate > testCase.atomic {
			t.Errorf("case %d: this rule reports where the rule it narrows does not", index)
		}
		if (testCase.lostUpdate != testCase.atomic) != (testCase.reason != "") {
			t.Errorf("case %d: a difference needs a reason, and only a difference has one", index)
		}
		result := rule_testing.RunTyped(t, ConcurrencyNoLostUpdate, concurrencyNoLostUpdateFile, testCase.source)
		if len(result.Diagnostics) != testCase.lostUpdate {
			t.Errorf("case %d: got %d findings, want %d (%v)", index, len(result.Diagnostics), testCase.lostUpdate, result.MessageIds())
		}
	}
	// The control: a replay in which nothing reports would pass every silent row above.
	if reporting < 20 {
		t.Fatalf("only %d replayed cases report; the corpus's read-modify-writes are not being judged", reporting)
	}
}
