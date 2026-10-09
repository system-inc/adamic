package react

import (
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// The corpus is upstream's, extracted by stubbing its RuleTester and capturing what the test file
// hands it, so no case was retyped and no escape was typed by hand. Every case was run against
// the installed build first and the expectations here are what that build reported.
//
// Seventeen of the eighty cases are excluded and each exclusion is recorded rather than dropped;
// see TestBooleanPropNamingCasesThisPortCannotExpress for the reasons and the counts.
const booleanPropNamingFile = "/repository/source/BooleanPropNaming.tsx"

func TestBooleanPropNamingStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name, sourceText string
		options          BooleanPropNamingOptions
	}{
		{"upstream valid 0", "\n        var Hello = createReactClass({\n          propTypes: {isSomething: PropTypes.bool, hasValue: PropTypes.bool},\n          render: function() { return <div />; }\n        });\n      ", BooleanPropNamingOptions{Rule: "^(is|has)[A-Z]([A-Za-z0-9]?)+", PropTypeNames: []string{"bool"}, ValidateNested: false, Message: ""}},
		{"upstream valid 1", "\n        var Hello = createReactClass({\n          propTypes: {isSomething: PropTypes.bool},\n          render: function() { return <div />; }\n        });\n      ", BooleanPropNamingOptions{Rule: "^is[A-Z]([A-Za-z0-9]?)+", PropTypeNames: []string{"bool"}, ValidateNested: false, Message: ""}},
		{"upstream valid 2", "\n        var Hello = createReactClass({\n          propTypes: {isSomething: React.PropTypes.bool},\n          render: function() { return <div />; }\n        });\n      ", BooleanPropNamingOptions{Rule: "^is[A-Z]([A-Za-z0-9]?)+", PropTypeNames: []string{"bool"}, ValidateNested: false, Message: ""}},
		{"upstream valid 5", "\n        class Hello extends React.Component {\n          render () { return <div />; }\n        }\n        Hello.propTypes = {isSomething: PropTypes.bool}\n      ", BooleanPropNamingOptions{Rule: "^is[A-Z]([A-Za-z0-9]?)+", PropTypeNames: []string{"bool"}, ValidateNested: false, Message: ""}},
		{"upstream valid 6", "\n        class Hello extends React.Component {\n          render () { return <div />; }\n        }\n        Hello.propTypes = wrap({ a: PropTypes.bool })\n      ", BooleanPropNamingOptions{Rule: "^is[A-Z]([A-Za-z0-9]?)+", PropTypeNames: []string{"bool"}, ValidateNested: false, Message: ""}},
		{"upstream valid 7", "\n        class Hello extends React.Component {\n          render () { return <div />; }\n        }\n        Hello.propTypes = {something: PropTypes.any}\n      ", BooleanPropNamingOptions{Rule: "^is[A-Z]([A-Za-z0-9]?)+", PropTypeNames: []string{"bool"}, ValidateNested: false, Message: ""}},
		{"upstream valid 8", "\n        class Hello extends Component {\n          render () { return <div />; }\n        }\n        Hello.propTypes = {isSomething: PropTypes.bool}\n      ", BooleanPropNamingOptions{Rule: "^is[A-Z]([A-Za-z0-9]?)+", PropTypeNames: []string{"bool"}, ValidateNested: false, Message: ""}},
		{"upstream valid 9", "\n        class Hello extends React.Component {\n          static propTypes = {isSomething: PropTypes.bool};\n          render () { return <div />; }\n        }\n      ", BooleanPropNamingOptions{Rule: "^is[A-Z]([A-Za-z0-9]?)+", PropTypeNames: []string{"bool"}, ValidateNested: false, Message: ""}},
		{"upstream valid 10", "\n        const spreadProps = { aSpreadProp: PropTypes.string };\n        class Hello extends React.Component {\n          static propTypes = {isSomething: PropTypes.bool, ...spreadProps};\n          render () { return <div />; }\n        }\n      ", BooleanPropNamingOptions{Rule: "^is[A-Z]([A-Za-z0-9]?)+", PropTypeNames: []string{"bool"}, ValidateNested: false, Message: ""}},
		{"upstream valid 11", "\n        const spreadProps = { aSpreadProp: PropTypes.string };\n        class Hello extends Component {\n          render () { return <div />; }\n        }\n        Hello.propTypes = {isSomething: PropTypes.bool, ...spreadProps}\n      ", BooleanPropNamingOptions{Rule: "^is[A-Z]([A-Za-z0-9]?)+", PropTypeNames: []string{"bool"}, ValidateNested: false, Message: ""}},
		{"upstream valid 12", "\n        class Hello extends React.Component {\n          static propTypes = {isSomething: React.PropTypes.bool};\n          render () { return <div />; }\n        }\n      ", BooleanPropNamingOptions{Rule: "^is[A-Z]([A-Za-z0-9]?)+", PropTypeNames: []string{"bool"}, ValidateNested: false, Message: ""}},
		{"upstream valid 13", "\n        class Hello extends React.Component {\n          static propTypes = {something: PropTypes.any};\n          render () { return <div />; }\n        }\n      ", BooleanPropNamingOptions{Rule: "^is[A-Z]([A-Za-z0-9]?)+", PropTypeNames: []string{"bool"}, ValidateNested: false, Message: ""}},
		{"upstream valid 16", "\n        var Hello = ({isSomething}) => { return <div /> }\n        Hello.propTypes = {isSomething: PropTypes.bool};\n      ", BooleanPropNamingOptions{Rule: "^is[A-Z]([A-Za-z0-9]?)+", PropTypeNames: []string{"bool"}, ValidateNested: false, Message: ""}},
		{"upstream valid 18", "\n        class Hello extends React.Component {\n          static propTypes = {\n            isSomething: PropTypes.mutuallyExclusiveTrueProps,\n            something: PropTypes.bool\n          };\n          render () { return <div />; }\n        }\n      ", BooleanPropNamingOptions{Rule: "^is[A-Z]([A-Za-z0-9]?)+", PropTypeNames: []string{"mutuallyExclusiveTrueProps"}, ValidateNested: false, Message: ""}},
		{"upstream valid 19", "\n        class Hello extends React.Component {\n          static propTypes = {\n            isSomething: mutuallyExclusiveTrueProps,\n            isSomethingElse: bool\n          };\n          render () { return <div />; }\n        }\n      ", BooleanPropNamingOptions{Rule: "^is[A-Z]([A-Za-z0-9]?)+", PropTypeNames: []string{"bool", "mutuallyExclusiveTrueProps"}, ValidateNested: false, Message: ""}},
		{"upstream valid 20", "\n        var x = {a: 1}\n        var y = {...x}\n      ", BooleanPropNamingOptions{Rule: "^is[A-Z]([A-Za-z0-9]?)+", PropTypeNames: []string{"bool"}, ValidateNested: false, Message: ""}},
		{"upstream valid 22", "\n      function Card(props) {\n        return <div>{props.showScore ? 'yeh' : 'no'}</div>;\n      }\n      Card.propTypes = merge({}, Card.propTypes, {\n          showScore: PropTypes.bool\n      });", BooleanPropNamingOptions{Rule: "^(is|has)[A-Z]([A-Za-z0-9]?)+", PropTypeNames: []string{"bool"}, ValidateNested: false, Message: ""}},
		{"upstream valid 23", "\n        var Hello = createReactClass({\n          propTypes: {isSomething: PropTypes.bool.isRequired, hasValue: PropTypes.bool.isRequired},\n          render: function() { return <div />; }\n        });\n      ", BooleanPropNamingOptions{Rule: "^(is|has)[A-Z]([A-Za-z0-9]?)+", PropTypeNames: []string{"bool"}, ValidateNested: false, Message: ""}},
		{"upstream valid 24", "\n        class Hello extends React.Component {\n          static propTypes = {\n            isSomething: PropTypes.bool.isRequired,\n            hasValue: PropTypes.bool.isRequired\n          };\n\n          render() {\n            return (\n              <div />\n            );\n          }\n        }\n      ", BooleanPropNamingOptions{Rule: "^(is|has)[A-Z]([A-Za-z0-9]?)+", PropTypeNames: []string{"bool"}, ValidateNested: false, Message: ""}},
		{"upstream valid 25", "\n        class Hello extends React.Component {\n          render() {\n            return (\n              <div />\n            );\n          }\n        }\n\n        Hello.propTypes = {\n          isSomething: PropTypes.bool.isRequired,\n          hasValue: PropTypes.bool.isRequired\n        }\n      ", BooleanPropNamingOptions{Rule: "^(is|has)[A-Z]([A-Za-z0-9]?)+", PropTypeNames: []string{"bool"}, ValidateNested: false, Message: ""}},
		{"upstream valid 26", "\n        var Hello = createReactClass({\n          propTypes: {something: PropTypes.shape({}).isRequired},\n          render: function() { return <div />; }\n        });\n      ", BooleanPropNamingOptions{Rule: "^is[A-Z]([A-Za-z0-9]?)+", PropTypeNames: []string{"bool"}, ValidateNested: false, Message: ""}},
		{"upstream valid 28", "\n        class Hello extends React.Component {\n          render() {\n            return (\n              <div />\n            );\n          }\n        }\n\n        Hello.propTypes = {\n          isSomething: PropTypes.bool.isRequired,\n          nested: PropTypes.shape({\n            isWorking: PropTypes.bool\n          })\n        };\n      ", BooleanPropNamingOptions{Rule: "^(is|has)[A-Z]([A-Za-z0-9]?)+", PropTypeNames: []string{"bool"}, ValidateNested: false, Message: ""}},
		{"upstream valid 29", "\n        class Hello extends React.Component {\n          render() {\n            return (\n              <div />\n            );\n          }\n        }\n\n        Hello.propTypes = {\n          isSomething: PropTypes.bool.isRequired,\n          nested: PropTypes.shape({\n            nested: PropTypes.shape({\n              isWorking: PropTypes.bool\n            })\n          })\n        };\n      ", BooleanPropNamingOptions{Rule: "^is[A-Z]([A-Za-z0-9]?)+", PropTypeNames: []string{"bool"}, ValidateNested: true, Message: ""}},
		{"upstream valid 30", "\n        type TestFNType = {\n          isEnabled: boolean\n        }\n        const HelloNew = (props: TestFNType) => { return <div /> };\n      ", BooleanPropNamingOptions{Rule: "^is[A-Z]([A-Za-z0-9]?)+", PropTypeNames: []string{"bool"}, ValidateNested: false, Message: ""}},
		{"upstream valid 31", "\n        type Props = {\n          isEnabled: boolean\n        } & OtherProps\n        const HelloNew = (props: Props) => { return <div /> };\n      ", BooleanPropNamingOptions{Rule: "^is[A-Z]([A-Za-z0-9]?)+", PropTypeNames: []string{"bool"}, ValidateNested: false, Message: ""}},
		{"upstream valid 32", "\n        type Props = {\n          isEnabled: boolean\n        } & {\n          hasLOL: boolean\n        } & OtherProps\n        const HelloNew = (props: Props) => { return <div /> };\n      ", BooleanPropNamingOptions{Rule: "(is|has)[A-Z]([A-Za-z0-9]?)+", PropTypeNames: []string{"bool"}, ValidateNested: false, Message: ""}},
		{"upstream valid 33", "\n        type Props = {\n          isEnabled: boolean\n        }\n\n        const HelloNew: React.FC<Props> = (props) => { return <div /> };\n      ", BooleanPropNamingOptions{Rule: "^is[A-Z]([A-Za-z0-9]?)+", PropTypeNames: []string{"bool"}, ValidateNested: false, Message: ""}},
		{"upstream valid 34", "\n        type Props = {\n          isEnabled: boolean\n        } & {\n          hasLOL: boolean\n        }\n\n        const HelloNew: React.FC<Props> = (props) => { return <div /> };\n      ", BooleanPropNamingOptions{Rule: "^(is|has)[A-Z]([A-Za-z0-9]?)+", PropTypeNames: []string{"bool"}, ValidateNested: false, Message: ""}},
		{"upstream valid 35", "\n        type Props = {\n          isEnabled: boolean\n        } | {\n          hasLOL: boolean\n        }\n\n        const HelloNew = (props: Props) => { return <div /> };\n      ", BooleanPropNamingOptions{Rule: "^(is|has)[A-Z]([A-Za-z0-9]?)+", PropTypeNames: []string{"bool"}, ValidateNested: false, Message: ""}},
		{"upstream valid 36", "\n        type Props = {\n          isEnabled: boolean\n        } & ({\n          hasLOL: boolean\n        } | {\n          isLOL: boolean\n        })\n\n        const HelloNew = (props: Props) => { return <div /> };\n      ", BooleanPropNamingOptions{Rule: "^(is|has)[A-Z]([A-Za-z0-9]?)+", PropTypeNames: []string{"bool"}, ValidateNested: false, Message: ""}},
		{"upstream valid 37", "\n        export const DataRow = (props: { label: string; value: string; } & React.HTMLAttributes<HTMLDivElement>) => {\n            const { label, value, ...otherProps } = props;\n            return (\n                <div {...otherProps}>\n                    <span>{label}</span>\n                    <span>{value}</span>\n                </div>\n            );\n        };\n      ", BooleanPropNamingOptions{Rule: "(^(is|has|should|without)[A-Z]([A-Za-z0-9]?)+|disabled|required|checked|defaultChecked)", PropTypeNames: []string{"bool"}, ValidateNested: false, Message: ""}},
		{"upstream valid 38", "\n        // Strip @jsx comments, see https://github.com/microsoft/fluentui/issues/29126\n        const resultCode = result.code\n          .replace('/** @jsxRuntime automatic */', '')\n          .replace('/** @jsxImportSource @fluentui/react-jsx-runtime */', '');\n      ", BooleanPropNamingOptions{Rule: "^(is|has)[A-Z]([A-Za-z0-9]?)+", PropTypeNames: []string{"bool"}, ValidateNested: false, Message: ""}},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTypedWithOptions(t, BooleanPropNaming, booleanPropNamingFile, testCase.sourceText, testCase.options)
			rule_testing.ExpectClean(t, result)
		})
	}
}

// TestBooleanPropNamingFires asserts the count AND the rendered message of every finding. The
// message interpolates the prop name and the pattern, so a message-id assertion could not see
// an interpolation defect at all.
func TestBooleanPropNamingFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name, sourceText string
		options          BooleanPropNamingOptions
		wantMessages     []string
	}{
		{"upstream invalid 0", "\n        var Hello = createReactClass({\n          propTypes: {something: PropTypes.bool},\n          render: function() { return <div />; }\n        });\n      ", BooleanPropNamingOptions{Rule: "^is[A-Z]([A-Za-z0-9]?)+", PropTypeNames: []string{"bool"}, ValidateNested: false, Message: ""}, []string{
			"Prop name `something` doesn’t match rule `^is[A-Z]([A-Za-z0-9]?)+`",
		}},
		{"upstream invalid 1", "\n        var Hello = createReactClass({\n          propTypes: {something: React.PropTypes.bool},\n          render: function() { return <div />; }\n        });\n      ", BooleanPropNamingOptions{Rule: "^is[A-Z]([A-Za-z0-9]?)+", PropTypeNames: []string{"bool"}, ValidateNested: false, Message: ""}, []string{
			"Prop name `something` doesn’t match rule `^is[A-Z]([A-Za-z0-9]?)+`",
		}},
		{"upstream invalid 3", "\n        class Hello extends React.Component {\n          render () { return <div />; }\n        }\n        Hello.propTypes = {something: PropTypes.bool}\n      ", BooleanPropNamingOptions{Rule: "^is[A-Z]([A-Za-z0-9]?)+", PropTypeNames: []string{"bool"}, ValidateNested: false, Message: ""}, []string{
			"Prop name `something` doesn’t match rule `^is[A-Z]([A-Za-z0-9]?)+`",
		}},
		{"upstream invalid 4", "\n        class Hello extends Component {\n          render () { return <div />; }\n        }\n        Hello.propTypes = {something: PropTypes.bool}\n      ", BooleanPropNamingOptions{Rule: "^is[A-Z]([A-Za-z0-9]?)+", PropTypeNames: []string{"bool"}, ValidateNested: false, Message: ""}, []string{
			"Prop name `something` doesn’t match rule `^is[A-Z]([A-Za-z0-9]?)+`",
		}},
		{"upstream invalid 5", "\n        class Hello extends React.Component {\n          static propTypes = {something: PropTypes.bool};\n          render () { return <div />; }\n        }\n      ", BooleanPropNamingOptions{Rule: "^is[A-Z]([A-Za-z0-9]?)+", PropTypeNames: []string{"bool"}, ValidateNested: false, Message: ""}, []string{
			"Prop name `something` doesn’t match rule `^is[A-Z]([A-Za-z0-9]?)+`",
		}},
		{"upstream invalid 6", "\n        const spreadProps = { aSpreadProp: PropTypes.string };\n        class Hello extends Component {\n          render () { return <div />; }\n        }\n        Hello.propTypes = {something: PropTypes.bool, ...spreadProps}\n      ", BooleanPropNamingOptions{Rule: "^is[A-Z]([A-Za-z0-9]?)+", PropTypeNames: []string{"bool"}, ValidateNested: false, Message: ""}, []string{
			"Prop name `something` doesn’t match rule `^is[A-Z]([A-Za-z0-9]?)+`",
		}},
		{"upstream invalid 7", "\n        const spreadProps = { aSpreadProp: PropTypes.string };\n        class Hello extends React.Component {\n          static propTypes = {something: PropTypes.bool, ...spreadProps};\n          render () { return <div />; }\n        }\n      ", BooleanPropNamingOptions{Rule: "^is[A-Z]([A-Za-z0-9]?)+", PropTypeNames: []string{"bool"}, ValidateNested: false, Message: ""}, []string{
			"Prop name `something` doesn’t match rule `^is[A-Z]([A-Za-z0-9]?)+`",
		}},
		{"upstream invalid 9", "\n        var Hello = ({something}) => { return <div /> }\n        Hello.propTypes = {something: PropTypes.bool};\n      ", BooleanPropNamingOptions{Rule: "^is[A-Z]([A-Za-z0-9]?)+", PropTypeNames: []string{"bool"}, ValidateNested: false, Message: ""}, []string{
			"Prop name `something` doesn’t match rule `^is[A-Z]([A-Za-z0-9]?)+`",
		}},
		{"upstream invalid 10", "\n        type Props = {\n          something: boolean;\n        };\n        function Hello(props: Props): React.Element { return <div /> }\n      ", BooleanPropNamingOptions{Rule: "^is[A-Z]([A-Za-z0-9]?)+", PropTypeNames: []string{"bool"}, ValidateNested: false, Message: ""}, []string{
			"Prop name `something` doesn’t match rule `^is[A-Z]([A-Za-z0-9]?)+`",
		}},
		{"upstream invalid 12", "\n        class Hello extends React.Component {\n          static propTypes = {\n            something: PropTypes.mutuallyExclusiveTrueProps,\n            somethingElse: PropTypes.bool\n          };\n          render () { return <div />; }\n        }\n      ", BooleanPropNamingOptions{Rule: "^is[A-Z]([A-Za-z0-9]?)+", PropTypeNames: []string{"bool", "mutuallyExclusiveTrueProps"}, ValidateNested: false, Message: ""}, []string{
			"Prop name `something` doesn’t match rule `^is[A-Z]([A-Za-z0-9]?)+`",
			"Prop name `somethingElse` doesn’t match rule `^is[A-Z]([A-Za-z0-9]?)+`",
		}},
		{"upstream invalid 13", "\n        class Hello extends React.Component {\n          static propTypes = {\n            something: mutuallyExclusiveTrueProps,\n            somethingElse: bool\n          };\n          render () { return <div />; }\n        }\n      ", BooleanPropNamingOptions{Rule: "^is[A-Z]([A-Za-z0-9]?)+", PropTypeNames: []string{"bool", "mutuallyExclusiveTrueProps"}, ValidateNested: false, Message: ""}, []string{
			"Prop name `something` doesn’t match rule `^is[A-Z]([A-Za-z0-9]?)+`",
			"Prop name `somethingElse` doesn’t match rule `^is[A-Z]([A-Za-z0-9]?)+`",
		}},
		{"upstream invalid 20", "\n        class Hello extends React.Component {\n          render () { return <div />; }\n        }\n        Hello.propTypes = {something: PropTypes.bool}\n      ", BooleanPropNamingOptions{Rule: "^is[A-Z]([A-Za-z0-9]?)+", PropTypeNames: []string{"bool"}, ValidateNested: false, Message: "Boolean prop names must begin with either 'is' or 'has'"}, []string{
			"Boolean prop names must begin with either 'is' or 'has'",
		}},
		{"upstream invalid 21", "\n        class Hello extends React.Component {\n          render () { return <div />; }\n        }\n        Hello.propTypes = {something: PropTypes.bool}\n      ", BooleanPropNamingOptions{Rule: "^is[A-Z]([A-Za-z0-9]?)+", PropTypeNames: []string{"bool"}, ValidateNested: false, Message: "It is better if your prop ({{ propName }}) matches this pattern: ({{ pattern }})"}, []string{
			"It is better if your prop (something) matches this pattern: (^is[A-Z]([A-Za-z0-9]?)+)",
		}},
		{"upstream invalid 22", "\n        var Hello = createReactClass({\n          propTypes: {something: PropTypes.bool.isRequired},\n          render: function() { return <div />; }\n        });\n      ", BooleanPropNamingOptions{Rule: "^is[A-Z]([A-Za-z0-9]?)+", PropTypeNames: []string{"bool"}, ValidateNested: false, Message: ""}, []string{
			"Prop name `something` doesn’t match rule `^is[A-Z]([A-Za-z0-9]?)+`",
		}},
		{"upstream invalid 23", "\n        class Hello extends React.Component {\n          static propTypes = {\n            something: PropTypes.bool.isRequired\n          };\n\n          render() {\n            return (\n              <div />\n            );\n          }\n        }\n      ", BooleanPropNamingOptions{Rule: "^is[A-Z]([A-Za-z0-9]?)+", PropTypeNames: []string{"bool"}, ValidateNested: false, Message: ""}, []string{
			"Prop name `something` doesn’t match rule `^is[A-Z]([A-Za-z0-9]?)+`",
		}},
		{"upstream invalid 24", "\n        class Hello extends React.Component {\n          render() {\n            return (\n              <div />\n            );\n          }\n        }\n\n        Hello.propTypes = {\n          something: PropTypes.bool.isRequired\n        }\n      ", BooleanPropNamingOptions{Rule: "^is[A-Z]([A-Za-z0-9]?)+", PropTypeNames: []string{"bool"}, ValidateNested: false, Message: ""}, []string{
			"Prop name `something` doesn’t match rule `^is[A-Z]([A-Za-z0-9]?)+`",
		}},
		{"upstream invalid 26", "\n        class Hello extends React.Component {\n          render() {\n            return (\n              <div />\n            );\n          }\n        }\n\n        Hello.propTypes = {\n          isSomething: PropTypes.bool.isRequired,\n          nested: PropTypes.shape({\n            failingItIs: PropTypes.bool\n          })\n        };\n      ", BooleanPropNamingOptions{Rule: "^is[A-Z]([A-Za-z0-9]?)+", PropTypeNames: []string{"bool"}, ValidateNested: true, Message: ""}, []string{
			"Prop name `failingItIs` doesn’t match rule `^is[A-Z]([A-Za-z0-9]?)+`",
		}},
		{"upstream invalid 27", "\n        class Hello extends React.Component {\n          render() {\n            return (\n              <div />\n            );\n          }\n        }\n\n        Hello.propTypes = {\n          isSomething: PropTypes.bool.isRequired,\n          nested: PropTypes.shape({\n            nested: PropTypes.shape({\n              failingItIs: PropTypes.bool\n            })\n          })\n        };\n      ", BooleanPropNamingOptions{Rule: "^is[A-Z]([A-Za-z0-9]?)+", PropTypeNames: []string{"bool"}, ValidateNested: true, Message: ""}, []string{
			"Prop name `failingItIs` doesn’t match rule `^is[A-Z]([A-Za-z0-9]?)+`",
		}},
		{"upstream invalid 28", "\n      import { bool } from 'prop-types';\n        var Hello = createReactClass({\n          propTypes: {something: bool},\n          render: function() { return <div />; }\n        });\n      ", BooleanPropNamingOptions{Rule: "^is[A-Z]([A-Za-z0-9]?)+", PropTypeNames: []string{"bool"}, ValidateNested: true, Message: ""}, []string{
			"Prop name `something` doesn’t match rule `^is[A-Z]([A-Za-z0-9]?)+`",
		}},
		{"upstream invalid 29", "\n        type TestConstType = {\n          enabled: boolean\n        }\n        const HelloNew = (props: TestConstType) => { return <div /> };\n      ", BooleanPropNamingOptions{Rule: "^is[A-Z]([A-Za-z0-9]?)+", PropTypeNames: []string{"bool"}, ValidateNested: false, Message: ""}, []string{
			"Prop name `enabled` doesn’t match rule `^is[A-Z]([A-Za-z0-9]?)+`",
		}},
		{"upstream invalid 30", "\n        type TestFNType = {\n          enabled: boolean\n        }\n        const HelloNew = (props: TestFNType) => { return <div /> };\n      ", BooleanPropNamingOptions{Rule: "^is[A-Z]([A-Za-z0-9]?)+", PropTypeNames: []string{"bool"}, ValidateNested: false, Message: ""}, []string{
			"Prop name `enabled` doesn’t match rule `^is[A-Z]([A-Za-z0-9]?)+`",
		}},
		{"upstream invalid 31", "\n        type Props = {\n          enabled: boolean\n        } & OtherProps\n\n        const HelloNew = (props: Props) => { return <div /> };\n      ", BooleanPropNamingOptions{Rule: "^is[A-Z]([A-Za-z0-9]?)+", PropTypeNames: []string{"bool"}, ValidateNested: false, Message: ""}, []string{
			"Prop name `enabled` doesn’t match rule `^is[A-Z]([A-Za-z0-9]?)+`",
		}},
		{"upstream invalid 32", "\n        type Props = {\n          enabled: boolean\n        } & {\n          hasLOL: boolean\n        } & OtherProps\n\n        const HelloNew = (props: Props) => { return <div /> };\n      ", BooleanPropNamingOptions{Rule: "^(is|has)[A-Z]([A-Za-z0-9]?)+", PropTypeNames: []string{"bool"}, ValidateNested: false, Message: ""}, []string{
			"Prop name `enabled` doesn’t match rule `^(is|has)[A-Z]([A-Za-z0-9]?)+`",
		}},
		{"upstream invalid 33", "\n        type Props = {\n          enabled: boolean\n        }\n\n        const HelloNew: React.FC<Props> = (props) => { return <div /> };\n      ", BooleanPropNamingOptions{Rule: "^is[A-Z]([A-Za-z0-9]?)+", PropTypeNames: []string{"bool"}, ValidateNested: false, Message: ""}, []string{
			"Prop name `enabled` doesn’t match rule `^is[A-Z]([A-Za-z0-9]?)+`",
		}},
		{"upstream invalid 34", "\n        type Props = {\n          enabled: boolean\n        } & {\n          hasLOL: boolean\n        }\n\n        const HelloNew: React.FC<Props> = (props) => { return <div /> };\n      ", BooleanPropNamingOptions{Rule: "^(is|has)[A-Z]([A-Za-z0-9]?)+", PropTypeNames: []string{"bool"}, ValidateNested: false, Message: ""}, []string{
			"Prop name `enabled` doesn’t match rule `^(is|has)[A-Z]([A-Za-z0-9]?)+`",
		}},
		{"upstream invalid 35", "\n        type Props = {\n          enabled: boolean\n        } | {\n          hasLOL: boolean\n        }\n\n        const HelloNew = (props: Props) => { return <div /> };\n      ", BooleanPropNamingOptions{Rule: "^(is|has)[A-Z]([A-Za-z0-9]?)+", PropTypeNames: []string{"bool"}, ValidateNested: false, Message: ""}, []string{
			"Prop name `enabled` doesn’t match rule `^(is|has)[A-Z]([A-Za-z0-9]?)+`",
		}},
		{"upstream invalid 36", "\n        type Props = {\n          enabled: boolean\n        } & ({\n          hasLOL: boolean\n        } | {\n          lol: boolean\n        })\n\n        const HelloNew = (props: Props) => { return <div /> };\n      ", BooleanPropNamingOptions{Rule: "^(is|has)[A-Z]([A-Za-z0-9]?)+", PropTypeNames: []string{"bool"}, ValidateNested: false, Message: ""}, []string{
			"Prop name `enabled` doesn’t match rule `^(is|has)[A-Z]([A-Za-z0-9]?)+`",
			"Prop name `lol` doesn’t match rule `^(is|has)[A-Z]([A-Za-z0-9]?)+`",
		}},
		{"upstream invalid 37", "\n        interface TestFNType {\n          enabled: boolean\n        }\n        const HelloNew = (props: TestFNType) => { return <div /> };\n      ", BooleanPropNamingOptions{Rule: "^is[A-Z]([A-Za-z0-9]?)+", PropTypeNames: []string{"bool"}, ValidateNested: false, Message: ""}, []string{
			"Prop name `enabled` doesn’t match rule `^is[A-Z]([A-Za-z0-9]?)+`",
		}},
		{"upstream invalid 38", "const Hello = (props: {enabled:boolean}) => <div />;", BooleanPropNamingOptions{Rule: "^(is|has)[A-Z]([A-Za-z0-9]?)+", PropTypeNames: []string{"bool"}, ValidateNested: false, Message: ""}, []string{
			"Prop name `enabled` doesn’t match rule `^(is|has)[A-Z]([A-Za-z0-9]?)+`",
		}},
		{"upstream invalid 39", "\n        type Props = {\n          enabled: boolean\n        };\n        type BaseProps = {\n          semi: boolean\n        };\n\n        const Hello = (props: Props & BaseProps) => <div />;\n      ", BooleanPropNamingOptions{Rule: "^(is|has)[A-Z]([A-Za-z0-9]?)+", PropTypeNames: []string{"bool"}, ValidateNested: false, Message: ""}, []string{
			"Prop name `enabled` doesn’t match rule `^(is|has)[A-Z]([A-Za-z0-9]?)+`",
			"Prop name `semi` doesn’t match rule `^(is|has)[A-Z]([A-Za-z0-9]?)+`",
		}},
		{"upstream invalid 40", "\n        type Props = {\n          enabled: boolean\n        };\n\n        const Hello = (props: Props & {\n          semi: boolean\n        }) => <div />;\n      ", BooleanPropNamingOptions{Rule: "^(is|has)[A-Z]([A-Za-z0-9]?)+", PropTypeNames: []string{"bool"}, ValidateNested: false, Message: ""}, []string{
			"Prop name `enabled` doesn’t match rule `^(is|has)[A-Z]([A-Za-z0-9]?)+`",
			"Prop name `semi` doesn’t match rule `^(is|has)[A-Z]([A-Za-z0-9]?)+`",
		}},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTypedWithOptions(t, BooleanPropNaming, booleanPropNamingFile, testCase.sourceText, testCase.options)
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

// TestBooleanPropNamingMatchesTheInstalledRuleOnInputsTheCorpusDoesNotWrite is a differential table.
//
// Every row was run against the installed build on 2026-08-28 and the expectation is the exact
// message text that build produced, so a wrong count and a wrong interpolation both fail here.
// The rows pin five things the corpus states thinly or not at all:
//
//	the prop-type spellings   `PropTypes.bool`, `React.PropTypes.bool`, a bare `bool`, and one
//	                          `isRequired` peel all count; a SECOND peel does not
//	the key defect            a quoted or numeric key renders the word undefined and ALWAYS
//	                          reports, so `'isEnabled'` cannot pass. A computed key is different
//	                          and renders the identifier underneath.
//	wrapper calls             every call form is silent with no propWrapperFunctions configured,
//	                          which is the only state we can express
//	nesting                   `PropTypes.shape({...})` is reached only under validateNested, and
//	                          upstream tests only that the value is a CALL, so `arrayOf` too
//	the type shapes           an interface, an inline literal, a type alias, an intersection, a
//	                          union, and a PARENTHESIZED intersection all reach their members
func TestBooleanPropNamingMatchesTheInstalledRuleOnInputsTheCorpusDoesNotWrite(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name         string
		sourceText   string
		options      BooleanPropNamingOptions
		wantMessages []string
	}{
		{"default pattern accepts isX", "class C extends React.Component { render(){return <div/>;} }\nC.propTypes = { isEnabled: PropTypes.bool };", DefaultBooleanPropNamingOptions(), []string{}},
		{"default pattern accepts hasX", "class C extends React.Component { render(){return <div/>;} }\nC.propTypes = { hasThing: PropTypes.bool };", DefaultBooleanPropNamingOptions(), []string{}},
		{"default pattern rejects a bare name", "class C extends React.Component { render(){return <div/>;} }\nC.propTypes = { enabled: PropTypes.bool };", DefaultBooleanPropNamingOptions(), []string{"Prop name `enabled` doesn\u2019t match rule `^(is|has)[A-Z]([A-Za-z0-9]?)+`"}},
		{"default pattern rejects isx lowercase", "class C extends React.Component { render(){return <div/>;} }\nC.propTypes = { isenabled: PropTypes.bool };", DefaultBooleanPropNamingOptions(), []string{"Prop name `isenabled` doesn\u2019t match rule `^(is|has)[A-Z]([A-Za-z0-9]?)+`"}},
		{"PropTypes.bool", "class C extends React.Component { render(){return <div/>;} }\nC.propTypes = { enabled: PropTypes.bool };", BooleanPropNamingOptions{Rule: "^(is|has)[A-Z]([A-Za-z0-9]?)+", PropTypeNames: []string{"bool"}, ValidateNested: false}, []string{"Prop name `enabled` doesn’t match rule `^(is|has)[A-Z]([A-Za-z0-9]?)+`"}},
		{"React.PropTypes.bool", "class C extends React.Component { render(){return <div/>;} }\nC.propTypes = { enabled: React.PropTypes.bool };", BooleanPropNamingOptions{Rule: "^(is|has)[A-Z]([A-Za-z0-9]?)+", PropTypeNames: []string{"bool"}, ValidateNested: false}, []string{"Prop name `enabled` doesn’t match rule `^(is|has)[A-Z]([A-Za-z0-9]?)+`"}},
		{"bare bool identifier", "class C extends React.Component { render(){return <div/>;} }\nC.propTypes = { enabled: bool };", BooleanPropNamingOptions{Rule: "^(is|has)[A-Z]([A-Za-z0-9]?)+", PropTypeNames: []string{"bool"}, ValidateNested: false}, []string{"Prop name `enabled` doesn’t match rule `^(is|has)[A-Z]([A-Za-z0-9]?)+`"}},
		{"isRequired peel", "class C extends React.Component { render(){return <div/>;} }\nC.propTypes = { enabled: PropTypes.bool.isRequired };", BooleanPropNamingOptions{Rule: "^(is|has)[A-Z]([A-Za-z0-9]?)+", PropTypeNames: []string{"bool"}, ValidateNested: false}, []string{"Prop name `enabled` doesn’t match rule `^(is|has)[A-Z]([A-Za-z0-9]?)+`"}},
		{"double isRequired", "class C extends React.Component { render(){return <div/>;} }\nC.propTypes = { enabled: PropTypes.bool.isRequired.isRequired };", BooleanPropNamingOptions{Rule: "^(is|has)[A-Z]([A-Za-z0-9]?)+", PropTypeNames: []string{"bool"}, ValidateNested: false}, []string{}},
		{"not boolean", "class C extends React.Component { render(){return <div/>;} }\nC.propTypes = { enabled: PropTypes.string };", BooleanPropNamingOptions{Rule: "^(is|has)[A-Z]([A-Za-z0-9]?)+", PropTypeNames: []string{"bool"}, ValidateNested: false}, []string{}},
		{"spread element", "class C extends React.Component { render(){return <div/>;} }\nC.propTypes = { ...other, enabled: PropTypes.bool };", BooleanPropNamingOptions{Rule: "^(is|has)[A-Z]([A-Za-z0-9]?)+", PropTypeNames: []string{"bool"}, ValidateNested: false}, []string{"Prop name `enabled` doesn’t match rule `^(is|has)[A-Z]([A-Za-z0-9]?)+`"}},
		{"computed key renders the identifier", "class C extends React.Component { render(){return <div/>;} }\nC.propTypes = { [k]: PropTypes.bool };", BooleanPropNamingOptions{Rule: "^(is|has)[A-Z]([A-Za-z0-9]?)+", PropTypeNames: []string{"bool"}, ValidateNested: false}, []string{"Prop name `k` doesn’t match rule `^(is|has)[A-Z]([A-Za-z0-9]?)+`"}},
		{"quoted invalid key", "class C extends React.Component { render(){return <div/>;} }\nC.propTypes = { 'enabled': PropTypes.bool };", BooleanPropNamingOptions{Rule: "^(is|has)[A-Z]([A-Za-z0-9]?)+", PropTypeNames: []string{"bool"}, ValidateNested: false}, []string{"Prop name `undefined` doesn’t match rule `^(is|has)[A-Z]([A-Za-z0-9]?)+`"}},
		{"quoted VALID key still reports", "class C extends React.Component { render(){return <div/>;} }\nC.propTypes = { 'isEnabled': PropTypes.bool };", BooleanPropNamingOptions{Rule: "^(is|has)[A-Z]([A-Za-z0-9]?)+", PropTypeNames: []string{"bool"}, ValidateNested: false}, []string{"Prop name `undefined` doesn’t match rule `^(is|has)[A-Z]([A-Za-z0-9]?)+`"}},
		{"numeric key", "class C extends React.Component { render(){return <div/>;} }\nC.propTypes = { 1: PropTypes.bool };", BooleanPropNamingOptions{Rule: "^(is|has)[A-Z]([A-Za-z0-9]?)+", PropTypeNames: []string{"bool"}, ValidateNested: false}, []string{"Prop name `undefined` doesn’t match rule `^(is|has)[A-Z]([A-Za-z0-9]?)+`"}},
		{"valid name passes", "class C extends React.Component { render(){return <div/>;} }\nC.propTypes = { isEnabled: PropTypes.bool };", BooleanPropNamingOptions{Rule: "^(is|has)[A-Z]([A-Za-z0-9]?)+", PropTypeNames: []string{"bool"}, ValidateNested: false}, []string{}},
		{"hasX passes", "class C extends React.Component { render(){return <div/>;} }\nC.propTypes = { hasThing: PropTypes.bool };", BooleanPropNamingOptions{Rule: "^(is|has)[A-Z]([A-Za-z0-9]?)+", PropTypeNames: []string{"bool"}, ValidateNested: false}, []string{}},
		{"wrap call is silent", "class C extends React.Component { render(){return <div/>;} }\nC.propTypes = wrap({ enabled: PropTypes.bool });", BooleanPropNamingOptions{Rule: "^(is|has)[A-Z]([A-Za-z0-9]?)+", PropTypeNames: []string{"bool"}, ValidateNested: false}, []string{}},
		{"forbidExtraProps is silent", "class C extends React.Component { render(){return <div/>;} }\nC.propTypes = forbidExtraProps({ enabled: PropTypes.bool });", BooleanPropNamingOptions{Rule: "^(is|has)[A-Z]([A-Za-z0-9]?)+", PropTypeNames: []string{"bool"}, ValidateNested: false}, []string{}},
		{"shape nested off", "class C extends React.Component { render(){return <div/>;} }\nC.propTypes = { nested: PropTypes.shape({ enabled: PropTypes.bool }) };", BooleanPropNamingOptions{Rule: "^(is|has)[A-Z]([A-Za-z0-9]?)+", PropTypeNames: []string{"bool"}, ValidateNested: false}, []string{}},
		{"shape nested on", "class C extends React.Component { render(){return <div/>;} }\nC.propTypes = { nested: PropTypes.shape({ enabled: PropTypes.bool }) };", BooleanPropNamingOptions{Rule: "^(is|has)[A-Z]([A-Za-z0-9]?)+", PropTypeNames: []string{"bool"}, ValidateNested: true}, []string{"Prop name `enabled` doesn’t match rule `^(is|has)[A-Z]([A-Za-z0-9]?)+`"}},
		{"arrayOf nested on", "class C extends React.Component { render(){return <div/>;} }\nC.propTypes = { n: PropTypes.arrayOf({ enabled: PropTypes.bool }) };", BooleanPropNamingOptions{Rule: "^(is|has)[A-Z]([A-Za-z0-9]?)+", PropTypeNames: []string{"bool"}, ValidateNested: true}, []string{"Prop name `enabled` doesn’t match rule `^(is|has)[A-Z]([A-Za-z0-9]?)+`"}},
		{"static propTypes", "class C extends React.Component { static propTypes = { enabled: PropTypes.bool }; render(){return <div/>;} }", BooleanPropNamingOptions{Rule: "^(is|has)[A-Z]([A-Za-z0-9]?)+", PropTypeNames: []string{"bool"}, ValidateNested: false}, []string{"Prop name `enabled` doesn’t match rule `^(is|has)[A-Z]([A-Za-z0-9]?)+`"}},
		{"interface reference", "interface Props { enabled: boolean }\nfunction C(p: Props) { return <div/>; }", BooleanPropNamingOptions{Rule: "^(is|has)[A-Z]([A-Za-z0-9]?)+", PropTypeNames: []string{"bool"}, ValidateNested: false}, []string{"Prop name `enabled` doesn’t match rule `^(is|has)[A-Z]([A-Za-z0-9]?)+`"}},
		{"interface valid name", "interface Props { isEnabled: boolean }\nfunction C(p: Props) { return <div/>; }", BooleanPropNamingOptions{Rule: "^(is|has)[A-Z]([A-Za-z0-9]?)+", PropTypeNames: []string{"bool"}, ValidateNested: false}, []string{}},
		{"inline type literal", "function C(p: { enabled: boolean }) { return <div/>; }", BooleanPropNamingOptions{Rule: "^(is|has)[A-Z]([A-Za-z0-9]?)+", PropTypeNames: []string{"bool"}, ValidateNested: false}, []string{"Prop name `enabled` doesn’t match rule `^(is|has)[A-Z]([A-Za-z0-9]?)+`"}},
		{"type alias reference", "type Props = { enabled: boolean };\nfunction C(p: Props) { return <div/>; }", BooleanPropNamingOptions{Rule: "^(is|has)[A-Z]([A-Za-z0-9]?)+", PropTypeNames: []string{"bool"}, ValidateNested: false}, []string{"Prop name `enabled` doesn’t match rule `^(is|has)[A-Z]([A-Za-z0-9]?)+`"}},
		{"parenthesized intersection", "type Props = ({ a: boolean } & { b: boolean });\nfunction C(p: Props) { return <div/>; }", BooleanPropNamingOptions{Rule: "^(is|has)[A-Z]([A-Za-z0-9]?)+", PropTypeNames: []string{"bool"}, ValidateNested: false}, []string{"Prop name `a` doesn’t match rule `^(is|has)[A-Z]([A-Za-z0-9]?)+`", "Prop name `b` doesn’t match rule `^(is|has)[A-Z]([A-Za-z0-9]?)+`"}},
		{"bare intersection", "type Props = { a: boolean } & { b: boolean };\nfunction C(p: Props) { return <div/>; }", BooleanPropNamingOptions{Rule: "^(is|has)[A-Z]([A-Za-z0-9]?)+", PropTypeNames: []string{"bool"}, ValidateNested: false}, []string{"Prop name `a` doesn’t match rule `^(is|has)[A-Z]([A-Za-z0-9]?)+`", "Prop name `b` doesn’t match rule `^(is|has)[A-Z]([A-Za-z0-9]?)+`"}},
		{"union", "type Props = { a: boolean } | { b: boolean };\nfunction C(p: Props) { return <div/>; }", BooleanPropNamingOptions{Rule: "^(is|has)[A-Z]([A-Za-z0-9]?)+", PropTypeNames: []string{"bool"}, ValidateNested: false}, []string{"Prop name `a` doesn’t match rule `^(is|has)[A-Z]([A-Za-z0-9]?)+`", "Prop name `b` doesn’t match rule `^(is|has)[A-Z]([A-Za-z0-9]?)+`"}},
		{"optional boolean", "interface Props { enabled?: boolean }\nfunction C(p: Props) { return <div/>; }", BooleanPropNamingOptions{Rule: "^(is|has)[A-Z]([A-Za-z0-9]?)+", PropTypeNames: []string{"bool"}, ValidateNested: false}, []string{"Prop name `enabled` doesn’t match rule `^(is|has)[A-Z]([A-Za-z0-9]?)+`"}},
		{"boolean or undefined", "interface Props { enabled: boolean | undefined }\nfunction C(p: Props) { return <div/>; }", BooleanPropNamingOptions{Rule: "^(is|has)[A-Z]([A-Za-z0-9]?)+", PropTypeNames: []string{"bool"}, ValidateNested: false}, []string{}},
		{"non-boolean member", "interface Props { enabled: string }\nfunction C(p: Props) { return <div/>; }", BooleanPropNamingOptions{Rule: "^(is|has)[A-Z]([A-Za-z0-9]?)+", PropTypeNames: []string{"bool"}, ValidateNested: false}, []string{}},
		{"arrow component interface", "interface Props { enabled: boolean }\nconst C = (p: Props) => <div/>;", BooleanPropNamingOptions{Rule: "^(is|has)[A-Z]([A-Za-z0-9]?)+", PropTypeNames: []string{"bool"}, ValidateNested: false}, []string{"Prop name `enabled` doesn’t match rule `^(is|has)[A-Z]([A-Za-z0-9]?)+`"}},
		{"self-referential interface", "interface Props { enabled: boolean; self: Props }\nfunction C(p: Props) { return <div/>; }", BooleanPropNamingOptions{Rule: "^(is|has)[A-Z]([A-Za-z0-9]?)+", PropTypeNames: []string{"bool"}, ValidateNested: false}, []string{"Prop name `enabled` doesn’t match rule `^(is|has)[A-Z]([A-Za-z0-9]?)+`"}},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTypedWithOptions(t, BooleanPropNaming, booleanPropNamingFile,
				testCase.sourceText, testCase.options)
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

// TestDecodeBooleanPropNamingOptions routes configuration through the rule's own decoder.
//
// Every table above builds the options struct directly, which leaves the decoder untested: two
// mutations dropping its fallbacks survived the entire suite until this existed. Those two lines
// are the ones with no upstream counterpart to check them against, and both defaults are non-zero,
// so the generic `rule.DecodeOptionsInto` would silence the rule rather than merely narrow it.
//
// The empty-pattern row is the reason the `rule` fallback cannot be skipped: Go's zero value for a
// string is "", and an empty pattern matches every name, so a rule decoded that way reports nothing
// while looking configured.
func TestDecodeBooleanPropNamingOptions(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name              string
		raw               string
		wantRule          string
		wantPropTypeNames []string
		wantNested        bool
		wantMessage       string
		wantErr           bool
	}{
		{
			// Empty input is a rule configured as a bare `"error"`. The decoder REFUSES it, which
			// is what gives `RequiresOptions` its effect: that flag only changes behaviour when the
			// decoder errors. A decoder returning defaults here would make the flag inert and the
			// rule would run tree-wide on a configuration upstream treats as off.
			name: "empty input is refused, which is what makes RequiresOptions work",
			raw:  "", wantErr: true,
		},
		{
			name: "empty object resolves to the documented defaults",
			raw:  `{}`, wantRule: DefaultBooleanPropNamingRulePattern,
			wantPropTypeNames: []string{"bool"},
		},
		{
			name: "an explicit pattern wins",
			raw:  `{"rule": "^is[A-Z]"}`, wantRule: "^is[A-Z]",
			wantPropTypeNames: []string{"bool"},
		},
		{
			name:              "explicit propTypeNames replace the default",
			raw:               `{"propTypeNames": ["mutuallyExclusiveTrueProps"]}`,
			wantRule:          DefaultBooleanPropNamingRulePattern,
			wantPropTypeNames: []string{"mutuallyExclusiveTrueProps"},
		},
		{
			name: "an explicitly empty propTypeNames list falls back, matching upstream's or",
			raw:  `{"propTypeNames": []}`, wantRule: DefaultBooleanPropNamingRulePattern,
			wantPropTypeNames: []string{"bool"},
		},
		{
			name: "validateNested", raw: `{"validateNested": true}`,
			wantRule: DefaultBooleanPropNamingRulePattern, wantPropTypeNames: []string{"bool"},
			wantNested: true,
		},
		{
			name: "a custom message", raw: `{"message": "bad {{ propName }}"}`,
			wantRule: DefaultBooleanPropNamingRulePattern, wantPropTypeNames: []string{"bool"},
			wantMessage: "bad {{ propName }}",
		},
		{name: "an uncompilable pattern is an error", raw: `{"rule": "^(["}`, wantErr: true},
		{name: "a wrong type is an error", raw: `{"validateNested": "yes"}`, wantErr: true},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			decoded, err := DecodeBooleanPropNamingOptions([]byte(testCase.raw))
			if testCase.wantErr {
				if err == nil {
					t.Fatalf("want an error, got %#v", decoded)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			options, isOptions := decoded.(BooleanPropNamingOptions)
			if !isOptions {
				t.Fatalf("decoder returned %T, want BooleanPropNamingOptions", decoded)
			}
			if options.Rule != testCase.wantRule {
				t.Errorf("Rule is %q, want %q", options.Rule, testCase.wantRule)
			}
			if len(options.PropTypeNames) != len(testCase.wantPropTypeNames) {
				t.Fatalf("PropTypeNames is %v, want %v", options.PropTypeNames, testCase.wantPropTypeNames)
			}
			for index, wanted := range testCase.wantPropTypeNames {
				if options.PropTypeNames[index] != wanted {
					t.Errorf("PropTypeNames[%d] is %q, want %q", index, options.PropTypeNames[index], wanted)
				}
			}
			if options.ValidateNested != testCase.wantNested {
				t.Errorf("ValidateNested is %v, want %v", options.ValidateNested, testCase.wantNested)
			}
			if options.Message != testCase.wantMessage {
				t.Errorf("Message is %q, want %q", options.Message, testCase.wantMessage)
			}
		})
	}
}

// TestBooleanPropNamingDeclinesWithoutOptions pins the RequiresOptions contract.
//
// Upstream reads its pattern from `context.options[0]` and ESLint fills a schema default only into
// an options object that is PRESENT. Measured on the installed build with one violating input three
// ways: no options at all reports zero, `{}` reports one, an explicit pattern reports one. So the
// rule is entirely inert when configured as a bare `"error"`.
//
// This rule is registered with `RequiresOptions` so the config layer refuses that state rather than
// reproducing it. The harness can still reach `Run` with nil, and declining there is what keeps the
// rule from inventing a default nobody configured.
func TestBooleanPropNamingDeclinesWithoutOptions(t *testing.T) {
	t.Parallel()

	sourceText := "class C extends React.Component { render(){return <div/>;} }\nC.propTypes = { enabled: PropTypes.bool };"

	// The control. With options this input reports, so the zero below means the decline fired
	// rather than that the input was uninteresting.
	configured := rule_testing.RunTypedWithOptions(t, BooleanPropNaming, booleanPropNamingFile,
		sourceText, DefaultBooleanPropNamingOptions())
	rule_testing.ExpectFindings(t, configured, "patternMismatch")

	unconfigured := rule_testing.RunTypedWithOptions(t, BooleanPropNaming, booleanPropNamingFile,
		sourceText, nil)
	rule_testing.ExpectClean(t, unconfigured)
}

// TestBooleanPropNamingRequiresTheTypedHarness asserts the rule declines rather than reports with
// no checker.
//
// `NeedsTypeChecker` governs the registration path only; a rule.Context can be built by hand with a
// nil checker, and `TestNoRegisteredRuleCrashesOnAbsentOptionalNodes` does exactly that. The guard
// lives in Run rather than in the listener so the file is declined once instead of once per node.
func TestBooleanPropNamingRequiresTheTypedHarness(t *testing.T) {
	t.Parallel()

	sourceText := "class C extends React.Component { render(){return <div/>;} }\nC.propTypes = { enabled: PropTypes.bool };"

	typed := rule_testing.RunTypedWithOptions(t, BooleanPropNaming, booleanPropNamingFile,
		sourceText, DefaultBooleanPropNamingOptions())
	rule_testing.ExpectFindings(t, typed, "patternMismatch")

	untyped := rule_testing.RunWithOptions(t, BooleanPropNaming, booleanPropNamingFile,
		sourceText, DefaultBooleanPropNamingOptions())
	rule_testing.ExpectClean(t, untyped)
}

// TestBooleanPropNamingAnchorsOnTheWholeProperty pins where each finding points.
//
// `ExpectFindings` asserts ids and counts and can see none of this. Upstream reports on the
// property node, so the span covers the name, the colon and the declared type rather than the name
// alone. Measured against the installed build on 2026-08-28 by slicing its reported range.
func TestBooleanPropNamingAnchorsOnTheWholeProperty(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		wantSpans  []string
	}{
		{
			name:       "a propTypes property, including its declared type",
			sourceText: "class C extends React.Component { render(){return <div/>;} }\nC.propTypes = { enabled: PropTypes.bool };",
			wantSpans:  []string{"enabled: PropTypes.bool"},
		},
		{
			name:       "an interface member",
			sourceText: "interface Props { enabled: boolean }\nfunction C(p: Props) { return <div/>; }",
			wantSpans:  []string{"enabled: boolean"},
		},
		{
			name:       "two props report in source order",
			sourceText: "class C extends React.Component { render(){return <div/>;} }\nC.propTypes = { enabled: PropTypes.bool, active: PropTypes.bool };",
			wantSpans:  []string{"enabled: PropTypes.bool", "active: PropTypes.bool"},
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTypedWithOptions(t, BooleanPropNaming, booleanPropNamingFile,
				testCase.sourceText, DefaultBooleanPropNamingOptions())
			if len(result.Diagnostics) != len(testCase.wantSpans) {
				t.Fatalf("want %d findings, got %d", len(testCase.wantSpans), len(result.Diagnostics))
			}
			for index, diagnostic := range result.Diagnostics {
				reported := testCase.sourceText[diagnostic.Range.Pos():diagnostic.Range.End()]
				if reported != testCase.wantSpans[index] {
					t.Errorf("finding %d points at %q, want %q",
						index, reported, testCase.wantSpans[index])
				}
			}
		})
	}
}

// TestBooleanPropNamingMessageUsesUpstreamsApostrophe pins the exact rendered text.
//
// The message carries a U+2019 right single quotation mark in "doesn't", which is upstream's own
// byte. The rule file writes it as the escape ’ rather than as a literal so it survives review
// and cannot be mistaken for the editor damage this project has twice cleaned up; this test is what
// proves the escape produces the same character upstream emits.
//
// Asserted against a literal built here from the same escape rather than against the rule's own
// constant, because a message-text mutation would move both together.
func TestBooleanPropNamingMessageUsesUpstreamsApostrophe(t *testing.T) {
	t.Parallel()

	sourceText := "class C extends React.Component { render(){return <div/>;} }\nC.propTypes = { enabled: PropTypes.bool };"
	result := rule_testing.RunTypedWithOptions(t, BooleanPropNaming, booleanPropNamingFile,
		sourceText, DefaultBooleanPropNamingOptions())
	rule_testing.ExpectFindings(t, result, "patternMismatch")

	want := "Prop name `enabled` doesn’t match rule `^(is|has)[A-Z]([A-Za-z0-9]?)+`"
	if result.Diagnostics[0].Message.Description != want {
		t.Errorf("renders\n  got  %q\n  want %q", result.Diagnostics[0].Message.Description, want)
	}
	// A straight apostrophe would be a different string, and this is the assertion that says so.
	straight := "Prop name `enabled` doesn't match rule `^(is|has)[A-Z]([A-Za-z0-9]?)+`"
	if result.Diagnostics[0].Message.Description == straight {
		t.Error("rendered a straight apostrophe, which diverges from upstream's own message")
	}
}

// TestBooleanPropNamingDoesNotPanicOnNameShapesThatCannotBeRead pins a crash class.
//
// `ast.Node.Text()` panics on several node shapes, a binding pattern among them, and this rule
// reads a NAME off every prop key and every parameter it inspects. The walk recovers per FILE
// rather than per rule, so one such call would cost every rule in this package its verdict on that
// file while the run still printed a plausible summary.
//
// Every row below puts an unreadable or unusual name somewhere the rule looks. Reaching the end of
// each subtest at all is the assertion, because no findings assertion can see a panic.
//
// The row set was checked against a control: making the rule read a raw name without a kind guard
// makes the destructured-parameter rows fail, so these are not passing vacuously.
func TestBooleanPropNamingDoesNotPanicOnNameShapesThatCannotBeRead(t *testing.T) {
	t.Parallel()

	sources := []string{
		// A destructured parameter, whose name node is a binding pattern.
		"interface Props { enabled: boolean }\nfunction C({ enabled }: Props) { return <div/>; }",
		// An array-pattern parameter.
		"function C([first]: boolean[]) { return <div/>; }",
		// A rest parameter, whose name IS an identifier and is the control for the two above.
		"function C(...rest: boolean[]) { return <div/>; }",
		// A computed propTypes key.
		"class C extends React.Component { render(){return <div/>;} }\nC.propTypes = { [key]: PropTypes.bool };",
		// A computed key whose inner expression is not an identifier.
		"class C extends React.Component { render(){return <div/>;} }\nC.propTypes = { [a.b]: PropTypes.bool };",
		// A quoted and a numeric key.
		"class C extends React.Component { render(){return <div/>;} }\nC.propTypes = { 'enabled': PropTypes.bool, 1: PropTypes.bool };",
		// A computed member name in a props interface.
		"interface Props { [key: string]: boolean }\nfunction C(p: Props) { return <div/>; }",
		// A spread inside propTypes, which carries no key at all.
		"class C extends React.Component { render(){return <div/>;} }\nC.propTypes = { ...base, enabled: PropTypes.bool };",
		// A parameter with no type annotation at all.
		"function C(p) { return <div/>; }",
	}

	for _, sourceText := range sources {
		t.Run(sourceText, func(t *testing.T) {
			t.Parallel()
			// Reaching the next line without a panic is the assertion.
			rule_testing.RunTypedWithOptions(t, BooleanPropNaming, booleanPropNamingFile,
				sourceText, DefaultBooleanPropNamingOptions())
		})
	}
}
