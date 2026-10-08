package javascript

import "testing"

func TestViewCallableShapeNode(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, setup, out, found string }{
		{"valid", "", "8\n", ""},
		{"wrong-arity", "recorded = {...recorded, parameters: []};", "", "function with arity 0"},
		{"wrong-parameter", "recorded = {...recorded, parameters: [2]};", "", "function with incompatible parameter representations"},
		{"wrong-result", "recorded = {...recorded, result: 2};", "", "function with incompatible result representation"},
		{"unknown-result", "expected.result = 0; recorded = {...recorded, result: 0};", "", "function with unknown signature"},
		{"unknown-parameter", "expected.parameters[0] = 0; recorded = {...recorded, parameters: [0]};", "", "function with incompatible parameter representations"},
		{"unknown-signature", "recorded = undefined;", "", "function with unknown signature"},
		{"number", "value = 7;", "", "number"},
		{"optional-absent", "value = undefined; optional = true;", "undefined\n", ""},
		{"required-absent", "value = undefined;", "", "undefined"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			source := viewTestRuntime + viewCallableShapeRuntime + `
const expected = {parameters: [1], result: 1, name: "function(number) -> number"};
let recorded = expected;
let value = new AdamicClosure((self, arguments_) => arguments_[0] + 1);
let optional = false;
` + probe.setup + `
const checked = adamicViewCallableShape(value, recorded, expected, "node.run", optional);
console.log(checked === undefined ? "undefined" : adamicCall(checked, [7]));
`
			exit, stderr := 0, ""
			if probe.found != "" {
				exit = 70
				stderr = "adamic: panic: field read failed: node.run expected function(number) -> number, found " + probe.found + "\n"
			}
			runViewNode(t, source, probe.out, stderr, exit)
		})
	}
	// Outside oracle for the positive call and optional field absence.
	runViewNode(t, `const object = {run: n => n + 1}; console.log(object.run(7));`, "8\n", "", 0)
	runViewNode(t, `const object = {}; console.log(object.run === undefined ? "undefined" : "present");`, "undefined\n", "", 0)
}
