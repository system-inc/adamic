package react

import (
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// noArrowFunctionLifecycleFile is where the fixtures pretend to live.
const noArrowFunctionLifecycleFile = "/repository/source/NoArrowFunctionLifecycle.tsx"

// The corpus is upstream's, extracted mechanically rather than retyped.
//
// Every case in the tables below comes from
// /tmp/lint-sources/eslint-plugin-react/tests/lib/rules/no-arrow-function-lifecycle.js, read by
// evaluating the two arrays in the tester with a stubbed RuleTester and emitting Go literals from
// the decoded values, so no escape sequence was typed on the way here. The sources and the fix
// outputs were then compared byte for byte against the decoded corpus with a script. Upstream
// carries 37 valid and 36 invalid cases, and every invalid case ships an `output`.
//
// All 73 were replayed against the installed build, eslint-plugin-react 7.37.5, through the ESLint
// Linter API before any Go was written, and it agreed on 72. The single disagreement is a harness
// artifact: upstream's `valid 36` writes a TypeScript type annotation that the default parser
// cannot read, so that run produced a parse error rather than a verdict. Our parser reads it, and
// it is carried at the clean verdict the corpus states.

// TestNoArrowFunctionLifecycleFires runs every reporting case and asserts the repair.
//
// The fix output is asserted for every case that ships one, which is what puts the fixer under
// test rather than only the judgment. `ExpectFixedSource` replays the fixes into the source and
// compares the whole rewritten file, so a repair with the right text over the wrong span fails
// here even though its message id is correct.
func TestNoArrowFunctionLifecycleFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		wantFixed  string
		wantCount  int
	}{
		{"invalid 0", `
        var Hello = createReactClass({
          render: () => { return <div />; }
        });
      `, `
        var Hello = createReactClass({
          render: function() { return <div />; }
        });
      `, 1},
		{"invalid 1", `
        var Hello = createReactClass({
          getDefaultProps: () => { return {}; },
          render: function() { return <div />; }
        });
      `, `
        var Hello = createReactClass({
          getDefaultProps: function() { return {}; },
          render: function() { return <div />; }
        });
      `, 1},
		{"invalid 2", `
        var Hello = createReactClass({
          getInitialState: () => { return {}; },
          render: function() { return <div />; }
        });
      `, `
        var Hello = createReactClass({
          getInitialState: function() { return {}; },
          render: function() { return <div />; }
        });
      `, 1},
		{"invalid 3", `
        var Hello = createReactClass({
          getChildContext: () => { return {}; },
          render: function() { return <div />; }
        });
      `, `
        var Hello = createReactClass({
          getChildContext: function() { return {}; },
          render: function() { return <div />; }
        });
      `, 1},
		{"invalid 4", `
        var Hello = createReactClass({
          componentWillMount: () => {},
          render: function() { return <div />; }
        });
      `, `
        var Hello = createReactClass({
          componentWillMount: function() {},
          render: function() { return <div />; }
        });
      `, 1},
		{"invalid 5", `
        var Hello = createReactClass({
          UNSAFE_componentWillMount: () => {},
          render: function() { return <div />; }
        });
      `, `
        var Hello = createReactClass({
          UNSAFE_componentWillMount: function() {},
          render: function() { return <div />; }
        });
      `, 1},
		{"invalid 6", `
        var Hello = createReactClass({
          componentDidMount: () => {},
          render: function() { return <div />; }
        });
      `, `
        var Hello = createReactClass({
          componentDidMount: function() {},
          render: function() { return <div />; }
        });
      `, 1},
		{"invalid 7", `
        var Hello = createReactClass({
          componentWillReceiveProps: () => {},
          render: function() { return <div />; }
        });
      `, `
        var Hello = createReactClass({
          componentWillReceiveProps: function() {},
          render: function() { return <div />; }
        });
      `, 1},
		{"invalid 8", `
        var Hello = createReactClass({
          UNSAFE_componentWillReceiveProps: () => {},
          render: function() { return <div />; }
        });
      `, `
        var Hello = createReactClass({
          UNSAFE_componentWillReceiveProps: function() {},
          render: function() { return <div />; }
        });
      `, 1},
		{"invalid 9", `
        var Hello = createReactClass({
          shouldComponentUpdate: () => { return true; },
          render: function() { return <div />; }
        });
      `, `
        var Hello = createReactClass({
          shouldComponentUpdate: function() { return true; },
          render: function() { return <div />; }
        });
      `, 1},
		{"invalid 10", `
        var Hello = createReactClass({
          componentWillUpdate: () => {},
          render: function() { return <div />; }
        });
      `, `
        var Hello = createReactClass({
          componentWillUpdate: function() {},
          render: function() { return <div />; }
        });
      `, 1},
		{"invalid 11", `
        var Hello = createReactClass({
          UNSAFE_componentWillUpdate: () => {},
          render: function() { return <div />; }
        });
      `, `
        var Hello = createReactClass({
          UNSAFE_componentWillUpdate: function() {},
          render: function() { return <div />; }
        });
      `, 1},
		{"invalid 12", `
        var Hello = createReactClass({
          getSnapshotBeforeUpdate: () => { return {}; },
          render: function() { return <div />; }
        });
      `, `
        var Hello = createReactClass({
          getSnapshotBeforeUpdate: function() { return {}; },
          render: function() { return <div />; }
        });
      `, 1},
		{"invalid 13", `
        var Hello = createReactClass({
          componentDidUpdate: () => {},
          render: function() { return <div />; }
        });
      `, `
        var Hello = createReactClass({
          componentDidUpdate: function() {},
          render: function() { return <div />; }
        });
      `, 1},
		{"invalid 14", `
        var Hello = createReactClass({
          componentDidCatch: () => {},
          render: function() { return <div />; }
        });
      `, `
        var Hello = createReactClass({
          componentDidCatch: function() {},
          render: function() { return <div />; }
        });
      `, 1},
		{"invalid 15", `
        var Hello = createReactClass({
          componentWillUnmount: () => {},
          render: function() { return <div />; }
        });
      `, `
        var Hello = createReactClass({
          componentWillUnmount: function() {},
          render: function() { return <div />; }
        });
      `, 1},
		{"invalid 16", `
        class Hello extends React.Component {
          handleEventMethods = () => {}
          render = () => { return <div />; }
        }
      `, `
        class Hello extends React.Component {
          handleEventMethods = () => {}
          render() { return <div />; }
        }
      `, 1},
		{"invalid 17", `
        class Hello extends React.Component {
          handleEventMethods = () => {}
          getDefaultProps = () => { return {}; }
          render = () => { return <div />; }
        }
      `, `
        class Hello extends React.Component {
          handleEventMethods = () => {}
          getDefaultProps() { return {}; }
          render() { return <div />; }
        }
      `, 2},
		{"invalid 18", `
        class Hello extends React.Component {
          handleEventMethods = () => {}
          getInitialState = () => { return {}; }
          render = () => { return <div />; }
        }
      `, `
        class Hello extends React.Component {
          handleEventMethods = () => {}
          getInitialState() { return {}; }
          render() { return <div />; }
        }
      `, 2},
		{"invalid 19", `
        class Hello extends React.Component {
          handleEventMethods = () => {}
          getChildContext = () => { return {}; }
          render = () => { return <div />; }
        }
      `, `
        class Hello extends React.Component {
          handleEventMethods = () => {}
          getChildContext() { return {}; }
          render() { return <div />; }
        }
      `, 2},
		{"invalid 20", `
        class Hello extends React.Component {
          handleEventMethods = () => {}
          static getDerivedStateFromProps = () => { return {}; }
          render = () => { return <div />; }
        }
      `, `
        class Hello extends React.Component {
          handleEventMethods = () => {}
          static getDerivedStateFromProps() { return {}; }
          render() { return <div />; }
        }
      `, 2},
		{"invalid 21", `
        class Hello extends React.Component {
          handleEventMethods = () => {}
          componentWillMount = () => {}
          render = () => { return <div />; }
        }
      `, `
        class Hello extends React.Component {
          handleEventMethods = () => {}
          componentWillMount() {}
          render() { return <div />; }
        }
      `, 2},
		{"invalid 22", `
        class Hello extends React.Component {
          handleEventMethods = () => {}
          UNSAFE_componentWillMount = () => {}
          render = () => { return <div />; }
        }
      `, `
        class Hello extends React.Component {
          handleEventMethods = () => {}
          UNSAFE_componentWillMount() {}
          render() { return <div />; }
        }
      `, 2},
		{"invalid 23", `
        class Hello extends React.Component {
          handleEventMethods = () => {}
          componentDidMount = () => {}
          render = () => { return <div />; }
        }
      `, `
        class Hello extends React.Component {
          handleEventMethods = () => {}
          componentDidMount() {}
          render() { return <div />; }
        }
      `, 2},
		{"invalid 24", `
        class Hello extends React.Component {
          handleEventMethods = () => {}
          componentWillReceiveProps = () => {}
          render = () => { return <div />; }
        }
      `, `
        class Hello extends React.Component {
          handleEventMethods = () => {}
          componentWillReceiveProps() {}
          render() { return <div />; }
        }
      `, 2},
		{"invalid 25", `
        class Hello extends React.Component {
          handleEventMethods = () => {}
          UNSAFE_componentWillReceiveProps = () => {}
          render = () => { return <div />; }
        }
      `, `
        class Hello extends React.Component {
          handleEventMethods = () => {}
          UNSAFE_componentWillReceiveProps() {}
          render() { return <div />; }
        }
      `, 2},
		{"invalid 26", `
        class Hello extends React.Component {
          handleEventMethods = () => {}
          shouldComponentUpdate = () => { return true; }
          render = () => { return <div />; }
        }
      `, `
        class Hello extends React.Component {
          handleEventMethods = () => {}
          shouldComponentUpdate() { return true; }
          render() { return <div />; }
        }
      `, 2},
		{"invalid 27", `
        class Hello extends React.Component {
          handleEventMethods = () => {}
          componentWillUpdate = () => {}
          render = () => { return <div />; }
        }
      `, `
        class Hello extends React.Component {
          handleEventMethods = () => {}
          componentWillUpdate() {}
          render() { return <div />; }
        }
      `, 2},
		{"invalid 28", `
        class Hello extends React.Component {
          handleEventMethods = () => {}
          UNSAFE_componentWillUpdate = () => {}
          render = () => { return <div />; }
        }
      `, `
        class Hello extends React.Component {
          handleEventMethods = () => {}
          UNSAFE_componentWillUpdate() {}
          render() { return <div />; }
        }
      `, 2},
		{"invalid 29", `
        class Hello extends React.Component {
          handleEventMethods = () => {}
          getSnapshotBeforeUpdate = () => { return {}; }
          render = () => { return <div />; }
        }
      `, `
        class Hello extends React.Component {
          handleEventMethods = () => {}
          getSnapshotBeforeUpdate() { return {}; }
          render() { return <div />; }
        }
      `, 2},
		{"invalid 30", `
        class Hello extends React.Component {
          handleEventMethods = () => {}
          componentDidUpdate = (prevProps) => {}
          render = () => { return <div />; }
        }
      `, `
        class Hello extends React.Component {
          handleEventMethods = () => {}
          componentDidUpdate(prevProps) {}
          render() { return <div />; }
        }
      `, 2},
		{"invalid 31", `
        class Hello extends React.Component {
          handleEventMethods = () => {}
          componentDidCatch = () => {}
          render = () => { return <div />; }
        }
      `, `
        class Hello extends React.Component {
          handleEventMethods = () => {}
          componentDidCatch() {}
          render() { return <div />; }
        }
      `, 2},
		{"invalid 32", `
        class Hello extends React.Component {
          handleEventMethods = () => {}
          componentWillUnmount = () => {}
          render = () => { return <div />; }
        }
      `, `
        class Hello extends React.Component {
          handleEventMethods = () => {}
          componentWillUnmount() {}
          render() { return <div />; }
        }
      `, 2},
		{"invalid 33", `
        class Hello extends React.Component {
          render = () => <div />
        }
      `, `
        class Hello extends React.Component {
          render() { return <div />; }
        }
      `, 1},
		{"invalid 34", `
        class Hello extends React.Component {
          render = () => /*first*/<div />/*second*/
        }
      `, `
        class Hello extends React.Component {
          render() { return /*first*/<div />/*second*/; }
        }
      `, 1},
		{"invalid 35", `
        export default class Root extends Component {
          getInitialState = () => ({
            errorImporting: null,
            errorParsing: null,
            errorUploading: null,
            file: null,
            fromExtension: false,
            importSuccess: false,
            isImporting: false,
            isParsing: false,
            isUploading: false,
            parsedResults: null,
            showLongRunningMessage: false,
          });
        }
      `, `
        export default class Root extends Component {
          getInitialState() { return {
            errorImporting: null,
            errorParsing: null,
            errorUploading: null,
            file: null,
            fromExtension: false,
            importSuccess: false,
            isImporting: false,
            isParsing: false,
            isUploading: false,
            parsedResults: null,
            showLongRunningMessage: false,
          }; }
        }
      `, 1},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoArrowFunctionLifecycle, noArrowFunctionLifecycleFile, testCase.sourceText)
			wantIds := make([]string, testCase.wantCount)
			for index := range wantIds {
				wantIds[index] = "lifecycle"
			}
			rule_testing.ExpectFindings(t, result, wantIds...)

			if testCase.wantFixed != "" {
				rule_testing.ExpectFixedSource(t, result, testCase.wantFixed)
			}
		})
	}
}

// TestNoArrowFunctionLifecycleStaysSilent runs every case this port declines.
func TestNoArrowFunctionLifecycleStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"valid 0", `
        var Hello = createReactClass({
          render: function() { return <div />; }
        });
      `},
		{"valid 1", `
        var Hello = createReactClass({
          getDefaultProps: function() { return {}; },
          render: function() { return <div />; }
        });
      `},
		{"valid 2", `
        var Hello = createReactClass({
          getInitialState: function() { return {}; },
          render: function() { return <div />; }
        });
      `},
		{"valid 3", `
        var Hello = createReactClass({
          getChildContext: function() { return {}; },
          render: function() { return <div />; }
        });
      `},
		{"valid 4", `
        var Hello = createReactClass({
          getDerivedStateFromProps: function() { return {}; },
          render: function() { return <div />; }
        });
      `},
		{"valid 5", `
        var Hello = createReactClass({
          componentWillMount: function() {},
          render: function() { return <div />; }
        });
      `},
		{"valid 6", `
        var Hello = createReactClass({
          UNSAFE_componentWillMount: function() {},
          render: function() { return <div />; }
        });
      `},
		{"valid 7", `
        var Hello = createReactClass({
          componentDidMount: function() {},
          render: function() { return <div />; }
        });
      `},
		{"valid 8", `
        var Hello = createReactClass({
          componentWillReceiveProps: function() {},
          render: function() { return <div />; }
        });
      `},
		{"valid 9", `
        var Hello = createReactClass({
          UNSAFE_componentWillReceiveProps: function() {},
          render: function() { return <div />; }
        });
      `},
		{"valid 10", `
        var Hello = createReactClass({
          shouldComponentUpdate: function() { return true; },
          render: function() { return <div />; }
        });
      `},
		{"valid 11", `
        var Hello = createReactClass({
          componentWillUpdate: function() {},
          render: function() { return <div />; }
        });
      `},
		{"valid 12", `
        var Hello = createReactClass({
          UNSAFE_componentWillUpdate: function() {},
          render: function() { return <div />; }
        });
      `},
		{"valid 13", `
        var Hello = createReactClass({
          getSnapshotBeforeUpdate: function() { return {}; },
          render: function() { return <div />; }
        });
      `},
		{"valid 14", `
        var Hello = createReactClass({
          componentDidUpdate: function() {},
          render: function() { return <div />; }
        });
      `},
		{"valid 15", `
        var Hello = createReactClass({
          componentDidCatch: function() {},
          render: function() { return <div />; }
        });
      `},
		{"valid 16", `
        var Hello = createReactClass({
          componentWillUnmount: function() {},
          render: function() { return <div />; }
        });
      `},
		{"valid 17", `
        class Hello extends React.Component {
          handleEventMethods = () => {}
          render() { return <div />; }
        }
      `},
		{"valid 18", `
        class Hello extends React.Component {
          handleEventMethods = () => {}
          getDefaultProps() { return {}; }
          render() { return <div />; }
        }
      `},
		{"valid 19", `
        class Hello extends React.Component {
          handleEventMethods = () => {}
          getInitialState() { return {}; }
          render() { return <div />; }
        }
      `},
		{"valid 20", `
        class Hello extends React.Component {
          handleEventMethods = () => {}
          getChildContext() { return {}; }
          render() { return <div />; }
        }
      `},
		{"valid 21", `
        class Hello extends React.Component {
          handleEventMethods = () => {}
          getDerivedStateFromProps() { return {}; }
          render() { return <div />; }
        }
      `},
		{"valid 22", `
        class Hello extends React.Component {
          handleEventMethods = () => {}
          componentWillMount() {}
          render() { return <div />; }
        }
      `},
		{"valid 23", `
        class Hello extends React.Component {
          handleEventMethods = () => {}
          UNSAFE_componentWillMount() {}
          render() { return <div />; }
        }
      `},
		{"valid 24", `
        class Hello extends React.Component {
          handleEventMethods = () => {}
          componentDidMount() {}
          render() { return <div />; }
        }
      `},
		{"valid 25", `
        class Hello extends React.Component {
          handleEventMethods = () => {}
          componentWillReceiveProps() {}
          render() { return <div />; }
        }
      `},
		{"valid 26", `
        class Hello extends React.Component {
          handleEventMethods = () => {}
          UNSAFE_componentWillReceiveProps() {}
          render() { return <div />; }
        }
      `},
		{"valid 27", `
        class Hello extends React.Component {
          handleEventMethods = () => {}
          shouldComponentUpdate() { return true; }
          render() { return <div />; }
        }
      `},
		{"valid 28", `
        class Hello extends React.Component {
          handleEventMethods = () => {}
          componentWillUpdate() {}
          render() { return <div />; }
        }
      `},
		{"valid 29", `
        class Hello extends React.Component {
          handleEventMethods = () => {}
          UNSAFE_componentWillUpdate() {}
          render() { return <div />; }
        }
      `},
		{"valid 30", `
        class Hello extends React.Component {
          handleEventMethods = () => {}
          getSnapshotBeforeUpdate() { return {}; }
          render() { return <div />; }
        }
      `},
		{"valid 31", `
        class Hello extends React.Component {
          handleEventMethods = () => {}
          componentDidUpdate() {}
          render() { return <div />; }
        }
      `},
		{"valid 32", `
        class Hello extends React.Component {
          handleEventMethods = () => {}
          componentDidCatch() {}
          render() { return <div />; }
        }
      `},
		{"valid 33", `
        class Hello extends React.Component {
          handleEventMethods = () => {}
          componentWillUnmount() {}
          render() { return <div />; }
        }
      `},
		{"valid 34", `
        class Hello extends React.Component {
          handleEventMethods = () => {}
          getDerivedStateFromProps = () => { return {}; } // not a lifecycle method
          static getDerivedStateFromProps() {}
          render() { return <div />; }
        }
      `},
		{"valid 35", `
        var Hello = createReactClass({
          getDerivedStateFromProps: () => { return {}; },
          render: function() { return <div />; }
        });
      `},
		{"valid 36", `
        class MyComponent extends React.Component {
          onChange: () => void;
        }
      `},

		// Cases upstream's corpus does not write, each measured against the installed build.

		// Component detection accepts `createReactClass` and nothing else. The shelf's
		// `react.IsEs5ComponentCall` also accepts `createClass`, both bare and on the React
		// object, so reaching for it here would report on both of these. Measured clean.
		{"createClass is not the factory", "var H = createClass({ render: () => <div /> });"},
		{"React.createClass is not the factory", "var H = React.createClass({ render: () => <div /> });"},
		{"a plain object is not a component", "var H = { render: () => <div /> };"},
		// The static and instance lists are separate, so these two invert against their partners
		// in the firing table.
		{"an instance getDerivedStateFromProps is not on the instance list", "class H extends React.Component { getDerivedStateFromProps = () => {}; render() { return <div />; } }"},
		{"a static render is not on the static list", "class H extends React.Component { static render = () => {}; }"},
		// An ES5 object property always uses the instance list, so the static name is clean there
		// while the same name as a static class field reports.
		{"an ES5 getDerivedStateFromProps uses the instance list", "var H = createReactClass({ getDerivedStateFromProps: () => {}, render: function(){ return <div/>; } });"},
		// A computed key has no readable name for the membership test.
		{"a computed lifecycle key", "class H extends React.Component { ['render'] = () => <div />; }"},
		// A function expression is not an arrow.
		{"a function expression class field", "class H extends React.Component { render = function() { return <div />; }; }"},
		// A class with no React base is not a component, even with a lifecycle-named arrow.
		{"a class extending something unrelated", "class H extends Foo { render = () => <div />; }"},
		{"a class with no heritage", "class H { render = () => <div />; }"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoArrowFunctionLifecycle, noArrowFunctionLifecycleFile, testCase.sourceText)
			rule_testing.ExpectClean(t, result)
		})
	}
}

// TestNoArrowFunctionLifecycleStaticSplit pins the four verdicts the two lists produce.
//
// A static class field is checked against the static list alone and an instance field against the
// instance list alone, so two of these four invert against the naive reading. Upstream's corpus
// writes only the reporting halves, so a port using one combined list would pass every imported
// case while reporting two shapes upstream ignores. All four measured on the installed build.
func TestNoArrowFunctionLifecycleStaticSplit(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		wantFires  bool
	}{
		{"static getDerivedStateFromProps", "class H extends React.Component { static getDerivedStateFromProps = () => {}; render() { return <div />; } }", true},
		{"instance getDerivedStateFromProps", "class H extends React.Component { getDerivedStateFromProps = () => {}; render() { return <div />; } }", false},
		{"static render", "class H extends React.Component { static render = () => {}; }", false},
		{"instance render", "class H extends React.Component { render = () => <div />; }", true},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoArrowFunctionLifecycle, noArrowFunctionLifecycleFile, testCase.sourceText)
			if testCase.wantFires {
				rule_testing.ExpectFindings(t, result, "lifecycle")
			} else {
				rule_testing.ExpectClean(t, result)
			}
		})
	}
}

// TestNoArrowFunctionLifecycleDeclinesUnrenderableParameters pins the deliberate fix decline.
//
// Upstream's fixer maps each parameter to its identifier name, which is undefined for anything
// that is not a plain identifier, so a destructured or defaulted parameter is written into the
// repaired source as the literal text `undefined`. That silently deletes what the parameter said,
// and the engine applies a fix unattended.
//
// This port reports those cases and declines the repair, which is the subset that can be shown
// correct. The typed-parameter row is the one our tree actually contains and upstream's corpus
// structurally cannot: rendering an annotated parameter by name alone would drop the annotation the
// way a shipped fixer once widened eight declarations to `any`.
func TestNoArrowFunctionLifecycleDeclinesUnrenderableParameters(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"a destructured parameter", "class H extends React.Component { componentDidUpdate = ({ a }) => { f(a); }; render() { return <div />; } }"},
		{"a defaulted parameter", "class H extends React.Component { componentDidUpdate = (a = 1) => { f(a); }; render() { return <div />; } }"},
		{"a rest parameter", "class H extends React.Component { componentDidUpdate = (...a) => { f(a); }; render() { return <div />; } }"},
		{"a typed parameter", "class H extends React.Component { componentDidUpdate = (previous: Properties) => { f(previous); }; render() { return <div />; } }"},

		// A RETURN annotation is lost by the same rewrite and the parameter check cannot see it,
		// because it sits outside the parameter list. Measured before the guard existed: the head
		// is rebuilt as `(params) `, replacing everything between the key and the body, so
		// `componentDidMount = (): void => {}` was repaired to `componentDidMount() {}` with the
		// `: void` silently gone. Untyped parameters walked straight past the check above.
		{"a return annotation", "class H extends React.Component { componentDidMount = (): void => { f(); }; render() { return <div />; } }"},
		{"a return annotation with untyped parameters", "class H extends React.Component { componentDidUpdate = (previous): boolean => { return f(previous); }; render() { return <div />; } }"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoArrowFunctionLifecycle, noArrowFunctionLifecycleFile, testCase.sourceText)
			rule_testing.ExpectFindings(t, result, "lifecycle")

			// The finding is real; the repair is what is withheld. Asserting the absence of a fix
			// is the whole point, and no message-id fixture can see it.
			if fixes := result.Diagnostics[0].Fixes; len(fixes) != 0 {
				t.Fatalf("a repair was offered for a parameter that cannot be rendered by name: %v", fixes)
			}
		})
	}
}

// TestNoArrowFunctionLifecycleKeepsTheHeadWhole pins what the rewrite carries and where it declines.
//
// The first row shipped broken. The scan for comments "before the body" started at the KEY, so a
// comment between `=` and the parameter list stopped the replaced head early and left `() =>`
// standing: `render = /* c */ () => {...}` became `render() /* c */ () => {...}`, which does not
// parse. Upstream's installed build writes `render() {...}`, dropping the comment; this declines
// instead, because the fix runs unattended and the comment is someone's.
//
// The remaining rows are what the rewrite must keep: a decorator, an accessibility modifier and the
// JSDoc above the field all sit before the key, outside the replaced head, and a comment after the
// `=>` still travels into the block as upstream's corpus asserts.
func TestNoArrowFunctionLifecycleKeepsTheHeadWhole(t *testing.T) {
	t.Parallel()

	t.Run("a comment before the arrow declines", func(t *testing.T) {
		t.Parallel()
		for _, sourceText := range []string{
			"class H extends React.Component {\n  render = /* c */ () => {\n    return <div />;\n  };\n}\n",
			"class H extends React.Component {\n  render /* c */ = () => {\n    return <div />;\n  };\n}\n",
			"class H extends React.Component {\n  render = () /* c */ => {\n    return <div />;\n  };\n}\n",
		} {
			result := rule_testing.Run(t, NoArrowFunctionLifecycle, noArrowFunctionLifecycleFile, sourceText)
			rule_testing.ExpectFindings(t, result, "lifecycle")
			if fixes := result.Diagnostics[0].Fixes; len(fixes) != 0 {
				t.Fatalf("a repair was offered that would delete a comment in the head of %q", sourceText)
			}
		}
	})

	cases := []struct {
		name       string
		sourceText string
		wantFixed  string
	}{
		{
			"a decorated public field with JSDoc",
			"declare const bound: any;\nclass H extends React.Component {\n  /** Draws. */\n  @bound\n  public render = () => {\n    return <div />;\n  };\n}\n",
			"declare const bound: any;\nclass H extends React.Component {\n  /** Draws. */\n  @bound\n  public render() {\n    return <div />;\n  };\n}\n",
		},
		{
			"a comment after the arrow travels into the block",
			"class H extends React.Component {\n  render = () => /* c */ <div />;\n}\n",
			"class H extends React.Component {\n  render() { return /* c */<div />; }\n}\n",
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoArrowFunctionLifecycle, noArrowFunctionLifecycleFile, testCase.sourceText)
			rule_testing.ExpectFindings(t, result, "lifecycle")
			rule_testing.ExpectFixedSource(t, result, testCase.wantFixed)
		})
	}
}

// TestNoArrowFunctionLifecycleMessageText asserts the rendered finding exactly.
//
// The message interpolates the property name, so an id assertion cannot see anything the format
// string does. Asserted against a literal typed here rather than against the rule's own constant,
// which would move with any mutation of it.
func TestNoArrowFunctionLifecycleMessageText(t *testing.T) {
	t.Parallel()

	result := rule_testing.Run(t, NoArrowFunctionLifecycle, noArrowFunctionLifecycleFile,
		"class H extends React.Component { componentDidMount = () => {}; render() { return <div />; } }")
	rule_testing.ExpectFindings(t, result, "lifecycle")

	const want = "componentDidMount is a React lifecycle method, and should not be an arrow " +
		"function or in a class field. Use an instance method instead."
	if got := result.Diagnostics[0].Message.Description; got != want {
		t.Fatalf("message is %q, want %q", got, want)
	}
}

// TestNoArrowFunctionLifecycleAnchor pins WHERE the finding points, at every report site.
//
// Upstream reports on the whole property, so the span covers the name, the arrow and the body
// together rather than any one of them. A fix that lands correctly while the finding points
// somewhere else shows the reader a line they were never told about.
//
// The rule has THREE report calls, not one: the block body, the concise body, and the declined
// repair. An earlier version of this test covered only the concise-body path, and a mutant moving
// the block-body anchor from the property to the arrow survived every fixture, because no other
// assertion in the suite looks at where a block-body finding points. Each site is a row here now.
func TestNoArrowFunctionLifecycleAnchor(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		want       string
	}{
		{
			"the concise-body report",
			"class H extends React.Component { render = () => <div />; }",
			"render = () => <div />;",
		},
		{
			"the block-body report",
			"class H extends React.Component { render = () => { return <div />; }; }",
			"render = () => { return <div />; };",
		},
		{
			"the declined-repair report",
			"class H extends React.Component { componentDidUpdate = ({ a }) => { f(a); }; }",
			"componentDidUpdate = ({ a }) => { f(a); };",
		},
		{
			"the ES5 property report",
			"var H = createReactClass({ render: () => <div /> });",
			"render: () => <div />",
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoArrowFunctionLifecycle, noArrowFunctionLifecycleFile, testCase.sourceText)
			rule_testing.ExpectFindings(t, result, "lifecycle")

			fixture := strings.TrimSpace(testCase.sourceText) + "\n"
			span := result.Diagnostics[0].Range
			if got := fixture[span.Pos():span.End()]; got != testCase.want {
				t.Fatalf("finding spans %q, want %q", got, testCase.want)
			}
		})
	}
}

// TestNoArrowFunctionLifecycleHasNoFileSuffixGate pins that a plain `.ts` file is examined.
//
// Three siblings in this package gate on `.tsx` and `.jsx`, which is oxc residue. Upstream has no
// such gate anywhere, and a lifecycle arrow with no JSX in it needs no JSX file.
func TestNoArrowFunctionLifecycleHasNoFileSuffixGate(t *testing.T) {
	t.Parallel()

	const source = "class H extends React.Component { componentDidMount = () => {}; }"

	for _, suffix := range []string{".tsx", ".ts"} {
		t.Run(suffix, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoArrowFunctionLifecycle, "/repository/source/SuffixProbe"+suffix, source)
			rule_testing.ExpectFindings(t, result, "lifecycle")
		})
	}
}
