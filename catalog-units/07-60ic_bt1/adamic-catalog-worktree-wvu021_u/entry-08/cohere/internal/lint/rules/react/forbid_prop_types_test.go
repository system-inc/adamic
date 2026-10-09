package react

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// forbidPropTypesFile is where the fixtures pretend to live.
//
// A .tsx extension because many cases return JSX. The rule has no suffix gate; the typed harness
// can only express one extension, which the suffix test below explains.
const forbidPropTypesFile = "/repository/source/PropTypes.tsx"

// The corpus is upstream's, extracted mechanically rather than retyped.
//
// Every case below comes from
// /tmp/lint-sources/eslint-plugin-react/tests/lib/rules/forbid-prop-types.js, read by stubbing the
// tester and capturing the object it was handed, so the `[].concat` shapes and the babel-eslint
// version gate inside it are resolved by the JavaScript engine rather than by a parser written
// here. That gate resolves to the modern branch because babel-eslint is not installed in this tree;
// its legacy branch adds three passing cases, each carrying upstream's own comment that the code is
// invalid and should not be validated.
//
// Upstream carries 48 valid and 63 invalid cases. Fourteen of the invalid ones are omitted here and
// the omission is deliberate rather than a gap in the import: they configure
// `settings.propWrapperFunctions`, a shared setting this tree has no path to. See the rule doc, and
// TestForbidPropTypesDoesNotUnwrapPropWrappers below, which pins the resulting silence as a stated
// divergence instead of leaving it unrecorded. What is imported is 48 passing and 49 reporting
// cases asserting 56 findings.
//
// Thirty-seven of the reporting cases state only a COUNT rather than per-error data, so these
// assert the count. `ExpectFindings` takes one id per finding and every finding this rule can
// produce carries the same id, so a repeated id is the faithful expectation.

// forbidPropTypesExpectedIds repeats the rule's single message id count times.
func forbidPropTypesExpectedIds(count int) []string {
	ids := make([]string, count)
	for index := range ids {
		ids[index] = "forbiddenPropType"
	}
	return ids
}

// runForbidPropTypes drives the rule through its own exported decoder.
//
// Routing through DecodeForbidPropTypesOptions rather than building the struct is what puts the
// absent-versus-empty distinction on `forbid` under test, which is the one line here with no
// upstream counterpart.
func runForbidPropTypes(t *testing.T, sourceText string, options *ForbidPropTypesOptions) rule_testing.Result {
	t.Helper()
	if options == nil {
		return rule_testing.RunTyped(t, ForbidPropTypes, forbidPropTypesFile, sourceText)
	}
	encoded, err := json.Marshal(options)
	if err != nil {
		t.Fatalf("encoding options: %v", err)
	}
	decoded, err := DecodeForbidPropTypesOptions(encoded)
	if err != nil {
		t.Fatalf("decoding options: %v", err)
	}
	return rule_testing.RunTypedWithOptions(t, ForbidPropTypes, forbidPropTypesFile, sourceText, decoded)
}

// TestForbidPropTypesFires runs the forty-nine reporting cases from upstream.
func TestForbidPropTypesFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		options    *ForbidPropTypesOptions
		wantCount  int
	}{
		{"invalid 0 var First = createReactClass({ propTypes: { a: PropTypes.a", `
        var First = createReactClass({
          propTypes: {
            a: PropTypes.any
          },
          render: function() {
            return <div />;
          }
        });
      `, nil, 1},
		{"invalid 1 var First = createReactClass({ propTypes: { n: PropTypes.n", `
        var First = createReactClass({
          propTypes: {
            n: PropTypes.number
          },
          render: function() {
            return <div />;
          }
        });
      `, &ForbidPropTypesOptions{Forbid: &[]string{"number"}}, 1},
		{"invalid 2 var First = createReactClass({ propTypes: { a: PropTypes.a", `
        var First = createReactClass({
          propTypes: {
            a: PropTypes.any.isRequired
          },
          render: function() {
            return <div />;
          }
        });
      `, nil, 1},
		{"invalid 3 var First = createReactClass({ propTypes: { a: PropTypes.a", `
        var First = createReactClass({
          propTypes: {
            a: PropTypes.array
          },
          render: function() {
            return <div />;
          }
        });
      `, nil, 1},
		{"invalid 4 var First = createReactClass({ propTypes: { a: PropTypes.a", `
        var First = createReactClass({
          propTypes: {
            a: PropTypes.array.isRequired
          },
          render: function() {
            return <div />;
          }
        });
      `, nil, 1},
		{"invalid 5 var First = createReactClass({ propTypes: { a: PropTypes.o", `
        var First = createReactClass({
          propTypes: {
            a: PropTypes.object
          },
          render: function() {
            return <div />;
          }
        });
      `, nil, 1},
		{"invalid 6 var First = createReactClass({ propTypes: { a: PropTypes.o", `
        var First = createReactClass({
          propTypes: {
            a: PropTypes.object.isRequired
          },
          render: function() {
            return <div />;
          }
        });
      `, nil, 1},
		{"invalid 7 var First = createReactClass({ propTypes: { a: PropTypes.a", `
        var First = createReactClass({
          propTypes: {
            a: PropTypes.array,
            o: PropTypes.object
          },
          render: function() {
            return <div />;
          }
        });
      `, nil, 2},
		{"invalid 8 var First = createReactClass({ propTypes: { a: PropTypes.a", `
        var First = createReactClass({
          propTypes: {
            a: PropTypes.array
          },
          render: function() {
            return <div />;
          }
        });
        var Second = createReactClass({
          propTypes: {
            o: PropTypes.object
          },
          render: function() {
            return <div />;
          }
        });
      `, nil, 2},
		{"invalid 9 class First extends React.Component { render() { return <d", `
        class First extends React.Component {
          render() {
            return <div />;
          }
        }
        First.propTypes = {
            a: PropTypes.array,
            o: PropTypes.object
        };
        class Second extends React.Component {
          render() {
            return <div />;
          }
        }
        Second.propTypes = {
            a: PropTypes.array,
            o: PropTypes.object
        };
      `, nil, 4},
		{"invalid 11 import { forbidExtraProps } from 'airbnb-prop-types'; expo", `
        import { forbidExtraProps } from "airbnb-prop-types";
        export const propTypes = {dpm: PropTypes.any};
        export default function Component() {}
        Component.propTypes = propTypes;
      `, nil, 1},
		{"invalid 13 class Component extends React.Component { static propTypes", `
        class Component extends React.Component {
          static propTypes = {
            a: PropTypes.array,
            o: PropTypes.object
          };
          render() {
            return <div />;
          }
        }
      `, nil, 2},
		{"invalid 14 class Component extends React.Component { static get propT", `
        class Component extends React.Component {
          static get propTypes() {
            return {
              a: PropTypes.array,
              o: PropTypes.object
            };
          };
          render() {
            return <div />;
          }
        }
      `, nil, 2},
		{"invalid 17 var Hello = createReactClass({ propTypes: { retailer: Prop", `
        var Hello = createReactClass({
          propTypes: {
            retailer: PropTypes.instanceOf(Map).isRequired,
            requestRetailer: PropTypes.func.isRequired
          },
          render: function() {
            return <div />;
          }
        });
      `, &ForbidPropTypesOptions{Forbid: &[]string{"instanceOf"}}, 1},
		{"invalid 18 var object = PropTypes.object; var Hello = createReactClas", `
        var object = PropTypes.object;
        var Hello = createReactClass({
          propTypes: {
            retailer: object,
          },
          render: function() {
            return <div />;
          }
        });
      `, &ForbidPropTypesOptions{Forbid: &[]string{"object"}}, 1},
		{"invalid 19 var First = createReactClass({ contextTypes: { a: PropType", `
        var First = createReactClass({
          contextTypes: {
            a: PropTypes.any
          },
          render: function() {
            return <div />;
          }
        });
      `, &ForbidPropTypesOptions{CheckContextTypes: true}, 1},
		{"invalid 20 class Foo extends Component { static contextTypes = { a: P", `
        class Foo extends Component {
          static contextTypes = {
            a: PropTypes.any
          }
          render() {
            return <div />;
          }
        }
      `, &ForbidPropTypesOptions{CheckContextTypes: true}, 1},
		{"invalid 21 class Foo extends Component { static get contextTypes() { ", `
        class Foo extends Component {
          static get contextTypes() {
            return {
              a: PropTypes.any
            };
          }
          render() {
            return <div />;
          }
        }
      `, &ForbidPropTypesOptions{CheckContextTypes: true}, 1},
		{"invalid 22 class Foo extends Component { render() { return <div />; }", `
        class Foo extends Component {
          render() {
            return <div />;
          }
        }
        Foo.contextTypes = {
          a: PropTypes.any
        };
      `, &ForbidPropTypesOptions{CheckContextTypes: true}, 1},
		{"invalid 23 function Foo(props) { return <div />; } Foo.contextTypes =", `
        function Foo(props) {
          return <div />;
        }
        Foo.contextTypes = {
          a: PropTypes.any
        };
      `, &ForbidPropTypesOptions{CheckContextTypes: true}, 1},
		{"invalid 24 const Foo = (props) => { return <div />; }; Foo.contextTyp", `
        const Foo = (props) => {
          return <div />;
        };
        Foo.contextTypes = {
          a: PropTypes.any
        };
      `, &ForbidPropTypesOptions{CheckContextTypes: true}, 1},
		{"invalid 30 var Hello = createReactClass({ contextTypes: { retailer: P", `
        var Hello = createReactClass({
          contextTypes: {
            retailer: PropTypes.instanceOf(Map).isRequired,
            requestRetailer: PropTypes.func.isRequired
          },
          render: function() {
            return <div />;
          }
        });
      `, &ForbidPropTypesOptions{Forbid: &[]string{"instanceOf"}, CheckContextTypes: true}, 1},
		{"invalid 31 class Component extends React.Component { static contextTy", `
        class Component extends React.Component {
          static contextTypes = {
            retailer: PropTypes.instanceOf(Map).isRequired,
            requestRetailer: PropTypes.func.isRequired
          }
          render() {
            return <div />;
          }
        }
      `, &ForbidPropTypesOptions{Forbid: &[]string{"instanceOf"}, CheckContextTypes: true}, 1},
		{"invalid 32 class Component extends React.Component { static get conte", `
        class Component extends React.Component {
          static get contextTypes() {
            return {
              retailer: PropTypes.instanceOf(Map).isRequired,
              requestRetailer: PropTypes.func.isRequired
            };
          }
          render() {
            return <div />;
          }
        }
      `, &ForbidPropTypesOptions{Forbid: &[]string{"instanceOf"}, CheckContextTypes: true}, 1},
		{"invalid 33 class Component extends React.Component { render() { retur", `
        class Component extends React.Component {
          render() {
            return <div />;
          }
        }
        Component.contextTypes = {
          retailer: PropTypes.instanceOf(Map).isRequired,
          requestRetailer: PropTypes.func.isRequired
        };
      `, &ForbidPropTypesOptions{Forbid: &[]string{"instanceOf"}, CheckContextTypes: true}, 1},
		{"invalid 34 function Component(props) { return <div />; } Component.co", `
        function Component(props) {
          return <div />;
        }
        Component.contextTypes = {
          retailer: PropTypes.instanceOf(Map).isRequired,
          requestRetailer: PropTypes.func.isRequired
        };
      `, &ForbidPropTypesOptions{Forbid: &[]string{"instanceOf"}, CheckContextTypes: true}, 1},
		{"invalid 35 const Component = (props) => { return <div />; }; Componen", `
        const Component = (props) => {
          return <div />;
        };
        Component.contextTypes = {
          retailer: PropTypes.instanceOf(Map).isRequired,
          requestRetailer: PropTypes.func.isRequired
        }
      `, &ForbidPropTypesOptions{Forbid: &[]string{"instanceOf"}, CheckContextTypes: true}, 1},
		{"invalid 36 var First = createReactClass({ childContextTypes: { a: Pro", `
        var First = createReactClass({
          childContextTypes: {
            a: PropTypes.any
          },
          render: function() {
            return <div />;
          }
        });
      `, &ForbidPropTypesOptions{CheckChildContextTypes: true}, 1},
		{"invalid 37 class Foo extends Component { static childContextTypes = {", `
        class Foo extends Component {
          static childContextTypes = {
            a: PropTypes.any
          }
          render() {
            return <div />;
          }
        }
      `, &ForbidPropTypesOptions{CheckChildContextTypes: true}, 1},
		{"invalid 38 class Foo extends Component { static get childContextTypes", `
        class Foo extends Component {
          static get childContextTypes() {
            return {
              a: PropTypes.any
            };
          }
          render() {
            return <div />;
          }
        }
      `, &ForbidPropTypesOptions{CheckChildContextTypes: true}, 1},
		{"invalid 39 class Foo extends Component { render() { return <div />; }", `
        class Foo extends Component {
          render() {
            return <div />;
          }
        }
        Foo.childContextTypes = {
          a: PropTypes.any
        };
      `, &ForbidPropTypesOptions{CheckChildContextTypes: true}, 1},
		{"invalid 40 function Foo(props) { return <div />; } Foo.childContextTy", `
        function Foo(props) {
          return <div />;
        }
        Foo.childContextTypes = {
          a: PropTypes.any
        };
      `, &ForbidPropTypesOptions{CheckChildContextTypes: true}, 1},
		{"invalid 41 const Foo = (props) => { return <div />; }; Foo.childConte", `
        const Foo = (props) => {
          return <div />;
        };
        Foo.childContextTypes = {
          a: PropTypes.any
        };
      `, &ForbidPropTypesOptions{CheckChildContextTypes: true}, 1},
		{"invalid 47 var Hello = createReactClass({ childContextTypes: { retail", `
        var Hello = createReactClass({
          childContextTypes: {
            retailer: PropTypes.instanceOf(Map).isRequired,
            requestRetailer: PropTypes.func.isRequired
          },
          render: function() {
            return <div />;
          }
        });
      `, &ForbidPropTypesOptions{Forbid: &[]string{"instanceOf"}, CheckChildContextTypes: true}, 1},
		{"invalid 48 class Component extends React.Component { static childCont", `
        class Component extends React.Component {
          static childContextTypes = {
            retailer: PropTypes.instanceOf(Map).isRequired,
            requestRetailer: PropTypes.func.isRequired
          }
          render() {
            return <div />;
          }
        }
      `, &ForbidPropTypesOptions{Forbid: &[]string{"instanceOf"}, CheckChildContextTypes: true}, 1},
		{"invalid 49 class Component extends React.Component { render() { retur", `
        class Component extends React.Component {
          render() {
            return <div />;
          }
        }
        Component.childContextTypes = {
          retailer: PropTypes.instanceOf(Map).isRequired,
          requestRetailer: PropTypes.func.isRequired
        };
      `, &ForbidPropTypesOptions{Forbid: &[]string{"instanceOf"}, CheckChildContextTypes: true}, 1},
		{"invalid 50 function Component(props) { return <div />; } Component.ch", `
        function Component(props) {
          return <div />;
        }
        Component.childContextTypes = {
          retailer: PropTypes.instanceOf(Map).isRequired,
          requestRetailer: PropTypes.func.isRequired
        };
      `, &ForbidPropTypesOptions{Forbid: &[]string{"instanceOf"}, CheckChildContextTypes: true}, 1},
		{"invalid 51 const Component = (props) => { return <div />; }; Componen", `
        const Component = (props) => {
          return <div />;
        };
        Component.childContextTypes = {
          retailer: PropTypes.instanceOf(Map).isRequired,
          requestRetailer: PropTypes.func.isRequired
        };
      `, &ForbidPropTypesOptions{Forbid: &[]string{"instanceOf"}, CheckChildContextTypes: true}, 1},
		{"invalid 52 import { object, string } from 'prop-types'; function C({ ", `
        import { object, string } from "prop-types";
        function C({ a, b }) { return [a, b]; }
        C.propTypes = {
          a: object,
          b: string
        };
      `, &ForbidPropTypesOptions{Forbid: &[]string{"object"}}, 1},
		{"invalid 53 import { objectOf, any } from 'prop-types'; function C({ a", `
        import { objectOf, any } from "prop-types";
        function C({ a }) { return a; }
        C.propTypes = {
          a: objectOf(any)
        };
      `, &ForbidPropTypesOptions{Forbid: &[]string{"any"}}, 1},
		{"invalid 54 import { objectOf, any } from 'prop-types'; function C({ a", `
        import { objectOf, any } from "prop-types";
        function C({ a }) { return a; }
        C.propTypes = {
          a: objectOf(any)
        };
      `, &ForbidPropTypesOptions{Forbid: &[]string{"objectOf"}}, 1},
		{"invalid 55 import { shape, any } from 'prop-types'; function C({ a })", `
        import { shape, any } from "prop-types";
        function C({ a }) { return a; }
        C.propTypes = {
          a: shape({
            b: any
          })
        };
      `, &ForbidPropTypesOptions{Forbid: &[]string{"any"}}, 1},
		{"invalid 56 import { any } from 'prop-types'; function C({ a }) { retu", `
        import { any } from "prop-types";
        function C({ a }) { return a; }
        C.propTypes = {
          a: PropTypes.shape({
            b: any
          })
        };
      `, &ForbidPropTypesOptions{Forbid: &[]string{"any"}}, 1},
		{"invalid 57 var First = createReactClass({ propTypes: { s: PropTypes.s", `
        var First = createReactClass({
          propTypes: {
            s: PropTypes.shape({
              o: PropTypes.object
            })
          },
          render: function() {
            return <div />;
          }
        });
      `, nil, 1},
		{"invalid 58 import React from './React'; import { arrayOf, object } fr", `
        import React from './React';

        import { arrayOf, object } from 'prop-types';

        const App = ({ foo }) => (
          <div>
            Hello world {foo}
          </div>
        );

        App.propTypes = {
          foo: arrayOf(object)
        }

        export default App;
      `, nil, 1},
		{"invalid 59 import React from './React'; import PropTypes, { arrayOf }", `
        import React from './React';

        import PropTypes, { arrayOf } from 'prop-types';

        const App = ({ foo }) => (
          <div>
            Hello world {foo}
          </div>
        );

        App.propTypes = {
          foo: arrayOf(PropTypes.object)
        }

        export default App;
      `, nil, 1},
		{"invalid 60 import CustomPropTypes from 'prop-types'; class Component ", `
        import CustomPropTypes from "prop-types";
        class Component extends React.Component {};
        Component.propTypes = {
          a: CustomPropTypes.shape({
            b: CustomPropTypes.String,
            c: CustomPropTypes.object.isRequired,
          })
        }
      `, nil, 1},
		{"invalid 61 import { PropTypes as CustomPropTypes } from 'react'; clas", `
        import { PropTypes as CustomPropTypes } from "react";
        class Component extends React.Component {};
        Component.propTypes = {
          a: CustomPropTypes.shape({
            b: CustomPropTypes.String,
            c: CustomPropTypes.object.isRequired,
          })
        }
      `, nil, 1},
		{"invalid 62 import CustomReact from 'react' class Component extends Re", `
        import CustomReact from "react"
        class Component extends React.Component {};
        Component.propTypes = {
          b: CustomReact.PropTypes.object,
        }
      `, nil, 1},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := runForbidPropTypes(t, testCase.sourceText, testCase.options)
			rule_testing.ExpectFindings(t, result, forbidPropTypesExpectedIds(testCase.wantCount)...)
		})
	}
}

// TestForbidPropTypesStaysSilent runs the forty-eight passing cases from upstream.
//
// These are the false positives upstream already thought about. Several turn on the forbid list
// being configurable rather than on the shape of the code, which is why the option travels with
// each case rather than being applied to the table as a whole.
func TestForbidPropTypesStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		options    *ForbidPropTypesOptions
	}{
		{"valid 0 var First = createReactClass({ render: function() { return", `
        var First = createReactClass({
          render: function() {
            return <div />;
          }
        });
      `, nil},
		{"valid 1 var First = createReactClass({ propTypes: externalPropType", `
        var First = createReactClass({
          propTypes: externalPropTypes,
          render: function() {
            return <div />;
          }
        });
      `, nil},
		{"valid 2 var First = createReactClass({ propTypes: { s: PropTypes.s", `
        var First = createReactClass({
          propTypes: {
            s: PropTypes.string,
            n: PropTypes.number,
            i: PropTypes.instanceOf,
            b: PropTypes.bool
          },
          render: function() {
            return <div />;
          }
        });
      `, nil},
		{"valid 3 var First = createReactClass({ propTypes: { a: PropTypes.a", `
        var First = createReactClass({
          propTypes: {
            a: PropTypes.array
          },
          render: function() {
            return <div />;
          }
        })
      `, &ForbidPropTypesOptions{Forbid: &[]string{"any", "object"}}},
		{"valid 4 var First = createReactClass({ propTypes: { o: PropTypes.o", `
        var First = createReactClass({
          propTypes: {
            o: PropTypes.object
          },
          render: function() {
            return <div />;
          }
        });
      `, &ForbidPropTypesOptions{Forbid: &[]string{"any", "array"}}},
		{"valid 5 var First = createReactClass({ propTypes: { o: PropTypes.o", `
        var First = createReactClass({
          propTypes: {
            o: PropTypes.object,
          },
          render: function() {
            return <div />;
          }
        });
      `, &ForbidPropTypesOptions{Forbid: &[]string{"any", "array"}}},
		{"valid 6 class First extends React.Component { render() { return <d", `
        class First extends React.Component {
          render() {
            return <div />;
          }
        }
        First.propTypes = {
          a: PropTypes.string,
          b: PropTypes.string
        };
        First.propTypes.justforcheck = PropTypes.string;
      `, nil},
		{"valid 7 class First extends React.Component { render() { return <d", `
        class First extends React.Component {
          render() {
            return <div />;
          }
        }
        First.propTypes = {
          elem: PropTypes.instanceOf(HTMLElement)
        };
      `, nil},
		{"valid 8 class Hello extends React.Component { render() { return <d", `
        class Hello extends React.Component {
          render() {
            return <div>Hello</div>;
          }
        }
        Hello.propTypes = {
          "aria-controls": PropTypes.string
        };
      `, nil},
		{"valid 9 var Hello = createReactClass({ render: function() { let { ", `
        var Hello = createReactClass({
          render: function() {
            let { a, ...b } = obj;
            let c = { ...d };
            return <div />;
          }
        });
      `, nil},
		{"valid 10 var Hello = createReactClass({ propTypes: { retailer: Prop", `
        var Hello = createReactClass({
          propTypes: {
            retailer: PropTypes.instanceOf(Map).isRequired,
            requestRetailer: PropTypes.func.isRequired
          },
          render: function() {
            return <div />;
          }
        });
      `, nil},
		{"valid 11 class Test extends React.component { static propTypes = { ", `
        class Test extends React.component {
          static propTypes = {
            intl: React.propTypes.number,
            ...propTypes
          };
        }
      `, nil},
		{"valid 12 class Test extends React.component { static get propTypes(", `
        class Test extends React.component {
          static get propTypes() {
            return {
              intl: React.propTypes.number,
              ...propTypes
            };
          };
        }
      `, nil},
		{"valid 13 var First = createReactClass({ childContextTypes: external", `
        var First = createReactClass({
          childContextTypes: externalPropTypes,
          render: function() {
            return <div />;
          }
        });
      `, &ForbidPropTypesOptions{CheckContextTypes: true}},
		{"valid 14 var First = createReactClass({ childContextTypes: { s: Pro", `
        var First = createReactClass({
          childContextTypes: {
            s: PropTypes.string,
            n: PropTypes.number,
            i: PropTypes.instanceOf,
            b: PropTypes.bool
          },
          render: function() {
            return <div />;
          }
        });
      `, &ForbidPropTypesOptions{CheckContextTypes: true}},
		{"valid 15 var First = createReactClass({ childContextTypes: { a: Pro", `
        var First = createReactClass({
          childContextTypes: {
            a: PropTypes.array
          },
          render: function() {
            return <div />;
          }
        });
      `, &ForbidPropTypesOptions{Forbid: &[]string{"any", "object"}, CheckContextTypes: true}},
		{"valid 16 var First = createReactClass({ childContextTypes: { o: Pro", `
        var First = createReactClass({
          childContextTypes: {
            o: PropTypes.object
          },
          render: function() {
            return <div />;
          }
        });
      `, &ForbidPropTypesOptions{Forbid: &[]string{"any", "array"}, CheckContextTypes: true}},
		{"valid 17 var First = createReactClass({ childContextTypes: { o: Pro", `
        var First = createReactClass({
          childContextTypes: {
            o: PropTypes.object,
          },
          render: function() {
            return <div />;
          }
        });
      `, &ForbidPropTypesOptions{Forbid: &[]string{"any", "array"}, CheckContextTypes: true}},
		{"valid 18 class First extends React.Component { render() { return <d", `
        class First extends React.Component {
          render() {
            return <div />;
          }
        }
        First.childContextTypes = {
          a: PropTypes.string,
          b: PropTypes.string
        };
        First.childContextTypes.justforcheck = PropTypes.string;
      `, &ForbidPropTypesOptions{CheckContextTypes: true}},
		{"valid 19 class First extends React.Component { render() { return <d", `
        class First extends React.Component {
          render() {
            return <div />;
          }
        }
        First.childContextTypes = {
          elem: PropTypes.instanceOf(HTMLElement)
        };
      `, &ForbidPropTypesOptions{CheckContextTypes: true}},
		{"valid 20 class Hello extends React.Component { render() { return <d", `
        class Hello extends React.Component {
          render() {
            return <div>Hello</div>;
          }
        }
        Hello.childContextTypes = {
          "aria-controls": PropTypes.string
        };
      `, &ForbidPropTypesOptions{CheckContextTypes: true}},
		{"valid 21 var Hello = createReactClass({ render: function() { let { ", `
        var Hello = createReactClass({
          render: function() {
            let { a, ...b } = obj;
            let c = { ...d };
            return <div />;
          }
        });
      `, &ForbidPropTypesOptions{CheckContextTypes: true}},
		{"valid 22 var Hello = createReactClass({ childContextTypes: { retail", `
        var Hello = createReactClass({
          childContextTypes: {
            retailer: PropTypes.instanceOf(Map).isRequired,
            requestRetailer: PropTypes.func.isRequired
          },
          render: function() {
            return <div />;
          }
        });
      `, &ForbidPropTypesOptions{CheckContextTypes: true}},
		{"valid 23 class Test extends React.component { static childContextTy", `
        class Test extends React.component {
          static childContextTypes = {
            intl: React.childContextTypes.number,
            ...childContextTypes
          };
        }
      `, &ForbidPropTypesOptions{CheckContextTypes: true}},
		{"valid 24 class Test extends React.component { static get childConte", `
        class Test extends React.component {
          static get childContextTypes() {
            return {
              intl: React.childContextTypes.number,
              ...childContextTypes
            };
          };
        }
      `, &ForbidPropTypesOptions{CheckContextTypes: true}},
		{"valid 25 var First = createReactClass({ childContextTypes: external", `
        var First = createReactClass({
          childContextTypes: externalPropTypes,
          render: function() {
            return <div />;
          }
        });
      `, &ForbidPropTypesOptions{CheckChildContextTypes: true}},
		{"valid 26 var First = createReactClass({ childContextTypes: { s: Pro", `
        var First = createReactClass({
          childContextTypes: {
            s: PropTypes.string,
            n: PropTypes.number,
            i: PropTypes.instanceOf,
            b: PropTypes.bool
          },
          render: function() {
            return <div />;
          }
        });
      `, &ForbidPropTypesOptions{CheckChildContextTypes: true}},
		{"valid 27 var First = createReactClass({ childContextTypes: { a: Pro", `
        var First = createReactClass({
          childContextTypes: {
            a: PropTypes.array
          },
          render: function() {
            return <div />;
          }
        });
      `, &ForbidPropTypesOptions{Forbid: &[]string{"any", "object"}, CheckChildContextTypes: true}},
		{"valid 28 var First = createReactClass({ childContextTypes: { o: Pro", `
        var First = createReactClass({
          childContextTypes: {
            o: PropTypes.object
          },
          render: function() {
            return <div />;
          }
        });
      `, &ForbidPropTypesOptions{Forbid: &[]string{"any", "array"}, CheckChildContextTypes: true}},
		{"valid 29 var First = createReactClass({ childContextTypes: { o: Pro", `
        var First = createReactClass({
          childContextTypes: {
            o: PropTypes.object,
          },
          render: function() {
            return <div />;
          }
        });
      `, &ForbidPropTypesOptions{Forbid: &[]string{"any", "array"}, CheckChildContextTypes: true}},
		{"valid 30 class First extends React.Component { render() { return <d", `
        class First extends React.Component {
          render() {
            return <div />;
          }
        }
        First.childContextTypes = {
          a: PropTypes.string,
          b: PropTypes.string
        };
        First.childContextTypes.justforcheck = PropTypes.string;
      `, &ForbidPropTypesOptions{CheckChildContextTypes: true}},
		{"valid 31 class First extends React.Component { render() { return <d", `
        class First extends React.Component {
          render() {
            return <div />;
          }
        }
        First.childContextTypes = {
          elem: PropTypes.instanceOf(HTMLElement)
        };
      `, &ForbidPropTypesOptions{CheckChildContextTypes: true}},
		{"valid 32 class Hello extends React.Component { render() { return <d", `
        class Hello extends React.Component {
          render() {
            return <div>Hello</div>;
          }
        }
        Hello.childContextTypes = {
          "aria-controls": PropTypes.string
        };
      `, &ForbidPropTypesOptions{CheckChildContextTypes: true}},
		{"valid 33 var Hello = createReactClass({ render: function() { let { ", `
        var Hello = createReactClass({
          render: function() {
            let { a, ...b } = obj;
            let c = { ...d };
            return <div />;
          }
        });
      `, &ForbidPropTypesOptions{CheckChildContextTypes: true}},
		{"valid 34 var Hello = createReactClass({ childContextTypes: { retail", `
        var Hello = createReactClass({
          childContextTypes: {
            retailer: PropTypes.instanceOf(Map).isRequired,
            requestRetailer: PropTypes.func.isRequired
          },
          render: function() {
            return <div />;
          }
        });
      `, &ForbidPropTypesOptions{CheckChildContextTypes: true}},
		{"valid 35 class Test extends React.component { static childContextTy", `
        class Test extends React.component {
          static childContextTypes = {
            intl: React.childContextTypes.number,
            ...childContextTypes
          };
        }
      `, &ForbidPropTypesOptions{CheckChildContextTypes: true}},
		{"valid 36 class Test extends React.component { static get childConte", `
        class Test extends React.component {
          static get childContextTypes() {
            return {
              intl: React.childContextTypes.number,
              ...childContextTypes
            };
          };
        }
      `, &ForbidPropTypesOptions{CheckChildContextTypes: true}},
		{"valid 37 class TestComponent extends React.Component { static defau", `
        class TestComponent extends React.Component {
          static defaultProps = function () {
            const date = new Date();
            return {
              date
            };
          }();
        }
      `, nil},
		{"valid 38 class HeroTeaserList extends React.Component { render() { ", `
        class HeroTeaserList extends React.Component {
          render() { return null; }
        }
        HeroTeaserList.propTypes = Object.assign({
          heroIndex: PropTypes.number,
          preview: PropTypes.bool,
        }, componentApi, teaserListProps);
      `, nil},
		{"valid 39 import PropTypes from 'prop-types'; const Foo = { foo: Pro", `
        import PropTypes from "prop-types";
        const Foo = {
          foo: PropTypes.string,
        };
        const Bar = {
          bar: PropTypes.shape(Foo),
        };
      `, nil},
		{"valid 40 import yup from 'yup' const formValidation = Yup.object().", `
        import yup from "yup"
        const formValidation = Yup.object().shape({
          name: Yup.string(),
          customer_ids: Yup.array()
        });
      `, nil},
		{"valid 41 import yup from 'Yup' const validation = yup.object().shap", `
        import yup from "Yup"
        const validation = yup.object().shape({
          address: yup.object({
            city: yup.string(),
            zip: yup.string(),
          })
        })
      `, &ForbidPropTypesOptions{Forbid: &[]string{"string", "object"}}},
		{"valid 42 import yup from 'yup' Yup.array( Yup.object().shape({ valu", `
        import yup from "yup"
        Yup.array(
          Yup.object().shape({
            value: Yup.number()
          })
        )
      `, &ForbidPropTypesOptions{Forbid: &[]string{"number"}}},
		{"valid 43 import CustomPropTypes from 'prop-types'; class Component ", `
        import CustomPropTypes from "prop-types";
        class Component extends React.Component {};
        Component.propTypes = {
          a: CustomPropTypes.shape({
            b: CustomPropTypes.String,
            c: CustomPropTypes.number.isRequired,
          })
        }
      `, nil},
		{"valid 44 import CustomReact from 'react' class Component extends Re", `
        import CustomReact from "react"
        class Component extends React.Component {};
        Component.propTypes = {
          b: CustomReact.PropTypes.string,
        }
      `, nil},
		{"valid 45 import PropTypes from 'yup' class Component extends React.", `
        import PropTypes from "yup"
        class Component extends React.Component {};
        Component.propTypes = {
          b: PropTypes.array(),
        }
      `, nil},
		{"valid 46 import { PropTypes, shape, any } from 'yup' class Componen", `
        import { PropTypes, shape, any } from "yup"
        class Component extends React.Component {};
        Component.propTypes = {
          b: PropTypes.any,
        }
      `, &ForbidPropTypesOptions{Forbid: &[]string{"any"}}},
		{"valid 47 import { PropTypes } from 'not-react' class Component exte", `
        import { PropTypes } from "not-react"
        class Component extends React.Component {};
        Component.propTypes = {
          b: PropTypes.array(),
        }
      `, nil},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := runForbidPropTypes(t, testCase.sourceText, testCase.options)
			rule_testing.ExpectClean(t, result)
		})
	}
}

// TestForbidPropTypesRequiresTheTypedHarnessForIndirection pins the one arm needing the checker.
//
// `C.propTypes = someIdentifier` resolves that identifier to the object it was declared with, which
// upstream answers from its own variable index. Every other arm is syntactic, so a nil checker
// costs exactly this shape rather than the rule, and the second half of this test is what makes
// that visible: the direct form still reports under the plain harness while the indirect form goes
// silent. A single-assertion test would have read as "the rule needs types" and hidden which part.
func TestForbidPropTypesRequiresTheTypedHarnessForIndirection(t *testing.T) {
	t.Parallel()

	indirect := "const shared = { a: PropTypes.any };\n" +
		"class C extends React.Component { render() { return null; } }\n" +
		"C.propTypes = shared;\n"
	direct := "class C extends React.Component { render() { return null; } }\n" +
		"C.propTypes = { a: PropTypes.any };\n"

	t.Run("the indirect form needs the checker", func(t *testing.T) {
		t.Parallel()
		typed := rule_testing.RunTyped(t, ForbidPropTypes, forbidPropTypesFile, indirect)
		rule_testing.ExpectFindings(t, typed, "forbiddenPropType")

		untyped := rule_testing.Run(t, ForbidPropTypes, forbidPropTypesFile, indirect)
		rule_testing.ExpectClean(t, untyped)
	})

	t.Run("the direct form does not", func(t *testing.T) {
		t.Parallel()
		untyped := rule_testing.Run(t, ForbidPropTypes, forbidPropTypesFile, direct)
		rule_testing.ExpectFindings(t, untyped, "forbiddenPropType")
	})
}

// TestForbidPropTypesHasNoFileSuffixGate pins that no filename decides anything.
//
// Measured on the installed build under .tsx, .jsx and .js: all three report. Only the .tsx row is
// expressible here, because `rule_testing.RunTyped` writes a tsconfig whose include list is
// `["**/*.ts", "**/*.tsx"]` (internal/rule_testing/program.go:30), so a .jsx or .js fixture is not
// in the program and the run fails before any rule is offered a file. That is a fact about the
// typed harness rather than about this rule.
func TestForbidPropTypesHasNoFileSuffixGate(t *testing.T) {
	t.Parallel()

	source := "class C extends React.Component { render() { return null; } }\n" +
		"C.propTypes = { a: PropTypes.any };\n"
	result := rule_testing.RunTyped(t, ForbidPropTypes, "/repository/source/Suffix.tsx", source)
	rule_testing.ExpectFindings(t, result, "forbiddenPropType")
}

// TestForbidPropTypesDeclarationShapes pins the six places a prop types object can live.
//
// Every arm is reachable and every row was measured against the installed build. The last two rows
// are the ones worth stating: this reads as a React rule and asks nothing about components, so a
// plain class and a bare object literal both report. A port that added a component gate would go
// silent on inputs upstream reports while every imported fixture stayed green, because the corpus
// writes real components throughout.
func TestForbidPropTypesDeclarationShapes(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		wantCount  int
	}{
		{
			"a static class property",
			"class C extends React.Component { static propTypes = { a: PropTypes.any }; render() { return null; } }\n",
			1,
		},
		{
			"an assignment to a member",
			"class C extends React.Component { render() { return null; } }\nC.propTypes = { a: PropTypes.any };\n",
			1,
		},
		{
			"a property of an object literal",
			"var C = createReactClass({ propTypes: { a: PropTypes.any }, render() { return null; } });\n",
			1,
		},
		{
			"a getter's return value",
			"class C extends React.Component { static get propTypes() { return { a: PropTypes.any }; } render() { return null; } }\n",
			1,
		},
		{
			"an identifier resolved to its object",
			"const shared = { a: PropTypes.any };\nclass C extends React.Component { render() { return null; } }\nC.propTypes = shared;\n",
			1,
		},
		{
			"nested inside a shape call",
			"class C extends React.Component { render() { return null; } }\n" +
				"C.propTypes = { a: PropTypes.shape({ b: PropTypes.any }) };\n",
			1,
		},

		// It asks nothing about components.
		{"a plain class with no heritage", "class C { static propTypes = { a: PropTypes.any }; }\n", 1},
		{"a bare object literal", "const o = {};\no.propTypes = { a: PropTypes.any };\n", 1},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := runForbidPropTypes(t, testCase.sourceText, nil)
			rule_testing.ExpectFindings(t, result, forbidPropTypesExpectedIds(testCase.wantCount)...)
		})
	}
}

// TestForbidPropTypesValueShapes pins how a property's value is unwrapped before its name is read.
//
// The three silent rows are the load-bearing half. `oneOfType([...])` is clean because its argument
// is an array literal, which contributes no name, so a port that read a name off any argument would
// report on it. And `PropTypes.string` is clean only because it is off the default list, which the
// forbid-list table below separates from the shape question.
func TestForbidPropTypesValueShapes(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name      string
		value     string
		wantCount int
	}{
		{"a plain member", "PropTypes.any", 1},
		{"an isRequired member peeled off", "PropTypes.any.isRequired", 1},
		{"a wrapper reporting on its argument", "PropTypes.arrayOf(PropTypes.object)", 1},
		{"a bare identifier read as the type name", "any", 1},
		{"a nested shape", "PropTypes.shape({ b: PropTypes.any })", 1},

		// The silent rows.
		//
		// A BARE nested object contributes nothing, because `checkProperties` does not recurse: a
		// nested `shape({...})` is reached by the shape arm instead, which visits every such call
		// in the file independently. The pair of rows below is what separates those two routes, and
		// a mutant adding recursion survived every imported case because the corpus writes no bare
		// nested object. Both measured on the installed build.
		{"a bare nested object literal", "{ b: PropTypes.any }", 0},
		{"a deeply bare nested object literal", "{ b: { c: PropTypes.any } }", 0},

		{"an allowed type", "PropTypes.string", 0},
		{"an array literal argument", "PropTypes.oneOfType([PropTypes.any])", 0},
		{"an allowed wrapper over an allowed type", "PropTypes.arrayOf(PropTypes.string)", 0},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			source := "class C extends React.Component { render() { return null; } }\n" +
				"C.propTypes = { a: " + testCase.value + " };\n"
			result := runForbidPropTypes(t, source, nil)
			rule_testing.ExpectFindings(t, result, forbidPropTypesExpectedIds(testCase.wantCount)...)
		})
	}
}

// TestForbidPropTypesForbidList pins the option that decides which names are refused.
//
// The empty-list row is the reason the wire field is a pointer. An absent `forbid` means upstream's
// three defaults and an explicit `[]` forbids nothing, and a decoder collapsing both to a nil slice
// would silently turn the second into the first. Measured on the installed build: `{forbid: []}`
// over `PropTypes.any` is clean.
func TestForbidPropTypesForbidList(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name      string
		value     string
		options   *ForbidPropTypesOptions
		wantCount int
	}{
		{"any is refused by default", "PropTypes.any", nil, 1},
		{"array is refused by default", "PropTypes.array", nil, 1},
		{"object is refused by default", "PropTypes.object", nil, 1},
		{"string is allowed by default", "PropTypes.string", nil, 0},

		// A custom list replaces the defaults rather than adding to them.
		{"a custom list refuses what it names", "PropTypes.string", &ForbidPropTypesOptions{Forbid: &[]string{"string"}}, 1},
		{"a custom list allows what it omits", "PropTypes.any", &ForbidPropTypesOptions{Forbid: &[]string{"string"}}, 0},

		// The distinction the pointer exists for.
		{"an explicitly empty list forbids nothing", "PropTypes.any", &ForbidPropTypesOptions{Forbid: &[]string{}}, 0},
		{"an absent list keeps the defaults", "PropTypes.any", &ForbidPropTypesOptions{}, 1},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			source := "class C extends React.Component { render() { return null; } }\n" +
				"C.propTypes = { a: " + testCase.value + " };\n"
			result := runForbidPropTypes(t, source, testCase.options)
			rule_testing.ExpectFindings(t, result, forbidPropTypesExpectedIds(testCase.wantCount)...)
		})
	}
}

// TestForbidPropTypesContextOptions pins the two declaration names that are off by default.
func TestForbidPropTypesContextOptions(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name        string
		declaration string
		options     *ForbidPropTypesOptions
		wantCount   int
	}{
		{"propTypes is always checked", "propTypes", nil, 1},
		{"contextTypes is off by default", "contextTypes", nil, 0},
		{"childContextTypes is off by default", "childContextTypes", nil, 0},
		{"contextTypes when asked for", "contextTypes", &ForbidPropTypesOptions{CheckContextTypes: true}, 1},
		{"childContextTypes when asked for", "childContextTypes", &ForbidPropTypesOptions{CheckChildContextTypes: true}, 1},

		// Each option governs only its own name.
		{"childContextTypes stays off under the other option", "childContextTypes", &ForbidPropTypesOptions{CheckContextTypes: true}, 0},
		{"contextTypes stays off under the other option", "contextTypes", &ForbidPropTypesOptions{CheckChildContextTypes: true}, 0},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			source := "class C extends React.Component { static " + testCase.declaration +
				" = { a: PropTypes.any }; render() { return null; } }\n"
			result := runForbidPropTypes(t, source, testCase.options)
			rule_testing.ExpectFindings(t, result, forbidPropTypesExpectedIds(testCase.wantCount)...)
		})
	}
}

// TestForbidPropTypesShapeReceiver pins the guard that keeps validation schemas out.
//
// `Yup.object().shape({...})` reads exactly like `PropTypes.shape({...})`, and three of upstream's
// passing cases are that shape. All three reported before this guard was written, which is the
// clearest thing the imported corpus did for this port: nothing about the rule's description
// suggests a validation library is the false positive to worry about.
//
// The discriminator is the RECEIVER, not the import. `isPropTypesPackage` accepts an identifier or
// a member expression and nothing else, so a call receiver such as `Yup.object()` falls through to
// false whatever is imported. Measured with and without the yup import: both clean.
func TestForbidPropTypesShapeReceiver(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		wantCount  int
	}{
		{
			"a shape reached through a prop types declaration",
			"X.propTypes = { a: PropTypes.shape({ b: PropTypes.any }) };\n",
			1,
		},
		{
			"a bare shape call",
			"shape({ b: PropTypes.any });\n",
			1,
		},

		// The call receiver is what declines these.
		{
			"a shape called on a call result, with the import",
			"import yup from \"yup\";\nYup.object().shape({ a: Yup.any });\n",
			0,
		},
		{
			"a shape called on a call result, without any import",
			"Yup.object().shape({ a: Yup.any });\n",
			0,
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := runForbidPropTypes(t, testCase.sourceText, nil)
			rule_testing.ExpectFindings(t, result, forbidPropTypesExpectedIds(testCase.wantCount)...)
		})
	}
}

// TestForbidPropTypesForeignPackageGuard pins the per-file import state.
//
// Importing anything named `PropTypes` from a module that is neither `react` nor `prop-types` makes
// the bare name in that file somebody else's, and every subsequent check then demands the recorded
// local name. Measured all four ways on the installed build.
//
// The state is per file and an import may sit after the usage it governs, so the imports are
// collected in a pass before anything is judged. The last case is what makes that pre-pass
// observable: the same code with the import moved below still reports.
func TestForbidPropTypesForeignPackageGuard(t *testing.T) {
	t.Parallel()

	body := "class C extends React.Component { static propTypes = { a: PropTypes.any }; render() { return null; } }\n"
	cases := []struct {
		name       string
		sourceText string
		wantCount  int
	}{
		{"no import at all", body, 1},
		{"imported from prop-types", "import PropTypes from 'prop-types';\n" + body, 1},
		{"imported from react", "import { PropTypes } from 'react';\n" + body, 1},
		{"imported from a third module", "import { PropTypes } from 'other-lib';\n" + body, 0},
		{
			"an aliased prop-types import used under its alias",
			"import PT from 'prop-types';\n" +
				"class C extends React.Component { static propTypes = { a: PT.any }; render() { return null; } }\n",
			1,
		},

		// The rows that need BOTH imports, which is what makes the recorded local name observable.
		// Without a foreign import the package predicate passes everything, so the recorded name is
		// never consulted and a mutant dropping the `import { PropTypes } from 'react'` branch
		// survives. These two are the same file shape with different sources for the alias, and
		// they have opposite verdicts. Both measured on the installed build.
		{
			"a react-aliased PropTypes beside a foreign import still reports",
			"import { PropTypes as PT } from 'react';\n" +
				"import { PropTypes } from 'other-lib';\n" +
				"class C extends React.Component { static propTypes = { a: PT.any }; render() { return null; } }\n",
			1,
		},
		{
			"an unrecorded alias beside a foreign import is silent",
			"import PT from 'prop-types';\n" +
				"import { PropTypes } from 'other-lib';\n" +
				"class C extends React.Component { static propTypes = { a: PT.any }; render() { return null; } }\n",
			0,
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := runForbidPropTypes(t, testCase.sourceText, nil)
			rule_testing.ExpectFindings(t, result, forbidPropTypesExpectedIds(testCase.wantCount)...)
		})
	}
}

// TestForbidPropTypesCollectsImportsBeforeJudging pins the ordering the pre-pass exists for.
//
// A foreign `PropTypes` import written BELOW the declaration it governs still silences it. Upstream
// gets this from ESLint visiting the whole program before the rule's other listeners can matter for
// a top-level import; our walk is a single pre-order pass, so the imports are collected explicitly
// first. Without that pass this case reports and the one above it does not, which is a difference no
// imported fixture can see because the corpus writes every import at the top.
func TestForbidPropTypesCollectsImportsBeforeJudging(t *testing.T) {
	t.Parallel()

	source := "class C extends React.Component { static propTypes = { a: PropTypes.any }; render() { return null; } }\n" +
		"import { PropTypes } from 'other-lib';\n"
	result := runForbidPropTypes(t, source, nil)
	rule_testing.ExpectClean(t, result)
}

// TestForbidPropTypesDoesNotUnwrapPropWrappers records a deliberate divergence rather than hiding it.
//
// Upstream also looks inside a "prop wrapper" call, so `C.propTypes = forbidExtraProps({...})` is
// checked. Which functions count comes from `settings.propWrapperFunctions`, a setting shared by
// four rules in this plugin and not present in this rule's own schema. Our `rule.Context` exposes
// SourceFile, Program, TypeChecker, Report and FileCache, and no path to configuration settings, so
// the arm has no substrate to stand on.
//
// Fourteen of upstream's 63 reporting cases configure that setting and are not imported; they
// assert 26 of the 82 findings. Every one is this shape. The divergence costs findings rather than
// inventing them, and this case pins the resulting silence so a later settings surface turns it
// into a failing test rather than leaving it unnoticed.
func TestForbidPropTypesDoesNotUnwrapPropWrappers(t *testing.T) {
	t.Parallel()

	source := "class C extends React.Component { render() { return null; } }\n" +
		"C.propTypes = forbidExtraProps({ a: PropTypes.any });\n"
	result := runForbidPropTypes(t, source, nil)
	rule_testing.ExpectClean(t, result)

	// The control: the same object without the wrapper reports, so the silence above is the wrapper
	// and not something else about the source.
	unwrapped := "class C extends React.Component { render() { return null; } }\n" +
		"C.propTypes = { a: PropTypes.any };\n"
	rule_testing.ExpectFindings(t, runForbidPropTypes(t, unwrapped, nil), "forbiddenPropType")
}

// TestForbidPropTypesSpans asserts where each finding points.
//
// The finding is anchored on the PROPERTY rather than on the value, so `a: PropTypes.any`
// underlines the whole property including its key. For a nested shape it underlines the inner
// property, which is the row a message-id fixture cannot see: both findings would otherwise look
// identical.
//
// `RunTyped` writes `strings.TrimSpace(source) + "\n"`, so the expectation is sliced from the same
// transform rather than from the Go literal.
func TestForbidPropTypesSpans(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		wantText   string
	}{
		{
			"a plain property underlines key and value",
			"X.propTypes = { a: PropTypes.any };\n",
			"a: PropTypes.any",
		},
		{
			"an isRequired property underlines the whole property",
			"X.propTypes = { a: PropTypes.any.isRequired };\n",
			"a: PropTypes.any.isRequired",
		},
		{
			"a wrapper underlines the outer property, not the argument",
			"X.propTypes = { a: PropTypes.arrayOf(PropTypes.object) };\n",
			"a: PropTypes.arrayOf(PropTypes.object)",
		},
		{
			"a nested shape underlines the INNER property",
			"X.propTypes = { a: PropTypes.shape({ b: PropTypes.any }) };\n",
			"b: PropTypes.any",
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := runForbidPropTypes(t, testCase.sourceText, nil)
			if len(result.Diagnostics) != 1 {
				t.Fatalf("wanted exactly one finding, got %d", len(result.Diagnostics))
			}
			written := strings.TrimSpace(testCase.sourceText) + "\n"
			diagnostic := result.Diagnostics[0]
			got := written[diagnostic.Range.Pos():diagnostic.Range.End()]
			if got != testCase.wantText {
				t.Errorf("span text = %q, wanted %q", got, testCase.wantText)
			}
		})
	}
}

// TestForbidPropTypesMessageText asserts the rendered message exactly.
//
// The message interpolates the offending type name, so the id assertion cannot see anything the
// format string does. Asserted by equality against a literal typed here rather than against the
// rule's own constant, which would move on both sides under mutation and could not fail.
func TestForbidPropTypesMessageText(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		value  string
		wanted string
	}{
		{"the type name is interpolated", "PropTypes.any", "any"},
		{"a wrapper names its argument's type", "PropTypes.arrayOf(PropTypes.object)", "object"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			source := "X.propTypes = { a: " + testCase.value + " };\n"
			result := runForbidPropTypes(t, source, nil)
			rule_testing.ExpectFindings(t, result, "forbiddenPropType")
			want := "Prop type \"" + testCase.wanted + "\" is forbidden here. A prop typed this " +
				"loosely documents nothing and checks nothing: every reader still has to find a " +
				"call site to learn what shape is really expected, and a wrong shape reaches the " +
				"component unreported. Declare the shape it actually takes instead."
			if got := result.Diagnostics[0].Message.Description; got != want {
				t.Errorf("message =\n%q\nwanted\n%q", got, want)
			}
		})
	}
}

// TestDecodeForbidPropTypesOptions pins the decoder, whose absent-versus-empty handling on `forbid`
// has no upstream counterpart.
func TestDecodeForbidPropTypesOptions(t *testing.T) {
	t.Parallel()

	t.Run("no options means the default list", func(t *testing.T) {
		t.Parallel()
		decoded, err := DecodeForbidPropTypesOptions(nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		options := decoded.(ForbidPropTypesOptions)
		if options.Forbid != nil {
			t.Errorf("Forbid = %v, wanted nil so the defaults apply", options.Forbid)
		}
		if got := options.forbiddenTypes(); len(got) != 3 {
			t.Errorf("effective list = %v, wanted the three defaults", got)
		}
	})

	t.Run("an empty object means the default list", func(t *testing.T) {
		t.Parallel()
		decoded, err := DecodeForbidPropTypesOptions([]byte(`{}`))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got := decoded.(ForbidPropTypesOptions).forbiddenTypes(); len(got) != 3 {
			t.Errorf("effective list = %v, wanted the three defaults", got)
		}
	})

	t.Run("an explicitly empty list forbids nothing", func(t *testing.T) {
		t.Parallel()
		decoded, err := DecodeForbidPropTypesOptions([]byte(`{"forbid":[]}`))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		options := decoded.(ForbidPropTypesOptions)
		if options.Forbid == nil {
			t.Fatal("Forbid is nil, so an explicit empty list was indistinguishable from absent")
		}
		if got := options.forbiddenTypes(); len(got) != 0 {
			t.Errorf("effective list = %v, wanted empty", got)
		}
	})

	t.Run("a custom list replaces the defaults", func(t *testing.T) {
		t.Parallel()
		decoded, err := DecodeForbidPropTypesOptions([]byte(`{"forbid":["string"]}`))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		got := decoded.(ForbidPropTypesOptions).forbiddenTypes()
		if len(got) != 1 || got[0] != "string" {
			t.Errorf("effective list = %v, wanted exactly [string]", got)
		}
	})

	t.Run("the two context flags decode", func(t *testing.T) {
		t.Parallel()
		decoded, err := DecodeForbidPropTypesOptions([]byte(`{"checkContextTypes":true,"checkChildContextTypes":true}`))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		options := decoded.(ForbidPropTypesOptions)
		if !options.CheckContextTypes || !options.CheckChildContextTypes {
			t.Errorf("flags = %v/%v, wanted both true", options.CheckContextTypes, options.CheckChildContextTypes)
		}
	})
}

// TestForbidPropTypesAssignmentOperators pins that every assignment operator qualifies.
//
// Upstream anchors on the member expression and reads `node.parent.right`, which any assignment
// carries, so a compound operator reports exactly like a plain one. This port narrowed it to `=` on
// the reasoning that a compound assignment has no meaning for a prop types object, and that
// reasoning was about what the code ought to mean rather than what upstream decides. It survived
// all 97 imported cases, because the corpus writes only plain assignment.
//
// All four rows measured against the installed build.
func TestForbidPropTypesAssignmentOperators(t *testing.T) {
	t.Parallel()

	for _, operator := range []string{"=", "||=", "??=", "+="} {
		t.Run(operator, func(t *testing.T) {
			t.Parallel()
			source := "class C extends React.Component { render() { return null; } }\n" +
				"C.propTypes " + operator + " { a: PropTypes.any };\n"
			result := runForbidPropTypes(t, source, nil)
			rule_testing.ExpectFindings(t, result, "forbiddenPropType")
		})
	}

	t.Run("a comparison is not an assignment", func(t *testing.T) {
		t.Parallel()
		// The control that keeps the operator test from being vacuous: without it, any binary
		// expression whose left side is a `.propTypes` access would be read as a declaration.
		source := "class C extends React.Component { render() { return null; } }\n" +
			"if (C.propTypes === { a: PropTypes.any }) {}\n"
		result := runForbidPropTypes(t, source, nil)
		rule_testing.ExpectClean(t, result)
	})
}

// TestForbidPropTypesGetterTakesTheLastReturn pins which return a multi-return getter uses.
//
// Upstream's `findReturnStatement` walks the body's statement list backwards and takes the LAST
// top-level return. Nothing in the corpus writes a getter with two returns, so a mutant reversing
// the search direction survived every imported case.
//
// The two rows have opposite verdicts from statement order alone, which is the cheapest possible
// distinguishing pair. Both measured on the installed build.
func TestForbidPropTypesGetterTakesTheLastReturn(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name      string
		body      string
		wantCount int
	}{
		// Both returns are TOP LEVEL statements, and that is load-bearing rather than incidental.
		// A first draft wrote the earlier return as `if (x) return ...`, which puts it inside an
		// if statement rather than in the body's own statement list, so only one return was at the
		// top level and the search direction could not matter. The mutant survived that fixture,
		// which is the brief's rule that a second failed fixture means the hypothesis needs
		// replacing rather than another case. The later return is unreachable at runtime and is
		// still a top-level statement, which is what the scan looks at.
		{
			"an allowed return followed by a forbidden one",
			"return { a: PropTypes.string };\n    return { a: PropTypes.any };",
			1,
		},
		{
			"a forbidden return followed by an allowed one",
			"return { a: PropTypes.any };\n    return { a: PropTypes.string };",
			0,
		},
		{
			// And a return nested inside an if is NOT top level, so it is not scanned at all.
			"a forbidden return nested inside an if, with an allowed one after",
			"if (x) return { a: PropTypes.any };\n    return { a: PropTypes.string };",
			0,
		},
		{
			// The control: one return, so order cannot be what decides.
			"a single forbidden return",
			"return { a: PropTypes.any };",
			1,
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			source := "declare const x: boolean;\n" +
				"class C extends React.Component {\n" +
				"  static get propTypes() {\n    " + testCase.body + "\n  }\n" +
				"  render() { return null; }\n}\n"
			result := runForbidPropTypes(t, source, nil)
			rule_testing.ExpectFindings(t, result, forbidPropTypesExpectedIds(testCase.wantCount)...)
		})
	}
}

// TestForbidPropTypesRecordedPackageNameReachesTheIdentifierArm pins the one route by which the
// recorded local name changes an answer.
//
// The predicate is not symmetric. Its member arm compares a receiver against `reactPackageName`
// only, so an aliased `prop-types` import never qualifies there; its identifier arm compares a bare
// value against `propTypesPackageName`. So the only shape on which the recorded name matters is a
// BARE prop type value whose name happens to equal that recorded local name, with a foreign
// `PropTypes` import present to close the general escape hatch.
//
// That took three attempts to locate. Two mutants survived here, one dropping the branch that
// records the name from `import { PropTypes } from 'react'` and one dropping the comparison in the
// identifier arm, and the first two fixtures written for them used a member value and could not
// reach either. This table is the shape that does.
//
// Contrived, and that is the point: everything less contrived cannot distinguish the versions, so a
// fixture over it would assert nothing. Every row measured against the installed build.
func TestForbidPropTypesRecordedPackageNameReachesTheIdentifierArm(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name      string
		imports   string
		value     string
		wantCount int
	}{
		{
			// The control: with no imports the predicate's last term passes everything.
			"a bare value with no imports at all",
			"",
			"any",
			1,
		},
		{
			// A foreign import closes that escape hatch, and no recorded name matches.
			"a bare value with only a foreign import",
			"import { PropTypes } from 'other-lib';\n",
			"any",
			0,
		},
		{
			// The row the recording branch exists for: `PropTypes` aliased to `any` from react
			// records `any`, which the identifier arm then accepts as the package.
			"a bare value matching a react-aliased package name",
			"import { PropTypes as any } from 'react';\nimport { PropTypes } from 'other-lib';\n",
			"any",
			1,
		},
		{
			// The same through a prop-types default import, which records the name the other way.
			"a bare value matching a prop-types aliased package name",
			"import any from 'prop-types';\nimport { PropTypes } from 'other-lib';\n",
			"any",
			1,
		},
		{
			// And a recorded name that does not match the value changes nothing.
			"a bare value not matching the recorded name",
			"import { PropTypes as PT } from 'react';\nimport { PropTypes } from 'other-lib';\n",
			"any",
			0,
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			source := testCase.imports +
				"class C extends React.Component { static propTypes = { a: " + testCase.value +
				" }; render() { return null; } }\n"
			result := runForbidPropTypes(t, source, nil)
			rule_testing.ExpectFindings(t, result, forbidPropTypesExpectedIds(testCase.wantCount)...)
		})
	}
}

// TestForbidPropTypesObjectLiteralArmDoesNotFollowIdentifiers pins an asymmetry between two arms
// that otherwise look interchangeable.
//
// The object-literal arm, which handles `createReactClass({ propTypes: ... })`, tests for an object
// expression and nothing else. The member-assignment arm routes through the full resolution, which
// also follows an identifier to what it was declared with. So the SAME identifier reports through
// one and is silent through the other.
//
// Nothing in the corpus writes a factory whose propTypes is an identifier, so a mutant widening the
// object-literal arm to accept any value survived every imported case. The first two rows are the
// same shape reached two ways with opposite verdicts, which is what makes this measurable at all.
// Every row measured against the installed build.
func TestForbidPropTypesObjectLiteralArmDoesNotFollowIdentifiers(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		wantCount  int
	}{
		{
			"a factory with an inline object",
			"var C = createReactClass({ propTypes: { a: PropTypes.any }, render() { return null; } });\n",
			1,
		},
		{
			"a factory whose propTypes is an identifier",
			"const shared = { a: PropTypes.any };\n" +
				"var C = createReactClass({ propTypes: shared, render() { return null; } });\n",
			0,
		},
		{
			"the same identifier through a member assignment",
			"const shared = { a: PropTypes.any };\n" +
				"class C extends React.Component { render() { return null; } }\nC.propTypes = shared;\n",
			1,
		},
		{
			"a factory whose propTypes is a call result",
			"declare function makeIt(): any;\n" +
				"var C = createReactClass({ propTypes: makeIt(), render() { return null; } });\n",
			0,
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := runForbidPropTypes(t, testCase.sourceText, nil)
			rule_testing.ExpectFindings(t, result, forbidPropTypesExpectedIds(testCase.wantCount)...)
		})
	}
}
