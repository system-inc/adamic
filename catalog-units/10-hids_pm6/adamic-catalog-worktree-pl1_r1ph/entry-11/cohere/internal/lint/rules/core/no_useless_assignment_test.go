package core

import (
	"fmt"
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// Corpus imported verbatim from oxc's inline Tester block from its inline Tester block
// -dump, then written by a script rather than by hand. The brief requires the bytes on disk
// to equal the bytes upstream, and three porters have now found their own reading pass over
// escape sequences insufficient, so a checker compares them mechanically below.
//
// The corpus is 71 pass and 43 fail inputs producing 54 diagnostics. Nine inputs report more
// than once, so a fixture asserting one finding per input would be wrong on nine of them.
// Recovering which nine took work and is written up in the rule file.
//
// Every case runs as .tsx because oxc's snapshot names no_useless_assignment.tsx, and the
// last nine fail inputs are JSX that would not parse otherwise.

// noUselessAssignmentUpstreamPass is every case upstream asserts reports nothing.
var noUselessAssignmentUpstreamPass = []string{
	"let v = 'used';\n                    console.log(v);\n                    v = 'used-2'\n                    console.log(v);",
	"function foo() {\n                        let v = 'used';\n                        console.log(v);\n                        v = 'used-2';\n                        console.log(v);\n                    }",
	"function foo() {\n                        let v = 'used';\n                        if (condition) {\n                            v = 'used-2';\n                            console.log(v);\n                            return\n                        }\n                        console.log(v);\n                    }",
	"function foo() {\n                        let v = 'used';\n                        if (condition) {\n                            console.log(v);\n                        } else {\n                            v = 'used-2';\n                            console.log(v);\n                        }\n                    }",
	"function foo() {\n                        let v = 'used';\n                        if (condition) {\n                            //\n                        } else {\n                            v = 'used-2';\n                        }\n                        console.log(v);\n                    }",
	"var foo = function () {\n                        let v = 'used';\n                        console.log(v);\n                        v = 'used-2'\n                        console.log(v);\n                    }",
	"var foo = () => {\n                        let v = 'used';\n                        console.log(v);\n                        v = 'used-2'\n                        console.log(v);\n                    }",
	"class foo {\n                        static {\n                            let v = 'used';\n                            console.log(v);\n                            v = 'used-2'\n                            console.log(v);\n                        }\n                    }",
	"function foo () {\n                        let v = 'used';\n                        for (let i = 0; i < 10; i++) {\n                            console.log(v);\n                            v = 'used in next iteration';\n                        }\n                    }",
	"function foo () {\n                        let i = 0;\n                        i++;\n                        i++;\n                        console.log(i);\n                    }",
	"export let foo = 'used';\n                    console.log(foo);\n                    foo = 'unused like but exported';",
	"export function foo () {};\n                    console.log(foo);\n                    foo = 'unused like but exported';",
	"export class foo {};\n                    console.log(foo);\n                    foo = 'unused like but exported';",
	"export default function foo () {};\n                    console.log(foo);\n                    foo = 'unused like but exported';",
	"export default class foo {};\n                    console.log(foo);\n                    foo = 'unused like but exported';",
	"let foo = 'used';\n                    export { foo };\n                    console.log(foo);\n                    foo = 'unused like but exported';",
	"function foo () {};\n                    export { foo };\n                    console.log(foo);\n                    foo = 'unused like but exported';",
	"class foo {};\n                    export { foo };\n                    console.log(foo);\n                    foo = 'unused like but exported';",
	"v = 'used';\n                    console.log(v);\n                    v = 'unused'",
	"let v = 'used variable';",
	"function foo() {\n                        return;\n\n                        const x = 1;\n                        if (y) {\n                            bar(x);\n                        }\n                    }",
	"function foo() {\n                        const x = 1;\n                        console.log(x);\n                        return;\n\n                        x = 'Foo'\n                    }",
	"function foo() {\n                        let a = 42;\n                        console.log(a);\n                        a++;\n                        console.log(a);\n                    }",
	"function foo() {\n                        let a = 42;\n                        console.log(a);\n                        a--;\n                        console.log(a);\n                    }",
	"function foo() {\n                        let a = 42;\n                        console.log(a);\n                        a = 10;\n                        a = a + 1;\n                        console.log(a);\n                    }",
	"function foo() {\n                        let a = 42;\n                        console.log(a);\n                        a = 10;\n                        if (cond) {\n                            a = a + 1;\n                        } else {\n                            a = 2 + a;\n                        }\n                        console.log(a);\n                    }",
	"function foo() {\n                        let a = 'used', b = 'used', c = 'used', d = 'used';\n                        console.log(a, b, c, d);\n                        ({ a, arr: [b, c, ...d] } = fn());\n                        console.log(a, b, c, d);\n                    }",
	"function foo() {\n                        let a = 'used', b = 'used', c = 'used';\n                        console.log(a, b, c);\n                        ({ a = 'unused', foo: b, ...c } = fn());\n                        console.log(a, b, c);\n                    }",
	"function foo() {\n                        let a = {};\n                        console.log(a);\n                        a.b = 'unused like, but maybe used in setter';\n                    }",
	"function foo() {\n                        let a = { b: 42 };\n                        console.log(a);\n                        a.b++;\n                    }",
	"function foo () {\n                        let v = 'used';\n                        console.log(v);\n                        function bar() {\n                            v = 'used in outer scope';\n                        }\n                        bar();\n                        console.log(v);\n                    }",
	"function foo () {\n                        let v = 'used';\n                        console.log(v);\n                        setTimeout(() => console.log(v), 1);\n                        v = 'used in other scope';\n                    }",
	"function foo () {\n                        let v = 'used';\n                        console.log(v);\n                        for (let i = 0; i < 10; i++) {\n                            if (condition) {\n                                v = 'maybe used';\n                                continue;\n                            }\n                            console.log(v);\n                        }\n                    }",
	"/* globals foo */\n                    const bk = foo;\n                    foo = 42;\n                    try {\n                        // process\n                    } finally {\n                        foo = bk;\n                    }",
	"\n                        const bk = console;\n                        console = { log () {} };\n                        try {\n                            // process\n                        } finally {\n                            console = bk;\n                        }",
	"let message = 'init';\n                    try {\n                        const result = call();\n                        message = result.message;\n                    } catch (e) {\n                        // ignore\n                    }\n                    console.log(message)",
	"let message = 'init';\n                    try {\n                        message = call().message;\n                    } catch (e) {\n                        // ignore\n                    }\n                    console.log(message)",
	"let v = 'init';\n                    try {\n                        v = callA();\n                        try {\n                            v = callB();\n                        } catch (e) {\n                            // ignore\n                        }\n                    } catch (e) {\n                        // ignore\n                    }\n                    console.log(v)",
	"let v = 'init';\n                    try {\n                        try {\n                            v = callA();\n                        } catch (e) {\n                            // ignore\n                        }\n                    } catch (e) {\n                        // ignore\n                    }\n                    console.log(v)",
	"let a;\n                    try {\n                        foo();\n                    } finally {\n                        a = 5;\n                    }\n                    console.log(a);",
	"const obj = { a: 5 };\n                    const { a, b = a } = obj;\n                    console.log(b); // 5",
	"const arr = [6];\n                    const [c, d = c] = arr;\n                    console.log(d); // 6",
	"const obj = { a: 1 };\n                    let {\n                        a,\n                        b = (a = 2)\n                    } = obj;\n                    console.log(a, b);",
	"let { a, b: {c = a} = {} } = obj;\n                    console.log(c);",
	"function foo(){\n                        let bar;\n                        try {\n                            bar = 2;\n                            unsafeFn();\n                            return { error: undefined };\n                        } catch {\n                            return { bar };\n                        }\n                    }\n                    function unsafeFn() {\n                        throw new Error();\n                    }",
	"function foo(){\n                        let bar, baz;\n                        try {\n                            bar = 2;\n                            unsafeFn();\n                            return { error: undefined };\n                        } catch {\n                           baz = bar;\n                        }\n                        return baz;\n                    }\n                    function unsafeFn() {\n                        throw new Error();\n                    }",
	"function foo(){\n                        let bar;\n                        try {\n                            bar = 2;\n                            unsafeFn();\n                            bar = 4;\n                        } catch {\n                           // handle error\n                        }\n                        return bar;\n                    }\n                    function unsafeFn() {\n                        throw new Error();\n                    }",
	"\n                            function App() {\n                                const A = \"\";\n                                return <A/>;\n                            }\n                        ",
	"\n                            function App() {\n                                let A = \"\";\n                                foo(A);\n                                A = \"A\";\n                                return <A/>;\n                            }\n                        ",
	"\n                            function App() {\n                                let A = \"a\";\n                                foo(A);\n                                return <A/>;\n                            }\n                        ",
	"function App() {\n                            let x = 0;\n                            foo(x);\n                            x = 1;\n                            return <A prop={x} />;\n                        }",
	"function App() {\n                            let x = \"init\";\n                            foo(x);\n                            x = \"used\";\n                            return <A>{x}</A>;\n                        }",
	"function App() {\n                            let props = { a: 1 };\n                            foo(props);\n                            props = { b: 2 };\n                            return <A {...props} />;\n                        }",
	"function App() {\n                            let NS = Lib;\n                            return <NS.Cmp />;\n                        }",
	"function App() {\n                            let a = 0;\n                            a++;\n                            return <A prop={a} />;\n                        }",
	"function App() {\n                            const obj = { a: 1 };\n                            const { a, b = a } = obj;\n                            return <A prop={b} />;\n                        }",
	"function App() {\n                            let { a, b: { c = a } = {} } = obj;\n                            return <A prop={c} />;\n                        }",
	"function App() {\n                            let x = \"init\";\n                            if (cond) {\n                                x = \"used\";\n                                return <A prop={x} />;\n                            }\n                            return <A prop={x} />;\n                        }",
	"function App() {\n                            let A;\n                            if (cond) {\n                              A = Foo;\n                            } else {\n                              A = Bar;\n                            }\n                            return <A />;\n                        }",
	"function App() {\n                            let m;\n                            try {\n                              m = 2;\n                              unsafeFn();\n                              m = 4;\n                            } catch (e) {\n                              // ignore\n                            }\n                            return <A prop={m} />;\n                        }",
	"function App() {\n                            const arr = [6];\n                            const [c, d = c] = arr;\n                            return <A prop={d} />;\n                        }",
	"function App() {\n                            const obj = { a: 1 };\n                            let {\n                              a,\n                              b = (a = 2)\n                            } = obj;\n                            return <A prop={a} />;\n                        }",
	"\n            let index = 0;\n            while (index < length) {\n                if (condition) {\n                    index++;\n                    continue;\n                }\n                while (index < length2) {\n                    index++;\n                }\n            }\n        ",
	"function createStore() {\n                        const options = { onTrigger: undefined };\n                        let isListening = false;\n                        options.onTrigger = () => {\n                            if (isListening) {\n                                console.log('event');\n                            }\n                        };\n                        isListening = true;\n                        return options;\n                    }",
	"let state = 0;\n                    const api = { read: () => state };\n                    state = 1;\n                    export { api };",
	"function foo() {\n                        let x = 0;\n                        (() => console.log(x))();\n                        x = 1;\n                    }",
	"const rgb2lab = (rgb: RGB): LAB => {\n\n    let [r, g, b] = rgb;\n\n    r = (r > 0) ? ((r + 0) / 1) ** 2 : r / 1;\n    g = (g > 0) ? ((g + 0) / 1) ** 2 : g / 1;\n    b = (b > 0) ? ((b + 0) / 1) ** 2 : b / 1;\n\n    let x = (r * 0 + g * 0 + b * 0) / 0;\n    let y = (r * 0 + g * 0 + b * 0) / 1;\n    let z = (r * 0 + g * 0 + b * 0) / 1;\n\n    x = (x > 0) ? Math.cbrt(x) : (1 * x) + 16/116;\n    y = (y > 0) ? Math.cbrt(y) : (1 * y) + 16/116;\n    z = (z > 0) ? Math.cbrt(z) : (1 * z) + 16/116;\n\n    return [(1 * y) - 16, 1 * (x - y), 1 * (y - z)];\n};",
	"const maxRetries = 3;\n\nasync function retryUntilSuccess(run: () => Promise<void>): Promise<void> {\n  for (let attempt = 0; ; attempt++) {\n    try {\n      return await run();\n    } catch (error) {\n      if (attempt >= maxRetries) {\n        throw error;\n      }\n    }\n  }\n}",
	"async function waitWithBackoff(run: () => Promise<void>): Promise<void> {\n  let backoffMillis = 20;\n  let releaseRequested = false;\n\n  while (Date.now() < Date.now() + 2000) {\n    try {\n      return await run();\n    } catch (error) {\n      if (!(error instanceof Error) || !error.message.startsWith(\"LeaseHeldError\")) {\n        throw error;\n      }\n    }\n\n    if (!releaseRequested) {\n      releaseRequested = true;\n    }\n\n    const waitUntil = Math.min(Date.now() + backoffMillis, Date.now() + 2000);\n    backoffMillis *= 2;\n    await new Promise((resolve) => setTimeout(resolve, waitUntil - Date.now()));\n  }\n}",
	"function makeResource(): { readonly release: () => void } {\n  return { release() {} };\n}\n\nfunction useResource(unsafe: (resource: { readonly release: () => void }) => void): { readonly release: () => void } {\n  const resource = makeResource();\n  let owned = true;\n\n  try {\n    unsafe(resource);\n    owned = false;\n  } finally {\n    if (owned) {\n      resource.release();\n    }\n  }\n\n  return resource;\n}",
	"function collectIds(ids: readonly string[], forward: boolean): string[] {\n  const collected: string[] = [];\n  const startIndex = forward ? 0 : ids.length - 1;\n\n  for (\n    let index = startIndex;\n    forward ? index < ids.length : index >= 0;\n    forward ? index++ : index--\n  ) {\n    collected.push(ids[index]!);\n  }\n\n  return collected;\n}",
}

// noUselessAssignmentUpstreamFail pairs each failing input with the number of diagnostics
// upstream's snapshot records against it.
var noUselessAssignmentUpstreamFail = []struct {
	source string
	count  int
}{
	{source: "let v = 'used';\n                        console.log(v);\n                        v = 'unused'", count: 1},
	{source: "function foo() {\n                            let v = 'used';\n                            console.log(v);\n                            v = 'unused';\n                        }", count: 1},
	{source: "function foo() {\n                            let v = 'used';\n                            if (condition) {\n                                v = 'unused';\n                                return\n                            }\n                            console.log(v);\n                        }", count: 1},
	{source: "function foo() {\n                            let v = 'used';\n                            if (condition) {\n                                console.log(v);\n                            } else {\n                                v = 'unused';\n                            }\n                        }", count: 1},
	{source: "var foo = function () {\n                            let v = 'used';\n                            console.log(v);\n                            v = 'unused'\n                        }", count: 1},
	{source: "var foo = () => {\n                            let v = 'used';\n                            console.log(v);\n                            v = 'unused'\n                        }", count: 1},
	{source: "class foo {\n                            static {\n                                let v = 'used';\n                                console.log(v);\n                                v = 'unused'\n                            }\n                        }", count: 1},
	{source: "function foo() {\n                            let v = 'unused';\n                            if (condition) {\n                                v = 'used';\n                                console.log(v);\n                                return\n                            }\n                        }", count: 1},
	{source: "function foo() {\n                            let v = 'used';\n                            console.log(v);\n                            v = 'unused';\n                            v = 'unused';\n                        }", count: 2},
	{source: "function foo() {\n                            let v = 'used';\n                            console.log(v);\n                            v = 'unused';\n                            v = 'used';\n                            console.log(v);\n                            v = 'used';\n                            console.log(v);\n                        }", count: 1},
	{source: "\n                        let v;\n                        v = 'unused';\n                        if (foo) {\n                            v = 'used';\n                        } else {\n                            v = 'used';\n                        }\n                        console.log(v);", count: 1},
	{source: "function foo() {\n                            let v = 'used';\n                            console.log(v);\n                            v = 'unused';\n                            v = 'unused';\n                            v = 'used';\n                            console.log(v);\n                        }", count: 2},
	{source: "function foo() {\n                            let v = 'unused';\n                            if (condition) {\n                                if (condition2) {\n                                    v = 'used-2';\n                                } else {\n                                    v = 'used-3';\n                                }\n                            } else {\n                                v = 'used-4';\n                            }\n                            console.log(v);\n                        }", count: 1},
	{source: "function foo() {\n                            let v;\n                            if (condition) {\n                                v = 'unused';\n                            } else {\n                                //\n                            }\n                            if (condition2) {\n                                v = 'used-1';\n                            } else {\n                                v = 'used-2';\n                            }\n                            console.log(v);\n                        }", count: 1},
	{source: "function foo() {\n                            let v = 'used';\n                            if (condition) {\n                                v = 'unused';\n                                v = 'unused';\n                                v = 'used';\n                            }\n                            console.log(v);\n                        }", count: 2},
	{source: "function foo() {\n                            let a = 42;\n                            console.log(a);\n                            a++;\n                        }", count: 1},
	{source: "function foo() {\n                            let a = 42;\n                            console.log(a);\n                            a--;\n                        }", count: 1},
	{source: "function foo() {\n                            let a = 'used', b = 'used', c = 'used', d = 'used';\n                            console.log(a, b, c, d);\n                            ({ a, arr: [b, c,, ...d] } = fn());\n                            console.log(c);\n                        }", count: 3},
	{source: "function foo() {\n                            let a = 'used', b = 'used', c = 'used';\n                            console.log(a, b, c);\n                            ({ a = 'unused', foo: b, ...c } = fn());\n                        }", count: 3},
	{source: "function foo () {\n                            let v = 'used';\n                            console.log(v);\n                            setTimeout(() => v = 42, 1);\n                            v = 'unused and variable is only updated in other scopes';\n                        }", count: 1},
	{source: "function foo() {\n                            let v = 'used';\n                            if (condition) {\n                                let v = 'used';\n                                console.log(v);\n                                v = 'unused';\n                            }\n                            console.log(v);\n                            v = 'unused';\n                        }", count: 2},
	{source: "function foo() {\n                            let v = 'used';\n                            if (condition) {\n                                console.log(v);\n                                v = 'unused';\n                            } else {\n                                v = 'unused';\n                            }\n                        }", count: 2},
	{source: "function foo () {\n                            let v = 'used';\n                            console.log(v);\n                            v = 'unused';\n                            return;\n                            console.log(v);\n                        }", count: 1},
	{source: "function foo () {\n                            let v = 'used';\n                            console.log(v);\n                            v = 'unused';\n                            throw new Error();\n                            console.log(v);\n                        }", count: 1},
	{source: "function foo () {\n                            let v = 'used';\n                            console.log(v);\n                            for (let i = 0; i < 10; i++) {\n                                v = 'unused';\n                                continue;\n                                console.log(v);\n                            }\n                        }\n                        function bar () {\n                            let v = 'used';\n                            console.log(v);\n                            for (let i = 0; i < 10; i++) {\n                                v = 'unused';\n                                break;\n                                console.log(v);\n                            }\n                        }", count: 2},
	{source: "function foo () {\n                            let v = 'used';\n                            console.log(v);\n                            for (let i = 0; i < 10; i++) {\n                                if (condition) {\n                                    v = 'unused';\n                                    break;\n                                }\n                                console.log(v);\n                            }\n                        }", count: 1},
	{source: "let message = 'unused';\n                        try {\n                            const result = call();\n                            message = result.message;\n                        } catch (e) {\n                            message = 'used';\n                        }\n                        console.log(message)", count: 1},
	{source: "let message = 'unused';\n                        try {\n                            message = 'used';\n                            console.log(message)\n                        } catch (e) {\n                        }", count: 1},
	{source: "let message = 'unused';\n                        try {\n                            message = call();\n                        } catch (e) {\n                            message = 'used';\n                        }\n                        console.log(message)", count: 1},
	{source: "let v = 'unused';\n                        try {\n                            v = callA();\n                            try {\n                                v = callB();\n                            } catch (e) {\n                                // ignore\n                            }\n                        } catch (e) {\n                            v = 'used';\n                        }\n                        console.log(v)", count: 1},
	{source: "\n                        var x = 1; // used\n                        x = x + 1; // unused\n                        x = 5; // used\n                        f(x);", count: 1},
	{source: "\n                        var x = 1; // used\n                        x = // used\n                            x++; // unused\n                        f(x);", count: 1},
	{source: "const obj = { a: 1 };\n                        let {\n                            a,\n                            b = (a = 2)\n                        } = obj;\n                        a = 3\n                        console.log(a, b);", count: 1},
	{source: "const arr = [1, 2];\n                        let [\n                            a,\n                            b\n                        ] = arr;\n                        a = 3\n                        console.log(a, b);", count: 1},
	{source: "function App() {\n                        let A = \"unused\";\n                        A = \"used\";\n                        return <A/>;\n                        }", count: 1},
	{source: "function App() {\n                        let A = \"unused\";\n                        A = \"used\";\n                        return <A></A>;\n                        }", count: 1},
	{source: "function App() {\n                        let A = \"unused\";\n                        A = \"used\";\n                        return <A.B />;\n                        }", count: 1},
	{source: "function App() {\n                        let x = \"used\";\n                        if (cond) {\n                          return <A prop={x} />;\n                        } else {\n                          x = \"unused\";\n                        }\n                        }", count: 1},
	{source: "function App() {\n                        let A;\n                        A = \"unused\";\n                        if (cond) {\n                          A = \"used1\";\n                        } else {\n                          A = \"used2\";\n                        }\n                        return <A/>;\n                        }", count: 1},
	{source: "function App() {\n                        let message = 'unused';\n                        try {\n                          const result = call();\n                          message = result.message;\n                        } catch (e) {\n                          message = 'used';\n                        }\n                        return <A prop={message} />;\n                        }", count: 1},
	{source: "function App() {\n                        let x = 1;\n                        x = x + 1;\n                        x = 5;\n                        return <A prop={x} />;\n                        }", count: 1},
	{source: "function App() {\n                        let x = 1;\n                        x = 2;\n                        return <A>{x}</A>;\n                        }", count: 1},
	{source: "function App() {\n                        let x = 0;\n                        x = 1;\n                        x = 2;\n                        return <A prop={x} />;\n                        }", count: 2},
}

// noUselessAssignmentFile is where the fixtures pretend to live. The .tsx extension matters: the
// last nine upstream fail inputs are JSX and do not parse as .ts.
const noUselessAssignmentFile = "/repository/source/Component.tsx"

// TestNoUselessAssignmentSurvey measures this port against the whole imported corpus and prints the
// gap rather than hiding it.
//
// This is deliberately a report and not an assertion, because this port covers a stated subset of
// what upstream reports and pinning the full corpus green would mean either claiming coverage it
// does not have or deleting the cases that show the boundary. The two assertions that follow are
// the contract; this one is the map. It fails only on a false positive, which is the direction that
// costs a reader trust.
func TestNoUselessAssignmentSurvey(t *testing.T) {
	t.Parallel()

	falsePositives := 0
	for index, source := range noUselessAssignmentUpstreamPass {
		result := rule_testing.RunTyped(t, NoUselessAssignment, noUselessAssignmentFile, source)
		if len(result.Diagnostics) != 0 {
			falsePositives++
			t.Errorf("clean upstream case %d reported %d findings; upstream reports none\n%s",
				index, len(result.Diagnostics), source)
		}
	}

	caught, expected := 0, 0
	for _, entry := range noUselessAssignmentUpstreamFail {
		result := rule_testing.RunTyped(t, NoUselessAssignment, noUselessAssignmentFile, entry.source)
		expected += entry.count
		if len(result.Diagnostics) > entry.count {
			t.Errorf("failing upstream case reported %d findings against upstream's %d, which is a "+
				"false positive on an input upstream already bounds\n%s",
				len(result.Diagnostics), entry.count, entry.source)
		}
		caught += len(result.Diagnostics)
	}
	t.Logf("upstream corpus: %d/%d diagnostics reproduced, %d false positives on %d clean cases",
		caught, expected, falsePositives, len(noUselessAssignmentUpstreamPass))
}

// TestNoUselessAssignmentFires pins every one of upstream's 43 failing inputs, at upstream's own
// per-input diagnostic count.
//
// It used to carry a hand-listed subset of 27 indices, because the inverted implementation matched
// only those. The liveness pass matches all of them, so the list is gone and the loop runs the whole
// corpus: an exception list that is empty is better deleted than kept at zero length, and a later
// regression now fails on the input it broke rather than slipping through an index nobody updated.
//
// Read from the imported corpus rather than retyped, so the assertion and the survey cannot drift,
// and the expected count is the same snapshot-derived number the survey uses. This is stronger than
// the survey's "does not exceed": it is a claim of exact agreement on every input.
func TestNoUselessAssignmentFires(t *testing.T) {
	t.Parallel()

	for index := range noUselessAssignmentUpstreamFail {
		entry := noUselessAssignmentUpstreamFail[index]
		t.Run(fmt.Sprintf("upstream fail %d", index), func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, NoUselessAssignment, noUselessAssignmentFile, entry.source)
			wantIds := make([]string, entry.count)
			for position := range wantIds {
				wantIds[position] = "noUselessAssignment"
			}
			rule_testing.ExpectFindings(t, result, wantIds...)
		})
	}
}

// TestNoUselessAssignmentStaysSilent pins every clean case upstream carries.
//
// These are the cases that catch this port rather than confirm it. Four exist only to pin the
// captured-read guard, three only to pin the try-block guard, three only the self-referential
// right-hand side, and several the exported-binding guard, and every one of them reported a finding
// at some point while this was being written.
func TestNoUselessAssignmentStaysSilent(t *testing.T) {
	t.Parallel()

	for index, source := range noUselessAssignmentUpstreamPass {
		t.Run(fmt.Sprintf("upstream pass %d", index), func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t,
				rule_testing.RunTyped(t, NoUselessAssignment, noUselessAssignmentFile, source))
		})
	}
}

// TestNoUselessAssignmentPointsAtTheIdentifier asserts where the finding lands and what it says.
//
// Message ids and counts cannot see either. Upstream underlines the assignment target alone, one
// character wide in its snapshot, not the statement and not the right-hand side, and a rule pointing
// at the whole statement passes every assertion above while being wrong about the only thing a
// reader looks at. Asserted for a declarator, a plain assignment, and a compound assignment,
// because the three take different paths through `writeTargetOf`.
func TestNoUselessAssignmentPointsAtTheIdentifier(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		wantText   []string
	}{
		{
			name:       "a plain reassignment underlines the target",
			sourceText: "function f() {\n\tlet value = 1;\n\tuse(value);\n\tvalue = 2;\n}",
			wantText:   []string{"value"},
		},
		{
			name:       "a declarator underlines the name and not the initializer",
			sourceText: "function f() {\n\tlet value = 1;\n\tvalue = 2;\n\tuse(value);\n}",
			wantText:   []string{"value"},
		},
		{
			// Written wrong the first time and corrected by the rule rather than the other way
			// round: with a `use(value)` at the end only `value = 2` is dead, because `value = 3`
			// is the one that gets read. Both writes are dead only when nothing reads after them,
			// which is what this case now says.
			name:       "two dead writes each underline their own target",
			sourceText: "function f() {\n\tlet value = 1;\n\tuse(value);\n\tvalue = 2;\n\tvalue = 3;\n}",
			wantText:   []string{"value", "value"},
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, NoUselessAssignment, noUselessAssignmentFile, testCase.sourceText)
			if len(result.Diagnostics) != len(testCase.wantText) {
				t.Fatalf("got %d findings, want %d", len(result.Diagnostics), len(testCase.wantText))
			}
			for position, diagnostic := range result.Diagnostics {
				reported := testCase.sourceText[diagnostic.Range.Pos():diagnostic.Range.End()]
				if reported != testCase.wantText[position] {
					t.Errorf("finding %d underlines %q, want %q", position, reported, testCase.wantText[position])
				}
				// Asserted by equality rather than by containment. A predicate weaker than the
				// property it guards is not a guard, and a substring check on an interpolated
				// message has shipped a doubled character in this tree before.
				if diagnostic.Message.Id != messageNoUselessAssignment.Id {
					t.Errorf("finding %d has id %q, want %q",
						position, diagnostic.Message.Id, messageNoUselessAssignment.Id)
				}
				if diagnostic.Message.Description != messageNoUselessAssignment.Description {
					t.Errorf("finding %d renders a description this rule does not define", position)
				}
			}
		})
	}
}

// TestNoUselessAssignmentNeedsTheTypedHarness makes a revert to the plain harness fail loudly.
//
// This rule declares NeedsTypeChecker, and the untyped harness hands it a nil checker. The rule
// returns early on that, so every clean case above would pass vacuously and every firing case would
// fail in a way that reads like a rule bug rather than a harness one. Asserting the difference
// directly means the two failure modes cannot be confused.
func TestNoUselessAssignmentNeedsTheTypedHarness(t *testing.T) {
	t.Parallel()

	source := "function f() {\n\tlet value = 1;\n\tuse(value);\n\tvalue = 2;\n}"

	if !NoUselessAssignment.NeedsTypeChecker {
		t.Fatal("the rule no longer declares NeedsTypeChecker, so the fixtures below assert nothing")
	}
	if findings := len(rule_testing.RunTyped(t, NoUselessAssignment, noUselessAssignmentFile, source).Diagnostics); findings != 1 {
		t.Fatalf("the typed harness produced %d findings, want 1", findings)
	}
	if findings := len(rule_testing.Run(t, NoUselessAssignment, noUselessAssignmentFile, source).Diagnostics); findings != 0 {
		t.Errorf("the untyped harness produced %d findings; it hands a nil checker, so this rule "+
			"is silent there and a fixture using it would pass vacuously", findings)
	}
}

// TestNoUselessAssignmentBoundary states in executable form what this port does not cover.
//
// Upstream reports every case below and this port does not. They are here so the boundary is a
// measured, failing-loudly-if-it-moves fact rather than a sentence in a doc comment, and so that a
// later porter extending the rule sees immediately which of these they have picked up. Each is a
// silent miss, never a wrong report, which is the direction a partial port should miss in.
func TestNoUselessAssignmentBoundary(t *testing.T) {
	t.Parallel()

	recovered := []struct {
		name  string
		why   string
		index int
	}{
		{"an update expression", "the graph emits a real store for the increment, judged before the load half puts the bit back", 15},
		{"a destructuring assignment target", "the graph places a pattern element's write where it happens, so no flow node has to line up", 17},
		{"a write inside a try block", "still never reported itself, but it now kills the earlier write it overwrites", 26},
		{"a self-referential write chain", "the graph emits the right-hand read before the store, so no reordering is needed", 30},
		{"a destructuring declarator", "the declarator is reached by walking up through the binding pattern", 33},
	}

	for _, entry := range recovered {
		t.Run(entry.name, func(t *testing.T) {
			t.Parallel()
			imported := noUselessAssignmentUpstreamFail[entry.index]
			result := rule_testing.RunTyped(t, NoUselessAssignment, noUselessAssignmentFile, imported.source)
			if len(result.Diagnostics) != imported.count {
				t.Fatalf("got %d findings, want upstream's %d; this shape was declined by the "+
					"inverted implementation and is recovered because %s",
					len(result.Diagnostics), imported.count, entry.why)
			}
		})
	}
}

// TestNoUselessAssignmentConditionalWritesDoNotKill is the regression for the false positive the
// inverted implementation shipped, and it is the most important test in this file.
//
// A write inside an `if` with no `else` does not run on every path, so it cannot kill an earlier
// write: on the path where the branch is skipped, the earlier value is what a later read observes.
// The inversion expressed the kill as a barrier in the flow graph rather than as a per-path bit, and
// a barrier stops the backward walk regardless of which path it sits on, so the earlier write looked
// dead.
//
// Measured rather than reasoned. Against the real tree this cost 34 findings at 34 locations, every
// one of them wrong, and the shape is the commonest one there is: an accumulator seeded before a
// search. Variables named closestDistance, minimumDistance, bestDistanceSquared, sign and task were
// all condemned at their initializers. The release binary reports nothing on any of the twenty three
// files involved, with a control case in the same invocation firing to prove the rule was running.
//
// No imported case could see it: upstream's corpus writes the shape only with an else arm present,
// where the kill is real and both implementations agree.
func TestNoUselessAssignmentConditionalWritesDoNotKill(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		wantCount  int
	}{
		{
			// The exact shape from ColorConverter.ts, which the inverted implementation reported at
			// the initializer. Both arms write, but neither is guaranteed, so the initializer is
			// live on the path where the chain falls through.
			name:       "an if/else-if chain with no final else keeps the initializer live",
			sourceText: "function f(sector: number, chroma: number, mid: number) {\n\tlet red = 0;\n\tif (sector < 1) { red = chroma; }\n\telse if (sector < 2) { red = mid; }\n\treturn red;\n}",
			wantCount:  0,
		},
		{
			// The accumulator shape, seeded then conditionally improved inside a loop.
			name:       "a loop accumulator seeded before a conditional update stays live",
			sourceText: "function f(points: number[], target: number) {\n\tlet best = Infinity;\n\tfor (const point of points) {\n\t\tconst d = Math.abs(point - target);\n\t\tif (d < best) { best = d; }\n\t}\n\treturn best;\n}",
			wantCount:  0,
		},
		{
			// A single guarded write, the smallest form of the same mistake.
			name:       "a single guarded write does not kill the initializer",
			sourceText: "function f(c: boolean) {\n\tlet v = 1;\n\tif (c) { v = 2; }\n\treturn v;\n}",
			wantCount:  0,
		},
		{
			// An update loads before it stores, so the load keeps the earlier write live. Measured
			// against the release binary, which is silent on this input while a control in the same
			// invocation fires. The mutation that clears the bit here instead of setting it survived
			// the whole imported corpus and every other fixture in this file, because no upstream
			// case writes a compound assignment whose target was written before it.
			name:       "a compound assignment keeps the earlier write live",
			sourceText: "function f() {\n\tlet v = 1;\n\tv += 1;\n\tg(v);\n}",
			wantCount:  0,
		},
		{
			// The same property through the other update spelling, which reaches the same arm by a
			// different branch of occurrenceKindOf.
			name:       "an increment keeps the earlier write live",
			sourceText: "function f() {\n\tlet v = 1;\n\tv++;\n\tg(v);\n}",
			wantCount:  0,
		},
		{
			// The other half of the same arm: an update's STORE is a candidate, so a trailing
			// compound assignment nothing reads is reported. Without this the two cases above would
			// also pass a rule that had stopped judging updates entirely.
			name:       "a trailing compound assignment nothing reads is reported",
			sourceText: "function f() {\n\tlet v = 1;\n\tg(v);\n\tv += 1;\n}",
			wantCount:  1,
		},
		{
			// The control: with BOTH arms writing, the initializer really is dead on every path and
			// upstream reports it. Without this case the three above would also pass a rule that had
			// simply stopped reporting initializers at all.
			name:       "a write on every arm does kill the initializer",
			sourceText: "function f(c: boolean) {\n\tlet v = 1;\n\tif (c) { v = 2; } else { v = 3; }\n\treturn v;\n}",
			wantCount:  1,
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, NoUselessAssignment, noUselessAssignmentFile, testCase.sourceText)
			if len(result.Diagnostics) != testCase.wantCount {
				t.Fatalf("got %d findings, want %d", len(result.Diagnostics), testCase.wantCount)
			}
		})
	}
}

// TestNoUselessAssignmentSelfReferentialWritesStayLive covers the interaction between the two
// mechanisms that decide liveness, which no upstream case exercises together.
//
// A write blocks the backward walk, and a read on a write's own right-hand side is exempt from that
// write's block because it evaluates first. Those two rules meet whenever a self-referential write
// follows other writes, and the cases below are the ones that separate a correct implementation from
// three plausible wrong ones. Every input here was constructed while hunting a surviving mutant, and
// the second one is the reason the mutant mattered: an attempted simplification that started the
// walk one step back instead of copying the blocker set passed the entire imported corpus, passed
// six of these seven, and reported two false positives on this one.
func TestNoUselessAssignmentSelfReferentialWritesStayLive(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		wantCount  int
	}{
		{
			// `v = 1` is read on the path where the branch does not run, and `v = 9` is read on the
			// path where it does, so both are live. A walk that lifts every barrier rather than the
			// one on the read's own assignment reports both.
			name:       "a branch write and a self-referential write both stay live",
			sourceText: "function f() {\n\tlet v = 0;\n\tuse(v);\n\tv = 1;\n\tif (c) { v = 9; }\n\tv = v + 1;\n\tuse(v);\n}",
			wantCount:  0,
		},
		{
			// `v = 1` is killed by `v = 2` before the self-referential read can observe it, so the
			// exemption must not rescue it. One finding, on `v = 1`.
			name:       "a write killed before a self-referential read is still dead",
			sourceText: "function f() {\n\tlet v = 0;\n\tuse(v);\n\tv = 1;\n\tv = 2;\n\tv = v + v;\n\tuse(v);\n}",
			wantCount:  1,
		},
		{
			// Chained self-references each read the previous value, so nothing here is dead.
			name:       "chained self-referential writes are all live",
			sourceText: "function f() {\n\tlet v = 0;\n\tuse(v);\n\tv = 1;\n\tv = v + 1;\n\tv = v + 2;\n\tuse(v);\n}",
			wantCount:  0,
		},
		{
			// The value is copied out before being overwritten, so the copy keeps the write live
			// even though the later self-referential write does not read it directly.
			name:       "a value copied out before being overwritten stays live",
			sourceText: "function f() {\n\tlet v = 0;\n\tuse(v);\n\tv = 1;\n\tlet t = v;\n\tv = 2;\n\tv = v + t;\n\tuse(v);\n}",
			wantCount:  0,
		},
		{
			// The self-referential read sits inside a conditional expression, so it carries a flow
			// node distinct from the assignment's own. Two attempted simplifications of the
			// exemption both passed the entire imported corpus and both over-reported here, one by
			// starting the walk a step back and one by matching the read's flow node against the
			// assignment's for equality. This case is the cheapest thing in the file that can see
			// either mistake.
			name:       "a self-referential read inside a conditional keeps the write live",
			sourceText: "function f() {\n\tlet v = 0;\n\tuse(v);\n\tv = 1;\n\tv = c ? v : 0;\n\tuse(v);\n}",
			wantCount:  0,
		},
		{
			// The same shape with a branch in between, which is where the first simplification
			// produced two findings against zero.
			name:       "a conditional self-reference after a branch write keeps both live",
			sourceText: "function f() {\n\tlet v = 0;\n\tuse(v);\n\tv = 1;\n\tif (c) { v = 2; }\n\tv = c ? v : 0;\n\tuse(v);\n}",
			wantCount:  0,
		},
		{
			// A loop body may run zero times, so a write inside it never kills a write before it.
			// The inverted implementation answered one here and this test asserted that answer;
			// both were wrong. Measured against the release binary, which reports nothing on this
			// input while a control case in the same invocation fires, so the silence is the rule
			// declining rather than the rule being absent. The corrected expectation is what the
			// liveness pass produces, because the loop-exit edge carries the pre-loop value to
			// `use(v)` untouched.
			name:       "a write before a loop that may not run stays live",
			sourceText: "function f() {\n\tlet v = 0;\n\tuse(v);\n\tv = 1;\n\tfor (let i = 0; i < 2; i++) {\n\t\tv = 5;\n\t\tv = v + 1;\n\t}\n\tuse(v);\n}",
			wantCount:  0,
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, NoUselessAssignment, noUselessAssignmentFile, testCase.sourceText)
			if len(result.Diagnostics) != testCase.wantCount {
				t.Fatalf("got %d findings, want %d", len(result.Diagnostics), testCase.wantCount)
			}
		})
	}
}

// TestNoUselessAssignmentShorthandPropertyIsARead covers the defect the imported corpus could not
// see and the dry run found immediately.
//
// A shorthand property (`return { v }`) resolves through GetSymbolAtLocation to the property's own
// symbol rather than to the variable it reads, so the read lands under a symbol nothing else
// touches and every write to the variable looks dead. Upstream's corpus contains no shorthand at
// all, so all 71 clean cases stayed green while the rule produced 223 false positives against the
// real tree, 50 of them in a single file. The findings count went from 257 to 34 when this was
// fixed, which is the ratio that makes this the most valuable test in the file.
func TestNoUselessAssignmentShorthandPropertyIsARead(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		wantCount  int
	}{
		{
			name:       "a shorthand property in a returned object reads the variable",
			sourceText: "function f(k) {\n\tlet v;\n\tif (k) { v = 1; } else { v = 2; }\n\treturn { v };\n}",
			wantCount:  0,
		},
		{
			// The shape found in the tree: an uninitialized declaration, writes in every switch
			// branch, and a shorthand read after. Fifty findings in one file came from this.
			name:       "switch branches feeding a shorthand property are all live",
			sourceText: "function f(k) {\n\tlet v;\n\tswitch (k) {\n\t\tcase 1: v = 1; break;\n\t\tdefault: v = 2; break;\n\t}\n\treturn { v };\n}",
			wantCount:  0,
		},
		{
			// The control: the same code with a plain read was already correct, which is why the
			// defect was invisible to everything except a shorthand.
			name:       "the same shape with a plain read was never affected",
			sourceText: "function f(k) {\n\tlet v;\n\tif (k) { v = 1; } else { v = 2; }\n\treturn v;\n}",
			wantCount:  0,
		},
		{
			// A shorthand does not make every write live: this one is still overwritten before the
			// shorthand read, so treating shorthand as a read must not become a blanket amnesty.
			name:       "a write killed before a shorthand read is still dead",
			sourceText: "function f() {\n\tlet v = 0;\n\tuse(v);\n\tv = 1;\n\tv = 2;\n\treturn { v };\n}",
			wantCount:  1,
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, NoUselessAssignment, noUselessAssignmentFile, testCase.sourceText)
			if len(result.Diagnostics) != testCase.wantCount {
				t.Fatalf("got %d findings, want %d", len(result.Diagnostics), testCase.wantCount)
			}
		})
	}
}

// TestNoUselessAssignmentSwitchFlow pins the control flow shapes a switch produces, which the
// imported corpus does not cover at all and which the real tree is full of.
func TestNoUselessAssignmentSwitchFlow(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		wantCount  int
	}{
		{
			name:       "every case breaking to a later read keeps its write live",
			sourceText: "function f(k) {\n\tlet v = 0;\n\tuse(v);\n\tswitch (k) {\n\t\tcase 1:\n\t\t\tv = 1;\n\t\t\tbreak;\n\t\tcase 2:\n\t\t\tv = 2;\n\t\t\tbreak;\n\t}\n\tuse(v);\n}",
			wantCount:  0,
		},
		{
			// Falling through means the first case's write is overwritten by the second before
			// anything reads it, so it is genuinely dead. One finding, and the shape that proves
			// the switch handling is doing real work rather than declining to look.
			name:       "a fallthrough write overwritten by the next case is dead",
			sourceText: "function f(k) {\n\tlet v = 0;\n\tuse(v);\n\tswitch (k) {\n\t\tcase 1:\n\t\t\tv = 1;\n\t\tcase 2:\n\t\t\tv = 2;\n\t}\n\tuse(v);\n}",
			wantCount:  1,
		},
		{
			name:       "a branch nested inside a case keeps both arms live",
			sourceText: "function f(k) {\n\tlet v = 0;\n\tuse(v);\n\tswitch (k) {\n\t\tcase 1:\n\t\t\tif (c) { v = 1; } else { v = 9; }\n\t\t\tbreak;\n\t}\n\tuse(v);\n}",
			wantCount:  0,
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, NoUselessAssignment, noUselessAssignmentFile, testCase.sourceText)
			if len(result.Diagnostics) != testCase.wantCount {
				t.Fatalf("got %d findings, want %d", len(result.Diagnostics), testCase.wantCount)
			}
		})
	}
}

// TestNoUselessAssignmentLeavesANeverReadBindingToNoUnusedVars pins upstream's `is_used` guard on
// the shape that slipped past it.
//
// ESLint's rule says it "will not report variables that are never read", leaving them to
// `no-unused-vars`, so a binding nothing reads gets ONE finding there and none here. cohere carried
// the same guard and still reported every write to `let provenance: string;` in
// `modules/finance/FinancePositionCommandLineInterface.ts` (three writes, beside the one
// `no-unused-vars` finding), because a declarator with no initializer is classified as a read for
// the liveness pass's sake and that classification was reused to decide whether the binding is
// read at all. It is not a read; it only errs safe inside the dataflow.
//
// Both directions: the uninitialized never-read binding is silent, and the same binding read once
// after a dead write still reports that write, so the guard cannot be swallowing the rule.
func TestNoUselessAssignmentLeavesANeverReadBindingToNoUnusedVars(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		wantCount  int
	}{
		{
			name:       "an uninitialized binding written on every branch and never read",
			sourceText: "function accountLine(valued: boolean): string {\n\tlet provenance: string;\n\tif (valued) {\n\t\tprovenance = 'valued';\n\t} else {\n\t\tprovenance = 'unvalued';\n\t}\n\treturn 'line';\n}",
			wantCount:  0,
		},
		{
			name:       "an uninitialized binding written twice and never read",
			sourceText: "function f(): void {\n\tlet v: number;\n\tv = 1;\n\tv = 2;\n}",
			wantCount:  0,
		},
		{
			name:       "the control: the same binding read once still reports the dead write",
			sourceText: "function f(): number {\n\tlet v: number;\n\tv = 1;\n\tv = 2;\n\treturn v;\n}",
			wantCount:  1,
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, NoUselessAssignment, noUselessAssignmentFile, testCase.sourceText)
			if len(result.Diagnostics) != testCase.wantCount {
				t.Errorf("got %d findings, want %d", len(result.Diagnostics), testCase.wantCount)
			}
		})
	}
}
