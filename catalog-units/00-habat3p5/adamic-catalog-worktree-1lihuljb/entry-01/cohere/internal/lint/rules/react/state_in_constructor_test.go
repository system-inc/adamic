package react

import (
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/internal/lint/testing"
)

// stateInConstructorFile is where the fixtures pretend to live.
//
// A .tsx extension because most corpus cases return JSX from a render method and would not parse
// otherwise. The rule itself has no suffix gate, which the suffix cases below pin by writing
// reporting source with no JSX in it to all four extensions.
const stateInConstructorFile = "/repository/source/StateInConstructor.tsx"

// The corpus is upstream's, extracted mechanically rather than retyped.
//
// Every case in the two tables below comes from
// /tmp/lint-sources/eslint-plugin-react/tests/lib/rules/state-in-constructor.js, read by evaluating
// the two arrays in the tester and emitting Go raw strings from the decoded values, so no escape
// sequence was ever typed on the way here. The upstream file carries 19 valid and 8 invalid cases,
// each invalid one naming exactly one messageId, which is 8 findings from 8 inputs.
//
// Upstream's "features: ['class fields']" markers carry no meaning for us; class fields are
// ordinary syntax to this parser, so the marker is dropped rather than modelled.
//
// The option is upstream's bare string, and every option fixture routes through the exported
// decoder in upstream's spelling, so the decoder is under test rather than bypassed.

// TestStateInConstructorFires runs the eight failing cases from upstream, one finding each.
//
// The two arms are exercised by the same table because the option is what selects which one runs,
// and the corpus makes that the point rather than an edge: three of these inputs are byte-identical
// to cases in the valid table, separated only by the option.
func TestStateInConstructorFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		mode       StateInConstructorMode
		sourceText string
		wantIds    []string
	}{
		{"invalid 0 this.state in constructor under never", StateInConstructorNever, `
        class Foo extends React.Component {
          constructor(props) {
            super(props)
            this.state = { bar: 0 }
          }
          render() {
            return <div>Foo</div>
          }
        }
      `, []string{"stateInitClassProp"}},
		{"invalid 1 this.state in constructor plus baz property under never", StateInConstructorNever, `
        class Foo extends React.Component {
          constructor(props) {
            super(props)
            this.state = { bar: 0 }
          }
          baz = { bar: 0 }
          render() {
            return <div>Foo</div>
          }
        }
      `, []string{"stateInitClassProp"}},
		{"invalid 2 state property by default", StateInConstructorAlways, `
        class Foo extends React.Component {
          state = { bar: 0 }
          render() {
            return <div>Foo</div>
          }
        }
      `, []string{"stateInitConstructor"}},
		{"invalid 3 state property plus baz property by default", StateInConstructorAlways, `
        class Foo extends React.Component {
          state = { bar: 0 }
          baz = { bar: 0 }
          render() {
            return <div>Foo</div>
          }
        }
      `, []string{"stateInitConstructor"}},
		{"invalid 4 this.baz in constructor plus state property by default", StateInConstructorAlways, `
        class Foo extends React.Component {
          constructor(props) {
            super(props)
            this.baz = { bar: 0 }
          }
          state = { baz: 0 }
          render() {
            return <div>Foo</div>
          }
        }
      `, []string{"stateInitConstructor"}},
		{"invalid 5 both styles at once by default", StateInConstructorAlways, `
        class Foo extends React.Component {
          constructor(props) {
            super(props)
            this.state = { bar: 0 }
          }
          state = { baz: 0 }
          render() {
            return <div>Foo</div>
          }
        }
      `, []string{"stateInitConstructor"}},
		{"invalid 6 both styles at once under never", StateInConstructorNever, `
        class Foo extends React.Component {
          constructor(props) {
            super(props)
            this.state = { bar: 0 }
          }
          state = { baz: 0 }
          render() {
            return <div>Foo</div>
          }
        }
      `, []string{"stateInitClassProp"}},
		{"invalid 7 this.state inside if in constructor under never", StateInConstructorNever, `
        class Foo extends React.Component {
          constructor(props) {
            super(props)
            if (foobar) {
              this.state = { bar: 0 }
            }
          }
          render() {
            return <div>Foo</div>
          }
        }
      `, []string{"stateInitClassProp"}},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := runStateInConstructor(t, testCase.mode, testCase.sourceText)
			rule_testing.ExpectFindings(t, result, testCase.wantIds...)
		})
	}
}

// TestStateInConstructorStaysSilent runs the nineteen passing cases from upstream.
//
// These are the false positives upstream already thought about. The most load-bearing members are
// the ones holding a property named something other than state beside a real state initialization,
// which is how the corpus says the rule keys on the name rather than on the presence of any
// property at all.
func TestStateInConstructorStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		mode       StateInConstructorMode
		sourceText string
	}{
		{"valid 0 render only by default", StateInConstructorAlways, `
        class Foo extends React.Component {
          render() {
            return <div>Foo</div>
          }
        }
      `},
		{"valid 1 render only under never", StateInConstructorNever, `
        class Foo extends React.Component {
          render() {
            return <div>Foo</div>
          }
        }
      `},
		{"valid 2 this.state in constructor by default", StateInConstructorAlways, `
        class Foo extends React.Component {
          constructor(props) {
            super(props)
            this.state = { bar: 0 }
          }
          render() {
            return <div>Foo</div>
          }
        }
      `},
		{"valid 3 this.state in constructor under always", StateInConstructorAlways, `
        class Foo extends React.Component {
          constructor(props) {
            super(props)
            this.state = { bar: 0 }
          }
          render() {
            return <div>Foo</div>
          }
        }
      `},
		{"valid 4 this.state in constructor plus baz property by default", StateInConstructorAlways, `
        class Foo extends React.Component {
          constructor(props) {
            super(props)
            this.state = { bar: 0 }
          }
          baz = { bar: 0 }
          render() {
            return <div>Foo</div>
          }
        }
      `},
		{"valid 5 this.baz in constructor by default", StateInConstructorAlways, `
        class Foo extends React.Component {
          constructor(props) {
            super(props)
            this.baz = { bar: 0 }
          }
          render() {
            return <div>Foo</div>
          }
        }
      `},
		{"valid 6 this.baz in constructor under never", StateInConstructorNever, `
        class Foo extends React.Component {
          constructor(props) {
            super(props)
            this.baz = { bar: 0 }
          }
          render() {
            return <div>Foo</div>
          }
        }
      `},
		{"valid 7 baz property only by default", StateInConstructorAlways, `
        class Foo extends React.Component {
          baz = { bar: 0 }
          render() {
            return <div>Foo</div>
          }
        }
      `},
		{"valid 8 baz property only under never", StateInConstructorNever, `
        class Foo extends React.Component {
          baz = { bar: 0 }
          render() {
            return <div>Foo</div>
          }
        }
      `},
		{"valid 9 arrow component by default", StateInConstructorAlways, `
        const Foo = () => <div>Foo</div>
      `},
		{"valid 10 arrow component under never", StateInConstructorNever, `
        const Foo = () => <div>Foo</div>
      `},
		{"valid 11 function component by default", StateInConstructorAlways, `
        function Foo () {
          return <div>Foo</div>
        }
      `},
		{"valid 12 function component under never", StateInConstructorNever, `
        function Foo () {
          return <div>Foo</div>
        }
      `},
		{"valid 13 state property under never", StateInConstructorNever, `
        class Foo extends React.Component {
          state = { bar: 0 }
          render() {
            return <div>Foo</div>
          }
        }
      `},
		{"valid 14 state property plus baz property under never", StateInConstructorNever, `
        class Foo extends React.Component {
          state = { bar: 0 }
          baz = { bar: 0 }
          render() {
            return <div>Foo</div>
          }
        }
      `},
		{"valid 15 this.baz in constructor plus state property under never", StateInConstructorNever, `
        class Foo extends React.Component {
          constructor(props) {
            super(props)
            this.baz = { bar: 0 }
          }
          state = { baz: 0 }
          render() {
            return <div>Foo</div>
          }
        }
      `},
		{"valid 16 this.state inside if in constructor by default", StateInConstructorAlways, `
        class Foo extends React.Component {
          constructor(props) {
            super(props)
            if (foobar) {
              this.state = { bar: 0 }
            }
          }
          render() {
            return <div>Foo</div>
          }
        }
      `},
		{"valid 17 bare foobar assignment in constructor by default", StateInConstructorAlways, `
        class Foo extends React.Component {
          constructor(props) {
            super(props)
            foobar = { bar: 0 }
          }
          render() {
            return <div>Foo</div>
          }
        }
      `},
		{"valid 18 bare foobar assignment in constructor under never", StateInConstructorNever, `
        class Foo extends React.Component {
          constructor(props) {
            super(props)
            foobar = { bar: 0 }
          }
          render() {
            return <div>Foo</div>
          }
        }
      `},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := runStateInConstructor(t, testCase.mode, testCase.sourceText)
			rule_testing.ExpectClean(t, result)
		})
	}
}

// runStateInConstructor drives the rule through its own exported decoder.
//
// Routing through DecodeStateInConstructorOptions rather than building the options struct directly
// is what puts the decoder under test: each mode reaches the rule as upstream writes it, `"always"`
// or `"never"`. A fixture handing RunWithOptions a struct would leave the decoder untested.
func runStateInConstructor(t *testing.T, mode StateInConstructorMode, sourceText string) rule_testing.Result {
	t.Helper()
	decoded, err := DecodeStateInConstructorOptions([]byte(`"` + strings.ToLower(string(mode)) + `"`))
	if err != nil {
		t.Fatalf("decoding mode %q: %v", mode, err)
	}
	return rule_testing.RunWithOptions(t, StateInConstructor, stateInConstructorFile, sourceText, decoded)
}

// TestStateInConstructorReportsWithoutOptions pins the nil-options fallback.
//
// A rule configured as bare "error" is handed nil, which type-asserts to the zero value whose Mode
// matches neither arm. Without the explicit fallback in Run the rule would register, pass every
// fixture above, and report nothing on the real tree, which is the wired-and-inert failure the
// brief describes. Every other fixture here reaches the rule through the decoder, so nothing else
// in this file can see it.
func TestStateInConstructorReportsWithoutOptions(t *testing.T) {
	t.Parallel()

	source := `
class Foo extends React.Component {
  state = { bar: 0 };
  render() { return null; }
}
`
	result := rule_testing.Run(t, StateInConstructor, stateInConstructorFile, source)
	rule_testing.ExpectFindings(t, result, "stateInitConstructor")
}

// TestStateInConstructorHasNoFileSuffixGate pins the absence of the gate three siblings carry.
//
// Those siblings were ported from oxc, which gates the whole rule on the file being read as JSX.
// The authority for this rule is `eslint-plugin-react`, which has no such gate anywhere. Measured
// by driving the installed build (7.37.5) over this exact source under all four suffixes: every one
// reported. The source deliberately holds no JSX, so a `.ts` file parses and the parser's opinion
// cannot be confused with the rule's.
func TestStateInConstructorHasNoFileSuffixGate(t *testing.T) {
	t.Parallel()

	source := `
class Foo extends React.Component {
  state = { bar: 0 };
  render() { return null; }
}
`
	for _, fileName := range []string{
		"/repository/source/Suffix.tsx",
		"/repository/source/Suffix.jsx",
		"/repository/source/Suffix.ts",
		"/repository/source/Suffix.js",
	} {
		t.Run(fileName, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, StateInConstructor, fileName, source)
			rule_testing.ExpectFindings(t, result, "stateInitConstructor")
		})
	}
}

// TestStateInConstructorKeyShapes pins the four key shapes the corpus never writes.
//
// Upstream's predicate is `node.key.name === 'state'`, which depends on which key shapes the parser
// gives a `.name` field. Our parser disagrees with ESLint's in two opposite directions, so both
// halves of this table are places a text-only comparison would ship a defect green: silent on
// `#state` where upstream reports, and reporting on `"state"` where upstream is silent. Every
// verdict below was measured by driving the installed rule.
func TestStateInConstructorKeyShapes(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		member  string
		wantIds []string
	}{
		// The plain identifier, as a control: if this stops reporting the table measures nothing.
		{"identifier key", `state = { bar: 0 };`, []string{"stateInitConstructor"}},

		// Excluded by upstream's explicit `!node.static`.
		{"static", `static state = { bar: 0 };`, nil},

		// A computed key is a Literal with no `.name` upstream, and a KindComputedPropertyName here.
		{"computed string key", `["state"] = { bar: 0 };`, nil},

		// A string-literal key is also a Literal with no `.name` upstream, but our parser hands
		// back the bare text `state`, so a text-only comparison would invent a finding here.
		{"string literal key", `"state" = { bar: 0 };`, nil},

		// A PrivateIdentifier's `.name` upstream is the bare `state`, so upstream reports. Our
		// Text() carries the hash, so a text-only comparison would lose the finding. Upstream
		// reporting here is a defect it does not know it has; it is reproduced rather than fixed.
		{"private name key", `#state = { bar: 0 };`, []string{"stateInitConstructor"}},

		// The rule never looks at the initializer, so a bare declaration reports.
		{"no initializer", `state;`, []string{"stateInitConstructor"}},

		// Some other property is not this rule's business.
		{"other name", `baz = { bar: 0 };`, nil},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			source := "\nclass Foo extends React.Component {\n  " + testCase.member +
				"\n  render() { return null; }\n}\n"
			result := runStateInConstructor(t, StateInConstructorAlways, source)
			rule_testing.ExpectFindings(t, result, testCase.wantIds...)
		})
	}
}

// TestStateInConstructorNeverTargets pins which assignment targets the Never arm accepts.
//
// Upstream's `isStateMemberExpression` requires a member expression whose object is a
// `ThisExpression` and whose property is named `state`, so the target has to be exactly
// `this.state`. The parenthesized rows are the ones our tree has to handle explicitly, because
// ESLint's AST does not materialize a parenthesized expression in either position and ours does;
// without the unwrapping both would go silent on inputs upstream reports. Every verdict measured
// on the installed rule, and none of these shapes is in the corpus.
func TestStateInConstructorNeverTargets(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name      string
		statement string
		wantIds   []string
	}{
		{"direct this.state", `this.state = { bar: 0 };`, []string{"stateInitClassProp"}},
		{"compound operator", `this.state += 1;`, []string{"stateInitClassProp"}},
		{"parenthesized target", `(this.state) = {};`, []string{"stateInitClassProp"}},
		{"parenthesized receiver", `(this).state = {};`, []string{"stateInitClassProp"}},

		// The target is `this.state.x`, whose object is `this.state` rather than `this`.
		{"nested property write", `this.state.x = 1;`, nil},

		// An element access is not a member expression carrying a name.
		{"computed this access", `this["state"] = {};`, nil},

		// Some other property on this.
		{"other property", `this.baz = {};`, nil},

		// Upstream's `inConstructor` walks the SCOPE chain, so a callback declared in the
		// constructor is still in the constructor. Measured: this reports upstream.
		{"inside a callback", `go(() => { this.state = {}; });`, []string{"stateInitClassProp"}},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			source := "\nclass Foo extends React.Component {\n" +
				"  constructor(props) {\n    super(props);\n    " + testCase.statement +
				"\n  }\n  render() { return null; }\n}\n"
			result := runStateInConstructor(t, StateInConstructorNever, source)
			rule_testing.ExpectFindings(t, result, testCase.wantIds...)
		})
	}
}

// TestStateInConstructorNeverRequiresTheConstructor pins that the Never arm is constructor-only.
//
// A `this.state` assignment in any other method is silent upstream, which is what makes this rule
// about initialization style rather than about state mutation. `no-direct-mutation-state` is the
// rule that watches the other methods.
func TestStateInConstructorNeverRequiresTheConstructor(t *testing.T) {
	t.Parallel()

	source := `
class Foo extends React.Component {
  someMethod() { this.state = { bar: 0 }; }
  render() { return null; }
}
`
	result := runStateInConstructor(t, StateInConstructorNever, source)
	rule_testing.ExpectClean(t, result)
}

// TestStateInConstructorNearestClassDecides pins that an inner class is not rescued by an outer one.
//
// Upstream's `getParentES6Component` walks to the NEAREST class scope and asks whether that class
// is a component, so a plain class nested inside a real component keeps its own verdict. Reaching
// for `react.EnclosingComponent` instead would skip past the plain class to the outer component and
// report on both of these, which is why this rule walks to the nearest class itself. Both measured
// silent on the installed rule.
func TestStateInConstructorNearestClassDecides(t *testing.T) {
	t.Parallel()

	t.Run("plain class inside a component, Always", func(t *testing.T) {
		t.Parallel()
		source := `
class Foo extends React.Component {
  render() {
    class Inner { state = { bar: 0 }; }
    return null;
  }
}
`
		result := runStateInConstructor(t, StateInConstructorAlways, source)
		rule_testing.ExpectClean(t, result)
	})

	t.Run("plain class inside a component constructor, Never", func(t *testing.T) {
		t.Parallel()
		// Silent because the NEAREST class is Inner, and Inner is not a component. This is the
		// component gate declining, not the constructor walk: see TestStateInConstructorWalkCrosses
		// a class boundary for the proof that the walk itself goes straight through.
		source := `
class Foo extends React.Component {
  constructor(props) {
    super(props);
    class Inner { constructor() { this.state = {}; } }
  }
  render() { return null; }
}
`
		result := runStateInConstructor(t, StateInConstructorNever, source)
		rule_testing.ExpectClean(t, result)
	})

	t.Run("component class inside a component reports on its own", func(t *testing.T) {
		t.Parallel()
		source := `
class Foo extends React.Component {
  render() {
    class Inner extends React.Component { state = { bar: 0 }; }
    return null;
  }
}
`
		result := runStateInConstructor(t, StateInConstructorAlways, source)
		rule_testing.ExpectFindings(t, result, "stateInitConstructor")
	})
}

// TestStateInConstructorComponentGate pins which classes count as components.
//
// The heritage clause is what decides, and the parenthesized row is the one worth stating: the
// shelf helper skips parentheses on the heritage receiver, and its doc comment flags that skip as a
// possible divergence measured on a different rule. Measured against THIS rule's authority,
// `extends (React.Component)` reports, so the skip is correct here rather than a divergence.
func TestStateInConstructorComponentGate(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		heritage string
		wantIds  []string
	}{
		{"React.Component", ` extends React.Component`, []string{"stateInitConstructor"}},
		{"React.PureComponent", ` extends React.PureComponent`, []string{"stateInitConstructor"}},
		{"bare Component", ` extends Component`, []string{"stateInitConstructor"}},
		{"bare PureComponent", ` extends PureComponent`, []string{"stateInitConstructor"}},
		{"parenthesized base", ` extends (React.Component)`, []string{"stateInitConstructor"}},
		{"doubly parenthesized base", ` extends ((React.Component))`, []string{"stateInitConstructor"}},
		{"parenthesized receiver", ` extends (React).Component`, []string{"stateInitConstructor"}},
		{"parenthesized bare base", ` extends (Component)`, []string{"stateInitConstructor"}},

		// The parenthesized path restates the base-name set rather than reusing the shelf's
		// unexported one, so PureComponent needs its own case on that path. A mutant dropping
		// PureComponent from the local set survived every other fixture here, because the
		// unparenthesized spellings never reach it.
		{"parenthesized qualified PureComponent", ` extends (React.PureComponent)`, []string{"stateInitConstructor"}},
		{"parenthesized bare PureComponent", ` extends (PureComponent)`, []string{"stateInitConstructor"}},

		// An implements clause names an interface, never a base class, and a component that
		// implements something React-shaped without extending it is not a component. Our parser
		// gives an implements clause KindTypeReference rather than KindExpressionWithTypeArguments,
		// so the kind test would decline it anyway; the token test says which question is being
		// asked instead of resting on that. All three measured silent upstream.
		{"implements a qualified component", ` implements React.Component`, nil},
		{"implements a bare component", ` implements Component`, nil},
		{"extends something else while implementing a component", ` extends Base implements Component`, nil},
		{"parenthesized non-component base", ` extends (Foo.Component)`, nil},
		{"parenthesized non-component bare", ` extends (Whatever)`, nil},
		{"no heritage", ``, nil},
		{"somebody else's Component", ` extends Foo.Component`, nil},
		{"a mixin call", ` extends mixin(React.Component)`, nil},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			source := "\nclass Foo" + testCase.heritage +
				" {\n  state = { bar: 0 };\n  render() { return null; }\n}\n"
			result := runStateInConstructor(t, StateInConstructorAlways, source)
			rule_testing.ExpectFindings(t, result, testCase.wantIds...)
		})
	}
}

// TestStateInConstructorClassExpression pins that a class expression counts.
//
// `const Foo = class extends React.Component {}` is the same component written where a declaration
// will not fit, and upstream's scope walk finds a class scope either way. Measured reporting on the
// installed rule in both arms.
func TestStateInConstructorClassExpression(t *testing.T) {
	t.Parallel()

	t.Run("Always", func(t *testing.T) {
		t.Parallel()
		source := `
const Foo = class extends React.Component {
  state = { bar: 0 };
  render() { return null; }
};
`
		result := runStateInConstructor(t, StateInConstructorAlways, source)
		rule_testing.ExpectFindings(t, result, "stateInitConstructor")
	})

	t.Run("Never", func(t *testing.T) {
		t.Parallel()
		source := `
const Foo = class extends React.Component {
  constructor(props) { super(props); this.state = {}; }
  render() { return null; }
};
`
		result := runStateInConstructor(t, StateInConstructorNever, source)
		rule_testing.ExpectFindings(t, result, "stateInitClassProp")
	})
}

// TestStateInConstructorSpans asserts where each finding points.
//
// Upstream passes the node itself to `report` in both arms, so the Always arm underlines the whole
// property declaration and the Never arm underlines the whole assignment expression, right-hand
// side included. A message-id assertion cannot see any of that. The reported text is sliced out of
// the source the harness actually wrote rather than out of the Go literal, because `rule_testing`
// trims the fixture and a slice of the literal would be off by one byte.
func TestStateInConstructorSpans(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		mode       StateInConstructorMode
		sourceText string
		wantText   string
	}{
		{
			"Always underlines the whole property including its initializer",
			StateInConstructorAlways,
			"\nclass Foo extends React.Component {\n  state = { bar: 0 };\n  render() { return null; }\n}\n",
			"state = { bar: 0 };",
		},
		{
			"Always underlines a property with no initializer",
			StateInConstructorAlways,
			"\nclass Foo extends React.Component {\n  state;\n  render() { return null; }\n}\n",
			"state;",
		},
		{
			"Never underlines the whole assignment including the right hand side",
			StateInConstructorNever,
			"\nclass Foo extends React.Component {\n  constructor(props) { super(props); this.state = { bar: 0 }; }\n  render() { return null; }\n}\n",
			"this.state = { bar: 0 }",
		},
		{
			"Never underlines a parenthesized target with its parentheses",
			StateInConstructorNever,
			"\nclass Foo extends React.Component {\n  constructor(props) { super(props); (this.state) = {}; }\n  render() { return null; }\n}\n",
			"(this.state) = {}",
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := runStateInConstructor(t, testCase.mode, testCase.sourceText)
			if len(result.Diagnostics) != 1 {
				t.Fatalf("wanted exactly one finding, got %d", len(result.Diagnostics))
			}
			// `rule_testing.Run` writes the source verbatim, unlike `RunTyped` which trims it, so
			// the literal here IS the file on disk and slicing it directly is correct. Asserting
			// that rather than assuming it: an accidental TrimSpace transform here shifts every
			// span by one byte and reads exactly like an off-by-one in the rule.
			diagnostic := result.Diagnostics[0]
			got := testCase.sourceText[diagnostic.Range.Pos():diagnostic.Range.End()]
			if got != testCase.wantText {
				t.Errorf("span text = %q, wanted %q", got, testCase.wantText)
			}
		})
	}
}

// TestStateInConstructorMessages asserts the two messages by identity.
//
// A rule.Message is {Id, Description} with no interpolation, so there is nothing to render and the
// assertion is on the constants themselves. Asserting the id against a literal string typed here
// rather than against the rule's own constant is deliberate: comparing a finding to the constant it
// was reported with is an equality that moves on both sides under mutation and cannot fail.
func TestStateInConstructorMessages(t *testing.T) {
	t.Parallel()

	if messageStateInConstructorInConstructor.Id != "stateInitConstructor" {
		t.Errorf("Always arm id = %q", messageStateInConstructorInConstructor.Id)
	}
	if messageStateInConstructorClassProperty.Id != "stateInitClassProp" {
		t.Errorf("Never arm id = %q", messageStateInConstructorClassProperty.Id)
	}
	for _, message := range []rule.Message{
		messageStateInConstructorInConstructor,
		messageStateInConstructorClassProperty,
	} {
		if len(message.Description) < 80 {
			t.Errorf("message %q has a description too short to say why: %q", message.Id, message.Description)
		}
	}
}

// TestDecodeStateInConstructorOptions pins the decoder against upstream's schema,
// `[{ enum: ["always", "never"] }]`.
func TestDecodeStateInConstructorOptions(t *testing.T) {
	t.Parallel()

	accepted := []struct {
		name string
		raw  string
		want StateInConstructorMode
	}{
		{"no options at all means Always", ``, StateInConstructorAlways},
		{"upstream's always", `"always"`, StateInConstructorAlways},
		{"upstream's never", `"never"`, StateInConstructorNever},
	}
	for _, testCase := range accepted {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			decoded, err := DecodeStateInConstructorOptions([]byte(testCase.raw))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if decoded.(StateInConstructorOptions).Mode != testCase.want {
				t.Errorf("mode = %q, wanted %q", decoded.(StateInConstructorOptions).Mode, testCase.want)
			}
		})
	}

	// The object form is the shape cohere invented before #d21war2 and no ESLint version accepts,
	// so it is refused in both of its old spellings. The rest are values upstream's enum refuses.
	refused := []struct {
		name string
		raw  string
	}{
		{"the old invented object", `{"mode":"Never"}`},
		{"the old invented object in upstream's casing", `{"mode":"never"}`},
		{"an empty object", `{}`},
		{"the old PascalCase spelling", `"Never"`},
		{"an unknown mode", `"sometimes"`},
		{"an empty string", `""`},
		{"null", `null`},
	}
	for _, testCase := range refused {
		t.Run(testCase.name+" is refused", func(t *testing.T) {
			t.Parallel()
			if _, err := DecodeStateInConstructorOptions([]byte(testCase.raw)); err == nil {
				t.Errorf("%s decoded; upstream's schema refuses it", testCase.raw)
			}
		})
	}
}

// TestStateInConstructorWalkCrossesAClassBoundary pins that the constructor walk does not stop at a
// class.
//
// Upstream's `inConstructor` walks the SCOPE chain, and a class body creates no barrier scope, so
// an enclosing constructor is reached straight through an intervening class. This rule stopped at
// the first class for a whole revision on the plausible reasoning that a nested class owns its own
// members. Nothing in the imported corpus could see it: the corpus writes no nested class at all.
// A mutation sweep removing the class-stop survived every fixture, and the survivor turned out to
// be the mutant being right rather than the fixtures being blind.
//
// The three rows are a measurement rather than a reading. All three were run against the installed
// build (eslint-plugin-react 7.37.5) through the ESLint Linter API. The first two have identical
// inner structure and differ only in whether the outer context is a constructor, which is what
// proves the walk crosses the class rather than something else explaining the verdict. The third
// keeps the walk from being vacuous by showing a constructor is still required somewhere.
func TestStateInConstructorWalkCrossesAClassBoundary(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		wantIds    []string
	}{
		{
			"component class in an outer constructor reports through the class",
			`
class Outer extends React.Component {
  constructor(props) {
    super(props);
    class Inner extends React.Component {
      someMethod() { this.state = {}; }
      render() { return null; }
    }
  }
  render() { return null; }
}
`,
			[]string{"stateInitClassProp"},
		},
		{
			"the same shape in an outer method is silent",
			`
class Outer extends React.Component {
  someMethod() {
    class Inner extends React.Component {
      m() { this.state = {}; }
      render() { return null; }
    }
  }
  render() { return null; }
}
`,
			nil,
		},
		{
			"a component method with no nesting at all is silent",
			`
class Outer extends React.Component {
  someMethod() { this.state = {}; }
  render() { return null; }
}
`,
			nil,
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := runStateInConstructor(t, StateInConstructorNever, testCase.sourceText)
			rule_testing.ExpectFindings(t, result, testCase.wantIds...)
		})
	}
}

// TestStateInConstructorNeverRequiresAnAssignment pins that a comparison is not a write.
//
// Upstream anchors the Never arm on `AssignmentExpression`, which in the ESLint AST is a node kind
// of its own. Ours is a binary expression covering comparison, logical and arithmetic operators as
// well, so the operator set is the only thing separating them and it is load-bearing rather than
// defensive. A mutation removing the operator test survived every other fixture in this file,
// because nothing here wrote `this.state` in a non-assignment position.
//
// Every row measured against the installed build: the assignment reports and none of the seven
// comparisons does.
func TestStateInConstructorNeverRequiresAnAssignment(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name      string
		statement string
		wantIds   []string
	}{
		// The control. If this stops reporting the table below measures nothing.
		{"an assignment", `this.state = {};`, []string{"stateInitClassProp"}},

		{"strict equality", `if (this.state === null) {}`, nil},
		{"strict inequality", `if (this.state !== null) {}`, nil},
		{"logical or", `const a = this.state || {};`, nil},
		{"logical and", `const b = this.state && 1;`, nil},
		{"addition", `const c = this.state + 1;`, nil},
		{"comma", `this.state, 1;`, nil},
		{"nullish coalescing", `this.state ?? {};`, nil},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			source := "\nclass Foo extends React.Component {\n" +
				"  constructor(props) {\n    super(props);\n    " + testCase.statement +
				"\n  }\n  render() { return null; }\n}\n"
			result := runStateInConstructor(t, StateInConstructorNever, source)
			rule_testing.ExpectFindings(t, result, testCase.wantIds...)
		})
	}
}
