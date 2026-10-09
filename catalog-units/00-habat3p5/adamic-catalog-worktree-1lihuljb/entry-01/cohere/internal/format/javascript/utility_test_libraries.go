package javascript

import "slices"

// utilities/test-libraries.js.

var testCallCalleePatterns = []string{
	"it",
	"it.only",
	"it.skip",
	"describe",
	"describe.only",
	"describe.skip",
	"test",
	"test.only",
	"test.skip",
	"test.fixme",
	"test.step",
	"test.describe",
	"test.describe.only",
	"test.describe.skip",
	"test.describe.fixme",
	"test.describe.parallel",
	"test.describe.parallel.only",
	"test.describe.serial",
	"test.describe.serial.only",
	"skip",
	"xit",
	"xdescribe",
	"xtest",
	"fit",
	"fdescribe",
	"ftest",
}

func isTestCallCallee(node Node) bool {
	return isNodeMatches(node, testCallCalleePatterns)
}

func isUnitTestSetupIdentifier(node Node) bool {
	return node.Is("Identifier") &&
		(node.String("name") == "beforeEach" ||
			node.String("name") == "beforeAll" ||
			node.String("name") == "afterEach" ||
			node.String("name") == "afterAll")
}

// isAngularTestWrapper is upstream's isAngularTestWrapper.
//
// Note: `inject` is used in AngularJS 1.x, `async` and `fakeAsync` in
// Angular 2+, although `async` is deprecated and replaced by `waitForAsync`
// since Angular 12.
func isAngularTestWrapper(node Node) bool {
	return isCallExpression(node) &&
		node.Child("callee").Is("Identifier") &&
		slices.Contains([]string{"async", "inject", "fakeAsync", "waitForAsync"}, node.Child("callee").String("name"))
}

func isFunctionOrArrowExpressionWithBody(node Node) bool {
	return node.Is("FunctionExpression") ||
		node.Is("ArrowFunctionExpression") &&
			node.Child("body").Is("BlockStatement")
}

// isTestCall is upstream's isTestCall, eg; `describe("some string", (done) => {})`. Upstream's
// one-argument call isTestCall(parent) passes a nil parent here.
func isTestCall(node Node, parent Node) bool {
	if !node.Is("CallExpression") || node.Truthy("optional") {
		return false
	}

	args := getCallArguments(node)

	if len(args) == 1 {
		if isAngularTestWrapper(node) && isTestCall(parent, nil) {
			return isFunctionOrArrowExpression(args[0])
		}

		if isUnitTestSetupIdentifier(node.Child("callee")) {
			return isAngularTestWrapper(args[0])
		}
	} else if (len(args) == 2 || len(args) == 3) &&
		(args[0].Is("TemplateLiteral") || isStringLiteral(args[0])) &&
		isTestCallCallee(node.Child("callee")) {
		// it("name", () => { ... }, 2500)
		if len(args) > 2 && args[2] != nil && !isNumericLiteral(args[2]) {
			return false
		}
		var isTestFunction bool
		if len(args) == 2 {
			isTestFunction = isFunctionOrArrowExpression(args[1])
		} else {
			isTestFunction = isFunctionOrArrowExpressionWithBody(args[1]) &&
				len(getFunctionParameters(args[1])) <= 1
		}
		return isTestFunction || isAngularTestWrapper(args[1])
	}
	return false
}
