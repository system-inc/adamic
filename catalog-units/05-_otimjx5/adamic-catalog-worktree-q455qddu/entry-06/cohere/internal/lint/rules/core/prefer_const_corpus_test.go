package core

import (
	"fmt"
	"sort"
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/internal/lint/testing"
)

/*
 * ESLint 10.8.1's prefer-const, run through the installed rule under @typescript-eslint/parser over
 * its whole test corpus and over the edges this port's own reading raised (#jjfa7qb), with every
 * finding written here where ESLint placed it and the file as ESLint's fix leaves it.
 *
 * The corpus rows are the registry's corpus file, verbatim, less the rows ESLint runs under configured
 * globals. The edge rows were written for this port: an assignment that is not the whole of its
 * statement, a destructuring assignment by every pattern shape, defaults standing where an element
 * goes, rest targets, a pattern's names from every kind of declaration, "all" grouping by the
 * assignment, and the read that moves a finding from the write to the declaration.
 *
 * Six edges are left out because the rule departs from ESLint there on purpose. Five fail toward
 * silence where the declaration ESLint invites would not compile: a pattern that also writes a `var`,
 * a function or an import from the same scope, an outer `let` nested below the pattern's top level,
 * which ESLint reads only at the top, and an array rest that is a property, `[a, ...o.p] = xs`, which
 * ESLint's member test does not descend into. The sixth is an uninitialized `let` written in a namespace, which
 * ESLint leaves since a namespace body is not a block it knows, and which converts there as it does in
 * any block; ESLint reports the initialized one. Their own tests say what the rule does instead.
 *
 * Each finding is its text and byte offset. The fixed column is the whole file after ESLint's fixes,
 * or empty where ESLint offers none.
 */
func TestPreferConstAgreesWithESLint(t *testing.T) {
	t.Parallel()

	decode := rule.DecodeOptionsInto[PreferConstOptions]()
	cases := []struct {
		origin  string
		options string
		code    string
		want    []string
		fixed   string
	}{
		{"corpus", "", "var x = 0;", []string{}, ""},
		{"corpus", "", "let x;", []string{}, ""},
		{"corpus", "", "let x; { x = 0; } foo(x);", []string{}, ""},
		{"corpus", "", "let x = 0; x = 1;", []string{}, ""},
		{"corpus", "", "using resource = fn();", []string{}, ""},
		{"corpus", "", "await using resource = fn();", []string{}, ""},
		{"corpus", "", "const x = 0;", []string{}, ""},
		{"corpus", "", "for (let i = 0, end = 10; i < end; ++i) {}", []string{}, ""},
		{"corpus", "", "for (let i in [1,2,3]) { i = 0; }", []string{}, ""},
		{"corpus", "", "for (let x of [1,2,3]) { x = 0; }", []string{}, ""},
		{"corpus", "", "(function() { var x = 0; })();", []string{}, ""},
		{"corpus", "", "(function() { let x; })();", []string{}, ""},
		{"corpus", "", "(function() { let x; { x = 0; } foo(x); })();", []string{}, ""},
		{"corpus", "", "(function() { let x = 0; x = 1; })();", []string{}, ""},
		{"corpus", "", "(function() { const x = 0; })();", []string{}, ""},
		{"corpus", "", "(function() { for (let i = 0, end = 10; i < end; ++i) {} })();", []string{}, ""},
		{"corpus", "", "(function() { for (let i in [1,2,3]) { i = 0; } })();", []string{}, ""},
		{"corpus", "", "(function() { for (let x of [1,2,3]) { x = 0; } })();", []string{}, ""},
		{"corpus", "", "(function(x = 0) { })();", []string{}, ""},
		{"corpus", "", "let a; while (a = foo());", []string{}, ""},
		{"corpus", "", "let a; do {} while (a = foo());", []string{}, ""},
		{"corpus", "", "let a; for (; a = foo(); );", []string{}, ""},
		{"corpus", "", "let a; for (;; ++a);", []string{}, ""},
		{"corpus", "", "let a; for (const {b = ++a} in foo());", []string{}, ""},
		{"corpus", "", "let a; for (const {b = ++a} of foo());", []string{}, ""},
		{"corpus", "", "let a; for (const x of [1,2,3]) { if (a) {} a = foo(); }", []string{}, ""},
		{"corpus", "", "let a; for (const x of [1,2,3]) { a = a || foo(); bar(a); }", []string{}, ""},
		{"corpus", "", "let a; for (const x of [1,2,3]) { foo(++a); }", []string{}, ""},
		{"corpus", "", "let a; function foo() { if (a) {} a = bar(); }", []string{}, ""},
		{"corpus", "", "let a; function foo() { a = a || bar(); baz(a); }", []string{}, ""},
		{"corpus", "", "let a; function foo() { bar(++a); }", []string{}, ""},
		{"corpus", "", "let id;\nfunction foo() {\n    if (typeof id !== 'undefined') {\n        return;\n    }\n    id = setInterval(() => {}, 250);\n}\nfoo();", []string{}, ""},
		{"corpus", "", "/*exported a*/ let a; function init() { a = foo(); }", []string{}, ""},
		{"corpus", "", "/*exported a*/ let a = 1", []string{"a@19"}, "/*exported a*/ const a = 1"},
		{"corpus", "", "let a; if (true) a = 0; foo(a);", []string{}, ""},
		{"corpus", "", "\n        (function (a) {\n            let b;\n            ({ a, b } = obj);\n        })();\n        ", []string{}, ""},
		{"corpus", "", "\n        (function (a) {\n            let b;\n            ([ a, b ] = obj);\n        })();\n        ", []string{}, ""},
		{"corpus", "", "var a; { var b; ({ a, b } = obj); }", []string{}, ""},
		{"corpus", "", "let a; { let b; ({ a, b } = obj); }", []string{}, ""},
		{"corpus", "", "var a; { var b; ([ a, b ] = obj); }", []string{}, ""},
		{"corpus", "", "let a; { let b; ([ a, b ] = obj); }", []string{}, ""},
		{"corpus", "", "let x; { x = 0; foo(x); }", []string{}, ""},
		{"corpus", "", "(function() { let x; { x = 0; foo(x); } })();", []string{}, ""},
		{"corpus", "", "let x; for (const a of [1,2,3]) { x = foo(); bar(x); }", []string{}, ""},
		{"corpus", "", "(function() { let x; for (const a of [1,2,3]) { x = foo(); bar(x); } })();", []string{}, ""},
		{"corpus", "", "let x; for (x of array) { x; }", []string{}, ""},
		{"corpus", "{\"destructuring\":\"all\"}", "let {a, b} = obj; b = 0;", []string{}, ""},
		{"corpus", "{\"destructuring\":\"all\"}", "let a, b; ({a, b} = obj); b++;", []string{}, ""},
		{"corpus", "{\"destructuring\":\"all\"}", "let { name, ...otherStuff } = obj; otherStuff = {};", []string{}, ""},
		{"corpus", "", "let predicate; [typeNode.returnType, predicate] = foo();", []string{}, ""},
		{"corpus", "", "let predicate; [typeNode.returnType, ...predicate] = foo();", []string{}, ""},
		{"corpus", "", "let predicate; [typeNode.returnType,, predicate] = foo();", []string{}, ""},
		{"corpus", "", "let predicate; [typeNode.returnType=5, predicate] = foo();", []string{}, ""},
		{"corpus", "", "let predicate; [[typeNode.returnType=5], predicate] = foo();", []string{}, ""},
		{"corpus", "", "let predicate; [[typeNode.returnType, predicate]] = foo();", []string{}, ""},
		{"corpus", "", "let predicate; [typeNode.returnType, [predicate]] = foo();", []string{}, ""},
		{"corpus", "", "let predicate; [, [typeNode.returnType, predicate]] = foo();", []string{}, ""},
		{"corpus", "", "let predicate; [, {foo:typeNode.returnType, predicate}] = foo();", []string{}, ""},
		{"corpus", "", "let predicate; [, {foo:typeNode.returnType, ...predicate}] = foo();", []string{}, ""},
		{"corpus", "", "let a; const b = {}; ({ a, c: b.c } = func());", []string{}, ""},
		{"corpus", "{\"ignoreReadBeforeAssign\":true}", "let x; function foo() { bar(x); } x = 0;", []string{}, ""},
		{"corpus", "", "const x = [1,2]; let y; [,y] = x; y = 0;", []string{}, ""},
		{"corpus", "", "const x = [1,2,3]; let y, z; [y,,z] = x; y = 0; z = 0;", []string{}, ""},
		{"corpus", "", "class C { static { let a = 1; a = 2; } }", []string{}, ""},
		{"corpus", "", "class C { static { let a; a = 1; a = 2; } }", []string{}, ""},
		{"corpus", "", "let a; class C { static { a = 1; } }", []string{}, ""},
		{"corpus", "", "class C { static { let a; if (foo) { a = 1; } } }", []string{}, ""},
		{"corpus", "", "class C { static { let a; if (foo) a = 1; } }", []string{}, ""},
		{"corpus", "", "class C { static { let a, b; if (foo) { ({ a, b } = foo); } } }", []string{}, ""},
		{"corpus", "", "class C { static { let a, b; if (foo) ({ a, b } = foo); } }", []string{}, ""},
		{"corpus", "{\"ignoreReadBeforeAssign\":true}", "class C { static { a; } } let a = 1; ", []string{}, ""},
		{"corpus", "{\"ignoreReadBeforeAssign\":true}", "class C { static { () => a; let a = 1; } };", []string{}, ""},
		{"corpus", "", "let x = 1; foo(x);", []string{"x@4"}, "const x = 1; foo(x);"},
		{"corpus", "", "for (let i in [1,2,3]) { foo(i); }", []string{"i@9"}, "for (const i in [1,2,3]) { foo(i); }"},
		{"corpus", "", "for (let x of [1,2,3]) { foo(x); }", []string{"x@9"}, "for (const x of [1,2,3]) { foo(x); }"},
		{"corpus", "", "let [x = -1, y] = [1,2]; y = 0;", []string{"x@5"}, ""},
		{"corpus", "", "let {a: x = -1, b: y} = {a:1,b:2}; y = 0;", []string{"x@8"}, ""},
		{"corpus", "", "(function() { let x = 1; foo(x); })();", []string{"x@18"}, "(function() { const x = 1; foo(x); })();"},
		{"corpus", "", "(function() { for (let i in [1,2,3]) { foo(i); } })();", []string{"i@23"}, "(function() { for (const i in [1,2,3]) { foo(i); } })();"},
		{"corpus", "", "(function() { for (let x of [1,2,3]) { foo(x); } })();", []string{"x@23"}, "(function() { for (const x of [1,2,3]) { foo(x); } })();"},
		{"corpus", "", "(function() { let [x = -1, y] = [1,2]; y = 0; })();", []string{"x@19"}, ""},
		{"corpus", "", "let f = (function() { let g = x; })(); f = 1;", []string{"g@26"}, "let f = (function() { const g = x; })(); f = 1;"},
		{"corpus", "", "(function() { let {a: x = -1, b: y} = {a:1,b:2}; y = 0; })();", []string{"x@22"}, ""},
		{"corpus", "", "let x = 0; { let x = 1; foo(x); } x = 0;", []string{"x@17"}, "let x = 0; { const x = 1; foo(x); } x = 0;"},
		{"corpus", "", "for (let i = 0; i < 10; ++i) { let x = 1; foo(x); }", []string{"x@35"}, "for (let i = 0; i < 10; ++i) { const x = 1; foo(x); }"},
		{"corpus", "", "for (let i in [1,2,3]) { let x = 1; foo(x); }", []string{"i@9", "x@29"}, "for (const i in [1,2,3]) { const x = 1; foo(x); }"},
		{"corpus", "", "var foo = function() {\n    for (const b of c) {\n       let a;\n       a = 1;\n   }\n};", []string{"a@69"}, ""},
		{"corpus", "", "var foo = function() {\n    for (const b of c) {\n       let a;\n       ({a} = 1);\n   }\n};", []string{"a@71"}, ""},
		{"corpus", "", "let x; x = 0;", []string{"x@7"}, ""},
		{"corpus", "", "switch (a) { case 0: let x; x = 0; }", []string{"x@28"}, ""},
		{"corpus", "", "(function() { let x; x = 1; })();", []string{"x@21"}, ""},
		{"corpus", "{\"destructuring\":\"any\"}", "let {a = 0, b} = obj; b = 0; foo(a, b);", []string{"a@5"}, ""},
		{"corpus", "{\"destructuring\":\"any\"}", "let {a: {b, c}} = {a: {b: 1, c: 2}}; b = 3;", []string{"c@12"}, ""},
		{"corpus", "{\"destructuring\":\"all\"}", "let {a: {b, c}} = {a: {b: 1, c: 2}}", []string{"b@9", "c@12"}, "const {a: {b, c}} = {a: {b: 1, c: 2}}"},
		{"corpus", "{\"destructuring\":\"any\"}", "let a, b; ({a = 0, b} = obj); b = 0; foo(a, b);", []string{"a@12"}, ""},
		{"corpus", "{\"destructuring\":\"all\"}", "let {a = 0, b} = obj; foo(a, b);", []string{"a@5", "b@12"}, "const {a = 0, b} = obj; foo(a, b);"},
		{"corpus", "", "let [a] = [1]", []string{"a@5"}, "const [a] = [1]"},
		{"corpus", "", "let {a} = obj", []string{"a@5"}, "const {a} = obj"},
		{"corpus", "{\"destructuring\":\"all\"}", "let a, b; ({a = 0, b} = obj); foo(a, b);", []string{"a@12", "b@19"}, ""},
		{"corpus", "{\"destructuring\":\"any\"}", "let {a = 0, b} = obj, c = a; b = a;", []string{"a@5", "c@22"}, ""},
		{"corpus", "{\"destructuring\":\"all\"}", "let {a = 0, b} = obj, c = a; b = a;", []string{"c@22"}, ""},
		{"corpus", "{\"destructuring\":\"any\"}", "let { name, ...otherStuff } = obj; otherStuff = {};", []string{"name@6"}, ""},
		{"corpus", "", "let x; function foo() { bar(x); } x = 0;", []string{"x@4"}, ""},
		{"corpus", "", "/*eslint custom/use-x:error*/ let x = 1", []string{"x@34"}, "/*eslint custom/use-x:error*/ const x = 1"},
		{"corpus", "", "/*eslint custom/use-x:error*/ { let x = 1 }", []string{"x@36"}, "/*eslint custom/use-x:error*/ { const x = 1 }"},
		{"corpus", "", "let { foo, bar } = baz;", []string{"foo@6", "bar@11"}, "const { foo, bar } = baz;"},
		{"corpus", "", "const x = [1,2]; let [,y] = x;", []string{"y@23"}, "const x = [1,2]; const [,y] = x;"},
		{"corpus", "", "const x = [1,2,3]; let [y,,z] = x;", []string{"y@24", "z@27"}, "const x = [1,2,3]; const [y,,z] = x;"},
		{"corpus", "", "let predicate; [, {foo:returnType, predicate}] = foo();", []string{"predicate@35"}, ""},
		{"corpus", "", "let predicate; [, {foo:returnType, predicate}, ...bar ] = foo();", []string{"predicate@35"}, ""},
		{"corpus", "", "let predicate; [, {foo:returnType, ...predicate} ] = foo();", []string{"predicate@38"}, ""},
		{"corpus", "", "let x = 'x', y = 'y';", []string{"x@4", "y@13"}, "const x = 'x', y = 'y';"},
		{"corpus", "", "let x = 'x', y = 'y'; x = 1", []string{"y@13"}, ""},
		{"corpus", "", "let x = 1, y = 'y'; let z = 1;", []string{"x@4", "y@11", "z@24"}, "const x = 1, y = 'y'; const z = 1;"},
		{"corpus", "", "let { a, b, c} = obj; let { x, y, z} = anotherObj; x = 2;", []string{"a@6", "b@9", "c@12", "y@31", "z@34"}, "const { a, b, c} = obj; let { x, y, z} = anotherObj; x = 2;"},
		{"corpus", "", "let x = 'x', y = 'y'; function someFunc() { let a = 1, b = 2; foo(a, b) }", []string{"x@4", "y@13", "a@48", "b@55"}, "const x = 'x', y = 'y'; function someFunc() { const a = 1, b = 2; foo(a, b) }"},
		{"corpus", "", "let someFunc = () => { let a = 1, b = 2; foo(a, b) }", []string{"someFunc@4", "a@27", "b@34"}, "const someFunc = () => { const a = 1, b = 2; foo(a, b) }"},
		{"corpus", "", "let {a, b} = c, d;", []string{"a@5", "b@8"}, ""},
		{"corpus", "", "let {a, b, c} = {}, e, f;", []string{"a@5", "b@8", "c@11"}, ""},
		{"corpus", "", "function a() {\nlet foo = 0,\n  bar = 1;\nfoo = 1;\n}\nfunction b() {\nlet foo = 0,\n  bar = 2;\nfoo = 2;\n}", []string{"bar@30", "bar@80"}, ""},
		{"corpus", "", "/*eslint no-undef-init:error*/ let foo = undefined;", []string{"foo@35"}, "/*eslint no-undef-init:error*/ const foo = undefined;"},
		{"corpus", "", "let a = 1; class C { static { a; } }", []string{"a@4"}, "const a = 1; class C { static { a; } }"},
		{"corpus", "", "class C { static { a; } } let a = 1;", []string{"a@30"}, "class C { static { a; } } const a = 1;"},
		{"corpus", "", "class C { static { let a = 1; } }", []string{"a@23"}, "class C { static { const a = 1; } }"},
		{"corpus", "", "class C { static { if (foo) { let a = 1; } } }", []string{"a@34"}, "class C { static { if (foo) { const a = 1; } } }"},
		{"corpus", "", "class C { static { let a = 1; if (foo) { a; } } }", []string{"a@23"}, "class C { static { const a = 1; if (foo) { a; } } }"},
		{"corpus", "", "class C { static { if (foo) { let a; a = 1; } } }", []string{"a@37"}, ""},
		{"corpus", "", "class C { static { let a; a = 1; } }", []string{"a@26"}, ""},
		{"corpus", "", "class C { static { let { a, b } = foo; } }", []string{"a@25", "b@28"}, "class C { static { const { a, b } = foo; } }"},
		{"corpus", "", "class C { static { let a, b; ({ a, b } = foo); } }", []string{"a@32", "b@35"}, ""},
		{"corpus", "", "class C { static { let a; let b; ({ a, b } = foo); } }", []string{"a@36", "b@39"}, ""},
		{"corpus", "", "class C { static { let a; a = 0; console.log(a); } }", []string{"a@26"}, ""},
		{"corpus", "{\"destructuring\":\"any\",\"ignoreReadBeforeAssign\":true}", "\n            let { itemId, list } = {},\n            obj = [],\n            total = 0;\n            total = 9;\n            console.log(itemId, list, obj, total);\n            ", []string{"itemId@19", "list@27", "obj@52"}, ""},
		{"corpus", "{\"destructuring\":\"any\",\"ignoreReadBeforeAssign\":true}", "\n            let { itemId, list } = {},\n            obj = [];\n            console.log(itemId, list, obj);\n            ", []string{"itemId@19", "list@27", "obj@52"}, "\n            const { itemId, list } = {},\n            obj = [];\n            console.log(itemId, list, obj);\n            "},
		{"corpus", "{\"destructuring\":\"any\",\"ignoreReadBeforeAssign\":true}", "\n            let [ itemId, list ] = [],\n            total = 0;\n            total = 9;\n            console.log(itemId, list, total);\n            ", []string{"itemId@19", "list@27"}, ""},
		{"corpus", "{\"destructuring\":\"any\",\"ignoreReadBeforeAssign\":true}", "\n            let [ itemId, list ] = [],\n            obj = [];\n            console.log(itemId, list, obj);\n            ", []string{"itemId@19", "list@27", "obj@52"}, "\n            const [ itemId, list ] = [],\n            obj = [];\n            console.log(itemId, list, obj);\n            "},
		{"edge", "", "let x; foo() || (x = 0);", []string{}, ""},
		{"edge", "", "let x; foo() && (x = 0); bar(x);", []string{}, ""},
		{"edge", "", "let x; y = x = 0;", []string{}, ""},
		{"edge", "", "let x, y; x = y = 0;", []string{"x@10"}, ""},
		{"edge", "", "let x; foo(), x = 0;", []string{}, ""},
		{"edge", "", "let x; label: x = 0;", []string{}, ""},
		{"edge", "", "let x; (x = 0);", []string{"x@8"}, ""},
		{"edge", "", "let x; ((x)) = 0;", []string{"x@9"}, ""},
		{"edge", "", "let x; foo(x = 0);", []string{}, ""},
		{"edge", "", "let a: number; a = 1;", []string{"a@15"}, ""},
		{"edge", "", "let a; [a = 0] = xs;", []string{"a@8"}, ""},
		{"edge", "", "let a; ({k: a = 0} = o);", []string{"a@12"}, ""},
		{"edge", "", "let a; ({a = 1} = o);", []string{"a@9"}, ""},
		{"edge", "", "let a; [a] = [1]; foo(a);", []string{"a@8"}, ""},
		{"edge", "", "let a; foo([a = 0]);", []string{}, ""},
		{"edge", "", "let a; [a = 0];", []string{}, ""},
		{"edge", "", "let a; for ({a} of xs) {}", []string{}, ""},
		{"edge", "", "let a; for ([a] of xs) {}", []string{}, ""},
		{"edge", "", "let a; ({a} = obj);", []string{"a@9"}, ""},
		{"edge", "", "let a; [a] = xs;", []string{"a@8"}, ""},
		{"edge", "", "let a, b; ({a, b} = obj);", []string{"a@12", "b@15"}, ""},
		{"edge", "", "let a; let b; [a, b] = xs;", []string{"a@15", "b@18"}, ""},
		{"edge", "", "let a, b; [a, [b]] = xs;", []string{"a@11", "b@15"}, ""},
		{"edge", "", "for (const b of c) { let a; ({a} = b); }", []string{"a@30"}, ""},
		{"edge", "", "let w; [...w] = [];", []string{"w@11"}, ""},
		{"edge", "", "let w; ({...w} = {});", []string{"w@12"}, ""},
		{"edge", "", "let w; [a, ...w] = [];", []string{"w@14"}, ""},
		{"edge", "", "let w; ({a, ...w} = {});", []string{"w@15"}, ""},
		{"edge", "", "let w; ({ x: [...w] } = {});", []string{"w@17"}, ""},
		{"edge", "", "let w; [...(w)] = [];", []string{"w@12"}, ""},
		{"edge", "", "let w; ({...(w)} = {});", []string{"w@13"}, ""},
		{"edge", "", "let w; [, [...w]] = [];", []string{"w@14"}, ""},
		{"edge", "", "let w = []; [...w] = [];", []string{}, ""},
		{"edge", "", "let w = []; for ([...w] of pairs) {}", []string{}, ""},
		{"edge", "", "let o; { let a; [o, a] = foo(); }", []string{}, ""},
		{"edge", "", "let a; [a = 0, o.p] = foo();", []string{}, ""},
		{"edge", "", "let a; { [a] = xs; }", []string{}, ""},
		{"edge", "", "let a; if (c) [a] = xs;", []string{}, ""},
		{"edge", "", "let a; const b = [a] = xs;", []string{}, ""},
		{"edge", "", "let a, b; ({a, b} = obj); b = 0;", []string{"a@12"}, ""},
		{"edge", "{\"destructuring\":\"all\"}", "let a, b; ({a, b} = obj); b = 0;", []string{}, ""},
		{"edge", "{\"destructuring\":\"all\"}", "let a; let b = 1; [a, b] = xs;", []string{}, ""},
		{"edge", "", "let a; let b = 1; [a, b] = xs;", []string{"a@19"}, ""},
		{"edge", "{\"destructuring\":\"all\"}", "let a, b; ({a, b} = obj);", []string{"a@12", "b@15"}, ""},
		{"edge", "{\"destructuring\":\"all\"}", "let a, b; [a, [b]] = xs;", []string{"a@11", "b@15"}, ""},
		{"edge", "{\"destructuring\":\"all\"}", "let a; [a, u] = xs;", []string{"a@8"}, ""},
		{"edge", "", "let x; foo(x); x = 0;", []string{"x@4"}, ""},
		{"edge", "", "let x; function f() { x; } x = 0;", []string{"x@4"}, ""},
		{"edge", "", "let x; x = 0; foo(x);", []string{"x@7"}, ""},
		{"edge", "{\"ignoreReadBeforeAssign\":true}", "let x; x = 0;", []string{"x@7"}, ""},
		{"edge", "{\"ignoreReadBeforeAssign\":true}", "(function() { let x; x = 1; })();", []string{"x@21"}, ""},
		{"edge", "{\"ignoreReadBeforeAssign\":true}", "let x; foo(x); x = 0;", []string{}, ""},
		{"edge", "", "namespace N { let x = 0; foo(x); }", []string{"x@18"}, "namespace N { const x = 0; foo(x); }"},
		{"edge", "", "let x; x = 0; x;", []string{"x@7"}, ""},
		{"edge", "", "let a; ({a, ...o.p} = obj);", []string{}, ""},
		{"edge", "", "let a; [a, escape] = xs;", []string{}, ""},
		{"edge", "", "let a; ({ a } = obj as any);", []string{"a@10"}, ""},
	}

	for index, testCase := range cases {
		var options any
		if testCase.options != "" {
			decoded, err := decode([]byte(testCase.options))
			if err != nil {
				t.Fatalf("case %d (%s): ESLint accepts %s and the decoder refused it: %v", index, testCase.origin, testCase.options, err)
			}
			options = decoded
		}
		result := rule_testing.RunTypedVerbatimWithOptions(t, PreferConst, "input.ts", testCase.code, options)
		source := result.SourceFile.Text()
		got := []string{}
		type pending struct {
			start, end int
			text       string
		}
		fixes := []pending{}
		for _, diagnostic := range result.Diagnostics {
			got = append(got, fmt.Sprintf("%s@%d", source[diagnostic.Range.Pos():diagnostic.Range.End()], diagnostic.Range.Pos()))
			for _, fix := range diagnostic.Fixes {
				fixes = append(fixes, pending{fix.Range.Pos(), fix.Range.End(), fix.Text})
			}
		}
		want := append([]string{}, testCase.want...)
		sort.Strings(got)
		sort.Strings(want)
		if strings.Join(got, "\n") != strings.Join(want, "\n") {
			t.Errorf("case %d (%s, %s): %q\n got: %q\nwant: %q", index, testCase.origin, testCase.options, testCase.code, got, want)
			continue
		}

		fixed := ""
		if len(fixes) > 0 {
			sort.Slice(fixes, func(first, second int) bool { return fixes[first].start > fixes[second].start })
			fixed = source
			for _, fix := range fixes {
				fixed = fixed[:fix.start] + fix.text + fixed[fix.end:]
			}
		}
		if fixed != testCase.fixed {
			t.Errorf("case %d (%s, %s): %q\nfixed to: %q\n  ESLint: %q", index, testCase.origin, testCase.options, testCase.code, fixed, testCase.fixed)
		}
	}
}
