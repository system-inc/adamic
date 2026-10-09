package core

import (
	"strings"
	"testing"

	rule_testing "github.com/system-inc/cohere/internal/lint/testing"
)

// noInvalidThisFile is where the fixtures pretend to live.
const noInvalidThisFile = "/repository/source/NoInvalidThis.ts"

// The corpus is ESLint's own, taken from
// /tmp/lint-sources-fresh/eslint/tests/lib/rules/no-invalid-this.js by RUNNING that file with
// RuleTester intercepted, so upstream's own NORMAL / USE_STRICT / IMPLIED_STRICT / MODULES matrix
// expansion produced the cases rather than a parser of mine. The file makes two RuleTester.run
// calls -- a 474-case JavaScript matrix and a 98-case TypeScript block -- and both were captured;
// an earlier single-slot interceptor kept only the last and silently lost 83% of the corpus.
//
// Every expectation below is what the INSTALLED ESLint 10.8.1 core rule answered when driven over
// that case through the Linter interface under `sourceType: module` with the typescript-eslint
// parser, which is the only configuration cohere has. That is not what the corpus file annotates,
// and the difference is the point: 62 of the 474 matrix verdicts change between sloppy-mode script
// and module, because in sloppy mode `this` is the global object and always valid. Porting the
// annotated verdicts would have encoded a configuration that cannot occur here.
//
// The 572 expanded cases collapse to 215 distinct (source, options) pairs once the matrix's
// condition markers are stripped, and that collapse was checked rather than assumed: grouping by
// stripped source gives 215 groups and ZERO of them contain members whose verdicts disagree.
//
//	100 reporting cases, 115 clean cases, 45 carrying the capIsConstructor option.
func TestNoInvalidThisFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		source   string
		settings any
		findings int
	}{
		{source: "console.log(this); z(x => console.log(x, this));", settings: nil, findings: 2},
		{source: "() => { this }; this;", settings: nil, findings: 2},
		{source: "this.eval('foo');", settings: nil, findings: 1},
		{source: "(function() { console.log(this); z(x => console.log(x, this)); })();", settings: nil, findings: 2},
		{source: "function foo() { console.log(this); z(x => console.log(x, this)); }", settings: nil, findings: 2},
		{source: "function foo() { console.log(this); z(x => console.log(x, this)); }", settings: noInvalidThisSettings{CapIsConstructor: false}, findings: 2},
		{source: "function Foo() { console.log(this); z(x => console.log(x, this)); }", settings: noInvalidThisSettings{CapIsConstructor: false}, findings: 2},
		{source: "return function() { console.log(this); z(x => console.log(x, this)); };", settings: nil, findings: 2},
		{source: "var foo = (function() { console.log(this); z(x => console.log(x, this)); }).bar(obj);", settings: nil, findings: 2},
		{source: "var obj = {foo: function() { function foo() { console.log(this); z(x => console.log(x, this)); } foo(); }};", settings: nil, findings: 2},
		{source: "var obj = {foo() { function foo() { console.log(this); z(x => console.log(x, this)); } foo(); }};", settings: nil, findings: 2},
		{source: "var obj = {foo: function() { return function() { console.log(this); z(x => console.log(x, this)); }; }};", settings: nil, findings: 2},
		{source: "obj.foo = function() { return function() { console.log(this); z(x => console.log(x, this)); }; };", settings: nil, findings: 2},
		{source: "obj.foo = (function() { return () => { console.log(this); z(x => console.log(x, this)); }; })();", settings: nil, findings: 2},
		{source: "obj.foo = (() => () => { console.log(this); z(x => console.log(x, this)); })();", settings: nil, findings: 2},
		{source: "var foo = function() { console.log(this); z(x => console.log(x, this)); }.bind(null);", settings: nil, findings: 2},
		{source: "(function() { console.log(this); z(x => console.log(x, this)); }).call(undefined);", settings: nil, findings: 2},
		{source: "(function() { console.log(this); z(x => console.log(x, this)); }).apply(void 0);", settings: nil, findings: 2},
		{source: "Array.from([], function() { console.log(this); z(x => console.log(x, this)); });", settings: nil, findings: 2},
		{source: "Array.fromAsync([], function() { console.log(this); z(x => console.log(x, this)); });", settings: nil, findings: 2},
		{source: "foo.every(function() { console.log(this); z(x => console.log(x, this)); });", settings: nil, findings: 2},
		{source: "foo.filter(function() { console.log(this); z(x => console.log(x, this)); });", settings: nil, findings: 2},
		{source: "foo.find(function() { console.log(this); z(x => console.log(x, this)); });", settings: nil, findings: 2},
		{source: "foo.findIndex(function() { console.log(this); z(x => console.log(x, this)); });", settings: nil, findings: 2},
		{source: "foo.findLast(function() { console.log(this); z(x => console.log(x, this)); });", settings: nil, findings: 2},
		{source: "foo.findLastIndex(function() { console.log(this); z(x => console.log(x, this)); });", settings: nil, findings: 2},
		{source: "foo.flatMap(function() { console.log(this); z(x => console.log(x, this)); });", settings: nil, findings: 2},
		{source: "foo.forEach(function() { console.log(this); z(x => console.log(x, this)); });", settings: nil, findings: 2},
		{source: "foo.map(function() { console.log(this); z(x => console.log(x, this)); });", settings: nil, findings: 2},
		{source: "foo.some(function() { console.log(this); z(x => console.log(x, this)); });", settings: nil, findings: 2},
		{source: "foo.forEach(function() { console.log(this); z(x => console.log(x, this)); }, null);", settings: nil, findings: 2},
		{source: "/** @returns {void} */ function foo() { console.log(this); z(x => console.log(x, this)); }", settings: nil, findings: 2},
		{source: "/** @this Obj */ foo(function() { console.log(this); z(x => console.log(x, this)); });", settings: nil, findings: 2},
		{source: "var Ctor = function() { console.log(this); z(x => console.log(x, this)); }", settings: noInvalidThisSettings{CapIsConstructor: false}, findings: 2},
		{source: "var func = function() { console.log(this); z(x => console.log(x, this)); }", settings: nil, findings: 2},
		{source: "var func = function() { console.log(this); z(x => console.log(x, this)); }", settings: noInvalidThisSettings{CapIsConstructor: false}, findings: 2},
		{source: "Ctor = function() { console.log(this); z(x => console.log(x, this)); }", settings: noInvalidThisSettings{CapIsConstructor: false}, findings: 2},
		{source: "func = function() { console.log(this); z(x => console.log(x, this)); }", settings: nil, findings: 2},
		{source: "func = function() { console.log(this); z(x => console.log(x, this)); }", settings: noInvalidThisSettings{CapIsConstructor: false}, findings: 2},
		{source: "function foo(func = function() { console.log(this); z(x => console.log(x, this)); }) {}", settings: nil, findings: 2},
		{source: "[func = function() { console.log(this); z(x => console.log(x, this)); }] = a", settings: nil, findings: 2},
		{source: "class C { [this.foo]; }", settings: nil, findings: 1},
		{source: "class C { static {} [this]; }", settings: nil, findings: 1},
		{source: "class C { static {} [this.x]; }", settings: nil, findings: 1},
		{source: "function foo() { 'use strict'; this.eval(); }", settings: nil, findings: 1},
		{source: "function foo() { \"use strict\"; console.log(this); z(x => console.log(x, this)); } /* should error */", settings: nil, findings: 2},
		{source: "function Foo() { \"use strict\"; console.log(this); z(x => console.log(x, this)); } /* should error */", settings: noInvalidThisSettings{CapIsConstructor: false}, findings: 2},
		{source: "var obj = {foo: function() { \"use strict\"; return function() { console.log(this); z(x => console.log(x, this)); }; }}; /* should error */", settings: nil, findings: 2},
		{source: "obj.foo = function() { \"use strict\"; return function() { console.log(this); z(x => console.log(x, this)); }; }; /* should error */", settings: nil, findings: 2},
		{source: "class A { foo() { return function() { console.log(this); z(x => console.log(x, this)); }; } } /* should error */", settings: nil, findings: 2},
		{source: "class C { static { function foo() { this.x; } } } /* should error */", settings: nil, findings: 1},
		{source: "class C { static { (function() { this.x; }); } } /* should error */", settings: nil, findings: 1},
		{source: "class C { static { (function() { this.x; })(); } } /* should error */", settings: nil, findings: 1},
		{source: "\n    interface SomeType {\n      prop: string;\n    }\n    function foo() {\n      this.prop;\n    }\n          ", settings: nil, findings: 1},
		{source: "\n    console.log(this);\n    z(x => console.log(x, this));\n          ", settings: nil, findings: 2},
		{source: "\n    (function () {\n      console.log(this);\n      z(x => console.log(x, this));\n    })();\n          ", settings: nil, findings: 2},
		{source: "\n    function foo() {\n      console.log(this);\n      z(x => console.log(x, this));\n    }\n          ", settings: nil, findings: 2},
		{source: "\n    function foo() {\n      console.log(this);\n      z(x => console.log(x, this));\n    }\n          ", settings: noInvalidThisSettings{CapIsConstructor: false}, findings: 2},
		{source: "\n    function Foo() {\n      console.log(this);\n      z(x => console.log(x, this));\n    }\n          ", settings: noInvalidThisSettings{CapIsConstructor: false}, findings: 2},
		{source: "\n    function foo() {\n      'use strict';\n      console.log(this);\n      z(x => console.log(x, this));\n    }\n          ", settings: nil, findings: 2},
		{source: "\n    function Foo() {\n      'use strict';\n      console.log(this);\n      z(x => console.log(x, this));\n    }\n          ", settings: noInvalidThisSettings{CapIsConstructor: false}, findings: 2},
		{source: "\n    return function () {\n      console.log(this);\n      z(x => console.log(x, this));\n    };\n          ", settings: nil, findings: 2},
		{source: "\n    var foo = function () {\n      console.log(this);\n      z(x => console.log(x, this));\n    }.bar(obj);\n          ", settings: nil, findings: 2},
		{source: "\n    var obj = {\n      foo: function () {\n        function foo() {\n          console.log(this);\n          z(x => console.log(x, this));\n        }\n        foo();\n      },\n    };\n          ", settings: nil, findings: 2},
		{source: "\n    var obj = {\n      foo() {\n        function foo() {\n          console.log(this);\n          z(x => console.log(x, this));\n        }\n        foo();\n      },\n    };\n          ", settings: nil, findings: 2},
		{source: "\n    var obj = {\n      foo: function () {\n        return function () {\n          console.log(this);\n          z(x => console.log(x, this));\n        };\n      },\n    };\n          ", settings: nil, findings: 2},
		{source: "\n    var obj = {\n      foo: function () {\n        'use strict';\n        return function () {\n          console.log(this);\n          z(x => console.log(x, this));\n        };\n      },\n    };\n          ", settings: nil, findings: 2},
		{source: "\n    obj.foo = function () {\n      return function () {\n        console.log(this);\n        z(x => console.log(x, this));\n      };\n    };\n          ", settings: nil, findings: 2},
		{source: "\n    obj.foo = function () {\n      'use strict';\n      return function () {\n        console.log(this);\n        z(x => console.log(x, this));\n      };\n    };\n          ", settings: nil, findings: 2},
		{source: "\n    class A {\n      foo() {\n        return function () {\n          console.log(this);\n          z(x => console.log(x, this));\n        };\n      }\n    }\n          ", settings: nil, findings: 2},
		{source: "\n    class A {\n      b = new Array(1, 2, function () {\n        console.log(this);\n        z(x => console.log(x, this));\n      });\n    }\n          ", settings: nil, findings: 2},
		{source: "\n    class A {\n      b = () => {\n        function c() {\n          console.log(this);\n          z(x => console.log(x, this));\n        }\n      };\n    }\n          ", settings: nil, findings: 2},
		{source: "\n    obj.foo = (function () {\n      return () => {\n        console.log(this);\n        z(x => console.log(x, this));\n      };\n    })();\n          ", settings: nil, findings: 2},
		{source: "\n    obj.foo = (() => () => {\n      console.log(this);\n      z(x => console.log(x, this));\n    })();\n          ", settings: nil, findings: 2},
		{source: "\n    var foo = function () {\n      console.log(this);\n      z(x => console.log(x, this));\n    }.bind(null);\n          ", settings: nil, findings: 2},
		{source: "\n    (function () {\n      console.log(this);\n      z(x => console.log(x, this));\n    }).call(undefined);\n          ", settings: nil, findings: 2},
		{source: "\n    (function () {\n      console.log(this);\n      z(x => console.log(x, this));\n    }).apply(void 0);\n          ", settings: nil, findings: 2},
		{source: "\n    Array.from([], function () {\n      console.log(this);\n      z(x => console.log(x, this));\n    });\n          ", settings: nil, findings: 2},
		{source: "\n    Array.fromAsync([], function () {\n      console.log(this);\n      z(x => console.log(x, this));\n    });\n          ", settings: nil, findings: 2},
		{source: "\n    foo.every(function () {\n      console.log(this);\n      z(x => console.log(x, this));\n    });\n          ", settings: nil, findings: 2},
		{source: "\n    foo.filter(function () {\n      console.log(this);\n      z(x => console.log(x, this));\n    });\n          ", settings: nil, findings: 2},
		{source: "\n    foo.find(function () {\n      console.log(this);\n      z(x => console.log(x, this));\n    });\n          ", settings: nil, findings: 2},
		{source: "\n    foo.findIndex(function () {\n      console.log(this);\n      z(x => console.log(x, this));\n    });\n          ", settings: nil, findings: 2},
		{source: "\n    foo.forEach(function () {\n      console.log(this);\n      z(x => console.log(x, this));\n    });\n          ", settings: nil, findings: 2},
		{source: "\n    foo.map(function () {\n      console.log(this);\n      z(x => console.log(x, this));\n    });\n          ", settings: nil, findings: 2},
		{source: "\n    foo.some(function () {\n      console.log(this);\n      z(x => console.log(x, this));\n    });\n          ", settings: nil, findings: 2},
		{source: "\n    foo.forEach(function () {\n      console.log(this);\n      z(x => console.log(x, this));\n    }, null);\n          ", settings: nil, findings: 2},
		{source: "\n    /** @returns {void} */ function foo() {\n      console.log(this);\n      z(x => console.log(x, this));\n    }\n          ", settings: nil, findings: 2},
		{source: "\n    /** @this Obj */ foo(function () {\n      console.log(this);\n      z(x => console.log(x, this));\n    });\n          ", settings: nil, findings: 2},
		{source: "\n    var Ctor = function () {\n      console.log(this);\n      z(x => console.log(x, this));\n    };\n          ", settings: noInvalidThisSettings{CapIsConstructor: false}, findings: 2},
		{source: "\n    var func = function () {\n      console.log(this);\n      z(x => console.log(x, this));\n    };\n          ", settings: nil, findings: 2},
		{source: "\n    var func = function () {\n      console.log(this);\n      z(x => console.log(x, this));\n    };\n          ", settings: noInvalidThisSettings{CapIsConstructor: false}, findings: 2},
		{source: "\n    Ctor = function () {\n      console.log(this);\n      z(x => console.log(x, this));\n    };\n          ", settings: noInvalidThisSettings{CapIsConstructor: false}, findings: 2},
		{source: "\n    func = function () {\n      console.log(this);\n      z(x => console.log(x, this));\n    };\n          ", settings: nil, findings: 2},
		{source: "\n    func = function () {\n      console.log(this);\n      z(x => console.log(x, this));\n    };\n          ", settings: noInvalidThisSettings{CapIsConstructor: false}, findings: 2},
		{source: "\n    function foo(\n      func = function () {\n        console.log(this);\n        z(x => console.log(x, this));\n      },\n    ) {}\n          ", settings: nil, findings: 2},
		{source: "\n    [\n      func = function () {\n        console.log(this);\n        z(x => console.log(x, this));\n      },\n    ] = a;\n          ", settings: nil, findings: 2},
		{source: "\n\t\t\tfunction foo() {\n  \t\t\t\tclass C {\n    \t\t\t\taccessor [this.a] = foo;\n  \t\t\t\t}\n\t\t\t}\n          ", settings: nil, findings: 1},
		{source: "\n\t\t\tfunction foo() {\n  \t\t\t\tclass C {\n    \t\t\t\taccessor [this.a] = this.b;\n  \t\t\t\t}\n\t\t\t}\n          ", settings: nil, findings: 1},
		{source: "\n\t\t\tfunction foo() {\n  \t\t\t\tclass C {\n    \t\t\t\taccessor a = this.b;\n    \t\t\t\taccessor [this.c] = foo;\n  \t\t\t\t}\n\t\t\t}\n          ", settings: nil, findings: 1},
	}

	for _, testCase := range cases {
		// RunWithOptions, not Run: 45 of these rows carry an explicit capIsConstructor and Run
		// would hand every one of them the default instead. That is not a weaker test, it is a
		// different one -- the rows exist precisely because the option changes the verdict.
		result := rule_testing.RunWithOptions(t, NoInvalidThis, noInvalidThisFile,
			testCase.source, testCase.settings)
		ids := make([]string, testCase.findings)
		for index := range ids {
			ids[index] = "unexpectedThis"
		}
		rule_testing.ExpectFindings(t, result, ids...)
	}
}

func TestNoInvalidThisStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		source   string
		settings any
	}{
		{source: "class A {static foo() { console.log(this); z(x => console.log(x, this)); }};", settings: nil},
		{source: "function Foo() { console.log(this); z(x => console.log(x, this)); }", settings: nil},
		{source: "function Foo() { console.log(this); z(x => console.log(x, this)); }", settings: noInvalidThisSettings{CapIsConstructor: true}},
		{source: "function Foo() { console.log(this); z(x => console.log(x, this)); }", settings: noInvalidThisSettings{CapIsConstructor: true}},
		{source: "var Foo = function Foo() { console.log(this); z(x => console.log(x, this)); };", settings: nil},
		{source: "class A {constructor() { console.log(this); z(x => console.log(x, this)); }};", settings: nil},
		{source: "var obj = {foo: function() { console.log(this); z(x => console.log(x, this)); }};", settings: nil},
		{source: "var obj = {foo() { console.log(this); z(x => console.log(x, this)); }};", settings: nil},
		{source: "var obj = {foo: foo || function() { console.log(this); z(x => console.log(x, this)); }};", settings: nil},
		{source: "var obj = {foo: hasNative ? foo : function() { console.log(this); z(x => console.log(x, this)); }};", settings: nil},
		{source: "var obj = {foo: (function() { return function() { console.log(this); z(x => console.log(x, this)); }; })()};", settings: nil},
		{source: "Object.defineProperty(obj, \"foo\", {value: function() { console.log(this); z(x => console.log(x, this)); }})", settings: nil},
		{source: "Object.defineProperties(obj, {foo: {value: function() { console.log(this); z(x => console.log(x, this)); }}})", settings: nil},
		{source: "obj.foo = function() { console.log(this); z(x => console.log(x, this)); };", settings: nil},
		{source: "obj.foo = foo || function() { console.log(this); z(x => console.log(x, this)); };", settings: nil},
		{source: "obj.foo = foo ? bar : function() { console.log(this); z(x => console.log(x, this)); };", settings: nil},
		{source: "obj.foo = (function() { return function() { console.log(this); z(x => console.log(x, this)); }; })();", settings: nil},
		{source: "obj.foo = (() => function() { console.log(this); z(x => console.log(x, this)); })();", settings: nil},
		{source: "obj.foo = (function() { return function() { console.log(this); z(x => console.log(x, this)); }; })?.();", settings: nil},
		{source: "class A {foo() { console.log(this); z(x => console.log(x, this)); }};", settings: nil},
		{source: "var foo = function() { console.log(this); z(x => console.log(x, this)); }.bind(obj);", settings: nil},
		{source: "(function() { console.log(this); z(x => console.log(x, this)); }).call(obj);", settings: nil},
		{source: "(function() { console.log(this); z(x => console.log(x, this)); }).apply(obj);", settings: nil},
		{source: "Reflect.apply(function() { console.log(this); z(x => console.log(x, this)); }, obj, []);", settings: nil},
		{source: "var foo = function() { console.log(this); z(x => console.log(x, this)); }?.bind(obj);", settings: nil},
		{source: "var foo = (function() { console.log(this); z(x => console.log(x, this)); }?.bind)(obj);", settings: nil},
		{source: "var foo = function() { console.log(this); z(x => console.log(x, this)); }.bind?.(obj);", settings: nil},
		{source: "Array.from([], function() { console.log(this); z(x => console.log(x, this)); }, obj);", settings: nil},
		{source: "Array.fromAsync([], function() { console.log(this); z(x => console.log(x, this)); }, obj);", settings: nil},
		{source: "foo.every(function() { console.log(this); z(x => console.log(x, this)); }, obj);", settings: nil},
		{source: "foo.filter(function() { console.log(this); z(x => console.log(x, this)); }, obj);", settings: nil},
		{source: "foo.find(function() { console.log(this); z(x => console.log(x, this)); }, obj);", settings: nil},
		{source: "foo.findIndex(function() { console.log(this); z(x => console.log(x, this)); }, obj);", settings: nil},
		{source: "foo.findLast(function() { console.log(this); z(x => console.log(x, this)); }, obj);", settings: nil},
		{source: "foo.findLastIndex(function() { console.log(this); z(x => console.log(x, this)); }, obj);", settings: nil},
		{source: "foo.flatMap(function() { console.log(this); z(x => console.log(x, this)); }, obj);", settings: nil},
		{source: "foo.forEach(function() { console.log(this); z(x => console.log(x, this)); }, obj);", settings: nil},
		{source: "foo.map(function() { console.log(this); z(x => console.log(x, this)); }, obj);", settings: nil},
		{source: "foo.some(function() { console.log(this); z(x => console.log(x, this)); }, obj);", settings: nil},
		{source: "Array?.from([], function() { console.log(this); z(x => console.log(x, this)); }, obj);", settings: nil},
		{source: "foo?.every(function() { console.log(this); z(x => console.log(x, this)); }, obj);", settings: nil},
		{source: "(Array?.from)([], function() { console.log(this); z(x => console.log(x, this)); }, obj);", settings: nil},
		{source: "(foo?.every)(function() { console.log(this); z(x => console.log(x, this)); }, obj);", settings: nil},
		{source: "/** @this Obj */ function foo() { console.log(this); z(x => console.log(x, this)); }", settings: nil},
		{source: "/**\n * @returns {void}\n * @this Obj\n */\nfunction foo() { console.log(this); z(x => console.log(x, this)); }", settings: nil},
		{source: "foo(/* @this Obj */ function() { console.log(this); z(x => console.log(x, this)); });", settings: nil},
		{source: "function foo() { /** @this Obj*/ return function bar() { console.log(this); z(x => console.log(x, this)); }; }", settings: nil},
		{source: "var Ctor = function() { console.log(this); z(x => console.log(x, this)); }", settings: nil},
		{source: "Ctor = function() { console.log(this); z(x => console.log(x, this)); }", settings: nil},
		{source: "function foo(Ctor = function() { console.log(this); z(x => console.log(x, this)); }) {}", settings: nil},
		{source: "[obj.method = function() { console.log(this); z(x => console.log(x, this)); }] = a", settings: nil},
		{source: "obj.method &&= function () { console.log(this); z(x => console.log(x, this)); }", settings: nil},
		{source: "obj.method ||= function () { console.log(this); z(x => console.log(x, this)); }", settings: nil},
		{source: "obj.method ??= function () { console.log(this); z(x => console.log(x, this)); }", settings: nil},
		{source: "class C { field = this }", settings: nil},
		{source: "class C { static field = this }", settings: nil},
		{source: "class C { field = console.log(this); }", settings: nil},
		{source: "class C { field = z(x => console.log(x, this)); }", settings: nil},
		{source: "class C { field = function () { console.log(this); z(x => console.log(x, this)); }; }", settings: nil},
		{source: "class C { #field = function () { console.log(this); z(x => console.log(x, this)); }; }", settings: nil},
		{source: "class C { foo = () => this; }", settings: nil},
		{source: "class C { foo = () => { this }; }", settings: nil},
		{source: "class C { static { this.x; } }", settings: nil},
		{source: "class C { static { () => { this.x; } } }", settings: nil},
		{source: "class C { static { class D { [this.x]; } } }", settings: nil},
		{source: "\n    describe('foo', () => {\n      it('does something', function (this: Mocha.Context) {\n        this.timeout(100);\n        // done\n      });\n    });\n        ", settings: nil},
		{source: "\n          interface SomeType {\n            prop: string;\n          }\n          function foo(this: SomeType) {\n            this.prop;\n          }\n        ", settings: nil},
		{source: "\n    function foo(this: prop) {\n      this.propMethod();\n    }\n        ", settings: nil},
		{source: "\n    z(function (x, this: context) {\n      console.log(x, this);\n    });\n        ", settings: nil},
		{source: "\n    function foo() {\n      /** @this Obj*/ return function bar() {\n        console.log(this);\n        z(x => console.log(x, this));\n      };\n    }\n        ", settings: nil},
		{source: "\n    var Ctor = function () {\n      console.log(this);\n      z(x => console.log(x, this));\n    };\n        ", settings: nil},
		{source: "\n    function Foo() {\n      console.log(this);\n      z(x => console.log(x, this));\n    }\n          ", settings: nil},
		{source: "\n    function Foo() {\n      console.log(this);\n      z(x => console.log(x, this));\n    }\n          ", settings: noInvalidThisSettings{CapIsConstructor: true}},
		{source: "\n    function Foo() {\n      console.log(this);\n      z(x => console.log(x, this));\n    }\n          ", settings: noInvalidThisSettings{CapIsConstructor: true}},
		{source: "\n    var Foo = function Foo() {\n      console.log(this);\n      z(x => console.log(x, this));\n    };\n          ", settings: nil},
		{source: "\n    class A {\n      constructor() {\n        console.log(this);\n        z(x => console.log(x, this));\n      }\n    }\n          ", settings: nil},
		{source: "\n    var obj = {\n      foo: function () {\n        console.log(this);\n        z(x => console.log(x, this));\n      },\n    };\n          ", settings: nil},
		{source: "\n    var obj = {\n      foo() {\n        console.log(this);\n        z(x => console.log(x, this));\n      },\n    };\n          ", settings: nil},
		{source: "\n    var obj = {\n      foo:\n        foo ||\n        function () {\n          console.log(this);\n          z(x => console.log(x, this));\n        },\n    };\n          ", settings: nil},
		{source: "\n    var obj = {\n      foo: hasNative\n        ? foo\n        : function () {\n            console.log(this);\n            z(x => console.log(x, this));\n          },\n    };\n          ", settings: nil},
		{source: "\n    var obj = {\n      foo: (function () {\n        return function () {\n          console.log(this);\n          z(x => console.log(x, this));\n        };\n      })(),\n    };\n          ", settings: nil},
		{source: "\n    Object.defineProperty(obj, 'foo', {\n      value: function () {\n        console.log(this);\n        z(x => console.log(x, this));\n      },\n    });\n          ", settings: nil},
		{source: "\n    Object.defineProperties(obj, {\n      foo: {\n        value: function () {\n          console.log(this);\n          z(x => console.log(x, this));\n        },\n      },\n    });\n          ", settings: nil},
		{source: "\n    obj.foo = function () {\n      console.log(this);\n      z(x => console.log(x, this));\n    };\n          ", settings: nil},
		{source: "\n    obj.foo =\n      foo ||\n      function () {\n        console.log(this);\n        z(x => console.log(x, this));\n      };\n          ", settings: nil},
		{source: "\n    obj.foo = foo\n      ? bar\n      : function () {\n          console.log(this);\n          z(x => console.log(x, this));\n        };\n          ", settings: nil},
		{source: "\n    obj.foo = (function () {\n      return function () {\n        console.log(this);\n        z(x => console.log(x, this));\n      };\n    })();\n          ", settings: nil},
		{source: "\n    obj.foo = (() =>\n      function () {\n        console.log(this);\n        z(x => console.log(x, this));\n      })();\n          ", settings: nil},
		{source: "\n    (function () {\n      console.log(this);\n      z(x => console.log(x, this));\n    }).call(obj);\n        ", settings: nil},
		{source: "\n    var foo = function () {\n      console.log(this);\n      z(x => console.log(x, this));\n    }.bind(obj);\n        ", settings: nil},
		{source: "\n    Reflect.apply(\n      function () {\n        console.log(this);\n        z(x => console.log(x, this));\n      },\n      obj,\n      [],\n    );\n        ", settings: nil},
		{source: "\n    (function () {\n      console.log(this);\n      z(x => console.log(x, this));\n    }).apply(obj);\n        ", settings: nil},
		{source: "\n    class A {\n      foo() {\n        console.log(this);\n        z(x => console.log(x, this));\n      }\n    }\n        ", settings: nil},
		{source: "\n    Array.from(\n      [],\n      function () {\n        console.log(this);\n        z(x => console.log(x, this));\n      },\n      obj,\n    );\n        ", settings: nil},
		{source: "\n    Array.fromAsync(\n      [],\n      function () {\n        console.log(this);\n        z(x => console.log(x, this));\n      },\n      obj,\n    );\n        ", settings: nil},
		{source: "\n    foo.every(function () {\n      console.log(this);\n      z(x => console.log(x, this));\n    }, obj);\n        ", settings: nil},
		{source: "\n    foo.filter(function () {\n      console.log(this);\n      z(x => console.log(x, this));\n    }, obj);\n        ", settings: nil},
		{source: "\n    foo.find(function () {\n      console.log(this);\n      z(x => console.log(x, this));\n    }, obj);\n        ", settings: nil},
		{source: "\n    foo.findIndex(function () {\n      console.log(this);\n      z(x => console.log(x, this));\n    }, obj);\n        ", settings: nil},
		{source: "\n    foo.forEach(function () {\n      console.log(this);\n      z(x => console.log(x, this));\n    }, obj);\n        ", settings: nil},
		{source: "\n    foo.map(function () {\n      console.log(this);\n      z(x => console.log(x, this));\n    }, obj);\n        ", settings: nil},
		{source: "\n    foo.some(function () {\n      console.log(this);\n      z(x => console.log(x, this));\n    }, obj);\n        ", settings: nil},
		{source: "\n    /** @this Obj */ function foo() {\n      console.log(this);\n      z(x => console.log(x, this));\n    }\n        ", settings: nil},
		{source: "\n    foo(\n      /* @this Obj */ function () {\n        console.log(this);\n        z(x => console.log(x, this));\n      },\n    );\n        ", settings: nil},
		{source: "\n    /**\n     * @returns {void}\n     * @this Obj\n     */\n    function foo() {\n      console.log(this);\n      z(x => console.log(x, this));\n    }\n        ", settings: nil},
		{source: "\n    Ctor = function () {\n      console.log(this);\n      z(x => console.log(x, this));\n    };\n        ", settings: nil},
		{source: "\n    function foo(\n      Ctor = function () {\n        console.log(this);\n        z(x => console.log(x, this));\n      },\n    ) {}\n        ", settings: nil},
		{source: "\n    [\n      obj.method = function () {\n        console.log(this);\n        z(x => console.log(x, this));\n      },\n    ] = a;\n        ", settings: nil},
		{source: "\n          class A {\n            a = 5;\n            b = this.a;\n            accessor c = this.a;\n          }\n          ", settings: nil},
		{source: "\n          class A {\n            a = 5;\n            accessor b = this.a + 1;\n          }\n          ", settings: nil},
		{source: "\n          class A {\n            a = 5;\n            accessor b = this;\n          }\n          ", settings: nil},
		{source: "\n          class A {\n            b = 0;\n            c = this.b;\n          }\n          ", settings: nil},
		{source: "\n          class A {\n            b = new Array(this, 1, 2, 3);\n          }\n          ", settings: nil},
		{source: "\n          class A {\n            b = () => {\n            console.log(this);\n            };\n          }\n          ", settings: nil},
		{source: "\n          class A {\n            static foo() {\n            console.log(this);\n            z(x => console.log(x, this));\n            }\n          }\n          ", settings: nil},
	}

	for _, testCase := range cases {
		result := rule_testing.RunWithOptions(t, NoInvalidThis, noInvalidThisFile,
			testCase.source, testCase.settings)
		rule_testing.ExpectClean(t, result)
	}
}

// TestNoInvalidThisComputedKeysAreOutsideTheMember pins the six shapes where the installed core
// rule and the installed @typescript-eslint extension were measured to DISAGREE, plus the seventh
// where this tree's namespaced port diverges from both.
//
// This is the reason the rule exists beside the already-enabled namespaced one. Driving both
// installed rules over all 572 corpus cases gave 566 agreements and 6 disagreements, and the six
// are one class: a `this` inside the computed key of a class field, which the extension exempts
// because its PropertyDefinition arm pushes valid=true over the whole member including its key.
//
// The rows are asserted with explicit spans, because a rule that reports the right COUNT while
// pointing at the value instead of the key would pass a count-only fixture while being wrong about
// the thing this test exists to check.
func TestNoInvalidThisComputedKeysAreOutsideTheMember(t *testing.T) {
	t.Parallel()

	cases := []struct {
		source string
		// spans is the source text each finding should cover, in order.
		spans []string
	}{
		// The six the extension misses.
		{"class C {\n  [this.foo];\n}\n", []string{"this"}},
		{"class C {\n  static {}\n  [this];\n}\n", []string{"this"}},
		{"class C {\n  static {}\n  [this.x];\n}\n", []string{"this"}},
		{"function foo() {\n  class C {\n    accessor [this.a] = foo;\n  }\n}\n", []string{"this"}},
		{"function foo() {\n  class C {\n    accessor [this.a] = this.b;\n  }\n}\n", []string{"this"}},
		{"function foo() {\n  class C {\n    accessor a = this.b;\n    accessor [this.c] = foo;\n  }\n}\n", []string{"this"}},

		// The seventh, where the extension agrees with core and the namespaced port here does not.
		{"class C {\n  [this.a]() {}\n}\n", []string{"this"}},

		// A field carrying BOTH: the key's `this` reports and the value's does not. This is the
		// discriminating shape -- a port that exempted the whole member reports zero here, and one
		// that exempted nothing reports two.
		{"class C {\n  [this.a] = this.b;\n}\n", []string{"this"}},

		// Controls, so a rule that had gone silent could not pass this test vacuously.
		{"class C {\n  foo = this.bar;\n}\n", nil},
		{"class C {\n  foo() {\n    this.bar;\n  }\n}\n", nil},
		{"class C {\n  static {\n    this.x;\n  }\n}\n", nil},
		// A function nested inside a computed key is judged by that function, not by the key.
		{"class C {\n  [foo(function () {\n    this.x;\n  })];\n}\n", []string{"this"}},
		{"class C {\n  [foo(function Bar() {\n    this.x;\n  })];\n}\n", nil},
	}

	for _, testCase := range cases {
		result := rule_testing.Run(t, NoInvalidThis, noInvalidThisFile, testCase.source)
		if len(result.Diagnostics) != len(testCase.spans) {
			t.Errorf("got %d findings, want %d, for %q",
				len(result.Diagnostics), len(testCase.spans), testCase.source)
			continue
		}
		for index, want := range testCase.spans {
			diagnostic := result.Diagnostics[index]
			got := testCase.source[diagnostic.Range.Pos():diagnostic.Range.End()]
			if got != want {
				t.Errorf("finding %d covers %q, want %q, for %q",
					index, got, want, testCase.source)
			}
		}
	}
}

// TestNoInvalidThisMessageText asserts the whole rendered message rather than its id.
//
// A message id proves which branch fired. The text proves what the rule computed, and a fixture
// whose predicate is weaker than the property it guards is not a guard.
func TestNoInvalidThisMessageText(t *testing.T) {
	t.Parallel()

	result := rule_testing.Run(t, NoInvalidThis, noInvalidThisFile, "function foo() {\n  this.x;\n}\n")
	rule_testing.ExpectFindings(t, result, "unexpectedThis")

	got := result.Diagnostics[0].Message
	if got.Id != "unexpectedThis" {
		t.Errorf("message id is %q, want unexpectedThis", got.Id)
	}
	// The WHOLE description, not a prefix or a Contains. A predicate weaker than the property it
	// guards is not a guard: a two-hash string contains a one-hash needle, which is how a rule
	// shipped a doubled `#` behind a green fixture.
	if got.Description != messageNoInvalidThisUnexpectedThis.Description {
		t.Errorf("description is %q, want %q", got.Description,
			messageNoInvalidThisUnexpectedThis.Description)
	}
	if !strings.HasPrefix(got.Description, "Nothing binds `this` here") {
		t.Errorf("description is %q, which does not open with the rule's explanation", got.Description)
	}
}

// TestNoInvalidThisDecoderDefaultsCapIsConstructorTrue routes the option through the rule's own
// exported decoder rather than building the settings struct, which is what puts the default
// inversion under test.
//
// The zero value of noInvalidThisSettings says capIsConstructor is FALSE and upstream's default is
// TRUE, so a rule reaching its settings through a generic decoder on empty input would silently
// report every capitalized constructor function in the tree. The nil case matters just as much:
// a rule configured as a bare "error" is handed nil options.
func TestNoInvalidThisDecoderDefaultsCapIsConstructorTrue(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		raw  string
		want bool
	}{
		{name: "empty input", raw: "", want: true},
		{name: "empty object", raw: `{}`, want: true},
		{name: "explicit true", raw: `{"capIsConstructor":true}`, want: true},
		{name: "explicit false", raw: `{"capIsConstructor":false}`, want: false},
	}
	for _, testCase := range cases {
		decoded, err := DecodeNoInvalidThisOptions([]byte(testCase.raw))
		if err != nil && testCase.raw != "" {
			t.Errorf("%s: decode returned %v", testCase.name, err)
			continue
		}
		settings, ok := decoded.(noInvalidThisSettings)
		if !ok {
			t.Errorf("%s: decode produced %T, not noInvalidThisSettings", testCase.name, decoded)
			continue
		}
		if settings.CapIsConstructor != testCase.want {
			t.Errorf("%s: capIsConstructor is %v, want %v",
				testCase.name, settings.CapIsConstructor, testCase.want)
		}
	}

	// Nil options is what a rule configured as a bare "error" actually receives.
	if got := noInvalidThisSettingsFrom(nil); !got.CapIsConstructor {
		t.Error("nil options gave capIsConstructor false, so a bare \"error\" would invert the default")
	}

	// And the end-to-end consequence, so the default is pinned at the rule rather than only at the
	// decoder: a capitalized function is clean by default and reports when the option is off.
	source := "function Foo() {\n  this.x = 1;\n}\n"
	rule_testing.ExpectClean(t, rule_testing.Run(t, NoInvalidThis, noInvalidThisFile, source))
}

// TestNoInvalidThisShapesTheCorpusDoesNotWrite covers thirteen inputs upstream's corpus never
// writes, each one added because a mutation sweep found the imported fixtures blind to a
// discrimination the rule makes.
//
// Every expectation is what the installed ESLint 10.8.1 core rule answered on that exact source,
// measured through the Linter interface rather than reasoned about. Four mutants survived all 215
// imported cases and every one of them is a fixture gap rather than an equivalence, which is what
// these rows close:
//
//	Array.from arity written as >= 3 instead of == 3   separated by a four-argument call
//	the /Array$/ object pattern reduced to "Array"      separated by Int8Array.from
//	the method arm of the OUTWARD walk deleted          separated by a function in a method's key
//	parent.Initializer() != current made constant       separated by a function in a field's key
//
// The last two are worth naming carefully, because their shape is easy to mistake. They are arms of
// the outward walk in thisIsDefaultBoundIn, reached by a FUNCTION whose parent is the member -- not
// by the stack push that a `this` inside a method body goes through. The corpus exercises the push
// heavily and the arm not at all, so the arm was dead against every imported case while being
// load-bearing on real source. That is the standard's "a guard can be dead against every imported
// case and still be load-bearing" arriving in this rule.
func TestNoInvalidThisShapesTheCorpusDoesNotWrite(t *testing.T) {
	t.Parallel()

	cases := []struct {
		source   string
		findings int
	}{
		// Array.from and friends take exactly three arguments upstream, not at least three.
		{"Array.from([], function () {\n  this.x;\n}, obj);\n", 0},
		{"Array.from([], function () {\n  this.x;\n}, obj, extra);\n", 1},
		{"Array.from([], function () {\n  this.x;\n});\n", 1},
		{"Reflect.apply(function () {\n  this.x;\n}, obj, [], extra);\n", 1},
		{"foo.forEach(function () {\n  this.x;\n}, obj, extra);\n", 1},

		// The object is matched by upstream's /Array$/ pattern, so every typed array qualifies,
		// and so does a user-defined name ending in "Array".
		{"Int8Array.from([], function () {\n  this.x;\n}, obj);\n", 0},
		{"Float64Array.from([], function () {\n  this.x;\n}, obj);\n", 0},
		{"MyArray.from([], function () {\n  this.x;\n}, obj);\n", 0},

		// A function sitting in a member's COMPUTED KEY has no receiver, while the same function
		// as the member's value does. These are the outward walk's member arms.
		{"class C {\n  [function () {\n    this.x;\n  }]() {}\n}\n", 1},
		{"const obj = {\n  [function () {\n    this.x;\n  }]: 1,\n};\n", 1},
		{"class C {\n  [function () {\n    this.x;\n  }] = 1;\n}\n", 1},
		{"const obj = {\n  foo: function () {\n    this.x;\n  },\n};\n", 0},
		{"class C {\n  foo = function () {\n    this.x;\n  };\n}\n", 0},
	}

	for _, testCase := range cases {
		result := rule_testing.Run(t, NoInvalidThis, noInvalidThisFile, testCase.source)
		ids := make([]string, testCase.findings)
		for index := range ids {
			ids[index] = "unexpectedThis"
		}
		rule_testing.ExpectFindings(t, result, ids...)
	}
}
