package core

// Generated from ESLint's own no-loop-func corpus and verified byte for byte against it.
//
// The corpus was loaded by running upstream's tester file with its RuleTester stubbed, so the cases
// are the file's own values rather than retyped, and each was then driven through the installed
// eslint 10.8.1 build to record what it reports.
//
// Two testers run in that file. The second supplies @typescript-eslint/parser as its base
// languageOptions, and its 12 cases are marked so a reader can tell which half a case came from.
//
// Four cases were fatal parse errors under the ecmaVersion upstream gave them, all of them
// `using` or `await using` at the top level of a script. Upstream lists them as valid, but the
// parser is what makes them clean there rather than the rule. Re-measured through a parser that
// accepts the syntax, so the RULE decides: all four report nothing, which is the same verdict for
// the reason upstream intended.

// loopFuncFiringCases are the corpus inputs upstream reports on, with the rendered message it
// produced for each. The message names the unsafe variables, so the text is the assertion rather
// than the id.
var loopFuncFiringCases = []struct {
	source          string
	unsafeVariables string
	typeScriptOnly  bool
}{
	{"for (var i=0; i<l; i++) { (function() { i; }) }", "'i'", false},
	{"for (var i=0; i<l; i++) { for (var j=0; j<m; j++) { (function() { i+j; }) } }", "'i', 'j'", false},
	{"for (var i in {}) { (function() { i; }) }", "'i'", false},
	{"for (var i of {}) { (function() { i; }) }", "'i'", false},
	{"for (var i=0; i < l; i++) { (() => { i; }) }", "'i'", false},
	{"for (var i=0; i < l; i++) { var a = function() { i; } }", "'i'", false},
	{"for (var i=0; i < l; i++) { function a() { i; }; a(); }", "'i'", false},
	{"let a; for (let i=0; i<l; i++) { a = 1; (function() { a; });}", "'a'", false},
	{"let a; for (let i in {}) { (function() { a; }); a = 1; }", "'a'", false},
	{"let a; for (let i of {}) { (function() { a; }); } a = 1; ", "'a'", false},
	{"let a; for (let i=0; i<l; i++) { (function() { (function() { a; }); }); a = 1; }", "'a'", false},
	{"let a; for (let i in {}) { a = 1; function foo() { (function() { a; }); } }", "'a'", false},
	{"let a; for (let i of {}) { (() => { (function() { a; }); }); } a = 1;", "'a'", false},
	{"for (var i = 0; i < 10; ++i) { for (let x in xs.filter(x => x != i)) {  } }", "'i'", false},
	{"for (let x of xs) { let a; for (let y of ys) { a = 1; (function() { a; }); } }", "'a'", false},
	{"for (var x of xs) { for (let y of ys) { (function() { x; }); } }", "'x'", false},
	{"for (var x of xs) { (function() { x; }); }", "'x'", false},
	{"var a; for (let x of xs) { a = 1; (function() { a; }); }", "'a'", false},
	{"var a; for (let x of xs) { (function() { a; }); a = 1; }", "'a'", false},
	{"let a; function foo() { a = 10; } for (let x of xs) { (function() { a; }); } foo();", "'a'", false},
	{"let a; function foo() { a = 10; for (let x of xs) { (function() { a; }); } } foo();", "'a'", false},
	{"let a; for (var i=0; i<l; i++) { (function* (){i;})() }", "'i'", false},
	{"let a; for (var i=0; i<l; i++) { (async function (){i;})() }", "'i'", false},
	{"\n            let current = getStart();\n            const arr = [];\n            while (current) {\n                (function f() {\n                    current;\n                    arr.push(f);\n                })();\n\n                current = current.upper;\n            }\n            ", "'current'", false},
	{"\n            var arr = [];\n\n            for (var i = 0; i < 5; i++) {\n                (function fun () {\n                    if (arr.includes(fun)) return i;\n                    else arr.push(fun);\n                })();\n            }\n            ", "'i'", false},
	{"\n            let current = getStart();\n            const arr = [];\n            while (current) {\n                const p = (async () => {\n                    await someDelay();\n                    current;\n                })();\n\n                arr.push(p);\n                current = current.upper;\n            }\n            ", "'current'", false},
	{"\n            var arr = [];\n\n            for (var i = 0; i < 5; i++) {\n                arr.push((f => f)(\n                    () => i\n                ));\n            }\n            ", "'i'", false},
	{"\n            var arr = [];\n\n            for (var i = 0; i < 5; i++) {\n                arr.push((() => {\n                    return () => i;\n                })());\n            }\n            ", "'i'", false},
	{"\n            var arr = [];\n\n            for (var i = 0; i < 5; i++) {\n                arr.push((() => {\n                    return () => { return i };\n                })());\n            }\n            ", "'i'", false},
	{"\n            var arr = [];\n\n            for (var i = 0; i < 5; i++) {\n                arr.push((() => {\n                    return () => {\n                        return () => i\n                    };\n                })());\n            }\n            ", "'i'", false},
	{"\n            var arr = [];\n\n            for (var i = 0; i < 5; i++) {\n                arr.push((() => {\n                    return () =>\n                        (() => i)();\n                })());\n            }\n            ", "'i'", false},
	{"\n            var arr = [];\n\n            for (var i = 0; i < 5; i ++) {\n                (() => {\n                    arr.push((async () => {\n                        await 1;\n                        return i;\n                    })());\n                })();\n            }\n            ", "'i'", false},
	{"\n            var arr = [];\n\n            for (var i = 0; i < 5; i ++) {\n                (() => {\n                    (function f() {\n                        if (!arr.includes(f)) {\n                            arr.push(f);\n                        }\n                        return i;\n                    })();\n                })();\n\n            }\n            ", "'i'", false},
	{"\n            var arr1 = [], arr2 = [];\n\n            for (var [i, j] of [\"a\", \"b\", \"c\"].entries()) {\n                (() => {\n                    arr1.push((() => i)());\n                    arr2.push(() => j);\n                })();\n            }\n            ", "'j'", false},
	{"\n            var arr = [];\n\n            for (var i = 0; i < 5; i ++) {\n                ((f) => {\n                    arr.push(f);\n                })(() => {\n                    return (() => i)();\n                });\n\n            }\n            ", "'i'", false},
	{"\n            for (var i = 0; i < 5; i++) {\n                (async () => {\n                    () => i;\n                })();\n            }\n            ", "'i'", false},
	{"\n            for (var i = 0; i < 10; i++) {\n\t\t\t\titems.push({\n\t\t\t\t\tid: i,\n\t\t\t\t\tname: \"Item \" + i\n\t\t\t\t});\n\n\t\t\t\tconst process = function (callback){\n\t\t\t\t\tcallback({ id: i, name: \"Item \" + i });\n\t\t\t\t};\n\t\t\t}\n            ", "'i'", false},
	{"\n  for (var i = 0; i < 10; i++) {\n    function foo() {\n      console.log(i);\n    }\n  }\n\t\t\t", "'i'", true},
	{"\n  for (var i = 0; i < 10; i++) {\n    const handler = (event: Event) => {\n      console.log(i);\n    };\n  }\n\t\t\t", "'i'", true},
	{"\n  interface Item {\n    id: number;\n    name: string;\n  }\n\n  const items: Item[] = [];\n  for (var i = 0; i < 10; i++) {\n    items.push({\n      id: i,\n      name: \"Item \" + i\n    });\n\n    const process = function(callback: (item: Item) => void): void {\n      callback({ id: i, name: \"Item \" + i });\n    };\n  }\n\t\t\t", "'i'", true},
	{"\n  type Processor<T> = (item: T) => void;\n\n  for (var i = 0; i < 10; i++) {\n    const processor: Processor<number> = (item) => {\n      return item + i;\n    };\n  }\n\t\t\t", "'i'", true},
	{"\n      for (var i = 0; i < 10; i++) {\n        // UnconfiguredGlobalType is not defined anywhere\n        // But the function still references i which makes it unsafe\n        const process = (item: UnconfiguredGlobalType) => {\n          console.log(i, item.value);\n        };\n      }\n      ", "'i'", true},
}

// loopFuncCleanCases are the corpus inputs upstream stays silent on. These are the false positives
// upstream already thought about, and they carry the whole discrimination: an undeclared name, a
// constant binding, a `let` per iteration, a variable nothing writes, and an immediately invoked
// function are each represented.
var loopFuncCleanCases = []struct {
	source         string
	typeScriptOnly bool
}{
	{"string = 'function a() {}';", false},
	{"for (var i=0; i<l; i++) { } var a = function() { i; };", false},
	{"for (var i=0, a=function() { i; }; i<l; i++) { }", false},
	{"for (var x in xs.filter(function(x) { return x != upper; })) { }", false},
	{"for (var x of xs.filter(function(x) { return x != upper; })) { }", false},
	{"for (var i=0; i<l; i++) { (function() {}) }", false},
	{"for (var i in {}) { (function() {}) }", false},
	{"for (var i of {}) { (function() {}) }", false},
	{"for (let i=0; i<l; i++) { (function() { i; }) }", false},
	{"for (let i in {}) { i = 7; (function() { i; }) }", false},
	{"for (const i of {}) { (function() { i; }) }", false},
	{"for (using i of foo) { (function() { i; }) }", false},
	{"for (await using i of foo) { (function() { i; }) }", false},
	{"for (var i = 0; i < 10; ++i) { using foo = bar(i); (function() { foo; }) }", false},
	{"for (var i = 0; i < 10; ++i) { await using foo = bar(i); (function() { foo; }) }", false},
	{"for (let i = 0; i < 10; ++i) { for (let x in xs.filter(x => x != i)) {  } }", false},
	{"let a = 0; for (let i=0; i<l; i++) { (function() { a; }); }", false},
	{"let a = 0; for (let i in {}) { (function() { a; }); }", false},
	{"let a = 0; for (let i of {}) { (function() { a; }); }", false},
	{"let a = 0; for (let i=0; i<l; i++) { (function() { (function() { a; }); }); }", false},
	{"let a = 0; for (let i in {}) { function foo() { (function() { a; }); } }", false},
	{"let a = 0; for (let i of {}) { (() => { (function() { a; }); }); }", false},
	{"var a = 0; for (let i=0; i<l; i++) { (function() { a; }); }", false},
	{"var a = 0; for (let i in {}) { (function() { a; }); }", false},
	{"var a = 0; for (let i of {}) { (function() { a; }); }", false},
	{"let result = {};\nfor (const score in scores) {\n  const letters = scores[score];\n  letters.split('').forEach(letter => {\n    result[letter] = score;\n  });\n}\nresult.__default = 6;", false},
	{"while (true) {\n    (function() { a; });\n}\nlet a;", false},
	{"while(i) { (function() { i; }) }", false},
	{"do { (function() { i; }) } while (i)", false},
	{"var i; while(i) { (function() { i; }) }", false},
	{"var i; do { (function() { i; }) } while (i)", false},
	{"for (var i=0; i<l; i++) { (function() { undeclared; }) }", false},
	{"for (let i=0; i<l; i++) { (function() { undeclared; }) }", false},
	{"for (var i in {}) { i = 7; (function() { undeclared; }) }", false},
	{"for (let i in {}) { i = 7; (function() { undeclared; }) }", false},
	{"for (const i of {}) { (function() { undeclared; }) }", false},
	{"for (let i = 0; i < 10; ++i) { for (let x in xs.filter(x => x != undeclared)) {  } }", false},
	{"\n            let current = getStart();\n            while (current) {\n            (() => {\n                current;\n                current.a;\n                current.b;\n                current.c;\n                current.d;\n            })();\n\n            current = current.upper;\n            }\n            ", false},
	{"for (var i=0; (function() { i; })(), i<l; i++) { }", false},
	{"for (var i=0; i<l; (function() { i; })(), i++) { }", false},
	{"for (var i = 0; i < 10; ++i) { (()=>{ i;})() }", false},
	{"for (var i = 0; i < 10; ++i) { (function a(){i;})() }", false},
	{"\n            var arr = [];\n\n            for (var i = 0; i < 5; i++) {\n                arr.push((f => f)((() => i)()));\n            }\n            ", false},
	{"\n            var arr = [];\n\n            for (var i = 0; i < 5; i++) {\n                arr.push((() => {\n                    return (() => i)();\n                })());\n            }\n            ", false},
	{"\n            const foo = bar;\n\n            for (var i = 0; i < 5; i++) {\n                arr.push(() => foo);\n            }\n\n\t\t\tfoo = baz; // This is a runtime error, but not concern of this rule. For this rule, variable 'foo' is constant.\n            ", false},
	{"\n            using foo = bar;\n\n            for (var i = 0; i < 5; i++) {\n                arr.push(() => foo);\n            }\n\n\t\t\tfoo = baz; // This is a runtime error, but not concern of this rule. For this rule, variable 'foo' is constant.\n            ", false},
	{"\n            await using foo = bar;\n\n            for (var i = 0; i < 5; i++) {\n                arr.push(() => foo);\n            }\n\n\t\t\tfoo = baz; // This is a runtime error, but not concern of this rule. For this rule, variable 'foo' is constant.\n            ", false},
	{"\n  for (let i = 0; i < 10; i++) {\n\tfunction foo() {\n\t  console.log('A');\n\t}\n  }\n\t  ", true},
	{"\n  let someArray: MyType[] = [];\n  for (let i = 0; i < 10; i += 1) {\n\tsomeArray = someArray.filter((item: MyType) => !!item);\n  }\n\t  ", true},
	{"\n  let someArray: MyType[] = [];\n  for (let i = 0; i < 10; i += 1) {\n\tsomeArray = someArray.filter((item: MyType) => !!item);\n  }\n\t\t", true},
	{"\n  let someArray: MyType[] = [];\n  for (let i = 0; i < 10; i += 1) {\n\tsomeArray = someArray.filter((item: MyType) => !!item);\n  }\n\t\t", true},
	{"\n  type MyType = 1;\n  let someArray: MyType[] = [];\n  for (let i = 0; i < 10; i += 1) {\n\tsomeArray = someArray.filter((item: MyType) => !!item);\n  }\n\t  ", true},
	{"\n    // UnconfiguredGlobalType is not defined anywhere or configured in globals\n    for (var i = 0; i < 10; i++) {\n      const process = (item: UnconfiguredGlobalType) => {\n        // This is valid because the type reference is considered safe\n        // even though UnconfiguredGlobalType is not configured\n        return item.id;\n      };\n    }\n    ", true},
	{"\n    for (var i = 0; i < 10; i++) {\n      // ConfiguredType is in globals, UnconfiguredType is not\n      // Both should be considered safe as they are type references\n      const process = (configItem: ConfiguredType, unconfigItem: UnconfiguredType) => {\n        return {\n          config: configItem.value,\n          unconfig: unconfigItem.value\n        };\n      };\n    }\n      ", true},
}
