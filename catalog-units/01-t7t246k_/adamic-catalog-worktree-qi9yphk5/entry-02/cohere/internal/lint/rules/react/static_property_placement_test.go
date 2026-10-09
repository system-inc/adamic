package react

import (
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// The corpus is upstream's seventy seven cases, extracted by stubbing its RuleTester and capturing
// the case objects the test file hands it, so no case was retyped. All seventy seven were replayed
// against the installed build before being written down and every one reproduced its upstream
// verdict exactly, with no substitution needed.
//
// Each row carries its options as the RAW JSON the config layer would deliver, routed through this
// rule's own decoder rather than built as a struct. That matters more here than usual: upstream's
// option surface is a POSITIONAL pair whose two slots have different types, and flattening it into
// a per-property map is this port's own code with no upstream counterpart. Building the struct
// directly in a fixture would leave the whole decoder untested.
const staticPropertyPlacementFile = "/repository/source/StaticPropertyPlacement.tsx"

// staticPropertyPlacementDecode routes a fixture's raw options through the exported decoder.
func staticPropertyPlacementDecode(t *testing.T, raw string) any {
	t.Helper()
	decoded, err := DecodeStaticPropertyPlacementOptions([]byte(raw))
	if err != nil {
		t.Fatalf("decoding %s returned %v", raw, err)
	}
	return decoded
}

func TestStaticPropertyPlacementStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct{ name, sourceText, rawOptions string }{
		{"upstream valid 0", "\n        var MyComponent = createReactClass({\n          childContextTypes: {\n            something: PropTypes.bool\n          },\n\n          contextTypes: {\n            something: PropTypes.bool\n          },\n\n          getDefaultProps: function() {\n            name: 'Bob'\n          },\n\n          displayName: 'Hello',\n\n          propTypes: {\n            something: PropTypes.bool\n          },\n\n          render: function() {\n            return null;\n          },\n        });\n      ", "[\"property assignment\"]"},
		{"upstream valid 1", "\n        var MyComponent = React.createClass({\n          childContextTypes: {\n            something: PropTypes.bool\n          },\n\n          contextTypes: {\n            something: PropTypes.bool\n          },\n\n          getDefaultProps: function() {\n            name: 'Bob'\n          },\n\n          displayName: 'Hello',\n\n          propTypes: {\n            something: PropTypes.bool\n          },\n\n          render: function() {\n            return null;\n          },\n        });\n      ", "[\"property assignment\"]"},
		{"upstream valid 2", "\n        const MyComponent = () => {\n            return <div>Hello</div>;\n        };\n\n        MyComponent.childContextTypes = {\n          something: PropTypes.bool\n        };\n\n        MyComponent.contextTypes = {\n          something: PropTypes.bool\n        };\n\n        MyComponent.defaultProps = {\n          something: 'Bob'\n        };\n\n        MyComponent.displayName = 'Hello';\n\n        MyComponent.propTypes = {\n          something: PropTypes.bool\n        };\n      ", ""},
		{"upstream valid 3", "\n        const MyComponent = () => (<div>Hello</div>);\n\n        MyComponent.childContextTypes = {\n          something: PropTypes.bool\n        };\n\n        MyComponent.contextTypes = {\n          something: PropTypes.bool\n        };\n\n        MyComponent.defaultProps = {\n          something: 'Bob'\n        };\n\n        MyComponent.displayName = 'Hello';\n\n        MyComponent.propTypes = {\n          something: PropTypes.bool\n        };\n      ", ""},
		{"upstream valid 4", "\n        export function MyComponent () {\n            return <div>Hello</div>;\n        };\n\n        MyComponent.childContextTypes = {\n          something: PropTypes.bool\n        };\n\n        MyComponent.contextTypes = {\n          something: PropTypes.bool\n        };\n\n        MyComponent.defaultProps = {\n          something: 'Bob'\n        };\n\n        MyComponent.displayName = 'Hello';\n\n        MyComponent.propTypes = {\n          something: PropTypes.bool\n        };\n      ", ""},
		{"upstream valid 5", "\n        class Foo {\n          static get propTypes() {}\n        }\n      ", ""},
		{"upstream valid 6", "\n        class Foo {\n          static propTypes = {}\n        }\n      ", "[\"property assignment\"]"},
		{"upstream valid 7", "\n        class MyComponent extends React.Component {\n          render() {\n            return null;\n          }\n        }\n      ", ""},
		{"upstream valid 8", "\n        class MyComponent extends React.Component {\n          static randomlyNamed = {\n            name: 'random'\n          }\n        }\n      ", ""},
		{"upstream valid 9", "\n        class MyComponent extends React.Component {\n          static randomlyNamed = {\n            name: 'random'\n          }\n        }\n      ", "[\"property assignment\"]"},
		{"upstream valid 10", "\n        class MyComponent extends React.Component {\n          render() {\n            return null;\n          }\n        }\n\n        MyComponent.randomlyNamed = {\n          name: 'random'\n        }\n      ", "[\"property assignment\"]"},
		{"upstream valid 11", "\n        class MyComponent extends React.Component {\n          render() {\n            return null;\n          }\n        }\n\n        MyComponent.randomlyNamed = {\n          name: 'random'\n        }\n      ", ""},
		{"upstream valid 12", "\n        class MyComponent extends React.Component {\n          static childContextTypes = {\n            something: PropTypes.bool\n          };\n        }\n      ", ""},
		{"upstream valid 13", "\n        class MyComponent extends React.Component {\n          static childContextTypes = {\n            something: PropTypes.bool\n          };\n        }\n      ", "[\"property assignment\", {\"childContextTypes\": \"static public field\"}]"},
		{"upstream valid 14", "\n        class MyComponent extends React.Component {\n          static get childContextTypes() {\n            return {\n              something: PropTypes.bool\n            };\n          }\n        }\n      ", "[\"static getter\"]"},
		{"upstream valid 15", "\n        class MyComponent extends React.Component {\n          static get childContextTypes() {\n            return {\n              something: PropTypes.bool\n            };\n          }\n        }\n      ", "[\"property assignment\", {\"childContextTypes\": \"static getter\"}]"},
		{"upstream valid 16", "\n        class MyComponent extends React.Component {\n          render() {\n            return null;\n          }\n        }\n\n        MyComponent.childContextTypes = {\n          name: PropTypes.string.isRequired\n        }\n      ", "[\"property assignment\"]"},
		{"upstream valid 17", "\n        class MyComponent extends React.Component {\n          render() {\n            return null;\n          }\n        }\n\n        MyComponent.childContextTypes = {\n          name: PropTypes.string.isRequired\n        }\n      ", "[\"static public field\", {\"childContextTypes\": \"property assignment\"}]"},
		{"upstream valid 18", "\n        class MyComponent extends React.Component {\n          static contextTypes = {\n            something: PropTypes.bool\n          };\n        }\n      ", ""},
		{"upstream valid 19", "\n        class MyComponent extends React.Component {\n          static contextTypes = {\n            something: PropTypes.bool\n          };\n        }\n      ", "[\"property assignment\", {\"contextTypes\": \"static public field\"}]"},
		{"upstream valid 20", "\n        class MyComponent extends React.Component {\n          static get contextTypes() {\n            return {\n              something: PropTypes.bool\n            };\n          }\n        }\n      ", "[\"static getter\"]"},
		{"upstream valid 21", "\n        class MyComponent extends React.Component {\n          static get contextTypes() {\n            return {\n              something: PropTypes.bool\n            };\n          }\n        }\n      ", "[\"property assignment\", {\"contextTypes\": \"static getter\"}]"},
		{"upstream valid 22", "\n        class MyComponent extends React.Component {\n          render() {\n            return null;\n          }\n        }\n\n        MyComponent.contextTypes = {\n          name: PropTypes.string.isRequired\n        }\n      ", "[\"property assignment\"]"},
		{"upstream valid 23", "\n        class MyComponent extends React.Component {\n          render() {\n            return null;\n          }\n        }\n\n        MyComponent.contextTypes = {\n          name: PropTypes.string.isRequired\n        }\n      ", "[\"static public field\", {\"contextTypes\": \"property assignment\"}]"},
		{"upstream valid 24", "\n        class MyComponent extends React.Component {\n          static contextType = MyContext;\n        }\n      ", ""},
		{"upstream valid 25", "\n        class MyComponent extends React.Component {\n          static contextType = MyContext;\n        }\n      ", "[\"property assignment\", {\"contextType\": \"static public field\"}]"},
		{"upstream valid 26", "\n        class MyComponent extends React.Component {\n          static get contextType() {\n             return MyContext;\n          }\n        }\n      ", "[\"static getter\"]"},
		{"upstream valid 27", "\n        class MyComponent extends React.Component {\n          static get contextType() {\n             return MyContext;\n          }\n        }\n      ", "[\"property assignment\", {\"contextType\": \"static getter\"}]"},
		{"upstream valid 28", "\n        class MyComponent extends React.Component {\n          render() {\n            return null;\n          }\n        }\n\n        MyComponent.contextType = MyContext;\n      ", "[\"property assignment\"]"},
		{"upstream valid 29", "\n        class MyComponent extends React.Component {\n          render() {\n            return null;\n          }\n        }\n\n        MyComponent.contextType = MyContext;\n      ", "[\"static public field\", {\"contextType\": \"property assignment\"}]"},
		{"upstream valid 30", "\n        class MyComponent extends React.Component {\n          static displayName = \"Hello\";\n        }\n      ", ""},
		{"upstream valid 31", "\n        class MyComponent extends React.Component {\n          static displayName = \"Hello\";\n        }\n      ", "[\"property assignment\", {\"displayName\": \"static public field\"}]"},
		{"upstream valid 32", "\n        class MyComponent extends React.Component {\n          static get displayName() {\n            return \"Hello\";\n          }\n        }\n      ", "[\"static getter\"]"},
		{"upstream valid 33", "\n        class MyComponent extends React.Component {\n          static get displayName() {\n            return \"Hello\";\n          }\n        }\n      ", "[\"property assignment\", {\"displayName\": \"static getter\"}]"},
		{"upstream valid 34", "\n        class MyComponent extends React.Component {\n          render() {\n            return null;\n          }\n        }\n\n        MyComponent.displayName = \"Hello\";\n      ", "[\"property assignment\"]"},
		{"upstream valid 35", "\n        class MyComponent extends React.Component {\n          render() {\n            return null;\n          }\n        }\n\n        MyComponent.displayName = \"Hello\";\n      ", "[\"static public field\", {\"displayName\": \"property assignment\"}]"},
		{"upstream valid 36", "\n        class MyComponent extends React.Component {\n          static defaultProps = {\n            something: 'Bob'\n          };\n        }\n      ", ""},
		{"upstream valid 37", "\n        class MyComponent extends React.Component {\n          static defaultProps = {\n            something: 'Bob'\n          };\n        }\n      ", "[\"property assignment\", {\"defaultProps\": \"static public field\"}]"},
		{"upstream valid 38", "\n        class MyComponent extends React.Component {\n          static get defaultProps() {\n            return {\n              something: 'Bob'\n            };\n          }\n        }\n      ", "[\"static getter\"]"},
		{"upstream valid 39", "\n        class MyComponent extends React.Component {\n          static get defaultProps() {\n            return {\n              something: 'Bob'\n            };\n          }\n        }\n      ", "[\"property assignment\", {\"defaultProps\": \"static getter\"}]"},
		{"upstream valid 40", "\n        class MyComponent extends React.Component {\n          render() {\n            return null;\n          }\n        }\n\n        MyComponent.defaultProps = {\n          name: 'Bob'\n        }\n      ", "[\"property assignment\"]"},
		{"upstream valid 41", "\n        class MyComponent extends React.Component {\n          render() {\n            return null;\n          }\n        }\n\n        MyComponent.defaultProps = {\n          name: 'Bob'\n        }\n      ", "[\"static public field\", {\"defaultProps\": \"property assignment\"}]"},
		{"upstream valid 42", "\n        class MyComponent extends React.Component {\n          static propTypes = {\n            something: PropTypes.bool\n          };\n        }\n      ", ""},
		{"upstream valid 43", "\n        class MyComponent extends React.Component {\n          static propTypes = {\n            something: PropTypes.bool\n          };\n        }\n      ", "[\"property assignment\", {\"propTypes\": \"static public field\"}]"},
		{"upstream valid 44", "\n        class MyComponent extends React.Component {\n          static get propTypes() {\n            return {\n              something: PropTypes.bool\n            };\n          }\n        }\n      ", "[\"static getter\"]"},
		{"upstream valid 45", "\n        class MyComponent extends React.Component {\n          static get propTypes() {\n            return {\n              something: PropTypes.bool\n            };\n          }\n        }\n      ", "[\"property assignment\", {\"propTypes\": \"static getter\"}]"},
		{"upstream valid 46", "\n        class MyComponent extends React.Component {\n          render() {\n            return null;\n          }\n        }\n\n        MyComponent.propTypes = {\n          name: PropTypes.string.isRequired\n        }\n      ", "[\"property assignment\"]"},
		{"upstream valid 47", "\n        class MyComponent extends React.Component {\n          render() {\n            return null;\n          }\n        }\n\n        MyComponent.propTypes = {\n          name: PropTypes.string.isRequired\n        }\n      ", "[\"static public field\", {\"propTypes\": \"property assignment\"}]"},
		{"upstream valid 48", "\n        class MyComponent extends React.Component {\n          static childContextTypes = {\n            something: PropTypes.bool\n          };\n\n          static contextTypes = {\n            something: PropTypes.bool\n          };\n\n          static contextType = MyContext;\n\n          static displayName = \"Hello\";\n\n          static defaultProps = {\n            something: 'Bob'\n          };\n\n          static propTypes = {\n            something: PropTypes.bool\n          };\n        }\n      ", ""},
		{"upstream valid 49", "\n        class MyComponent extends React.Component {\n          static childContextTypes = {\n            something: PropTypes.bool\n          };\n\n          static contextTypes = {\n            something: PropTypes.bool\n          };\n\n          static contextType = MyContext;\n\n          static displayName = \"Hello\";\n\n          static defaultProps = {\n            something: 'Bob'\n          };\n\n          static propTypes = {\n            something: PropTypes.bool\n          };\n        }\n      ", "[\"property assignment\", {\"childContextTypes\": \"static public field\", \"contextTypes\": \"static public field\", \"contextType\": \"static public field\", \"displayName\": \"static public field\", \"defaultProps\": \"static public field\", \"propTypes\": \"static public field\"}]"},
		{"upstream valid 50", "\n        class MyComponent extends React.Component {\n          static get childContextTypes() {\n            return {\n              something: PropTypes.bool\n            };\n          }\n\n          static get contextTypes() {\n            return {\n              something: PropTypes.bool\n            };\n          }\n\n          static get contextType() {\n            return MyContext;\n          }\n\n          static get displayName() {\n            return \"Hello\";\n          }\n\n          static get defaultProps() {\n            return {\n              something: PropTypes.bool\n            };\n          }\n\n          static get propTypes() {\n            return {\n              something: PropTypes.bool\n            };\n          }\n        }\n      ", "[\"static getter\"]"},
		{"upstream valid 51", "\n        class MyComponent extends React.Component {\n          static get childContextTypes() {\n            return {\n              something: PropTypes.bool\n            };\n          }\n\n          static get contextTypes() {\n            return {\n              something: PropTypes.bool\n            };\n          }\n\n          static get contextType() {\n            return MyContext;\n          }\n\n          static get displayName() {\n            return \"Hello\";\n          }\n\n          static get defaultProps() {\n            return {\n              something: PropTypes.bool\n            };\n          }\n\n          static get propTypes() {\n            return {\n              something: PropTypes.bool\n            };\n          }\n        }\n      ", "[\"property assignment\", {\"childContextTypes\": \"static getter\", \"contextTypes\": \"static getter\", \"contextType\": \"static getter\", \"displayName\": \"static getter\", \"defaultProps\": \"static getter\", \"propTypes\": \"static getter\"}]"},
		{"upstream valid 52", "\n        class MyComponent extends React.Component {\n          render() {\n            return null;\n          }\n        }\n\n        MyComponent.childContextTypes = {\n          name: PropTypes.string.isRequired\n        }\n\n        MyComponent.contextTypes = {\n          name: PropTypes.string.isRequired\n        }\n\n        MyComponent.displayName = \"Hello\";\n\n        MyComponent.defaultProps = {\n          name: 'Bob'\n        }\n\n        MyComponent.propTypes = {\n          name: PropTypes.string.isRequired\n        }\n      ", "[\"property assignment\"]"},
		{"upstream valid 53", "\n        class MyComponent extends React.Component {\n          render() {\n            return null;\n          }\n        }\n\n        MyComponent.childContextTypes = {\n          name: PropTypes.string.isRequired\n        }\n\n        MyComponent.contextTypes = {\n          name: PropTypes.string.isRequired\n        }\n\n        MyComponent.displayName = \"Hello\";\n\n        MyComponent.defaultProps = {\n          name: 'Bob'\n        }\n\n        MyComponent.propTypes = {\n          name: PropTypes.string.isRequired\n        }\n      ", "[\"static public field\", {\"childContextTypes\": \"property assignment\", \"contextTypes\": \"property assignment\", \"displayName\": \"property assignment\", \"defaultProps\": \"property assignment\", \"propTypes\": \"property assignment\"}]"},
		{"upstream valid 54", "\n        class MyComponent extends React.Component {\n          static childContextTypes = {\n            name: PropTypes.string.isRequired\n          }\n\n          static contextTypes = {\n            name: PropTypes.string.isRequired\n          }\n\n          static get displayName() {\n            return \"Hello\"\n          }\n        }\n\n        MyComponent.defaultProps = {\n          name: 'Bob'\n        }\n\n        MyComponent.propTypes = {\n          name: PropTypes.string.isRequired\n        }\n      ", "[\"static public field\", {\"displayName\": \"static getter\", \"defaultProps\": \"property assignment\", \"propTypes\": \"property assignment\"}]"},
		{"upstream valid 55", "\n        class MyComponent extends React.Component {\n          static childContextTypes = {\n            name: PropTypes.string.isRequired\n          }\n\n          static contextTypes = {\n            name: PropTypes.string.isRequired\n          }\n\n          static get displayName() {\n            return \"Hello\"\n          }\n        }\n\n        MyComponent.defaultProps = {\n          name: 'Bob'\n        }\n\n        MyComponent.propTypes = {\n          name: PropTypes.string.isRequired\n        }\n      ", "[\"property assignment\", {\"childContextTypes\": \"static public field\", \"contextTypes\": \"static public field\", \"displayName\": \"static getter\"}]"},
		{"upstream valid 56", "\n        class MyComponent extends React.Component {\n          static childContextTypes = {\n            name: PropTypes.string.isRequired\n          }\n\n          static contextTypes = {\n            name: PropTypes.string.isRequired\n          }\n\n          static displayName = \"Hello\";\n        }\n\n        const OtherComponent = () => (<div>Hello</div>);\n\n        OtherComponent.defaultProps = {\n          name: 'Bob'\n        }\n\n        OtherComponent.propTypes = {\n          name: PropTypes.string.isRequired\n        }\n      ", ""},
		{"upstream valid 57", "\n        class MyComponent extends React.Component {\n          static childContextTypes = {\n            name: PropTypes.string.isRequired\n          }\n\n          static contextTypes = {\n            name: PropTypes.string.isRequired\n          }\n\n          static displayName = \"Hello\";\n        }\n\n        class OtherComponent extends React.Component {\n          static defaultProps = {\n            name: 'Bob'\n          }\n\n          static propTypes = {\n            name: PropTypes.string.isRequired\n          }\n        }\n      ", ""},
		{"upstream valid 58", "\n        class MyComponent extends React.Component {\n          static displayName = \"Hello\";\n\n          myMethod() {\n            console.log(MyComponent.displayName);\n          }\n        }\n      ", "[\"static public field\"]"},
		{"upstream valid 59", "\n        class MyComponent extends React.Component {\n          static displayName = \"Hello\";\n\n          myMethod() {\n            MyComponent.displayName = \"Bonjour\";\n          }\n        }\n      ", "[\"static public field\"]"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTypedWithOptions(t, StaticPropertyPlacement,
				staticPropertyPlacementFile, testCase.sourceText,
				staticPropertyPlacementDecode(t, testCase.rawOptions))
			rule_testing.ExpectClean(t, result)
		})
	}
}

func TestStaticPropertyPlacementFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name, sourceText, rawOptions string
		messageIds                   []string
	}{
		{"upstream invalid 0", "\n        class MyComponent extends React.Component {\n          render() {\n            return null;\n          }\n        }\n\n        MyComponent.childContextTypes = {\n          name: PropTypes.string.isRequired\n        }\n\n        MyComponent.contextTypes = {\n          name: PropTypes.string.isRequired\n        }\n\n        MyComponent.contextType = MyContext;\n\n        MyComponent.displayName = \"Hello\";\n\n        MyComponent.defaultProps = {\n          name: 'Bob'\n        }\n\n        MyComponent.propTypes = {\n          name: PropTypes.string.isRequired\n        }\n      ", "", []string{"notStaticClassProp", "notStaticClassProp", "notStaticClassProp", "notStaticClassProp", "notStaticClassProp", "notStaticClassProp"}},
		{"upstream invalid 1", "\n        class MyComponent extends React.Component {\n          render() {\n            return null;\n          }\n        }\n\n        MyComponent.childContextTypes = {\n          name: PropTypes.string.isRequired\n        }\n\n        MyComponent.contextTypes = {\n          name: PropTypes.string.isRequired\n        }\n\n        MyComponent.contextType = MyContext;\n\n        MyComponent.displayName = \"Hello\";\n\n        MyComponent.defaultProps = {\n          name: 'Bob'\n        }\n\n        MyComponent.propTypes = {\n          name: PropTypes.string.isRequired\n        }\n      ", "[\"property assignment\", {\"childContextTypes\": \"static public field\", \"contextTypes\": \"static public field\", \"contextType\": \"static public field\", \"displayName\": \"static public field\", \"defaultProps\": \"static public field\", \"propTypes\": \"static public field\"}]", []string{"notStaticClassProp", "notStaticClassProp", "notStaticClassProp", "notStaticClassProp", "notStaticClassProp", "notStaticClassProp"}},
		{"upstream invalid 2", "\n        class MyComponent extends React.Component {\n          static get childContextTypes() {\n            return {\n              something: PropTypes.bool\n            };\n          }\n\n          static get contextTypes() {\n            return {\n              something: PropTypes.bool\n            };\n          }\n\n          static get contextType() {\n            return MyContext;\n          }\n\n          static get displayName() {\n            return \"Hello\";\n          }\n\n          static get defaultProps() {\n            return {\n              something: PropTypes.bool\n            };\n          }\n\n          static get propTypes() {\n            return {\n              something: PropTypes.bool\n            };\n          }\n        }\n      ", "", []string{"notStaticClassProp", "notStaticClassProp", "notStaticClassProp", "notStaticClassProp", "notStaticClassProp", "notStaticClassProp"}},
		{"upstream invalid 3", "\n        class MyComponent extends React.Component {\n          static get childContextTypes() {\n            return {\n              something: PropTypes.bool\n            };\n          }\n\n          static get contextTypes() {\n            return {\n              something: PropTypes.bool\n            };\n          }\n\n          static get contextType() {\n            return MyContext;\n          }\n\n          static get displayName() {\n            return \"Hello\";\n          }\n\n          static get defaultProps() {\n            return {\n              something: PropTypes.bool\n            };\n          }\n\n          static get propTypes() {\n            return {\n              something: PropTypes.bool\n            };\n          }\n        }\n      ", "[\"property assignment\", {\"childContextTypes\": \"static public field\", \"contextTypes\": \"static public field\", \"contextType\": \"static public field\", \"displayName\": \"static public field\", \"defaultProps\": \"static public field\", \"propTypes\": \"static public field\"}]", []string{"notStaticClassProp", "notStaticClassProp", "notStaticClassProp", "notStaticClassProp", "notStaticClassProp", "notStaticClassProp"}},
		{"upstream invalid 4", "\n        class MyComponent extends React.Component {\n          static childContextTypes = {\n            something: PropTypes.bool\n          };\n\n          static contextTypes = {\n            something: PropTypes.bool\n          };\n\n          static contextType = MyContext;\n\n          static displayName = \"Hello\";\n\n          static defaultProps = {\n            something: 'Bob'\n          };\n\n          static propTypes = {\n            something: PropTypes.bool\n          };\n        }\n      ", "[\"property assignment\"]", []string{"declareOutsideClass", "declareOutsideClass", "declareOutsideClass", "declareOutsideClass", "declareOutsideClass", "declareOutsideClass"}},
		{"upstream invalid 5", "\n        class MyComponent extends React.Component {\n          static childContextTypes = {\n            something: PropTypes.bool\n          };\n\n          static contextTypes = {\n            something: PropTypes.bool\n          };\n\n          static contextType = MyContext;\n\n          static displayName = \"Hello\";\n\n          static defaultProps = {\n            something: 'Bob'\n          };\n\n          static propTypes = {\n            something: PropTypes.bool\n          };\n        }\n      ", "[\"static public field\", {\"childContextTypes\": \"property assignment\", \"contextTypes\": \"property assignment\", \"contextType\": \"property assignment\", \"displayName\": \"property assignment\", \"defaultProps\": \"property assignment\", \"propTypes\": \"property assignment\"}]", []string{"declareOutsideClass", "declareOutsideClass", "declareOutsideClass", "declareOutsideClass", "declareOutsideClass", "declareOutsideClass"}},
		{"upstream invalid 6", "\n        class MyComponent extends React.Component {\n          static get childContextTypes() {\n            return {\n              something: PropTypes.bool\n            };\n          }\n\n          static get contextTypes() {\n            return {\n              something: PropTypes.bool\n            };\n          }\n\n          static get contextType() {\n            return MyContext;\n          }\n\n          static get displayName() {\n            return \"Hello\";\n          }\n\n          static get defaultProps() {\n            return {\n              something: PropTypes.bool\n            };\n          }\n\n          static get propTypes() {\n            return {\n              something: PropTypes.bool\n            };\n          }\n        }\n      ", "[\"property assignment\"]", []string{"declareOutsideClass", "declareOutsideClass", "declareOutsideClass", "declareOutsideClass", "declareOutsideClass", "declareOutsideClass"}},
		{"upstream invalid 7", "\n        class MyComponent extends React.Component {\n          static get childContextTypes() {\n            return {\n              something: PropTypes.bool\n            };\n          }\n\n          static get contextTypes() {\n            return {\n              something: PropTypes.bool\n            };\n          }\n\n          static get contextType() {\n            return MyContext;\n          }\n\n          static get displayName() {\n            return \"Hello\";\n          }\n\n          static get defaultProps() {\n            return {\n              something: PropTypes.bool\n            };\n          }\n\n          static get propTypes() {\n            return {\n              something: PropTypes.bool\n            };\n          }\n        }\n      ", "[\"static getter\", {\"childContextTypes\": \"property assignment\", \"contextTypes\": \"property assignment\", \"contextType\": \"property assignment\", \"displayName\": \"property assignment\", \"defaultProps\": \"property assignment\", \"propTypes\": \"property assignment\"}]", []string{"declareOutsideClass", "declareOutsideClass", "declareOutsideClass", "declareOutsideClass", "declareOutsideClass", "declareOutsideClass"}},
		{"upstream invalid 8", "\n        class MyComponent extends React.Component {\n          static childContextTypes = {\n            something: PropTypes.bool\n          };\n\n          static contextTypes = {\n            something: PropTypes.bool\n          };\n\n          static contextType = MyContext;\n\n          static displayName = \"Hello\";\n\n          static defaultProps = {\n            something: 'Bob'\n          };\n\n          static propTypes = {\n            something: PropTypes.bool\n          };\n        }\n      ", "[\"static getter\"]", []string{"notGetterClassFunc", "notGetterClassFunc", "notGetterClassFunc", "notGetterClassFunc", "notGetterClassFunc", "notGetterClassFunc"}},
		{"upstream invalid 9", "\n        class MyComponent extends React.Component {\n          static childContextTypes = {\n            something: PropTypes.bool\n          };\n\n          static contextTypes = {\n            something: PropTypes.bool\n          };\n\n          static contextType = MyContext;\n\n          static displayName = \"Hello\";\n\n          static defaultProps = {\n            something: 'Bob'\n          };\n\n          static propTypes = {\n            something: PropTypes.bool\n          };\n        }\n      ", "[\"static public field\", {\"childContextTypes\": \"static getter\", \"contextTypes\": \"static getter\", \"contextType\": \"static getter\", \"displayName\": \"static getter\", \"defaultProps\": \"static getter\", \"propTypes\": \"static getter\"}]", []string{"notGetterClassFunc", "notGetterClassFunc", "notGetterClassFunc", "notGetterClassFunc", "notGetterClassFunc", "notGetterClassFunc"}},
		{"upstream invalid 10", "\n        class MyComponent extends React.Component {\n          render() {\n            return null;\n          }\n        }\n\n        MyComponent.childContextTypes = {\n          name: PropTypes.string.isRequired\n        }\n\n        MyComponent.contextTypes = {\n          name: PropTypes.string.isRequired\n        }\n\n        MyComponent.contextType = MyContext;\n\n        MyComponent.displayName = \"Hello\";\n\n        MyComponent.defaultProps = {\n          name: 'Bob'\n        }\n\n        MyComponent.propTypes = {\n          name: PropTypes.string.isRequired\n        }\n      ", "[\"static getter\"]", []string{"notGetterClassFunc", "notGetterClassFunc", "notGetterClassFunc", "notGetterClassFunc", "notGetterClassFunc", "notGetterClassFunc"}},
		{"upstream invalid 11", "\n        class MyComponent extends React.Component {\n          render() {\n            return null;\n          }\n        }\n\n        MyComponent.childContextTypes = {\n          name: PropTypes.string.isRequired\n        }\n\n        MyComponent.contextTypes = {\n          name: PropTypes.string.isRequired\n        }\n\n        MyComponent.contextType = MyContext;\n\n        MyComponent.displayName = \"Hello\";\n\n        MyComponent.defaultProps = {\n          name: 'Bob'\n        }\n\n        MyComponent.propTypes = {\n          name: PropTypes.string.isRequired\n        }\n      ", "[\"property assignment\", {\"childContextTypes\": \"static getter\", \"contextTypes\": \"static getter\", \"contextType\": \"static getter\", \"displayName\": \"static getter\", \"defaultProps\": \"static getter\", \"propTypes\": \"static getter\"}]", []string{"notGetterClassFunc", "notGetterClassFunc", "notGetterClassFunc", "notGetterClassFunc", "notGetterClassFunc", "notGetterClassFunc"}},
		{"upstream invalid 12", "\n        class MyComponent extends React.Component {\n          static childContextTypes = {\n            name: PropTypes.string.isRequired\n          }\n\n          static contextTypes = {\n            name: PropTypes.string.isRequired\n          }\n\n          static contextType = MyContext;\n\n          static get displayName() {\n            return \"Hello\";\n          }\n        }\n\n        MyComponent.defaultProps = {\n          name: 'Bob'\n        }\n\n        MyComponent.propTypes = {\n          name: PropTypes.string.isRequired\n        }\n      ", "[\"property assignment\", {\"defaultProps\": \"static getter\", \"propTypes\": \"static public field\", \"displayName\": \"static public field\"}]", []string{"declareOutsideClass", "declareOutsideClass", "declareOutsideClass", "notStaticClassProp", "notGetterClassFunc", "notStaticClassProp"}},
		{"upstream invalid 13", "\n        class MyComponent extends React.Component {\n          static childContextTypes = {\n            name: PropTypes.string.isRequired\n          }\n\n          static contextTypes = {\n            name: PropTypes.string.isRequired\n          }\n\n          static contextType = MyContext;\n\n          static get displayName() {\n            return \"Hello\";\n          }\n        }\n\n        MyComponent.defaultProps = {\n          name: 'Bob'\n        }\n\n        MyComponent.propTypes = {\n          name: PropTypes.string.isRequired\n        }\n      ", "[\"static getter\", {\"childContextTypes\": \"property assignment\", \"contextTypes\": \"property assignment\", \"contextType\": \"property assignment\", \"displayName\": \"property assignment\"}]", []string{"declareOutsideClass", "declareOutsideClass", "declareOutsideClass", "declareOutsideClass", "notGetterClassFunc", "notGetterClassFunc"}},
		{"upstream invalid 14", "\n        class MyComponent extends React.Component {\n          static childContextTypes = {\n            name: PropTypes.string.isRequired\n          }\n\n          static contextTypes = {\n            name: PropTypes.string.isRequired\n          }\n\n          static contextType = MyContext;\n\n          static get displayName() {\n            return \"Hello\";\n          }\n        }\n\n        const OtherComponent = () => (<div>Hello</div>);\n\n        OtherComponent.defaultProps = {\n          name: 'Bob'\n        }\n\n        OtherComponent.propTypes = {\n          name: PropTypes.string.isRequired\n        }\n      ", "[\"property assignment\", {\"defaultProps\": \"static public field\", \"propTypes\": \"static getter\"}]", []string{"declareOutsideClass", "declareOutsideClass", "declareOutsideClass", "declareOutsideClass"}},
		{"upstream invalid 15", "\n        class MyComponent extends React.Component {\n          static childContextTypes = {\n            name: PropTypes.string.isRequired\n          }\n\n          static contextTypes = {\n            name: PropTypes.string.isRequired\n          }\n\n          static contextType = MyContext;\n\n          static displayName = \"Hello\";\n        }\n\n        class OtherComponent extends React.Component {\n          static contextTypes = {\n            name: PropTypes.string.isRequired\n          }\n\n          static defaultProps = {\n            name: 'Bob'\n          }\n\n          static propTypes = {\n            name: PropTypes.string.isRequired\n          }\n\n          static get displayName() {\n            return \"Hello\";\n          }\n        }\n      ", "[\"property assignment\"]", []string{"declareOutsideClass", "declareOutsideClass", "declareOutsideClass", "declareOutsideClass", "declareOutsideClass", "declareOutsideClass", "declareOutsideClass", "declareOutsideClass"}},
		{"upstream invalid 16", "\n        class MyComponent extends React.Component {\n          displayName = 'Foo';\n        }\n      ", "[\"static public field\"]", []string{"notStaticClassProp"}},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTypedWithOptions(t, StaticPropertyPlacement,
				staticPropertyPlacementFile, testCase.sourceText,
				staticPropertyPlacementDecode(t, testCase.rawOptions))
			rule_testing.ExpectFindings(t, result, testCase.messageIds...)
		})
	}
}

// TestStaticPropertyPlacementPositionMatrix pins every arm against every configured position.
//
// All nine rows measured against the installed build on 2026-08-27. The two that would not occur to
// a reader are the last two: a NON-static getter is silent under every configuration, because
// upstream requires both `static` and the getter kind before it looks at the name at all, and an
// unrelated static field is silent because it is not one of the six names.
func TestStaticPropertyPlacementPositionMatrix(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name, sourceText, rawOptions string
		messageIds                   []string
	}{
		{
			"a static getter under the getter position is correct",
			"class MyComponent extends React.Component { static get propTypes(){ return {}; } render(){ return null; } }",
			`["static getter"]`, nil,
		},
		{
			"a static getter under the field position reports as a field",
			"class MyComponent extends React.Component { static get propTypes(){ return {}; } render(){ return null; } }",
			`["static public field"]`, []string{"notStaticClassProp"},
		},
		{
			"a non static field under the field position still reports",
			"class MyComponent extends React.Component { propTypes = {}; render(){ return null; } }",
			`["static public field"]`, []string{"notStaticClassProp"},
		},
		{
			"a non static field under the getter position reports as a getter",
			"class MyComponent extends React.Component { propTypes = {}; render(){ return null; } }",
			`["static getter"]`, []string{"notGetterClassFunc"},
		},
		{
			"a non static field under the assignment position reports as an assignment",
			"class MyComponent extends React.Component { propTypes = {}; render(){ return null; } }",
			`["property assignment"]`, []string{"declareOutsideClass"},
		},
		{
			"a static field under the assignment position reports",
			"class MyComponent extends React.Component { static propTypes = {}; render(){ return null; } }",
			`["property assignment"]`, []string{"declareOutsideClass"},
		},
		{
			"displayName is one of the six names",
			"class MyComponent extends React.Component { static displayName = \"x\"; render(){ return null; } }",
			`["property assignment"]`, []string{"declareOutsideClass"},
		},
		{
			"a NON static getter is silent under the getter position",
			"class MyComponent extends React.Component { get propTypes(){ return {}; } render(){ return null; } }",
			`["static getter"]`, nil,
		},
		{
			// The two rows below are what make the getter arm's static gate visible. Under the
			// getter position a non static getter is silent either way, because the arm passes
			// through and the configured position matches, so only a DIFFERENT configured position
			// separates a rule that gates from one that does not. Both measured silent upstream.
			"a NON static getter is silent under the field position too",
			"class MyComponent extends React.Component { get propTypes(){ return {}; } render(){ return null; } }",
			`["static public field"]`, nil,
		},
		{
			"a NON static getter is silent under the assignment position too",
			"class MyComponent extends React.Component { get propTypes(){ return {}; } render(){ return null; } }",
			`["property assignment"]`, nil,
		},
		{
			"a static field with an unrelated name is not judged",
			"class MyComponent extends React.Component { static other = {}; render(){ return null; } }",
			`["property assignment"]`, nil,
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTypedWithOptions(t, StaticPropertyPlacement,
				staticPropertyPlacementFile, testCase.sourceText,
				staticPropertyPlacementDecode(t, testCase.rawOptions))
			if len(testCase.messageIds) == 0 {
				rule_testing.ExpectClean(t, result)
				return
			}
			rule_testing.ExpectFindings(t, result, testCase.messageIds...)
		})
	}
}

// TestStaticPropertyPlacementAssignmentReceiver pins which assignment receivers resolve.
//
// Upstream's corpus writes exactly one receiver shape, a bare identifier naming a class declared in
// the same file, so nothing imported can distinguish a rule that resolves receivers from one that
// accepts every identifier. Each row below was measured against the installed build.
//
// The nested-path row matters because upstream walks a member PATH rather than a single name, and
// the corpus never writes one. The function-declaration row is the other direction: upstream
// requires an ES6 class specifically, so a function component with `propTypes` attached is silent
// even though that is the most common way the property is written in real code.
func TestStaticPropertyPlacementAssignmentReceiver(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name, sourceText string
		messageIds       []string
	}{
		{
			"a class declaration resolves",
			"class MyComponent extends React.Component { render(){ return null; } }\nMyComponent.propTypes = {};",
			[]string{"notStaticClassProp"},
		},
		{
			"a variable holding a class expression resolves",
			"const MyComponent = class extends React.Component { render(){ return null; } };\nMyComponent.propTypes = {};",
			[]string{"notStaticClassProp"},
		},
		{
			"a nested member path resolves",
			"const ns = { MyComponent: class extends React.Component { render(){ return null; } } };\nns.MyComponent.propTypes = {};",
			[]string{"notStaticClassProp"},
		},
		{
			"a class extending something other than React does not resolve",
			"class MyComponent extends Other { render(){ return null; } }\nMyComponent.propTypes = {};",
			nil,
		},
		{
			"a function declaration is not an ES6 component",
			"function MyComponent(){ return null; }\nMyComponent.propTypes = {};",
			nil,
		},
		{
			"an assignment written inside the class body is skipped",
			"class MyComponent extends React.Component { render(){ MyComponent.propTypes = {}; return null; } }",
			nil,
		},
		{
			"an unresolved receiver is silent",
			"Unknown.propTypes = {};",
			nil,
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTypedWithOptions(t, StaticPropertyPlacement,
				staticPropertyPlacementFile, testCase.sourceText,
				staticPropertyPlacementDecode(t, `["static public field"]`))
			if len(testCase.messageIds) == 0 {
				rule_testing.ExpectClean(t, result)
				return
			}
			rule_testing.ExpectFindings(t, result, testCase.messageIds...)
		})
	}
}

// TestStaticPropertyPlacementMessageText asserts each rendered message exactly.
//
// The property name is interpolated into all three, so a message-id assertion cannot see anything
// the interpolation does. Asserted against literals typed here rather than against the rule's own
// builder, which would move both sides together under mutation.
func TestStaticPropertyPlacementMessageText(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name, sourceText, rawOptions, wantId, wantText string
	}{
		{
			"the field message names the property",
			"class MyComponent extends React.Component { propTypes = {}; render(){ return null; } }",
			`["static public field"]`,
			"notStaticClassProp",
			"'propTypes' should be declared as a static class property. This file's configuration " +
				"puts React statics in the class body as static fields, and a reader looking for " +
				"them will look there.",
		},
		{
			"the getter message names the property",
			"class MyComponent extends React.Component { defaultProps = {}; render(){ return null; } }",
			`["static getter"]`,
			"notGetterClassFunc",
			"'defaultProps' should be declared as a static getter class function. This file's " +
				"configuration puts React statics in the class body as getters, and a reader " +
				"looking for them will look there.",
		},
		{
			"the assignment message names the property",
			"class MyComponent extends React.Component { static displayName = \"x\"; render(){ return null; } }",
			`["property assignment"]`,
			"declareOutsideClass",
			"'displayName' should be declared outside the class body. This file's configuration " +
				"puts React statics after the class as assignments, and a reader looking for them " +
				"will look there.",
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTypedWithOptions(t, StaticPropertyPlacement,
				staticPropertyPlacementFile, testCase.sourceText,
				staticPropertyPlacementDecode(t, testCase.rawOptions))
			if len(result.Diagnostics) != 1 {
				t.Fatalf("wanted one finding, got %d", len(result.Diagnostics))
			}
			if result.Diagnostics[0].Message.Id != testCase.wantId {
				t.Fatalf("id is %q, wanted %q", result.Diagnostics[0].Message.Id, testCase.wantId)
			}
			if result.Diagnostics[0].Message.Description != testCase.wantText {
				t.Fatalf("message is %q", result.Diagnostics[0].Message.Description)
			}
			if len(result.Diagnostics[0].Fixes) != 0 || len(result.Diagnostics[0].Suggestions) != 0 {
				t.Fatal("upstream marks this rule not fixable and neither does this port")
			}
		})
	}
}

// TestStaticPropertyPlacementSpans asserts where each finding points.
//
// The in-class arms anchor on the whole member and the assignment arm anchors on the LEFT side of
// the assignment rather than on the statement, which is upstream reporting the MemberExpression it
// was visiting. A message-id assertion cannot see either.
func TestStaticPropertyPlacementSpans(t *testing.T) {
	t.Parallel()

	cases := []struct{ name, sourceText, rawOptions, reported string }{
		{
			"a class field spans the whole member",
			"class MyComponent extends React.Component { propTypes = {}; render(){ return null; } }",
			`["static public field"]`,
			"propTypes = {};",
		},
		{
			"a static getter spans the whole member",
			"class MyComponent extends React.Component { static get propTypes(){ return {}; } render(){ return null; } }",
			`["static public field"]`,
			"static get propTypes(){ return {}; }",
		},
		{
			"an assignment spans the left side only",
			"class MyComponent extends React.Component { render(){ return null; } }\nMyComponent.propTypes = {};",
			`["static public field"]`,
			"MyComponent.propTypes",
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTypedWithOptions(t, StaticPropertyPlacement,
				staticPropertyPlacementFile, testCase.sourceText,
				staticPropertyPlacementDecode(t, testCase.rawOptions))
			if len(result.Diagnostics) != 1 {
				t.Fatalf("wanted one finding, got %d", len(result.Diagnostics))
			}
			// The typed harness writes `strings.TrimSpace(contents)+"\n"`, so the expectation is
			// transformed the same way the input was rather than sliced from the Go literal.
			onDisk := strings.TrimSpace(testCase.sourceText) + "\n"
			reported := onDisk[result.Diagnostics[0].Range.Pos():result.Diagnostics[0].Range.End()]
			if reported != testCase.reported {
				t.Fatalf("finding points at %q, wanted %q", reported, testCase.reported)
			}
		})
	}
}

// TestDecodeStaticPropertyPlacementOptions covers the decoder's accepted list shapes.
//
// This is the port's own code with no upstream counterpart: upstream reads two POSITIONAL options
// and applies the default lazily per property, while this flattens both into one map at decode
// time. The nil row is the one that matters most, because a rule configured as a bare "error" is
// handed empty input and a decoder without a fallback would give every property an unmatched
// position and silently invert the rule.
func TestDecodeStaticPropertyPlacementOptions(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name, raw string
		want      map[string]StaticPropertyPlacementPosition
	}{
		{
			"absent configuration expects a static public field everywhere",
			"",
			map[string]StaticPropertyPlacementPosition{
				"propTypes": StaticPublicField, "displayName": StaticPublicField,
			},
		},
		{
			"a positional array with only a default",
			`["property assignment"]`,
			map[string]StaticPropertyPlacementPosition{
				"propTypes": PropertyAssignment, "displayName": PropertyAssignment,
			},
		},
		{
			"a per property override wins over the default",
			`["property assignment", {"displayName": "static getter"}]`,
			map[string]StaticPropertyPlacementPosition{
				"propTypes": PropertyAssignment, "displayName": StaticGetter,
			},
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			decoded, err := DecodeStaticPropertyPlacementOptions([]byte(testCase.raw))
			if err != nil {
				t.Fatalf("decode returned %v", err)
			}
			options, isOptions := decoded.(StaticPropertyPlacementOptions)
			if !isOptions {
				t.Fatalf("decode returned %T", decoded)
			}
			if len(options.Positions) != len(staticPropertyPlacementProperties) {
				t.Fatalf("decoded %d positions, wanted %d",
					len(options.Positions), len(staticPropertyPlacementProperties))
			}
			for property, want := range testCase.want {
				if options.Positions[property] != want {
					t.Fatalf("%s is %q, wanted %q", property, options.Positions[property], want)
				}
			}
		})
	}
}

// TestDecodeStaticPropertyPlacementOptionsRefusesWhatUpstreamRefuses is the other half.
//
// Each of these used to decode to the default, so a typo in a position or a property name produced a
// rule enforcing something the author did not write. Upstream's schema refuses every one.
func TestDecodeStaticPropertyPlacementOptionsRefusesWhatUpstreamRefuses(t *testing.T) {
	t.Parallel()

	for _, raw := range []string{
		`["nonsense"]`,
		`["static getter", {"notAThing": "property assignment"}]`,
		`["static getter", {"displayName": "nonsense"}]`,
		`["static getter", {"displayName": "property assignment"}, "static public field"]`,
		// A bare string, which the config layer never delivers to a list rule.
		`"static getter"`,
	} {
		if decoded, err := DecodeStaticPropertyPlacementOptions([]byte(raw)); err == nil {
			t.Errorf("%s decoded to %+v; it must be refused", raw, decoded)
		}
	}
}

// TestStaticPropertyPlacementNilOptionsUsesTheDefault bypasses the decoder entirely.
//
// Every other fixture here reaches the rule through the decoder, so nothing else can see the
// rule's own nil fallback. A rule handed nil options would otherwise read a nil map, find no
// position for any property, and go silent on everything.
func TestStaticPropertyPlacementNilOptionsUsesTheDefault(t *testing.T) {
	t.Parallel()

	result := rule_testing.RunTyped(t, StaticPropertyPlacement, staticPropertyPlacementFile,
		"class MyComponent extends React.Component { propTypes = {}; render(){ return null; } }")
	rule_testing.ExpectFindings(t, result, "notStaticClassProp")
}

// TestStaticPropertyPlacementRequiresTheTypedHarness asserts the checker is genuinely needed.
//
// `GetSymbolAtLocation` on a nil checker returns nil rather than crashing, so a rule that lost its
// guard would go quietly narrower rather than announcing itself. The assignment arm is the half
// that needs resolution; the in-class arms do not, which is why the untyped run is asserted to be
// silent on an assignment and the typed run to report it.
func TestStaticPropertyPlacementRequiresTheTypedHarness(t *testing.T) {
	t.Parallel()

	if !StaticPropertyPlacement.NeedsTypeChecker {
		t.Fatal("the assignment arm resolves its receiver through the checker and must declare it")
	}
	source := "class MyComponent extends React.Component { render(){ return null; } }\nMyComponent.propTypes = {};"
	untyped := rule_testing.RunWithOptions(t, StaticPropertyPlacement, staticPropertyPlacementFile,
		source, DefaultStaticPropertyPlacementOptions())
	rule_testing.ExpectClean(t, untyped)

	typed := rule_testing.RunTypedWithOptions(t, StaticPropertyPlacement,
		staticPropertyPlacementFile, source, DefaultStaticPropertyPlacementOptions())
	rule_testing.ExpectFindings(t, typed, "notStaticClassProp")
}

// TestStaticPropertyPlacementFlowAliases pins the two names that map onto a different reported name.
//
// Upstream's `propTypes` predicate accepts a class field named `props` when it carries a type
// annotation, and its `contextTypes` predicate accepts one named `context` the same way. Both
// report under the MAPPED name rather than the written one, measured against the installed build:
// `props: {a: number}` produces a finding naming `propTypes`.
//
// The annotation is required. A field named `props` with no annotation is silent, which is what
// separates the alias from an ordinary field that happens to share the name. Upstream's corpus
// writes neither shape, so a mutation disabling the alias survived until these rows existed.
func TestStaticPropertyPlacementFlowAliases(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name, sourceText, wantId, wantName string
	}{
		{
			"an annotated props field reports as propTypes",
			"class MyComponent extends React.Component { props: { a: number }; render(){ return null; } }",
			"declareOutsideClass", "propTypes",
		},
		{
			"an annotated context field reports as contextTypes",
			"class MyComponent extends React.Component { context: { a: number }; render(){ return null; } }",
			"declareOutsideClass", "contextTypes",
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTypedWithOptions(t, StaticPropertyPlacement,
				staticPropertyPlacementFile, testCase.sourceText,
				staticPropertyPlacementDecode(t, `["property assignment"]`))
			rule_testing.ExpectFindings(t, result, testCase.wantId)
			if !strings.Contains(result.Diagnostics[0].Message.Description, "'"+testCase.wantName+"'") {
				t.Fatalf("message %q does not name %q",
					result.Diagnostics[0].Message.Description, testCase.wantName)
			}
		})
	}

	// An unannotated field named `props` is not the alias, so it is not one of the six names and is
	// silent. Measured.
	plain := rule_testing.RunTypedWithOptions(t, StaticPropertyPlacement,
		staticPropertyPlacementFile,
		"class MyComponent extends React.Component { props = {}; render(){ return null; } }",
		staticPropertyPlacementDecode(t, `["property assignment"]`))
	rule_testing.ExpectClean(t, plain)
}

// TestStaticPropertyPlacementMergedReceiverDeclaration pins the declaration loop.
//
// Declaration merging puts an `interface MyComponent` written above the class at declaration index
// zero, so a rule reading `symbol.Declarations[0]` instead of looping goes silent on it. Measured:
// the installed build reports, so the loop is what reproduces upstream and the index would be a
// silent false negative. This is the index-zero trap this tree names for six other rules.
func TestStaticPropertyPlacementMergedReceiverDeclaration(t *testing.T) {
	t.Parallel()

	result := rule_testing.RunTypedWithOptions(t, StaticPropertyPlacement,
		staticPropertyPlacementFile,
		"interface MyComponent { a: number }\nclass MyComponent extends React.Component { render(){ return null; } }\nMyComponent.propTypes = {};",
		staticPropertyPlacementDecode(t, `["static public field"]`))
	rule_testing.ExpectFindings(t, result, "notStaticClassProp")
}
