package typescript

import (
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// invalidThisFile is where the fixtures pretend to live.
const invalidThisFile = "/repository/source/InvalidThis.ts"

// The corpus is typescript-eslint's own, extracted from its test file through the TypeScript
// compiler API rather than by regular expression or by retyping, so no escape sequence passes
// through a shell or an editor on the way in. It carries 47 valid cases and
// 44 invalid ones.
//
// Every case below was first driven through the INSTALLED @typescript-eslint 8.67.0 rule with the
// ESLint Linter interface, and the expectations here are that run's output rather than the test
// file's annotations. The two agree on all 91. See the rule's doc
// comment for why they were run under sourceType module, and for the two cases whose declared
// `globalReturn` turns out not to be what makes them report.
func TestNoInvalidThisFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		source   string
		settings any
		findings int
	}{
		{source: "\ninterface SomeType {\n  prop: string;\n}\nfunction foo() {\n  this.prop;\n}\n      ", settings: nil, findings: 1},
		{source: "\nconsole.log(this);\nz(x => console.log(x, this));\n      ", settings: nil, findings: 2},
		{source: "\nconsole.log(this);\nz(x => console.log(x, this));\n      ", settings: nil, findings: 2},
		{source: "\n(function () {\n  console.log(this);\n  z(x => console.log(x, this));\n})();\n      ", settings: nil, findings: 2},
		{source: "\nfunction foo() {\n  console.log(this);\n  z(x => console.log(x, this));\n}\n      ", settings: nil, findings: 2},
		{source: "\nfunction foo() {\n  console.log(this);\n  z(x => console.log(x, this));\n}\n      ", settings: noInvalidThisSettings{CapIsConstructor: false}, findings: 2},
		{source: "\nfunction Foo() {\n  console.log(this);\n  z(x => console.log(x, this));\n}\n      ", settings: noInvalidThisSettings{CapIsConstructor: false}, findings: 2},
		{source: "\nfunction foo() {\n  'use strict';\n  console.log(this);\n  z(x => console.log(x, this));\n}\n      ", settings: nil, findings: 2},
		{source: "\nfunction Foo() {\n  'use strict';\n  console.log(this);\n  z(x => console.log(x, this));\n}\n      ", settings: noInvalidThisSettings{CapIsConstructor: false}, findings: 2},
		{source: "\nreturn function () {\n  console.log(this);\n  z(x => console.log(x, this));\n};\n      ", settings: nil, findings: 2},
		{source: "\nvar foo = function () {\n  console.log(this);\n  z(x => console.log(x, this));\n}.bar(obj);\n      ", settings: nil, findings: 2},
		{source: "\nvar obj = {\n  foo: function () {\n    function foo() {\n      console.log(this);\n      z(x => console.log(x, this));\n    }\n    foo();\n  },\n};\n      ", settings: nil, findings: 2},
		{source: "\nvar obj = {\n  foo() {\n    function foo() {\n      console.log(this);\n      z(x => console.log(x, this));\n    }\n    foo();\n  },\n};\n      ", settings: nil, findings: 2},
		{source: "\nvar obj = {\n  foo: function () {\n    return function () {\n      console.log(this);\n      z(x => console.log(x, this));\n    };\n  },\n};\n      ", settings: nil, findings: 2},
		{source: "\nvar obj = {\n  foo: function () {\n    'use strict';\n    return function () {\n      console.log(this);\n      z(x => console.log(x, this));\n    };\n  },\n};\n      ", settings: nil, findings: 2},
		{source: "\nobj.foo = function () {\n  return function () {\n    console.log(this);\n    z(x => console.log(x, this));\n  };\n};\n      ", settings: nil, findings: 2},
		{source: "\nobj.foo = function () {\n  'use strict';\n  return function () {\n    console.log(this);\n    z(x => console.log(x, this));\n  };\n};\n      ", settings: nil, findings: 2},
		{source: "\nclass A {\n  foo() {\n    return function () {\n      console.log(this);\n      z(x => console.log(x, this));\n    };\n  }\n}\n      ", settings: nil, findings: 2},
		{source: "\nclass A {\n  b = new Array(1, 2, function () {\n    console.log(this);\n    z(x => console.log(x, this));\n  });\n}\n      ", settings: nil, findings: 2},
		{source: "\nclass A {\n  b = () => {\n    function c() {\n      console.log(this);\n      z(x => console.log(x, this));\n    }\n  };\n}\n      ", settings: nil, findings: 2},
		{source: "\nobj.foo = (function () {\n  return () => {\n    console.log(this);\n    z(x => console.log(x, this));\n  };\n})();\n      ", settings: nil, findings: 2},
		{source: "\nobj.foo = (() => () => {\n  console.log(this);\n  z(x => console.log(x, this));\n})();\n      ", settings: nil, findings: 2},
		{source: "\nvar foo = function () {\n  console.log(this);\n  z(x => console.log(x, this));\n}.bind(null);\n      ", settings: nil, findings: 2},
		{source: "\n(function () {\n  console.log(this);\n  z(x => console.log(x, this));\n}).call(undefined);\n      ", settings: nil, findings: 2},
		{source: "\n(function () {\n  console.log(this);\n  z(x => console.log(x, this));\n}).apply(void 0);\n      ", settings: nil, findings: 2},
		{source: "\nArray.from([], function () {\n  console.log(this);\n  z(x => console.log(x, this));\n});\n      ", settings: nil, findings: 2},
		{source: "\nfoo.every(function () {\n  console.log(this);\n  z(x => console.log(x, this));\n});\n      ", settings: nil, findings: 2},
		{source: "\nfoo.filter(function () {\n  console.log(this);\n  z(x => console.log(x, this));\n});\n      ", settings: nil, findings: 2},
		{source: "\nfoo.find(function () {\n  console.log(this);\n  z(x => console.log(x, this));\n});\n      ", settings: nil, findings: 2},
		{source: "\nfoo.findIndex(function () {\n  console.log(this);\n  z(x => console.log(x, this));\n});\n      ", settings: nil, findings: 2},
		{source: "\nfoo.forEach(function () {\n  console.log(this);\n  z(x => console.log(x, this));\n});\n      ", settings: nil, findings: 2},
		{source: "\nfoo.map(function () {\n  console.log(this);\n  z(x => console.log(x, this));\n});\n      ", settings: nil, findings: 2},
		{source: "\nfoo.some(function () {\n  console.log(this);\n  z(x => console.log(x, this));\n});\n      ", settings: nil, findings: 2},
		{source: "\nfoo.forEach(function () {\n  console.log(this);\n  z(x => console.log(x, this));\n}, null);\n      ", settings: nil, findings: 2},
		{source: "\n/** @returns {void} */ function foo() {\n  console.log(this);\n  z(x => console.log(x, this));\n}\n      ", settings: nil, findings: 2},
		{source: "\n/** @this Obj */ foo(function () {\n  console.log(this);\n  z(x => console.log(x, this));\n});\n      ", settings: nil, findings: 2},
		{source: "\nvar Ctor = function () {\n  console.log(this);\n  z(x => console.log(x, this));\n};\n      ", settings: noInvalidThisSettings{CapIsConstructor: false}, findings: 2},
		{source: "\nvar func = function () {\n  console.log(this);\n  z(x => console.log(x, this));\n};\n      ", settings: nil, findings: 2},
		{source: "\nvar func = function () {\n  console.log(this);\n  z(x => console.log(x, this));\n};\n      ", settings: noInvalidThisSettings{CapIsConstructor: false}, findings: 2},
		{source: "\nCtor = function () {\n  console.log(this);\n  z(x => console.log(x, this));\n};\n      ", settings: noInvalidThisSettings{CapIsConstructor: false}, findings: 2},
		{source: "\nfunc = function () {\n  console.log(this);\n  z(x => console.log(x, this));\n};\n      ", settings: nil, findings: 2},
		{source: "\nfunc = function () {\n  console.log(this);\n  z(x => console.log(x, this));\n};\n      ", settings: noInvalidThisSettings{CapIsConstructor: false}, findings: 2},
		{source: "\nfunction foo(\n  func = function () {\n    console.log(this);\n    z(x => console.log(x, this));\n  },\n) {}\n      ", settings: nil, findings: 2},
		{source: "\n[\n  func = function () {\n    console.log(this);\n    z(x => console.log(x, this));\n  },\n] = a;\n      ", settings: nil, findings: 2},
	}

	for index, testCase := range cases {
		t.Run(strings.TrimSpace(firstLineOf(testCase.source)), func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunWithOptions(t, NoInvalidThis, invalidThisFile,
				testCase.source, testCase.settings)
			expected := make([]string, testCase.findings)
			for position := range expected {
				expected[position] = "unexpectedThis"
			}
			rule_testing.ExpectFindings(t, result, expected...)
			_ = index
		})
	}
}

func TestNoInvalidThisStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		source   string
		settings any
	}{
		{source: "\ndescribe('foo', () => {\n  it('does something', function (this: Mocha.Context) {\n    this.timeout(100);\n    // done\n  });\n});\n    ", settings: nil},
		{source: "\ninterface SomeType {\n  prop: string;\n}\nfunction foo(this: SomeType) {\n  this.prop;\n}\n    ", settings: nil},
		{source: "\nfunction foo(this: prop) {\n  this.propMethod();\n}\n    ", settings: nil},
		{source: "\nz(function (x, this: context) {\n  console.log(x, this);\n});\n    ", settings: nil},
		{source: "\nfunction foo() {\n  /** @this Obj*/ return function bar() {\n    console.log(this);\n    z(x => console.log(x, this));\n  };\n}\n    ", settings: nil},
		{source: "\nvar Ctor = function () {\n  console.log(this);\n  z(x => console.log(x, this));\n};\n    ", settings: nil},
		{source: "\nfunction Foo() {\n  console.log(this);\n  z(x => console.log(x, this));\n}\n      ", settings: nil},
		{source: "\nfunction Foo() {\n  console.log(this);\n  z(x => console.log(x, this));\n}\n      ", settings: DefaultNoInvalidThisSettings()},
		{source: "\nfunction Foo() {\n  console.log(this);\n  z(x => console.log(x, this));\n}\n      ", settings: noInvalidThisSettings{CapIsConstructor: true}},
		{source: "\nvar Foo = function Foo() {\n  console.log(this);\n  z(x => console.log(x, this));\n};\n      ", settings: nil},
		{source: "\nclass A {\n  constructor() {\n    console.log(this);\n    z(x => console.log(x, this));\n  }\n}\n      ", settings: nil},
		{source: "\nvar obj = {\n  foo: function () {\n    console.log(this);\n    z(x => console.log(x, this));\n  },\n};\n      ", settings: nil},
		{source: "\nvar obj = {\n  foo() {\n    console.log(this);\n    z(x => console.log(x, this));\n  },\n};\n      ", settings: nil},
		{source: "\nvar obj = {\n  foo:\n    foo ||\n    function () {\n      console.log(this);\n      z(x => console.log(x, this));\n    },\n};\n      ", settings: nil},
		{source: "\nvar obj = {\n  foo: hasNative\n    ? foo\n    : function () {\n        console.log(this);\n        z(x => console.log(x, this));\n      },\n};\n      ", settings: nil},
		{source: "\nvar obj = {\n  foo: (function () {\n    return function () {\n      console.log(this);\n      z(x => console.log(x, this));\n    };\n  })(),\n};\n      ", settings: nil},
		{source: "\nObject.defineProperty(obj, 'foo', {\n  value: function () {\n    console.log(this);\n    z(x => console.log(x, this));\n  },\n});\n      ", settings: nil},
		{source: "\nObject.defineProperties(obj, {\n  foo: {\n    value: function () {\n      console.log(this);\n      z(x => console.log(x, this));\n    },\n  },\n});\n      ", settings: nil},
		{source: "\nobj.foo = function () {\n  console.log(this);\n  z(x => console.log(x, this));\n};\n      ", settings: nil},
		{source: "\nobj.foo =\n  foo ||\n  function () {\n    console.log(this);\n    z(x => console.log(x, this));\n  };\n      ", settings: nil},
		{source: "\nobj.foo = foo\n  ? bar\n  : function () {\n      console.log(this);\n      z(x => console.log(x, this));\n    };\n      ", settings: nil},
		{source: "\nobj.foo = (function () {\n  return function () {\n    console.log(this);\n    z(x => console.log(x, this));\n  };\n})();\n      ", settings: nil},
		{source: "\nobj.foo = (() =>\n  function () {\n    console.log(this);\n    z(x => console.log(x, this));\n  })();\n      ", settings: nil},
		{source: "\n(function () {\n  console.log(this);\n  z(x => console.log(x, this));\n}).call(obj);\n    ", settings: nil},
		{source: "\nvar foo = function () {\n  console.log(this);\n  z(x => console.log(x, this));\n}.bind(obj);\n    ", settings: nil},
		{source: "\nReflect.apply(\n  function () {\n    console.log(this);\n    z(x => console.log(x, this));\n  },\n  obj,\n  [],\n);\n    ", settings: nil},
		{source: "\n(function () {\n  console.log(this);\n  z(x => console.log(x, this));\n}).apply(obj);\n    ", settings: nil},
		{source: "\nclass A {\n  foo() {\n    console.log(this);\n    z(x => console.log(x, this));\n  }\n}\n    ", settings: nil},
		{source: "\nclass A {\n  b = 0;\n  c = this.b;\n}\n    ", settings: nil},
		{source: "\nclass A {\n  b = new Array(this, 1, 2, 3);\n}\n    ", settings: nil},
		{source: "\nclass A {\n  b = () => {\n    console.log(this);\n  };\n}\n    ", settings: nil},
		{source: "\nArray.from(\n  [],\n  function () {\n    console.log(this);\n    z(x => console.log(x, this));\n  },\n  obj,\n);\n    ", settings: nil},
		{source: "\nfoo.every(function () {\n  console.log(this);\n  z(x => console.log(x, this));\n}, obj);\n    ", settings: nil},
		{source: "\nfoo.filter(function () {\n  console.log(this);\n  z(x => console.log(x, this));\n}, obj);\n    ", settings: nil},
		{source: "\nfoo.find(function () {\n  console.log(this);\n  z(x => console.log(x, this));\n}, obj);\n    ", settings: nil},
		{source: "\nfoo.findIndex(function () {\n  console.log(this);\n  z(x => console.log(x, this));\n}, obj);\n    ", settings: nil},
		{source: "\nfoo.forEach(function () {\n  console.log(this);\n  z(x => console.log(x, this));\n}, obj);\n    ", settings: nil},
		{source: "\nfoo.map(function () {\n  console.log(this);\n  z(x => console.log(x, this));\n}, obj);\n    ", settings: nil},
		{source: "\nfoo.some(function () {\n  console.log(this);\n  z(x => console.log(x, this));\n}, obj);\n    ", settings: nil},
		{source: "\n/** @this Obj */ function foo() {\n  console.log(this);\n  z(x => console.log(x, this));\n}\n    ", settings: nil},
		{source: "\nfoo(\n  /* @this Obj */ function () {\n    console.log(this);\n    z(x => console.log(x, this));\n  },\n);\n    ", settings: nil},
		{source: "\n/**\n * @returns {void}\n * @this Obj\n */\nfunction foo() {\n  console.log(this);\n  z(x => console.log(x, this));\n}\n    ", settings: nil},
		{source: "\nCtor = function () {\n  console.log(this);\n  z(x => console.log(x, this));\n};\n    ", settings: nil},
		{source: "\nfunction foo(\n  Ctor = function () {\n    console.log(this);\n    z(x => console.log(x, this));\n  },\n) {}\n    ", settings: nil},
		{source: "\n[\n  obj.method = function () {\n    console.log(this);\n    z(x => console.log(x, this));\n  },\n] = a;\n    ", settings: nil},
		{source: "\nclass A {\n  static foo() {\n    console.log(this);\n    z(x => console.log(x, this));\n  }\n}\n    ", settings: nil},
		{source: "\nclass A {\n  a = 5;\n  b = this.a;\n  accessor c = this.a;\n}\n    ", settings: nil},
	}

	for _, testCase := range cases {
		t.Run(strings.TrimSpace(firstLineOf(testCase.source)), func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.RunWithOptions(t, NoInvalidThis, invalidThisFile,
				testCase.source, testCase.settings))
		})
	}
}

// firstLineOf names a subtest after the first non-empty line of its fixture, because the fixtures
// are multi-line templates and a numeric name says nothing when one fails.
func firstLineOf(source string) string {
	for _, line := range strings.Split(source, "\n") {
		if strings.TrimSpace(line) != "" {
			return line
		}
	}
	return "empty"
}

// The cases below are not in the corpus. Each was measured against the installed
// @typescript-eslint 8.67.0 rule through the ESLint Linter interface before being written down, and
// each covers a decision the corpus leaves unpinned. The JSDoc block is the largest group because
// the tag lookup is the biggest piece of borrowed machinery in this rule and the corpus exercises
// exactly one of its shapes.
func TestNoInvalidThisJsDocAttachment(t *testing.T) {
	t.Parallel()

	silent := []struct {
		name   string
		source string
	}{
		{
			name:   "a callback's own leading comment",
			source: "z(/* @this Obj */ function () {\n  this.x;\n});",
		},
		{
			name:   "a line comment carries the tag",
			source: "// @this Obj\nfunction bar() {\n  this.x;\n}",
		},
		{
			name:   "the tag comes first of two comments",
			source: "/* @this Obj */ /* second */ function bar() {\n  this.x;\n}",
		},
		{
			name:   "the tag comes second of two comments",
			source: "/* second */ /* @this Obj */ function bar() {\n  this.x;\n}",
		},
		{
			name:   "the walk crosses a variable declaration",
			source: "function foo() {\n  /** @this Obj*/ var q = function bar() {\n    this.x;\n  };\n}",
		},
		{
			name:   "the walk crosses an if statement",
			source: "function foo() {\n  /** @this Obj*/ if (a) {\n    return function bar() {\n      this.x;\n    };\n  }\n}",
		},
		{
			name:   "the walk crosses a parenthesis",
			source: "function foo() {\n  /** @this Obj*/ return (function bar() {\n    this.x;\n  });\n}",
		},
		{
			name:   "the tag is a prefix of a longer word",
			source: "function foo() {\n  /** @thisx Obj*/ return function bar() {\n    this.x;\n  };\n}",
		},
		{
			name:   "a multi-line block with a leading asterisk",
			source: "function foo() {\n  /**\n   * @this Obj\n   */ return function bar() {\n    this.x;\n  };\n}",
		},
	}

	for _, testCase := range silent {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.RunWithOptions(t, NoInvalidThis, invalidThisFile,
				testCase.source, DefaultNoInvalidThisSettings()))
		})
	}

	// The other side of the same boundary. Each of these is one edit away from a case above and the
	// verdict moves, which is what makes the cases above evidence rather than coincidence.
	reports := []struct {
		name     string
		source   string
		findings int
	}{
		{
			name:     "the comment leads a different argument",
			source:   "z(/* @this Obj */ 1, function () {\n  this.x;\n});",
			findings: 1,
		},
		{
			name:     "a statement closes the gap before the return",
			source:   "function foo() {\n  /** @this Obj*/ const q = 1;\n  return function bar() {\n    this.x;\n  };\n}",
			findings: 1,
		},
		{
			name:     "the tag is not at the start of the line",
			source:   "function foo() {\n  /** @nothis Obj*/ return function bar() {\n    this.x;\n  };\n}",
			findings: 1,
		},
		{
			name:     "no comment at all",
			source:   "function foo() {\n  return function bar() {\n    this.x;\n  };\n}",
			findings: 1,
		},
	}

	for _, testCase := range reports {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunWithOptions(t, NoInvalidThis, invalidThisFile,
				testCase.source, DefaultNoInvalidThisSettings())
			expected := make([]string, testCase.findings)
			for position := range expected {
				expected[position] = "unexpectedThis"
			}
			rule_testing.ExpectFindings(t, result, expected...)
		})
	}
}

// typescript-go keeps parentheses as nodes and the parser upstream ports from does not, so every
// arm of the outward walk is reachable through a parenthesized form that upstream never sees. The
// corpus covers three of these; the rest would be silent divergences.
func TestNoInvalidThisSeesThroughParentheses(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		source string
	}{
		{
			name:   "a parenthesized call receiver",
			source: "(function () {\n  this.x;\n}).call(obj);",
		},
		{
			name:   "a parenthesized apply receiver",
			source: "(function () {\n  this.x;\n}).apply(obj);",
		},
		{
			name:   "a parenthesized bind receiver",
			source: "var foo = (function () {\n  this.x;\n}).bind(obj);",
		},
		{
			name:   "an immediately invoked function returning one",
			source: "obj.foo = (function () {\n  return function () {\n    this.x;\n  };\n})();",
		},
		{
			name:   "an immediately invoked arrow returning one",
			source: "obj.foo = (() =>\n  function () {\n    this.x;\n  })();",
		},
		{
			name:   "a parenthesized assignment target",
			source: "obj.foo = (function () {\n  this.x;\n});",
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.RunWithOptions(t, NoInvalidThis, invalidThisFile,
				testCase.source, DefaultNoInvalidThisSettings()))
		})
	}
}

// The option defaults to TRUE, so a decoder filling a zero-value struct inverts it. Routed through
// the exported decoder rather than by building the struct, because the inversion is exactly what a
// struct-built fixture cannot see.
func TestDecodeNoInvalidThisOptions(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		raw  string
		want bool
	}{
		// The wire shape here is the bare options object, not the array ESLint wraps it in: the
		// config layer unwraps before handing the payload to a rule's decoder. A first draft of
		// this test passed the ESLint spelling and failed on all three rows, which is exactly what
		// routing option fixtures through the exported decoder is for.
		{name: "an absent key keeps the default", raw: `{}`, want: true},
		{name: "an explicit true", raw: `{"capIsConstructor": true}`, want: true},
		{name: "an explicit false", raw: `{"capIsConstructor": false}`, want: false},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			decoded, err := DecodeNoInvalidThisOptions([]byte(testCase.raw))
			if err != nil {
				t.Fatalf("decode returned %v", err)
			}
			settings, ok := decoded.(noInvalidThisSettings)
			if !ok {
				t.Fatalf("decode returned %T", decoded)
			}
			if settings.CapIsConstructor != testCase.want {
				t.Fatalf("capIsConstructor was %v, want %v",
					settings.CapIsConstructor, testCase.want)
			}
		})
	}

	// A rule configured as a bare "error" is handed nil, which never reaches the decoder at all.
	// The fallback has to produce the documented default rather than the zero value, and a rule
	// that got this wrong would report every capitalized constructor function in the tree.
	t.Run("nil options fall back to the default", func(t *testing.T) {
		t.Parallel()
		rule_testing.ExpectClean(t, rule_testing.RunWithOptions(t, NoInvalidThis, invalidThisFile,
			"function Foo() {\n  this.x = 1;\n}", nil))
	})
}

// The span is the four characters of `this` and nothing else, which no message-id assertion can see.
// Asserted against a literal rather than against the rule's own constant, because a constant
// compared to itself moves under mutation.
func TestNoInvalidThisSpansTheKeyword(t *testing.T) {
	t.Parallel()

	const source = "function foo() {\n  console.log(this);\n}\n"
	result := rule_testing.RunWithOptions(t, NoInvalidThis, invalidThisFile, source,
		DefaultNoInvalidThisSettings())
	rule_testing.ExpectFindings(t, result, "unexpectedThis")

	reported := source[result.Diagnostics[0].Range.Pos():result.Diagnostics[0].Range.End()]
	if reported != "this" {
		t.Fatalf("reported span was %q", reported)
	}
}

// The message carries no format verbs, so there is nothing to render and the assertion is on the
// value itself.
func TestNoInvalidThisMessage(t *testing.T) {
	t.Parallel()

	if messageUnexpectedThis.Id != "unexpectedThis" {
		t.Fatalf("message id was %q", messageUnexpectedThis.Id)
	}
	if !strings.HasPrefix(messageUnexpectedThis.Description,
		"This function is called rather than constructed") {
		t.Fatalf("message description was %q", messageUnexpectedThis.Description)
	}
}

// An arrow inherits the enclosing binding rather than introducing one, which is why the walk pushes
// nothing for it. Both `this` reads below resolve against the same enclosing function, so a port
// that pushed for arrows would report the outer one and silently exempt the inner.
func TestNoInvalidThisArrowsInheritTheBinding(t *testing.T) {
	t.Parallel()

	result := rule_testing.RunWithOptions(t, NoInvalidThis, invalidThisFile,
		"function foo() {\n  console.log(this);\n  z(x => console.log(x, this));\n}",
		DefaultNoInvalidThisSettings())
	rule_testing.ExpectFindings(t, result, "unexpectedThis", "unexpectedThis")

	// And the same two reads inside a bound function are both silent, for the same reason.
	rule_testing.ExpectClean(t, rule_testing.RunWithOptions(t, NoInvalidThis, invalidThisFile,
		"class A {\n  foo() {\n    console.log(this);\n    z(x => console.log(x, this));\n  }\n}",
		DefaultNoInvalidThisSettings()))
}

// A receiver of `null` or `undefined` binds nothing, so the three thisArg shapes report even though
// the argument is present in the right position. The corpus writes only real receivers, so all three
// nullish checks were surviving mutants until these landed; each pair below was measured against the
// installed rule at 8.67.0 before it was written.
func TestNoInvalidThisNullishReceiversDoNotBind(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		source   string
		findings int
	}{
		{
			name:     "Reflect.apply with a real receiver",
			source:   "Reflect.apply(function () {\n  this.x;\n}, obj, []);",
			findings: 0,
		},
		{
			name:     "Reflect.apply with null",
			source:   "Reflect.apply(function () {\n  this.x;\n}, null, []);",
			findings: 1,
		},
		{
			name:     "Reflect.apply with undefined",
			source:   "Reflect.apply(function () {\n  this.x;\n}, undefined, []);",
			findings: 1,
		},
		{
			name:     "Array.from with a real receiver",
			source:   "Array.from([], function () {\n  this.x;\n}, obj);",
			findings: 0,
		},
		{
			name:     "Array.from with null",
			source:   "Array.from([], function () {\n  this.x;\n}, null);",
			findings: 1,
		},
		{
			name:     "Array.from with undefined",
			source:   "Array.from([], function () {\n  this.x;\n}, undefined);",
			findings: 1,
		},
		// fromAsync shares Array.from's arm and its arity, and nothing in the corpus reaches it.
		{
			name:     "Array.fromAsync with a real receiver",
			source:   "Array.fromAsync([], function () {\n  this.x;\n}, obj);",
			findings: 0,
		},
		{
			name:     "Array.fromAsync with null",
			source:   "Array.fromAsync([], function () {\n  this.x;\n}, null);",
			findings: 1,
		},
		{
			name:     "a thisArg method with null",
			source:   "foo.forEach(function () {\n  this.x;\n}, null);",
			findings: 1,
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunWithOptions(t, NoInvalidThis, invalidThisFile,
				testCase.source, DefaultNoInvalidThisSettings())
			expected := make([]string, testCase.findings)
			for position := range expected {
				expected[position] = "unexpectedThis"
			}
			rule_testing.ExpectFindings(t, result, expected...)
		})
	}
}

// The constructor-name test is upstream's `name[0] !== name[0].toLocaleLowerCase()`, which asks
// "does this character have a different lowercase form" rather than "is this character uppercase".
// The two answers diverge on TITLECASE letters, and the shelf's react.IsLikelyComponentName uses the
// IsUpper form, which is why it is the wrong helper here and this rule spells the test out.
//
// U+01C5 is the distinguishing character: unicode.ToLower moves it, so upstream reads it as a
// constructor name and stays silent, while unicode.IsUpper answers false, so the IsUpper spelling
// reports. Measured against the installed rule at 8.67.0, which is silent on both shapes below.
// Written as an escape so this file stays pure ASCII on disk.
func TestNoInvalidThisTitlecaseNamesReadAsConstructors(t *testing.T) {
	t.Parallel()

	const titlecase = "\u01C5"

	rule_testing.ExpectClean(t, rule_testing.RunWithOptions(t, NoInvalidThis, invalidThisFile,
		"function "+titlecase+"oo() {\n  this.x;\n}", DefaultNoInvalidThisSettings()))
	rule_testing.ExpectClean(t, rule_testing.RunWithOptions(t, NoInvalidThis, invalidThisFile,
		"var "+titlecase+"oo = function () {\n  this.x;\n};", DefaultNoInvalidThisSettings()))

	// The controls: a lowercase name reports in both positions, so the silence above is the name
	// test answering rather than the fixture failing to reach the rule.
	rule_testing.ExpectFindings(t, rule_testing.RunWithOptions(t, NoInvalidThis, invalidThisFile,
		"function foo() {\n  this.x;\n}", DefaultNoInvalidThisSettings()), "unexpectedThis")
	rule_testing.ExpectFindings(t, rule_testing.RunWithOptions(t, NoInvalidThis, invalidThisFile,
		"var foo = function () {\n  this.x;\n};", DefaultNoInvalidThisSettings()), "unexpectedThis")
}

// The three survivors that the corpus could not see, each pinned by the input that separates the
// rule from the mutation. All measured against the installed rule at 8.67.0.
func TestNoInvalidThisBindingShapesBeyondTheCorpus(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		source   string
		findings int
	}{
		// A computed member target binds a receiver exactly as a dotted one does. Upstream tests
		// `parent.left.type === 'MemberExpression'`, which covers both spellings; typescript-go
		// splits them into two kinds, so dropping either half is a silent divergence.
		{
			name:     "a string-subscript assignment target",
			source:   "obj['foo'] = function () {\n  this.x;\n};",
			findings: 0,
		},
		{
			name:     "a numeric-subscript assignment target",
			source:   "obj[0] = function () {\n  this.x;\n};",
			findings: 0,
		},
		// A property merely NAMED bind is not a call to Function.prototype.bind. It is silent for a
		// different reason than it looks: the parent is a PropertyAssignment rather than a property
		// access, so the bind arm is never reached at all and the property-value arm answers.
		{
			name:     "a property named bind is not a bind call",
			source:   "var q = {\n  bind: function () {\n    this.x;\n  },\n};",
			findings: 0,
		},
		// A property binds only its VALUE, and a computed key is not one. Both spellings below
		// report by reaching the walk's default arm rather than through the property arm, because
		// typescript-go wraps a computed key in its own node. See the rule's PropertyAssignment
		// case for the measurement.
		{
			name:     "a function as a property value",
			source:   "var obj = {\n  foo: function () {\n    this.x;\n  },\n};",
			findings: 0,
		},
		{
			name:     "a function in a computed property key",
			source:   "var obj = {\n  [function () {\n    this.x;\n  }]: 1,\n};",
			findings: 1,
		},
		{
			name:     "a function as a class field value",
			source:   "class A {\n  foo = function () {\n    this.x;\n  };\n}",
			findings: 0,
		},
		{
			name:     "a function in a computed class field key",
			source:   "class A {\n  [function () {\n    this.x;\n  }]: number = 1;\n}",
			findings: 1,
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunWithOptions(t, NoInvalidThis, invalidThisFile,
				testCase.source, DefaultNoInvalidThisSettings())
			expected := make([]string, testCase.findings)
			for position := range expected {
				expected[position] = "unexpectedThis"
			}
			rule_testing.ExpectFindings(t, result, expected...)
		})
	}
}
