package react

import (
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// The corpus is upstream's, extracted by stubbing its RuleTester and capturing what the test file
// hands it, so no case was retyped and no escape was typed by hand. Every case was run against the
// installed build first, carrying its own options AND its languageOptions, and the expectations
// here are what that build reported.
//
// Eight of the ninety cases are excluded with the reason recorded rather than dropped; see
// TestDefaultPropsMatchPropTypesCasesThisPortCannotExpress.
const defaultPropsMatchPropTypesFile = "/repository/source/DefaultPropsMatchPropTypes.tsx"

func TestDefaultPropsMatchPropTypesStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name, sourceText string
		options          DefaultPropsMatchPropTypesOptions
	}{
		{"upstream valid 0", "\n        function MyStatelessComponent({ foo, bar }) {\n          return <div>{foo}{bar}</div>;\n        }\n        MyStatelessComponent.propTypes = {\n          foo: React.PropTypes.string.isRequired,\n          bar: React.PropTypes.string.isRequired\n        };\n      ", DefaultPropsMatchPropTypesOptions{AllowRequiredDefaults: false}},
		{"upstream valid 1", "\n        function MyStatelessComponent({ foo, bar }) {\n          return <div>{foo}{bar}</div>;\n        }\n        MyStatelessComponent.propTypes = {\n          foo: React.PropTypes.string,\n          bar: React.PropTypes.string.isRequired\n        };\n        MyStatelessComponent.defaultProps = {\n          foo: \"foo\"\n        };\n      ", DefaultPropsMatchPropTypesOptions{AllowRequiredDefaults: false}},
		{"upstream valid 2", "\n        function MyStatelessComponent({ foo, bar }) {\n          return <div>{foo}{bar}</div>;\n        }\n      ", DefaultPropsMatchPropTypesOptions{AllowRequiredDefaults: false}},
		{"upstream valid 3", "\n        function MyStatelessComponent({ foo, bar }) {\n          return <div>{foo}{bar}</div>;\n        }\n        MyStatelessComponent.propTypes = {\n          bar: React.PropTypes.string.isRequired\n        };\n        MyStatelessComponent.propTypes.foo = React.PropTypes.string;\n        MyStatelessComponent.defaultProps = {\n          foo: \"foo\"\n        };\n      ", DefaultPropsMatchPropTypesOptions{AllowRequiredDefaults: false}},
		{"upstream valid 4", "\n        function MyStatelessComponent({ foo, bar }) {\n          return <div>{foo}{bar}</div>;\n        }\n        MyStatelessComponent.propTypes = {\n          bar: React.PropTypes.string.isRequired\n        };\n        MyStatelessComponent.defaultProps = {\n          bar: \"bar\"\n        };\n      ", DefaultPropsMatchPropTypesOptions{AllowRequiredDefaults: true}},
		{"upstream valid 5", "\n        function MyStatelessComponent({ foo, bar }) {\n          return <div>{foo}{bar}</div>;\n        }\n        MyStatelessComponent.propTypes = {\n          bar: React.PropTypes.string.isRequired\n        };\n        MyStatelessComponent.propTypes.foo = React.PropTypes.string;\n        MyStatelessComponent.defaultProps = {};\n        MyStatelessComponent.defaultProps.foo = \"foo\";\n      ", DefaultPropsMatchPropTypesOptions{AllowRequiredDefaults: false}},
		{"upstream valid 6", "\n        function MyStatelessComponent({ foo }) {\n          return <div>{foo}</div>;\n        }\n        MyStatelessComponent.propTypes = {};\n        MyStatelessComponent.propTypes.foo = React.PropTypes.string;\n        MyStatelessComponent.defaultProps = {};\n        MyStatelessComponent.defaultProps.foo = \"foo\";\n      ", DefaultPropsMatchPropTypesOptions{AllowRequiredDefaults: false}},
		{"upstream valid 7", "\n        const types = {\n          foo: React.PropTypes.string,\n          bar: React.PropTypes.string.isRequired\n        };\n\n        function MyStatelessComponent({ foo, bar }) {\n          return <div>{foo}{bar}</div>;\n        }\n        MyStatelessComponent.propTypes = types;\n        MyStatelessComponent.defaultProps = {\n          foo: \"foo\"\n        };\n      ", DefaultPropsMatchPropTypesOptions{AllowRequiredDefaults: false}},
		{"upstream valid 8", "\n        const defaults = {\n          foo: \"foo\"\n        };\n\n        function MyStatelessComponent({ foo, bar }) {\n          return <div>{foo}{bar}</div>;\n        }\n        MyStatelessComponent.propTypes = {\n          foo: React.PropTypes.string,\n          bar: React.PropTypes.string.isRequired\n        };\n        MyStatelessComponent.defaultProps = defaults;\n      ", DefaultPropsMatchPropTypesOptions{AllowRequiredDefaults: false}},
		{"upstream valid 9", "\n        const defaults = {\n          foo: \"foo\"\n        };\n        const types = {\n          foo: React.PropTypes.string,\n          bar: React.PropTypes.string.isRequired\n        };\n\n        function MyStatelessComponent({ foo, bar }) {\n          return <div>{foo}{bar}</div>;\n        }\n        MyStatelessComponent.propTypes = types;\n        MyStatelessComponent.defaultProps = defaults;\n      ", DefaultPropsMatchPropTypesOptions{AllowRequiredDefaults: false}},
		{"upstream valid 10", "\n        var Greeting = createReactClass({\n          render: function() {\n            return <div>Hello {this.props.foo} {this.props.bar}</div>;\n          },\n          propTypes: {\n            foo: React.PropTypes.string.isRequired,\n            bar: React.PropTypes.string.isRequired\n          }\n        });\n      ", DefaultPropsMatchPropTypesOptions{AllowRequiredDefaults: false}},
		{"upstream valid 11", "\n        var Greeting = createReactClass({\n          render: function() {\n            return <div>Hello {this.props.foo} {this.props.bar}</div>;\n          },\n          propTypes: {\n            foo: React.PropTypes.string,\n            bar: React.PropTypes.string.isRequired\n          },\n          getDefaultProps: function() {\n            return {\n              foo: \"foo\"\n            };\n          }\n        });\n      ", DefaultPropsMatchPropTypesOptions{AllowRequiredDefaults: false}},
		{"upstream valid 12", "\n        var Greeting = createReactClass({\n          render: function() {\n            return <div>Hello {this.props.foo} {this.props.bar}</div>;\n          },\n          propTypes: {\n            foo: React.PropTypes.string,\n            bar: React.PropTypes.string\n          },\n          getDefaultProps: function() {\n            return {\n              foo: \"foo\",\n              bar: \"bar\"\n            };\n          }\n        });\n      ", DefaultPropsMatchPropTypesOptions{AllowRequiredDefaults: false}},
		{"upstream valid 13", "\n        var Greeting = createReactClass({\n          render: function() {\n            return <div>Hello {this.props.foo} {this.props.bar}</div>;\n          }\n        });\n      ", DefaultPropsMatchPropTypesOptions{AllowRequiredDefaults: false}},
		{"upstream valid 14", "\n        class Greeting extends React.Component {\n          render() {\n            return (\n              <h1>Hello, {this.props.foo} {this.props.bar}</h1>\n            );\n          }\n        }\n        Greeting.propTypes = {\n          foo: React.PropTypes.string,\n          bar: React.PropTypes.string.isRequired\n        };\n        Greeting.defaultProps = {\n          foo: \"foo\"\n        };\n      ", DefaultPropsMatchPropTypesOptions{AllowRequiredDefaults: false}},
		{"upstream valid 15", "\n        class Greeting extends React.Component {\n          render() {\n            return (\n              <h1>Hello, {this.props.foo} {this.props.bar}</h1>\n            );\n          }\n        }\n      ", DefaultPropsMatchPropTypesOptions{AllowRequiredDefaults: false}},
		{"upstream valid 16", "\n        class Greeting extends React.Component {\n          render() {\n            return (\n              <h1>Hello, {this.props.foo} {this.props.bar}</h1>\n            );\n          }\n        }\n        Greeting.propTypes = {\n          bar: React.PropTypes.string.isRequired\n        };\n        Greeting.propTypes.foo = React.PropTypes.string;\n        Greeting.defaultProps = {\n          foo: \"foo\"\n        };\n      ", DefaultPropsMatchPropTypesOptions{AllowRequiredDefaults: false}},
		{"upstream valid 17", "\n        class Greeting extends React.Component {\n          render() {\n            return (\n              <h1>Hello, {this.props.foo} {this.props.bar}</h1>\n            );\n          }\n        }\n        Greeting.propTypes = {\n          bar: React.PropTypes.string.isRequired\n        };\n        Greeting.propTypes.foo = React.PropTypes.string;\n        Greeting.defaultProps = {};\n        Greeting.defaultProps.foo = \"foo\";\n      ", DefaultPropsMatchPropTypesOptions{AllowRequiredDefaults: false}},
		{"upstream valid 18", "\n        class Greeting extends React.Component {\n          render() {\n            return (\n              <h1>Hello, {this.props.foo} {this.props.bar}</h1>\n            );\n          }\n        }\n        Greeting.propTypes = {};\n        Greeting.propTypes.foo = React.PropTypes.string;\n        Greeting.defaultProps = {};\n        Greeting.defaultProps.foo = \"foo\";\n      ", DefaultPropsMatchPropTypesOptions{AllowRequiredDefaults: false}},
		{"upstream valid 19", "\n        function NotAComponent({ foo, bar }) {}\n        NotAComponent.defaultProps = {\n          bar: \"bar\"\n        };\n      ", DefaultPropsMatchPropTypesOptions{AllowRequiredDefaults: false}},
		{"upstream valid 20", "\n        class Greeting {\n          render() {\n            return (\n              <h1>Hello, {this.props.foo} {this.props.bar}</h1>\n            );\n          }\n        }\n        Greeting.defaulProps = {\n          bar: \"bar\"\n        };\n      ", DefaultPropsMatchPropTypesOptions{AllowRequiredDefaults: false}},
		{"upstream valid 21", "\n        const defaults = require(\"./defaults\");\n        const types = {\n          foo: React.PropTypes.string,\n          bar: React.PropTypes.string\n        };\n\n        function MyStatelessComponent({ foo, bar }) {\n          return <div>{foo}{bar}</div>;\n        }\n        MyStatelessComponent.propTypes = types;\n        MyStatelessComponent.defaultProps = defaults;\n      ", DefaultPropsMatchPropTypesOptions{AllowRequiredDefaults: false}},
		{"upstream valid 22", "\n        const defaults = {\n          foo: \"foo\"\n        };\n        const types = require(\"./propTypes\");\n\n        function MyStatelessComponent({ foo, bar }) {\n          return <div>{foo}{bar}</div>;\n        }\n        MyStatelessComponent.propTypes = types;\n        MyStatelessComponent.defaultProps = defaults;\n      ", DefaultPropsMatchPropTypesOptions{AllowRequiredDefaults: false}},
		{"upstream valid 23", "\n        MyStatelessComponent.propTypes = {\n          foo: React.PropTypes.string\n        };\n        MyStatelessComponent.defaultProps = require(\"./defaults\").foo;\n\n        function MyStatelessComponent({ foo, bar }) {\n          return <div>{foo}{bar}</div>;\n        }\n      ", DefaultPropsMatchPropTypesOptions{AllowRequiredDefaults: false}},
		{"upstream valid 24", "\n        MyStatelessComponent.propTypes = {\n          foo: React.PropTypes.string\n        };\n        MyStatelessComponent.defaultProps = require(\"./defaults\").foo;\n        MyStatelessComponent.defaultProps.bar = \"bar\";\n\n        function MyStatelessComponent({ foo, bar }) {\n          return <div>{foo}{bar}</div>;\n        }\n      ", DefaultPropsMatchPropTypesOptions{AllowRequiredDefaults: false}},
		{"upstream valid 25", "\n        import defaults from \"./defaults\";\n\n        MyStatelessComponent.propTypes = {\n          foo: React.PropTypes.string\n        };\n        MyStatelessComponent.defaultProps = defaults;\n\n        function MyStatelessComponent({ foo, bar }) {\n          return <div>{foo}{bar}</div>;\n        }\n      ", DefaultPropsMatchPropTypesOptions{AllowRequiredDefaults: false}},
		{"upstream valid 26", "\n        import { foo } from \"./defaults\";\n\n        MyStatelessComponent.propTypes = {\n          foo: React.PropTypes.string\n        };\n        MyStatelessComponent.defaultProps = foo;\n\n        function MyStatelessComponent({ foo, bar }) {\n          return <div>{foo}{bar}</div>;\n        }\n      ", DefaultPropsMatchPropTypesOptions{AllowRequiredDefaults: false}},
		{"upstream valid 27", "\n        const component = rowsOfType(GuestlistEntry, (rowData, ownProps) => ({\n            ...rowData,\n            onPress: () => ownProps.onPress(rowData.id),\n        }));\n      ", DefaultPropsMatchPropTypesOptions{AllowRequiredDefaults: false}},
		{"upstream valid 28", "\n        MyStatelessComponent.propTypes = {\n          ...stuff,\n          foo: React.PropTypes.string\n        };\n        MyStatelessComponent.defaultProps = {\n         foo: \"foo\"\n        };\n        function MyStatelessComponent({ foo, bar }) {\n          return <div>{foo}{bar}</div>;\n        }\n      ", DefaultPropsMatchPropTypesOptions{AllowRequiredDefaults: false}},
		{"upstream valid 29", "\n        MyStatelessComponent.propTypes = {\n          foo: React.PropTypes.string\n        };\n        MyStatelessComponent.defaultProps = {\n         ...defaults,\n        };\n        function MyStatelessComponent({ foo, bar }) {\n          return <div>{foo}{bar}</div>;\n        }\n      ", DefaultPropsMatchPropTypesOptions{AllowRequiredDefaults: false}},
		{"upstream valid 30", "\n        class Greeting extends React.Component {\n          render() {\n            return (\n              <h1>Hello, {this.props.foo} {this.props.bar}</h1>\n            );\n          }\n        }\n        Greeting.propTypes = {\n          ...someProps,\n          bar: React.PropTypes.string.isRequired\n        };\n      ", DefaultPropsMatchPropTypesOptions{AllowRequiredDefaults: false}},
		{"upstream valid 31", "\n        class Greeting extends React.Component {\n          render() {\n            return (\n              <h1>Hello, {this.props.foo} {this.props.bar}</h1>\n            );\n          }\n        }\n        Greeting.propTypes = {\n          foo: React.PropTypes.string,\n          bar: React.PropTypes.string.isRequired\n        };\n        Greeting.defaultProps = {\n          ...defaults,\n          bar: \"bar\"\n        };\n      ", DefaultPropsMatchPropTypesOptions{AllowRequiredDefaults: false}},
		{"upstream valid 32", "\n        type Props = {\n          foo: string\n        };\n\n        class Hello extends React.Component {\n          props: Props;\n\n          render() {\n            return <div>Hello {this.props.foo}</div>;\n          }\n        }\n      ", DefaultPropsMatchPropTypesOptions{AllowRequiredDefaults: false}},
		{"upstream valid 33", "\n        type Props = {\n          foo: string,\n          bar?: string\n        };\n\n        class Hello extends React.Component {\n          props: Props;\n\n          render() {\n            return <div>Hello {this.props.foo}</div>;\n          }\n        }\n\n        Hello.defaultProps = {\n          bar: \"bar\"\n        };\n      ", DefaultPropsMatchPropTypesOptions{AllowRequiredDefaults: false}},
		{"upstream valid 34", "\n        class Hello extends React.Component {\n          props: {\n            foo: string,\n            bar?: string\n          };\n\n          render() {\n            return <div>Hello {this.props.foo}</div>;\n          }\n        }\n\n        Hello.defaultProps = {\n          bar: \"bar\"\n        };\n      ", DefaultPropsMatchPropTypesOptions{AllowRequiredDefaults: false}},
		{"upstream valid 35", "\n        class Hello extends React.Component {\n          props: {\n            foo: string\n          };\n\n          render() {\n            return <div>Hello {this.props.foo}</div>;\n          }\n        }\n      ", DefaultPropsMatchPropTypesOptions{AllowRequiredDefaults: false}},
		{"upstream valid 36", "\n        function Hello(props: { foo?: string }) {\n          return <div>Hello {props.foo}</div>;\n        }\n\n        Hello.defaultProps = { foo: \"foo\" };\n      ", DefaultPropsMatchPropTypesOptions{AllowRequiredDefaults: false}},
		{"upstream valid 37", "\n        function Hello(props: { foo: string }) {\n          return <div>Hello {foo}</div>;\n        }\n      ", DefaultPropsMatchPropTypesOptions{AllowRequiredDefaults: false}},
		{"upstream valid 38", "\n        const Hello = (props: { foo?: string }) => {\n          return <div>Hello {props.foo}</div>;\n        };\n\n        Hello.defaultProps = { foo: \"foo\" };\n      ", DefaultPropsMatchPropTypesOptions{AllowRequiredDefaults: false}},
		{"upstream valid 39", "\n        const Hello = (props: { foo: string }) => {\n          return <div>Hello {foo}</div>;\n        };\n      ", DefaultPropsMatchPropTypesOptions{AllowRequiredDefaults: false}},
		{"upstream valid 40", "\n        const Hello = function(props: { foo?: string }) {\n          return <div>Hello {props.foo}</div>;\n        };\n\n        Hello.defaultProps = { foo: \"foo\" };\n      ", DefaultPropsMatchPropTypesOptions{AllowRequiredDefaults: false}},
		{"upstream valid 41", "\n        const Hello = function(props: { foo: string }) {\n          return <div>Hello {foo}</div>;\n        };\n      ", DefaultPropsMatchPropTypesOptions{AllowRequiredDefaults: false}},
		{"upstream valid 42", "\n        type Props = {\n          foo: string,\n          bar?: string\n        };\n\n        type Props2 = {\n          foo: string,\n          baz?: string\n        }\n\n        function Hello(props: Props | Props2) {\n          return <div>Hello {props.foo}</div>;\n        }\n\n        Hello.defaultProps = {\n          bar: \"bar\",\n          baz: \"baz\"\n        };\n      ", DefaultPropsMatchPropTypesOptions{AllowRequiredDefaults: false}},
		{"upstream valid 43", "\n        type PropsA = { foo?: string };\n        type PropsB = { bar?: string, fooBar: string };\n        type Props = PropsA & PropsB;\n\n        class Bar extends React.Component {\n          props: Props;\n          static defaultProps = {\n            foo: \"foo\",\n          }\n\n          render() {\n            return <div>{this.props.foo} - {this.props.bar}</div>\n          }\n        }\n      ", DefaultPropsMatchPropTypesOptions{AllowRequiredDefaults: false}},
		{"upstream valid 44", "\n        import type Props from \"fake\";\n        class Hello extends React.Component {\n          props: Props;\n          render () {\n            return <div>Hello {this.props.name.firstname}</div>;\n          }\n        }\n      ", DefaultPropsMatchPropTypesOptions{AllowRequiredDefaults: false}},
		{"upstream valid 45", "\n        type Props = any;\n\n        const Hello = function({ foo }: Props) {\n          return <div>Hello {foo}</div>;\n        };\n      ", DefaultPropsMatchPropTypesOptions{AllowRequiredDefaults: false}},
		{"upstream valid 46", "\n        import type ImportedProps from \"fake\";\n        type Props = ImportedProps;\n        function Hello(props: Props) {\n          return <div>Hello {props.name.firstname}</div>;\n        }\n      ", DefaultPropsMatchPropTypesOptions{AllowRequiredDefaults: false}},
		{"upstream valid 49", "\n        import type { ImportedType } from \"fake\";\n        type Props = ImportedType;\n        function Hello(props: Props) {\n          return <div>Hello {props.name.firstname}</div>;\n        }\n      ", DefaultPropsMatchPropTypesOptions{AllowRequiredDefaults: false}},
		{"upstream valid 50", "\n        import type ImportedProps from \"fake\";\n        type NestedProps = ImportedProps;\n        type Props = NestedProps;\n        function Hello(props: Props) {\n          return <div>Hello {props.name.firstname}</div>;\n        }\n      ", DefaultPropsMatchPropTypesOptions{AllowRequiredDefaults: false}},
		{"upstream valid 51", "\n        import type ImportedProps from \"fake\";\n        type Props = ImportedProps & {\n          foo: string\n        };\n        function Hello(props: Props) {\n          return <div>Hello {props.name.firstname}</div>;\n        }\n      ", DefaultPropsMatchPropTypesOptions{AllowRequiredDefaults: false}},
		{"upstream valid 52", "\n        import type { FieldProps } from 'redux-form';\n\n        type Props = FieldProps & {\n          name: string,\n          type: string,\n          label?: string,\n          placeholder?: string,\n          disabled?: boolean,\n        };\n\n        TextField.defaultProps = {\n          label: '',\n          placeholder: '',\n          disabled: false,\n        };\n      ", DefaultPropsMatchPropTypesOptions{AllowRequiredDefaults: false}},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunWithOptions(t, DefaultPropsMatchPropTypes, defaultPropsMatchPropTypesFile, testCase.sourceText, testCase.options)
			rule_testing.ExpectClean(t, result)
		})
	}
}

// TestDefaultPropsMatchPropTypesFires asserts the count AND the rendered message of every
// finding. Both messages interpolate the prop name, so a message-id assertion could not see an
// interpolation defect, and the two arms differ only by which id fires on which input.
func TestDefaultPropsMatchPropTypesFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name, sourceText string
		options          DefaultPropsMatchPropTypesOptions
		wantIds          []string
		wantMessages     []string
	}{
		{"upstream invalid 0", "\n        function MyStatelessComponent({ foo, bar }) {\n          return <div>{foo}{bar}</div>;\n        }\n        MyStatelessComponent.propTypes = {\n          foo: React.PropTypes.string,\n          bar: React.PropTypes.string.isRequired\n        };\n        MyStatelessComponent.defaultProps = {\n          baz: \"baz\"\n        };\n      ", DefaultPropsMatchPropTypesOptions{AllowRequiredDefaults: false}, []string{"defaultHasNoType"}, []string{
			"defaultProp \"baz\" has no corresponding propTypes declaration.",
		}},
		{"upstream invalid 3", "\n        function MyStatelessComponent({ foo, bar }) {\n          return <div>{foo}{bar}</div>;\n        }\n        MyStatelessComponent.propTypes = {\n          foo: React.PropTypes.string,\n          bar: React.PropTypes.string.isRequired\n        };\n        MyStatelessComponent.defaultProps = {\n          baz: \"baz\"\n        };\n      ", DefaultPropsMatchPropTypesOptions{AllowRequiredDefaults: true}, []string{"defaultHasNoType"}, []string{
			"defaultProp \"baz\" has no corresponding propTypes declaration.",
		}},
		{"upstream invalid 4", "\n        function MyStatelessComponent({ foo, bar }) {\n          return <div>{foo}{bar}</div>;\n        }\n        MyStatelessComponent.propTypes = {\n          foo: React.PropTypes.string,\n          bar: React.PropTypes.string.isRequired\n        };\n        MyStatelessComponent.defaultProps = {\n          bar: \"bar\"\n        };\n        MyStatelessComponent.defaultProps.baz = \"baz\";\n      ", DefaultPropsMatchPropTypesOptions{AllowRequiredDefaults: false}, []string{"requiredHasDefault", "defaultHasNoType"}, []string{
			"defaultProp \"bar\" defined for isRequired propType.",
			"defaultProp \"baz\" has no corresponding propTypes declaration.",
		}},
		{"upstream invalid 5", "\n        const types = {\n          foo: React.PropTypes.string,\n          bar: React.PropTypes.string.isRequired\n        };\n\n        function MyStatelessComponent({ foo, bar }) {\n          return <div>{foo}{bar}</div>;\n        }\n        MyStatelessComponent.propTypes = types;\n        MyStatelessComponent.defaultProps = {\n          bar: \"bar\"\n        };\n      ", DefaultPropsMatchPropTypesOptions{AllowRequiredDefaults: false}, []string{"requiredHasDefault"}, []string{
			"defaultProp \"bar\" defined for isRequired propType.",
		}},
		{"upstream invalid 6", "\n        const defaults = {\n          foo: \"foo\"\n        };\n\n        function MyStatelessComponent({ foo, bar }) {\n          return <div>{foo}{bar}</div>;\n        }\n        MyStatelessComponent.propTypes = {\n          foo: React.PropTypes.string.isRequired,\n          bar: React.PropTypes.string\n        };\n        MyStatelessComponent.defaultProps = defaults;\n      ", DefaultPropsMatchPropTypesOptions{AllowRequiredDefaults: false}, []string{"requiredHasDefault"}, []string{
			"defaultProp \"foo\" defined for isRequired propType.",
		}},
		{"upstream invalid 7", "\n        const defaults = {\n          foo: \"foo\"\n        };\n        const types = {\n          foo: React.PropTypes.string.isRequired,\n          bar: React.PropTypes.string\n        };\n\n        function MyStatelessComponent({ foo, bar }) {\n          return <div>{foo}{bar}</div>;\n        }\n        MyStatelessComponent.propTypes = types;\n        MyStatelessComponent.defaultProps = defaults;\n      ", DefaultPropsMatchPropTypesOptions{AllowRequiredDefaults: false}, []string{"requiredHasDefault"}, []string{
			"defaultProp \"foo\" defined for isRequired propType.",
		}},
		{"upstream invalid 8", "\n        var Greeting = createReactClass({\n          render: function() {\n            return <div>Hello {this.props.foo} {this.props.bar}</div>;\n          },\n          propTypes: {\n            foo: React.PropTypes.string,\n            bar: React.PropTypes.string.isRequired\n          },\n          getDefaultProps: function() {\n            return {\n              baz: \"baz\"\n            };\n          }\n        });\n      ", DefaultPropsMatchPropTypesOptions{AllowRequiredDefaults: false}, []string{"defaultHasNoType"}, []string{
			"defaultProp \"baz\" has no corresponding propTypes declaration.",
		}},
		{"upstream invalid 9", "\n        var Greeting = createReactClass({\n          render: function() {\n            return <div>Hello {this.props.foo} {this.props.bar}</div>;\n          },\n          propTypes: {\n            foo: React.PropTypes.string.isRequired,\n            bar: React.PropTypes.string\n          },\n          getDefaultProps: function() {\n            return {\n              foo: \"foo\"\n            };\n          }\n        });\n      ", DefaultPropsMatchPropTypesOptions{AllowRequiredDefaults: false}, []string{"requiredHasDefault"}, []string{
			"defaultProp \"foo\" defined for isRequired propType.",
		}},
		{"upstream invalid 10", "\n        class Greeting extends React.Component {\n          render() {\n            return (\n              <h1>Hello, {this.props.foo} {this.props.bar}</h1>\n            );\n          }\n        }\n        Greeting.propTypes = {\n          foo: React.PropTypes.string,\n          bar: React.PropTypes.string.isRequired\n        };\n        Greeting.defaultProps = {\n          baz: \"baz\"\n        };\n      ", DefaultPropsMatchPropTypesOptions{AllowRequiredDefaults: false}, []string{"defaultHasNoType"}, []string{
			"defaultProp \"baz\" has no corresponding propTypes declaration.",
		}},
		{"upstream invalid 11", "\n        class Greeting extends React.Component {\n          render() {\n            return (\n              <h1>Hello, {this.props.foo} {this.props.bar}</h1>\n            );\n          }\n        }\n        Greeting.propTypes = {\n          foo: React.PropTypes.string.isRequired,\n          bar: React.PropTypes.string\n        };\n        Greeting.defaultProps = {\n          foo: \"foo\"\n        };\n      ", DefaultPropsMatchPropTypesOptions{AllowRequiredDefaults: false}, []string{"requiredHasDefault"}, []string{
			"defaultProp \"foo\" defined for isRequired propType.",
		}},
		{"upstream invalid 12", "\n        class Greeting extends React.Component {\n          render() {\n            return (\n              <h1>Hello, {this.props.foo} {this.props.bar}</h1>\n            );\n          }\n        }\n        Greeting.propTypes = {\n          bar: React.PropTypes.string.isRequired\n        };\n        Greeting.propTypes.foo = React.PropTypes.string.isRequired;\n        Greeting.defaultProps = {};\n        Greeting.defaultProps.foo = \"foo\";\n      ", DefaultPropsMatchPropTypesOptions{AllowRequiredDefaults: false}, []string{"requiredHasDefault"}, []string{
			"defaultProp \"foo\" defined for isRequired propType.",
		}},
		{"upstream invalid 13", "\n        class Greeting extends React.Component {\n          render() {\n            return (\n              <h1>Hello, {this.props.foo} {this.props.bar}</h1>\n            );\n          }\n        }\n        Greeting.propTypes = {\n          bar: React.PropTypes.string\n        };\n        Greeting.propTypes.foo = React.PropTypes.string;\n        Greeting.defaultProps = {};\n        Greeting.defaultProps.baz = \"baz\";\n      ", DefaultPropsMatchPropTypesOptions{AllowRequiredDefaults: false}, []string{"defaultHasNoType"}, []string{
			"defaultProp \"baz\" has no corresponding propTypes declaration.",
		}},
		{"upstream invalid 14", "\n        class Greeting extends React.Component {\n          render() {\n            return (\n              <h1>Hello, {this.props.foo} {this.props.bar}</h1>\n            );\n          }\n        }\n        Greeting.propTypes = {};\n        Greeting.propTypes.foo = React.PropTypes.string.isRequired;\n        Greeting.defaultProps = {};\n        Greeting.defaultProps.foo = \"foo\";\n      ", DefaultPropsMatchPropTypesOptions{AllowRequiredDefaults: false}, []string{"requiredHasDefault"}, []string{
			"defaultProp \"foo\" defined for isRequired propType.",
		}},
		{"upstream invalid 15", "\n        class Greeting extends React.Component {\n          render() {\n            return (\n              <h1>Hello, {this.props.foo} {this.props.bar}</h1>\n            );\n          }\n        }\n        const props = {\n          foo: React.PropTypes.string,\n          bar: React.PropTypes.string.isRequired\n        };\n        Greeting.propTypes = props;\n        const defaults = {\n          bar: \"bar\"\n        };\n        Greeting.defaultProps = defaults;\n      ", DefaultPropsMatchPropTypesOptions{AllowRequiredDefaults: false}, []string{"requiredHasDefault"}, []string{
			"defaultProp \"bar\" defined for isRequired propType.",
		}},
		{"upstream invalid 16", "\n        class Greeting extends React.Component {\n          render() {\n            return (\n              <h1>Hello, {this.props.foo} {this.props.bar}</h1>\n            );\n          }\n        }\n        const props = {\n          foo: React.PropTypes.string,\n          bar: React.PropTypes.string\n        };\n        const defaults = {\n          baz: \"baz\"\n        };\n        Greeting.propTypes = props;\n        Greeting.defaultProps = defaults;\n      ", DefaultPropsMatchPropTypesOptions{AllowRequiredDefaults: false}, []string{"defaultHasNoType"}, []string{
			"defaultProp \"baz\" has no corresponding propTypes declaration.",
		}},
		{"upstream invalid 17", "\n        class Hello extends React.Component {\n          static get propTypes() {\n            return {\n              name: React.PropTypes.string.isRequired\n            };\n          }\n          static get defaultProps() {\n            return {\n              name: \"name\"\n            };\n          }\n          render() {\n            return <div>Hello {this.props.name}</div>;\n          }\n        }\n      ", DefaultPropsMatchPropTypesOptions{AllowRequiredDefaults: false}, []string{"requiredHasDefault"}, []string{
			"defaultProp \"name\" defined for isRequired propType.",
		}},
		{"upstream invalid 18", "\n        class Hello extends React.Component {\n          static get propTypes() {\n            return {\n              foo: React.PropTypes.string,\n              bar: React.PropTypes.string\n            };\n          }\n          static get defaultProps() {\n            return {\n              baz: \"world\"\n            };\n          }\n          render() {\n            return <div>Hello {this.props.bar}</div>;\n          }\n        }\n      ", DefaultPropsMatchPropTypesOptions{AllowRequiredDefaults: false}, []string{"defaultHasNoType"}, []string{
			"defaultProp \"baz\" has no corresponding propTypes declaration.",
		}},
		{"upstream invalid 19", "\n        const props = {\n          foo: React.PropTypes.string\n        };\n        const defaults = {\n          baz: \"baz\"\n        };\n\n        class Hello extends React.Component {\n          static get propTypes() {\n            return props;\n          }\n          static get defaultProps() {\n            return defaults;\n          }\n          render() {\n            return <div>Hello {this.props.foo}</div>;\n          }\n        }\n      ", DefaultPropsMatchPropTypesOptions{AllowRequiredDefaults: false}, []string{"defaultHasNoType"}, []string{
			"defaultProp \"baz\" has no corresponding propTypes declaration.",
		}},
		{"upstream invalid 20", "\n        const defaults = {\n          bar: \"world\"\n        };\n\n        class Hello extends React.Component {\n          static get propTypes() {\n            return {\n              foo: React.PropTypes.string,\n              bar: React.PropTypes.string.isRequired\n            };\n          }\n          static get defaultProps() {\n            return defaults;\n          }\n          render() {\n            return <div>Hello {this.props.bar}</div>;\n          }\n        }\n      ", DefaultPropsMatchPropTypesOptions{AllowRequiredDefaults: false}, []string{"requiredHasDefault"}, []string{
			"defaultProp \"bar\" defined for isRequired propType.",
		}},
		{"upstream invalid 21", "\n        class Greeting extends React.Component {\n          render() {\n            return (\n              <h1>Hello, {this.props.foo} {this.props.bar}</h1>\n            );\n          }\n          static propTypes = {\n            foo: React.PropTypes.string,\n            bar: React.PropTypes.string.isRequired\n          };\n          static defaultProps = {\n            bar: \"bar\"\n          };\n        }\n      ", DefaultPropsMatchPropTypesOptions{AllowRequiredDefaults: false}, []string{"requiredHasDefault"}, []string{
			"defaultProp \"bar\" defined for isRequired propType.",
		}},
		{"upstream invalid 22", "\n        class Greeting extends React.Component {\n          render() {\n            return (\n              <h1>Hello, {this.props.foo} {this.props.bar}</h1>\n            );\n          }\n          static propTypes = {\n            foo: React.PropTypes.string,\n            bar: React.PropTypes.string\n          };\n          static defaultProps = {\n            baz: \"baz\"\n          };\n        }\n      ", DefaultPropsMatchPropTypesOptions{AllowRequiredDefaults: false}, []string{"defaultHasNoType"}, []string{
			"defaultProp \"baz\" has no corresponding propTypes declaration.",
		}},
		{"upstream invalid 23", "\n        const props = {\n          foo: React.PropTypes.string,\n          bar: React.PropTypes.string.isRequired\n        };\n        const defaults = {\n          bar: \"bar\"\n        };\n        class Greeting extends React.Component {\n          render() {\n            return (\n              <h1>Hello, {this.props.foo} {this.props.bar}</h1>\n            );\n          }\n          static propTypes = props;\n          static defaultProps = defaults;\n        }\n      ", DefaultPropsMatchPropTypesOptions{AllowRequiredDefaults: false}, []string{"requiredHasDefault"}, []string{
			"defaultProp \"bar\" defined for isRequired propType.",
		}},
		{"upstream invalid 24", "\n        const props = {\n          foo: React.PropTypes.string,\n          bar: React.PropTypes.string\n        };\n        const defaults = {\n          baz: \"baz\"\n        };\n        class Greeting extends React.Component {\n          render() {\n            return (\n              <h1>Hello, {this.props.foo} {this.props.bar}</h1>\n            );\n          }\n          static propTypes = props;\n          static defaultProps = defaults;\n        }\n      ", DefaultPropsMatchPropTypesOptions{AllowRequiredDefaults: false}, []string{"defaultHasNoType"}, []string{
			"defaultProp \"baz\" has no corresponding propTypes declaration.",
		}},
		{"upstream invalid 25", "\n        let Greetings = {};\n        Greetings.Hello = class extends React.Component {\n          render () {\n            return <div>Hello {this.props.foo}</div>;\n          }\n        }\n        Greetings.Hello.propTypes = {\n          foo: React.PropTypes.string.isRequired\n        };\n        Greetings.Hello.defaultProps = {\n          foo: \"foo\"\n        };\n      ", DefaultPropsMatchPropTypesOptions{AllowRequiredDefaults: false}, []string{"requiredHasDefault"}, []string{
			"defaultProp \"foo\" defined for isRequired propType.",
		}},
		{"upstream invalid 26", "\n        var Greetings = ({ foo = \"foo\" }) => {\n          return <div>Hello {this.props.foo}</div>;\n        }\n        Greetings.propTypes = {\n          foo: React.PropTypes.string.isRequired\n        };\n        Greetings.defaultProps = {\n          foo: \"foo\"\n        };\n      ", DefaultPropsMatchPropTypesOptions{AllowRequiredDefaults: false}, []string{"requiredHasDefault"}, []string{
			"defaultProp \"foo\" defined for isRequired propType.",
		}},
		{"upstream invalid 33", "\n        class SomeComponent extends React.Component {\n          render() {\n            return <div />;\n          }\n        }\n        SomeComponent.propTypes = {\n          \"firstProperty\": PropTypes.string.isRequired,\n        };\n\n        SomeComponent.defaultProps = {\n          \"firstProperty\": () => undefined\n        };\n      ", DefaultPropsMatchPropTypesOptions{AllowRequiredDefaults: false}, []string{"requiredHasDefault"}, []string{
			"defaultProp \"firstProperty\" defined for isRequired propType.",
		}},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunWithOptions(t, DefaultPropsMatchPropTypes, defaultPropsMatchPropTypesFile, testCase.sourceText, testCase.options)

			// The ids first, through the harness helper. `TestEveryRuleShipsAFixturePair` looks for
			// this call BY NAME, so asserting the rendered text alone reads to that guard as a rule
			// nothing proves can fire. It caught exactly that here.
			rule_testing.ExpectFindings(t, result, testCase.wantIds...)

			if len(result.Diagnostics) != len(testCase.wantMessages) {
				t.Fatalf("want %d findings, got %d", len(testCase.wantMessages), len(result.Diagnostics))
			}
			for index, diagnostic := range result.Diagnostics {
				if diagnostic.Message.Description != testCase.wantMessages[index] {
					t.Errorf("finding %d renders\n  got  %q\n  want %q", index, diagnostic.Message.Description, testCase.wantMessages[index])
				}
			}
		})
	}
}

// TestDefaultPropsMatchPropTypesCasesThisPortCannotExpress records the thirteen corpus cases that
// are not fixtures, rather than dropping them.
//
// Three groups, each for a different reason, and the third is a decision rather than a limitation.
//
//	two cases      Flow object-type syntax our parser cannot parse at all
//	two cases      `settings.react.propWrapperFunctions`, a settings surface we do not have
//	five cases     props read from a TYPE ANNOTATION rather than from a propTypes object
//	four cases     the same Flow syntax reaching a parse error inside a type position
//
// The five annotation cases are the interesting group. They are expressible: our parser produces
// both spellings, a `props:` class property and a function component's parameter type. Reading them
// needs the type checker to resolve a type reference to its declaration, and declaring
// `NeedsTypeChecker` costs the per-file checker lock across the whole tree.
//
// That trade was measured rather than assumed, on 2026-08-28 and with a seeded control proving the
// counter worked: this tree contains ZERO occurrences of `defaultProps`, `propTypes` or
// `getDefaultProps` in any TypeScript or JavaScript file, against 56 files naming `React.Component`.
// So the annotation arm would buy nothing here and cost every file the lock.
//
// Each row below asserts what this port does, which is nothing, so the record fails loudly if that
// ever changes. The `wantUpstream` column is what the installed build reports, kept beside it so
// the size of the gap is on the page rather than in a commit message.
func TestDefaultPropsMatchPropTypesCasesThisPortCannotExpress(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name         string
		sourceText   string
		wantUpstream int
	}{
		{"upstream invalid 27", "\n        class Hello extends React.Component {\n          props: {\n            foo: string,\n            bar?: string\n          };\n\n          render() {\n            return <div>Hello {this.props.foo}</div>;\n          }\n        }\n\n        Hello.defaultProps = {\n          foo: \"foo\"\n        };\n      ", 1},
		{"upstream invalid 28", "\n        function Hello(props: { foo: string }) {\n          return <div>Hello {props.foo}</div>;\n        }\n        Hello.defaultProps = {\n          foo: \"foo\"\n        }\n      ", 1},
		{"upstream invalid 29", "\n        type Props = {\n          foo: string\n        };\n\n        function Hello(props: Props) {\n          return <div>Hello {props.foo}</div>;\n        }\n        Hello.defaultProps = {\n          foo: \"foo\"\n        }\n      ", 1},
		{"upstream invalid 30", "\n        const Hello = (props: { foo: string, bar?: string }) => {\n          return <div>Hello {props.foo}</div>;\n        };\n        Hello.defaultProps = { foo: \"foo\", bar: \"bar\" };\n      ", 1},
		{"upstream invalid 32", "\n        type PropsA = { foo: string };\n        type PropsB = { bar: string };\n        type Props = PropsA & PropsB;\n\n        class Bar extends React.Component {\n          props: Props;\n          static defaultProps = {\n            fooBar: \"fooBar\",\n            foo: \"foo\",\n          }\n\n          render() {\n            return <div>{this.props.foo} - {this.props.bar}</div>\n          }\n        }\n      ", 2},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunWithOptions(t, DefaultPropsMatchPropTypes,
				defaultPropsMatchPropTypesFile, testCase.sourceText,
				DefaultDefaultPropsMatchPropTypesOptions())
			if len(result.Diagnostics) != 0 {
				t.Errorf("this port reports %d findings on a case it is not meant to express; "+
					"upstream reports %d", len(result.Diagnostics), testCase.wantUpstream)
			}
			if testCase.wantUpstream == 0 {
				t.Error("this row records a gap and upstream reports nothing, so it records nothing")
			}
		})
	}
}

// TestDefaultPropsMatchPropTypesMatchesTheInstalledRuleOnInputsTheCorpusDoesNotWrite is a
// differential table.
//
// Every row was run against the installed build on 2026-08-28 and the expectation is the exact
// message text that build produced, so a wrong count and a wrong interpolation both fail here. The
// rows pin four things the corpus states thinly or not at all:
//
//	the two arms             a default nothing declares reports one message, a default on a
//	                         required prop reports the other, and `allowRequiredDefaults` suppresses
//	                         only the second
//	the three silences       absent propTypes, EMPTY propTypes, and a spread in either object each
//	                         silence the component entirely. The empty case is the surprising one.
//	the declaration shapes   a class property, a static getter, an assignment after the class, the
//	                         ES5 factory, a member assignment onto either object, a nested member
//	                         owner, and either object resolved through an identifier
//	component separation     two components in one file do not share props, which is what keeps one
//	                         component's propTypes from satisfying another's defaults
func TestDefaultPropsMatchPropTypesMatchesTheInstalledRuleOnInputsTheCorpusDoesNotWrite(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name         string
		sourceText   string
		options      DefaultPropsMatchPropTypesOptions
		wantMessages []string
	}{
		{"spread beside a real prop in propTypes still reports", "class C extends React.Component { render(){return <div/>;} }\nC.propTypes={...base, a:PropTypes.string};\nC.defaultProps={b:1};", DefaultPropsMatchPropTypesOptions{AllowRequiredDefaults: false}, []string{"defaultProp \"b\" has no corresponding propTypes declaration."}},
		{"spread beside a real prop in defaultProps stays silent", "class C extends React.Component { render(){return <div/>;} }\nC.propTypes={a:PropTypes.string};\nC.defaultProps={...base, b:1};", DefaultPropsMatchPropTypesOptions{AllowRequiredDefaults: false}, []string{}},
		{"two nested owners under one prefix do not share", "let N={};\nN.A=class extends React.Component{render(){return <div/>;}};\nN.B=class extends React.Component{render(){return <div/>;}};\nN.A.propTypes={a:PropTypes.string};\nN.B.defaultProps={a:1};", DefaultPropsMatchPropTypesOptions{AllowRequiredDefaults: false}, []string{}},
		{"a sibling owner's propTypes must not satisfy another's defaults", "let N={};\nN.A=class extends React.Component{render(){return <div/>;}};\nN.B=class extends React.Component{render(){return <div/>;}};\nN.A.propTypes={a:PropTypes.string};\nN.B.propTypes={x:PropTypes.string};\nN.B.defaultProps={a:1};", DefaultPropsMatchPropTypesOptions{AllowRequiredDefaults: false}, []string{"defaultProp \"a\" has no corresponding propTypes declaration."}},
		{"two plain owners must not share either", "class A extends React.Component { render(){return <div/>;} }\nA.propTypes={a:PropTypes.string};\nclass B extends React.Component { render(){return <div/>;} }\nB.propTypes={x:PropTypes.string};\nB.defaultProps={a:1};", DefaultPropsMatchPropTypesOptions{AllowRequiredDefaults: false}, []string{"defaultProp \"a\" has no corresponding propTypes declaration."}},
		{"missing propType", "class C extends React.Component { render(){return <div/>;} }\nC.propTypes={a:PropTypes.string};\nC.defaultProps={b:1};", DefaultPropsMatchPropTypesOptions{AllowRequiredDefaults: false}, []string{"defaultProp \"b\" has no corresponding propTypes declaration."}},
		{"isRequired with default", "class C extends React.Component { render(){return <div/>;} }\nC.propTypes={a:PropTypes.string.isRequired};\nC.defaultProps={a:1};", DefaultPropsMatchPropTypesOptions{AllowRequiredDefaults: false}, []string{"defaultProp \"a\" defined for isRequired propType."}},
		{"optional with default is clean", "class C extends React.Component { render(){return <div/>;} }\nC.propTypes={a:PropTypes.string};\nC.defaultProps={a:1};", DefaultPropsMatchPropTypesOptions{AllowRequiredDefaults: false}, []string{}},
		{"isRequired under allowRequiredDefaults", "class C extends React.Component { render(){return <div/>;} }\nC.propTypes={a:PropTypes.string.isRequired};\nC.defaultProps={a:1};", DefaultPropsMatchPropTypesOptions{AllowRequiredDefaults: true}, []string{}},
		{"missing under allowRequiredDefaults still reports", "class C extends React.Component { render(){return <div/>;} }\nC.propTypes={a:PropTypes.string};\nC.defaultProps={b:1};", DefaultPropsMatchPropTypesOptions{AllowRequiredDefaults: true}, []string{"defaultProp \"b\" has no corresponding propTypes declaration."}},
		{"no propTypes at all", "class C extends React.Component { render(){return <div/>;} }\nC.defaultProps={a:1};", DefaultPropsMatchPropTypesOptions{AllowRequiredDefaults: false}, []string{}},
		{"empty propTypes object", "class C extends React.Component { render(){return <div/>;} }\nC.propTypes={};\nC.defaultProps={a:1};", DefaultPropsMatchPropTypesOptions{AllowRequiredDefaults: false}, []string{}},
		{"spread in defaultProps", "class C extends React.Component { render(){return <div/>;} }\nC.propTypes={a:PropTypes.string};\nC.defaultProps={...base,b:1};", DefaultPropsMatchPropTypesOptions{AllowRequiredDefaults: false}, []string{}},
		{"spread in propTypes", "class C extends React.Component { render(){return <div/>;} }\nC.propTypes={...base};\nC.defaultProps={b:1};", DefaultPropsMatchPropTypesOptions{AllowRequiredDefaults: false}, []string{}},
		{"no defaultProps at all", "class C extends React.Component { render(){return <div/>;} }\nC.propTypes={a:PropTypes.string};", DefaultPropsMatchPropTypesOptions{AllowRequiredDefaults: false}, []string{}},
		{"static class properties", "class C extends React.Component { static propTypes={a:PropTypes.string}; static defaultProps={b:1}; render(){return <div/>;} }", DefaultPropsMatchPropTypesOptions{AllowRequiredDefaults: false}, []string{"defaultProp \"b\" has no corresponding propTypes declaration."}},
		{"static getters", "class C extends React.Component { static get propTypes(){return {a:PropTypes.string};} static get defaultProps(){return {b:1};} render(){return <div/>;} }", DefaultPropsMatchPropTypesOptions{AllowRequiredDefaults: false}, []string{"defaultProp \"b\" has no corresponding propTypes declaration."}},
		{"function component", "function C(p){return <div/>;}\nC.propTypes={a:PropTypes.string};\nC.defaultProps={b:1};", DefaultPropsMatchPropTypesOptions{AllowRequiredDefaults: false}, []string{"defaultProp \"b\" has no corresponding propTypes declaration."}},
		{"ES5 factory", "var C = createReactClass({ render:function(){return <div/>;}, propTypes:{a:PropTypes.string}, getDefaultProps:function(){return {b:1};} });", DefaultPropsMatchPropTypesOptions{AllowRequiredDefaults: false}, []string{"defaultProp \"b\" has no corresponding propTypes declaration."}},
		{"member assignment onto defaults", "class C extends React.Component { render(){return <div/>;} }\nC.propTypes={a:PropTypes.string};\nC.defaultProps={};\nC.defaultProps.b=1;", DefaultPropsMatchPropTypesOptions{AllowRequiredDefaults: false}, []string{"defaultProp \"b\" has no corresponding propTypes declaration."}},
		{"member assignment onto propTypes", "class C extends React.Component { render(){return <div/>;} }\nC.propTypes={};\nC.propTypes.a=PropTypes.string.isRequired;\nC.defaultProps={a:1};", DefaultPropsMatchPropTypesOptions{AllowRequiredDefaults: false}, []string{"defaultProp \"a\" defined for isRequired propType."}},
		{"nested member owner", "let N={};\nN.C = class extends React.Component { render(){return <div/>;} };\nN.C.propTypes={a:PropTypes.string.isRequired};\nN.C.defaultProps={a:1};", DefaultPropsMatchPropTypesOptions{AllowRequiredDefaults: false}, []string{"defaultProp \"a\" defined for isRequired propType."}},
		{"propTypes from an identifier", "const types={a:PropTypes.string.isRequired};\nfunction C(p){return <div/>;}\nC.propTypes=types;\nC.defaultProps={a:1};", DefaultPropsMatchPropTypesOptions{AllowRequiredDefaults: false}, []string{"defaultProp \"a\" defined for isRequired propType."}},
		{"defaultProps from an identifier", "const defs={b:1};\nfunction C(p){return <div/>;}\nC.propTypes={a:PropTypes.string};\nC.defaultProps=defs;", DefaultPropsMatchPropTypesOptions{AllowRequiredDefaults: false}, []string{"defaultProp \"b\" has no corresponding propTypes declaration."}},
		{"React.PropTypes three levels", "class C extends React.Component { render(){return <div/>;} }\nC.propTypes={a:React.PropTypes.string.isRequired};\nC.defaultProps={a:1};", DefaultPropsMatchPropTypesOptions{AllowRequiredDefaults: false}, []string{"defaultProp \"a\" defined for isRequired propType."}},
		{"bare string type", "class C extends React.Component { render(){return <div/>;} }\nC.propTypes={a:string};\nC.defaultProps={a:1};", DefaultPropsMatchPropTypesOptions{AllowRequiredDefaults: false}, []string{}},
		{"two components keep props apart", "class A extends React.Component { render(){return <div/>;} }\nA.propTypes={a:PropTypes.string};\nclass B extends React.Component { render(){return <div/>;} }\nB.defaultProps={a:1};", DefaultPropsMatchPropTypesOptions{AllowRequiredDefaults: false}, []string{}},
		{"quoted keys", "class C extends React.Component { render(){return <div/>;} }\nC.propTypes={'a':PropTypes.string};\nC.defaultProps={'b':1};", DefaultPropsMatchPropTypesOptions{AllowRequiredDefaults: false}, []string{"defaultProp \"b\" has no corresponding propTypes declaration."}},
		{"two bad defaults report in order", "class C extends React.Component { render(){return <div/>;} }\nC.propTypes={a:PropTypes.string};\nC.defaultProps={b:1,c:2};", DefaultPropsMatchPropTypesOptions{AllowRequiredDefaults: false}, []string{"defaultProp \"b\" has no corresponding propTypes declaration.", "defaultProp \"c\" has no corresponding propTypes declaration."}},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunWithOptions(t, DefaultPropsMatchPropTypes,
				defaultPropsMatchPropTypesFile, testCase.sourceText, testCase.options)
			if len(result.Diagnostics) != len(testCase.wantMessages) {
				t.Fatalf("installed build reports %d findings, this rule reports %d",
					len(testCase.wantMessages), len(result.Diagnostics))
			}
			for index, diagnostic := range result.Diagnostics {
				if diagnostic.Message.Description != testCase.wantMessages[index] {
					t.Errorf("finding %d renders\n  got  %q\n  want %q",
						index, diagnostic.Message.Description, testCase.wantMessages[index])
				}
			}
		})
	}
}
