// Code generated from upstream's corpus by /tmp/tjpst83/gen_rau_fixtures.js. Do not hand-edit;
// see requireAtomicUpdatesCorpusProvenance in require_atomic_updates_test.go.

package core

// requireAtomicUpdatesCase is one imported upstream case: its source, the options it was
// tested under, and the message ids upstream reports on it in order.
type requireAtomicUpdatesCase struct {
	source  string
	options any
	wantIds []string
}

var requireAtomicUpdatesCleanCases = []requireAtomicUpdatesCase{
	{
		source:  `let foo; async function x() { foo += bar; }`,
		options: nil,
	},
	{
		source:  `let foo; async function x() { foo = foo + bar; }`,
		options: nil,
	},
	{
		source:  `let foo; async function x() { foo = await bar + foo; }`,
		options: nil,
	},
	{
		source:  `async function x() { let foo; foo += await bar; }`,
		options: nil,
	},
	{
		source:  `let foo; async function x() { foo = (await result)(foo); }`,
		options: nil,
	},
	{
		source:  `let foo; async function x() { foo = bar(await something, foo) }`,
		options: nil,
	},
	{
		source:  `function* x() { let foo; foo += yield bar; }`,
		options: nil,
	},
	{
		source:  `const foo = {}; async function x() { foo.bar = await baz; }`,
		options: nil,
	},
	{
		source:  `const foo = []; async function x() { foo[x] += 1;  }`,
		options: nil,
	},
	{
		source:  `let foo; function* x() { foo = bar + foo; }`,
		options: nil,
	},
	{
		source:  `async function x() { let foo; bar(() => baz += 1); foo += await amount; }`,
		options: nil,
	},
	{
		source:  `let foo; async function x() { foo = condition ? foo : await bar; }`,
		options: nil,
	},
	{
		source:  `async function x() { let foo; bar(() => { let foo; blah(foo); }); foo += await result; }`,
		options: nil,
	},
	{
		source:  `let foo; async function x() { foo = foo + 1; await bar; }`,
		options: nil,
	},
	{
		source:  `async function x() { foo += await bar; }`,
		options: nil,
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
		options: nil,
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
		options: nil,
	},
	{
		source: `
            async function f() {
                let records
                records = await a.records
                g(() => { records })
            }
        `,
		options: nil,
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
		options: nil,
	},
	{
		source: `
            async function f(foo) {
                let bar = await get(foo.id);
                bar.prop = foo.prop;
            }
        `,
		options: nil,
	},
	{
		source: `
            async function f(foo) {
                let bar = await get(foo.id);
                foo = bar.prop;
            }
        `,
		options: nil,
	},
	{
		source: `
            async function f() {
                let foo = {}
                let bar = await get(foo.id);
                foo.prop = bar.prop;
            }
        `,
		options: nil,
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
		options: nil,
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
		options: nil,
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
		options: nil,
	},
	{
		source: `
            async function run() {
                await a;
                b = 1;
            }
        `,
		options: nil,
	},
	{
		source: `
                async function a(foo) {
                    if (foo.bar) {
                        foo.bar = await something;
                    }
                }
            `,
		options: RequireAtomicUpdatesOptions{AllowProperties: true},
	},
	{
		source: `
                function* g(foo) {
                    baz = foo.bar;
                    yield something;
                    foo.bar = 1;
                }
            `,
		options: RequireAtomicUpdatesOptions{AllowProperties: true},
	},
}

var requireAtomicUpdatesReportingCases = []requireAtomicUpdatesCase{
	{
		source:  `let foo; async function x() { foo += await amount; }`,
		options: nil,
		wantIds: []string{"nonAtomicUpdate"},
	},
	{
		source:  `if (1); let foo; async function x() { foo += await amount; }`,
		options: nil,
		wantIds: []string{"nonAtomicUpdate"},
	},
	{
		source:  `let foo; async function x() { while (condition) { foo += await amount; } }`,
		options: nil,
		wantIds: []string{"nonAtomicUpdate"},
	},
	{
		source:  `let foo; async function x() { foo = foo + await amount; }`,
		options: nil,
		wantIds: []string{"nonAtomicUpdate"},
	},
	{
		source:  `let foo; async function x() { foo = foo + (bar ? baz : await amount); }`,
		options: nil,
		wantIds: []string{"nonAtomicUpdate"},
	},
	{
		source:  `let foo; async function x() { foo = foo + (bar ? await amount : baz); }`,
		options: nil,
		wantIds: []string{"nonAtomicUpdate"},
	},
	{
		source:  `let foo; async function x() { foo = condition ? foo + await amount : somethingElse; }`,
		options: nil,
		wantIds: []string{"nonAtomicUpdate"},
	},
	{
		source:  `let foo; async function x() { foo = (condition ? foo : await bar) + await bar; }`,
		options: nil,
		wantIds: []string{"nonAtomicUpdate"},
	},
	{
		source:  `let foo; async function x() { foo += bar + await amount; }`,
		options: nil,
		wantIds: []string{"nonAtomicUpdate"},
	},
	{
		source:  `async function x() { let foo; bar(() => foo); foo += await amount; }`,
		options: nil,
		wantIds: []string{"nonAtomicUpdate"},
	},
	{
		source:  `let foo; function* x() { foo += yield baz }`,
		options: nil,
		wantIds: []string{"nonAtomicUpdate"},
	},
	{
		source:  `let foo; async function x() { foo = bar(foo, await something) }`,
		options: nil,
		wantIds: []string{"nonAtomicUpdate"},
	},
	{
		source:  `const foo = {}; async function x() { foo.bar += await baz }`,
		options: nil,
		wantIds: []string{"nonAtomicObjectUpdate"},
	},
	{
		source:  `const foo = []; async function x() { foo[bar].baz += await result;  }`,
		options: nil,
		wantIds: []string{"nonAtomicObjectUpdate"},
	},
	{
		source:  `const foo = {}; class C { #bar; async wrap() { foo.#bar += await baz } }`,
		options: nil,
		wantIds: []string{"nonAtomicObjectUpdate"},
	},
	{
		source:  `let foo; async function* x() { foo = (yield foo) + await bar; }`,
		options: nil,
		wantIds: []string{"nonAtomicUpdate"},
	},
	{
		source:  `let foo; async function x() { foo = foo + await result(foo); }`,
		options: nil,
		wantIds: []string{"nonAtomicUpdate"},
	},
	{
		source:  `let foo; async function x() { foo = await result(foo, await somethingElse); }`,
		options: nil,
		wantIds: []string{"nonAtomicUpdate"},
	},
	{
		source:  `function* x() { let foo; yield async function y() { foo += await bar; } }`,
		options: nil,
		wantIds: []string{"nonAtomicUpdate"},
	},
	{
		source:  `let foo; async function* x() { foo = await foo + (yield bar); }`,
		options: nil,
		wantIds: []string{"nonAtomicUpdate"},
	},
	{
		source:  `let foo; async function x() { foo = bar + await foo; }`,
		options: nil,
		wantIds: []string{"nonAtomicUpdate"},
	},
	{
		source:  `let foo = {}; async function x() { foo[bar].baz = await (foo.bar += await foo[bar].baz) }`,
		options: nil,
		wantIds: []string{"nonAtomicObjectUpdate", "nonAtomicObjectUpdate"},
	},
	{
		source:  `let foo = ''; async function x() { foo += await bar; }`,
		options: nil,
		wantIds: []string{"nonAtomicUpdate"},
	},
	{
		source:  `let foo = 0; async function x() { foo = (a ? b : foo) + await bar; if (baz); }`,
		options: nil,
		wantIds: []string{"nonAtomicUpdate"},
	},
	{
		source:  `let foo = 0; async function x() { foo = (a ? b ? c ? d ? foo : e : f : g : h) + await bar; if (baz); }`,
		options: nil,
		wantIds: []string{"nonAtomicUpdate"},
	},
	{
		source: `
                async function f(foo) {
                    let buz = await get(foo.id);
                    foo.bar = buz.bar;
                }
            `,
		options: nil,
		wantIds: []string{"nonAtomicObjectUpdate"},
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
		options: nil,
		wantIds: []string{"nonAtomicObjectUpdate", "nonAtomicObjectUpdate"},
	},
	{
		source: `
                async function a(foo) {
                    if (foo.bar) {
                        foo.bar = await something;
                    }
                }
            `,
		options: nil,
		wantIds: []string{"nonAtomicObjectUpdate"},
	},
	{
		source: `
                function* g(foo) {
                    baz = foo.bar;
                    yield something;
                    foo.bar = 1;
                }
            `,
		options: nil,
		wantIds: []string{"nonAtomicObjectUpdate"},
	},
	{
		source: `
                async function a(foo) {
                    if (foo.bar) {
                        foo.bar = await something;
                    }
                }
            `,
		options: RequireAtomicUpdatesOptions{},
		wantIds: []string{"nonAtomicObjectUpdate"},
	},
	{
		source: `
                function* g(foo) {
                    baz = foo.bar;
                    yield something;
                    foo.bar = 1;
                }
            `,
		options: RequireAtomicUpdatesOptions{},
		wantIds: []string{"nonAtomicObjectUpdate"},
	},
	{
		source: `
                async function a(foo) {
                    if (foo.bar) {
                        foo.bar = await something;
                    }
                }
            `,
		options: RequireAtomicUpdatesOptions{AllowProperties: false},
		wantIds: []string{"nonAtomicObjectUpdate"},
	},
	{
		source: `
                function* g(foo) {
                    baz = foo.bar;
                    yield something;
                    foo.bar = 1;
                }
            `,
		options: RequireAtomicUpdatesOptions{AllowProperties: false},
		wantIds: []string{"nonAtomicObjectUpdate"},
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
		options: RequireAtomicUpdatesOptions{AllowProperties: true},
		wantIds: []string{"nonAtomicUpdate"},
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
		options: RequireAtomicUpdatesOptions{AllowProperties: true},
		wantIds: []string{"nonAtomicUpdate"},
	},
}
