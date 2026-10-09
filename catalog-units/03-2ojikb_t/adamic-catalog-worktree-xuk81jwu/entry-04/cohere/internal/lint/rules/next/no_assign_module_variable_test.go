package next

import (
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

const assignModuleFile = "/repository/source/Thing.ts"

func TestNoAssignModuleVariableFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"a let declaration", "let module = {};\n"},
		{"a const declaration", "const module = {};\n"},
		{"a var declaration", "var module = {};\n"},
		// The declarator that matters is not always the first one, and a rule that inspected only
		// declarations[0] would pass every other fixture here while missing this.
		{"the second name in a list", "let a = 1, module = {};\n"},
		{"declared without an initializer", "let module;\n"},
		// Inside a function the statement is nested rather than top level, which is the shape a
		// listener keyed to source-file children rather than to the node kind would miss.
		{"inside a function body", "export function run() {\n    let module = {};\n}\n"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t, rule_testing.Run(t, NoAssignModuleVariable, assignModuleFile, testCase.sourceText),
				"noAssignModuleVariable")
		})
	}
}

func TestNoAssignModuleVariableStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		// The decision this rule makes is "does this declaration bind the name `module`", and the
		// way to get it wrong is to match the text `module` anywhere it appears. Every case below
		// contains the word and none of them binds it.
		{"a name that merely starts with it", "const moduleName = 'x';\nexport const Name = moduleName;\n"},
		{"a name that merely contains it", "const myModule = {};\nexport const Thing = myModule;\n"},
		{"a property called module", "export const Settings = { module: true };\n"},
		{"a string containing the word", "export const Message = 'module';\n"},
		{"an import binding", "import module from 'node:module';\nexport const Loaded = module;\n"},
		// A destructured binding never introduces the name by spelling it, so a rule that read the
		// name off a binding pattern rather than off an identifier would fire here wrongly.
		{"a destructured property named module", "const { module: renamed } = imported;\nexport const Value = renamed;\n"},
		{"a function parameter named module", "export function run(module: unknown) {\n    return module;\n}\n"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.Run(t, NoAssignModuleVariable, assignModuleFile, testCase.sourceText))
		})
	}
}
