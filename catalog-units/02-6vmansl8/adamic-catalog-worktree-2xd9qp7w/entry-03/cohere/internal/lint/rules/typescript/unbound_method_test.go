package typescript

import (
	"encoding/json"
	"strconv"
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// unboundMethodFile names the fixture file.
//
// It ends in .ts rather than .tsx because several upstream cases use an angle-bracket type
// assertion, which is not parseable in a .tsx file at all.
const unboundMethodFile = "/repository/source/Unbound.ts"

// unboundMethodCaseName numbers a row so a failure names which one.
func unboundMethodCaseName(index int) string {
	return "case" + strconv.Itoa(index)
}

// asTheTypedUnboundMethodHarnessWroteIt transforms a fixture the way RunTyped transforms its input.
//
// rule_testing/program.go writes each fixture as strings.TrimSpace(contents)+"\n", so the file on
// disk is offset from the string in the Go literal above it. Nearly every case in this corpus carries
// a leading newline and trailing indentation, so slicing the literal to check a span reports a result
// shifted at both ends, which reads exactly like an off-by-one in the rule.
func asTheTypedUnboundMethodHarnessWroteIt(text string) string {
	return strings.TrimSpace(text) + "\n"
}

// decodeUnboundMethodOptionsForTest routes a fixture through the rule's own decoder.
//
// Only one upstream case carries options, so this is the single line in the file that can tell a
// decoder reading its input from one returning a zero value. The key defaults to false, so a decoder
// that dropped it entirely would still satisfy the other two hundred and ten rows.
func decodeUnboundMethodOptionsForTest(t *testing.T, raw string) any {
	t.Helper()
	decoded, err := DecodeUnboundMethodOptions(json.RawMessage(raw))
	if err != nil {
		t.Fatalf("the decoder refused %s: %v", raw, err)
	}
	return decoded
}

// runUnboundMethod runs one case, with or without options.
func runUnboundMethod(t *testing.T, sourceText string, options string) rule_testing.Result {
	t.Helper()
	if options == "" {
		return rule_testing.RunTyped(t, UnboundMethod, unboundMethodFile, sourceText)
	}
	return rule_testing.RunTypedWithOptions(t, UnboundMethod, unboundMethodFile, sourceText,
		decodeUnboundMethodOptionsForTest(t, options))
}

// TestUnboundMethodStaysSilentOnUpstreamPassCases is the imported clean corpus, verbatim.
//
// All one hundred and forty-four of upstream's passing inputs, extracted from the clone's test file
// by parsing it with the TypeScript compiler rather than by reading it, then byte verified. Every one
// was replayed through the installed 8.x build with a real type checker, one program per case, and
// all one hundred and forty-four reported nothing.
//
// This list is the rule. A method reference that is safe is the entire false-positive surface, and
// upstream has thought about far more of them than a port would invent: every syntactic position
// where a reference is immediately consumed rather than detached, a `this: void` annotation, an
// arrow-function property, a natively bound global, and a long tail of assertion and chaining
// wrappers that have to be seen through.
func TestUnboundMethodStaysSilentOnUpstreamPassCases(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText string
		options    string
	}{
		{sourceText: "Promise.resolve().then(console.log);", options: ""},
		{sourceText: "['1', '2', '3'].map(Number.parseInt);", options: ""},
		{sourceText: "[5.2, 7.1, 3.6].map(Math.floor);", options: ""},
		{sourceText: "\nconst foo = Number;\n['1', '2', '3'].map(foo.parseInt);\n    ", options: ""},
		{sourceText: "\nconst foo = Math;\n[5.2, 7.1, 3.6].map(foo.floor);\n    ", options: ""},
		{sourceText: "['1', '2', '3'].map(Number['floor']);", options: ""},
		{sourceText: "const x = console.log;", options: ""},
		{sourceText: "const x = Object.defineProperty;", options: ""},
		{sourceText: "\nconst foo = Object;\nconst x = foo.defineProperty;\n    ", options: ""},
		{sourceText: "const x = String.fromCharCode;", options: ""},
		{sourceText: "\nconst foo = String;\nconst x = foo.fromCharCode;\n    ", options: ""},
		{sourceText: "const x = RegExp.prototype;", options: ""},
		{sourceText: "const x = Symbol.keyFor;", options: ""},
		{sourceText: "\nconst foo = Symbol;\nconst x = foo.keyFor;\n    ", options: ""},
		{sourceText: "const x = Array.isArray;", options: ""},
		{sourceText: "\nconst foo = Array;\nconst x = foo.isArray;\n    ", options: ""},
		{sourceText: "\nclass Foo extends Array {}\nconst x = Foo.isArray;\n    ", options: ""},
		{sourceText: "const x = Proxy.revocable;", options: ""},
		{sourceText: "\nconst foo = Proxy;\nconst x = foo.revocable;\n    ", options: ""},
		{sourceText: "const x = Date.parse;", options: ""},
		{sourceText: "\nconst foo = Date;\nconst x = foo.parse;\n    ", options: ""},
		{sourceText: "const x = Atomics.load;", options: ""},
		{sourceText: "\nconst foo = Atomics;\nconst x = foo.load;\n    ", options: ""},
		{sourceText: "const x = Reflect.deleteProperty;", options: ""},
		{sourceText: "const x = JSON.stringify;", options: ""},
		{sourceText: "\nconst foo = JSON;\nconst x = foo.stringify;\n    ", options: ""},
		{sourceText: "\nconst o = {\n  f: function (this: void) {},\n};\nconst f = o.f;\n    ", options: ""},
		{sourceText: "\nconst { alert } = window;\n    ", options: ""},
		{sourceText: "\nlet b = window.blur;\n    ", options: ""},
		{sourceText: "\nfunction foo() {}\nconst fooObject = { foo };\nconst { foo: bar } = fooObject;\n    ", options: ""},
		{sourceText: "\nclass ContainsMethods {\n  bound?: () => void;\n  unbound?(): void;\n\n  static boundStatic?: () => void;\n  static unboundStatic?(): void;\n}\n\nlet instance = new ContainsMethods();\n\nconst arith = {\n  double(this: void, x: number): number {\n    return x * 2;\n  },\n};\n\ninstance.bound();\n    ", options: ""},
		{sourceText: "\nclass ContainsMethods {\n  bound?: () => void;\n  unbound?(): void;\n\n  static boundStatic?: () => void;\n  static unboundStatic?(): void;\n}\n\nlet instance = new ContainsMethods();\n\nconst arith = {\n  double(this: void, x: number): number {\n    return x * 2;\n  },\n};\n\ninstance.unbound();\n    ", options: ""},
		{sourceText: "\nclass ContainsMethods {\n  bound?: () => void;\n  unbound?(): void;\n\n  static boundStatic?: () => void;\n  static unboundStatic?(): void;\n}\n\nlet instance = new ContainsMethods();\n\nconst arith = {\n  double(this: void, x: number): number {\n    return x * 2;\n  },\n};\n\nContainsMethods.boundStatic();\n    ", options: ""},
		{sourceText: "\nclass ContainsMethods {\n  bound?: () => void;\n  unbound?(): void;\n\n  static boundStatic?: () => void;\n  static unboundStatic?(): void;\n}\n\nlet instance = new ContainsMethods();\n\nconst arith = {\n  double(this: void, x: number): number {\n    return x * 2;\n  },\n};\n\nContainsMethods.unboundStatic();\n    ", options: ""},
		{sourceText: "\nclass ContainsMethods {\n  bound?: () => void;\n  unbound?(): void;\n\n  static boundStatic?: () => void;\n  static unboundStatic?(): void;\n}\n\nlet instance = new ContainsMethods();\n\nconst arith = {\n  double(this: void, x: number): number {\n    return x * 2;\n  },\n};\n\nconst bound = instance.bound;\n    ", options: ""},
		{sourceText: "\nclass ContainsMethods {\n  bound?: () => void;\n  unbound?(): void;\n\n  static boundStatic?: () => void;\n  static unboundStatic?(): void;\n}\n\nlet instance = new ContainsMethods();\n\nconst arith = {\n  double(this: void, x: number): number {\n    return x * 2;\n  },\n};\n\nconst boundStatic = ContainsMethods;\n    ", options: ""},
		{sourceText: "\nclass ContainsMethods {\n  bound?: () => void;\n  unbound?(): void;\n\n  static boundStatic?: () => void;\n  static unboundStatic?(): void;\n}\n\nlet instance = new ContainsMethods();\n\nconst arith = {\n  double(this: void, x: number): number {\n    return x * 2;\n  },\n};\n\nconst { bound } = instance;\n    ", options: ""},
		{sourceText: "\nclass ContainsMethods {\n  bound?: () => void;\n  unbound?(): void;\n\n  static boundStatic?: () => void;\n  static unboundStatic?(): void;\n}\n\nlet instance = new ContainsMethods();\n\nconst arith = {\n  double(this: void, x: number): number {\n    return x * 2;\n  },\n};\n\nconst { boundStatic } = ContainsMethods;\n    ", options: ""},
		{sourceText: "\nclass ContainsMethods {\n  bound?: () => void;\n}\n\nlet instance = new ContainsMethods();\n\nconst arith = {\n  double(this: void, x: number): number {\n    return x * 2;\n  },\n};\n\ninstance.bound();\n    ", options: ""},
		{sourceText: "\nclass ContainsMethods {\n  unbound?(): void;\n}\n\nlet instance = new ContainsMethods();\n\nconst arith = {\n  double(this: void, x: number): number {\n    return x * 2;\n  },\n};\n\ninstance.unbound();\n    ", options: ""},
		{sourceText: "\nclass ContainsMethods {\n  static boundStatic?: () => void;\n}\n\nlet instance = new ContainsMethods();\n\nconst arith = {\n  double(this: void, x: number): number {\n    return x * 2;\n  },\n};\n\nContainsMethods.boundStatic();\n    ", options: ""},
		{sourceText: "\nclass ContainsMethods {\n  static unboundStatic?(): void;\n}\n\nlet instance = new ContainsMethods();\n\nconst arith = {\n  double(this: void, x: number): number {\n    return x * 2;\n  },\n};\n\nContainsMethods.unboundStatic();\n    ", options: ""},
		{sourceText: "\nclass ContainsMethods {\n  bound?: () => void;\n  unbound?(): void;\n\n  static boundStatic?: () => void;\n  static unboundStatic?(): void;\n}\n\nlet instance = new ContainsMethods();\n\nconst arith = {\n  double(this: void, x: number): number {\n    return x * 2;\n  },\n};\n\ninstance.bound``;\n    ", options: ""},
		{sourceText: "\nclass ContainsMethods {\n  bound?: () => void;\n  unbound?(): void;\n\n  static boundStatic?: () => void;\n  static unboundStatic?(): void;\n}\n\nlet instance = new ContainsMethods();\n\nconst arith = {\n  double(this: void, x: number): number {\n    return x * 2;\n  },\n};\n\ninstance.unbound``;\n    ", options: ""},
		{sourceText: "\nclass ContainsMethods {\n  bound?: () => void;\n  unbound?(): void;\n\n  static boundStatic?: () => void;\n  static unboundStatic?(): void;\n}\n\nlet instance = new ContainsMethods();\n\nconst arith = {\n  double(this: void, x: number): number {\n    return x * 2;\n  },\n};\n\nif (instance.bound) {\n}\n    ", options: ""},
		{sourceText: "\nclass ContainsMethods {\n  bound?: () => void;\n  unbound?(): void;\n\n  static boundStatic?: () => void;\n  static unboundStatic?(): void;\n}\n\nlet instance = new ContainsMethods();\n\nconst arith = {\n  double(this: void, x: number): number {\n    return x * 2;\n  },\n};\n\nif (instance.unbound) {\n}\n    ", options: ""},
		{sourceText: "\nclass ContainsMethods {\n  bound?: () => void;\n  unbound?(): void;\n\n  static boundStatic?: () => void;\n  static unboundStatic?(): void;\n}\n\nlet instance = new ContainsMethods();\n\nconst arith = {\n  double(this: void, x: number): number {\n    return x * 2;\n  },\n};\n\nif (instance.bound !== undefined) {\n}\n    ", options: ""},
		{sourceText: "\nclass ContainsMethods {\n  bound?: () => void;\n  unbound?(): void;\n\n  static boundStatic?: () => void;\n  static unboundStatic?(): void;\n}\n\nlet instance = new ContainsMethods();\n\nconst arith = {\n  double(this: void, x: number): number {\n    return x * 2;\n  },\n};\n\nif (instance.unbound !== undefined) {\n}\n    ", options: ""},
		{sourceText: "\nclass ContainsMethods {\n  bound?: () => void;\n  unbound?(): void;\n\n  static boundStatic?: () => void;\n  static unboundStatic?(): void;\n}\n\nlet instance = new ContainsMethods();\n\nconst arith = {\n  double(this: void, x: number): number {\n    return x * 2;\n  },\n};\n\nif (ContainsMethods.boundStatic) {\n}\n    ", options: ""},
		{sourceText: "\nclass ContainsMethods {\n  bound?: () => void;\n  unbound?(): void;\n\n  static boundStatic?: () => void;\n  static unboundStatic?(): void;\n}\n\nlet instance = new ContainsMethods();\n\nconst arith = {\n  double(this: void, x: number): number {\n    return x * 2;\n  },\n};\n\nif (ContainsMethods.unboundStatic) {\n}\n    ", options: ""},
		{sourceText: "\nclass ContainsMethods {\n  bound?: () => void;\n  unbound?(): void;\n\n  static boundStatic?: () => void;\n  static unboundStatic?(): void;\n}\n\nlet instance = new ContainsMethods();\n\nconst arith = {\n  double(this: void, x: number): number {\n    return x * 2;\n  },\n};\n\nif (ContainsMethods.boundStatic !== undefined) {\n}\n    ", options: ""},
		{sourceText: "\nclass ContainsMethods {\n  bound?: () => void;\n  unbound?(): void;\n\n  static boundStatic?: () => void;\n  static unboundStatic?(): void;\n}\n\nlet instance = new ContainsMethods();\n\nconst arith = {\n  double(this: void, x: number): number {\n    return x * 2;\n  },\n};\n\nif (ContainsMethods.unboundStatic !== undefined) {\n}\n    ", options: ""},
		{sourceText: "\nclass ContainsMethods {\n  bound?: () => void;\n  unbound?(): void;\n\n  static boundStatic?: () => void;\n  static unboundStatic?(): void;\n}\n\nlet instance = new ContainsMethods();\n\nconst arith = {\n  double(this: void, x: number): number {\n    return x * 2;\n  },\n};\n\nif (ContainsMethods.boundStatic && instance) {\n}\n    ", options: ""},
		{sourceText: "\nclass ContainsMethods {\n  bound?: () => void;\n  unbound?(): void;\n\n  static boundStatic?: () => void;\n  static unboundStatic?(): void;\n}\n\nlet instance = new ContainsMethods();\n\nconst arith = {\n  double(this: void, x: number): number {\n    return x * 2;\n  },\n};\n\nif (ContainsMethods.unboundStatic && instance) {\n}\n    ", options: ""},
		{sourceText: "\nclass ContainsMethods {\n  bound?: () => void;\n  unbound?(): void;\n\n  static boundStatic?: () => void;\n  static unboundStatic?(): void;\n}\n\nlet instance = new ContainsMethods();\n\nconst arith = {\n  double(this: void, x: number): number {\n    return x * 2;\n  },\n};\n\nif (instance.bound || instance) {\n}\n    ", options: ""},
		{sourceText: "\nclass ContainsMethods {\n  bound?: () => void;\n  unbound?(): void;\n\n  static boundStatic?: () => void;\n  static unboundStatic?(): void;\n}\n\nlet instance = new ContainsMethods();\n\nconst arith = {\n  double(this: void, x: number): number {\n    return x * 2;\n  },\n};\n\nif (instance.unbound || instance) {\n}\n    ", options: ""},
		{sourceText: "\nclass ContainsMethods {\n  bound?: () => void;\n  unbound?(): void;\n\n  static boundStatic?: () => void;\n  static unboundStatic?(): void;\n}\n\nlet instance = new ContainsMethods();\n\nconst arith = {\n  double(this: void, x: number): number {\n    return x * 2;\n  },\n};\n\n(ContainsMethods.unboundStatic && 0) || ContainsMethods;\n    ", options: ""},
		{sourceText: "\nclass ContainsMethods {\n  bound?: () => void;\n  unbound?(): void;\n\n  static boundStatic?: () => void;\n  static unboundStatic?(): void;\n}\n\nlet instance = new ContainsMethods();\n\nconst arith = {\n  double(this: void, x: number): number {\n    return x * 2;\n  },\n};\n\ninstance.bound || instance ? 1 : 0;\n    ", options: ""},
		{sourceText: "\nclass ContainsMethods {\n  bound?: () => void;\n  unbound?(): void;\n\n  static boundStatic?: () => void;\n  static unboundStatic?(): void;\n}\n\nlet instance = new ContainsMethods();\n\nconst arith = {\n  double(this: void, x: number): number {\n    return x * 2;\n  },\n};\n\ninstance.unbound || instance ? 1 : 0;\n    ", options: ""},
		{sourceText: "\nclass ContainsMethods {\n  bound?: () => void;\n  unbound?(): void;\n\n  static boundStatic?: () => void;\n  static unboundStatic?(): void;\n}\n\nlet instance = new ContainsMethods();\n\nconst arith = {\n  double(this: void, x: number): number {\n    return x * 2;\n  },\n};\n\nwhile (instance.bound) {}\n    ", options: ""},
		{sourceText: "\nclass ContainsMethods {\n  bound?: () => void;\n  unbound?(): void;\n\n  static boundStatic?: () => void;\n  static unboundStatic?(): void;\n}\n\nlet instance = new ContainsMethods();\n\nconst arith = {\n  double(this: void, x: number): number {\n    return x * 2;\n  },\n};\n\nwhile (instance.unbound) {}\n    ", options: ""},
		{sourceText: "\nclass ContainsMethods {\n  bound?: () => void;\n  unbound?(): void;\n\n  static boundStatic?: () => void;\n  static unboundStatic?(): void;\n}\n\nlet instance = new ContainsMethods();\n\nconst arith = {\n  double(this: void, x: number): number {\n    return x * 2;\n  },\n};\n\nwhile (instance.bound !== undefined) {}\n    ", options: ""},
		{sourceText: "\nclass ContainsMethods {\n  bound?: () => void;\n  unbound?(): void;\n\n  static boundStatic?: () => void;\n  static unboundStatic?(): void;\n}\n\nlet instance = new ContainsMethods();\n\nconst arith = {\n  double(this: void, x: number): number {\n    return x * 2;\n  },\n};\n\nwhile (instance.unbound !== undefined) {}\n    ", options: ""},
		{sourceText: "\nclass ContainsMethods {\n  bound?: () => void;\n  unbound?(): void;\n\n  static boundStatic?: () => void;\n  static unboundStatic?(): void;\n}\n\nlet instance = new ContainsMethods();\n\nconst arith = {\n  double(this: void, x: number): number {\n    return x * 2;\n  },\n};\n\nwhile (ContainsMethods.boundStatic) {}\n    ", options: ""},
		{sourceText: "\nclass ContainsMethods {\n  bound?: () => void;\n  unbound?(): void;\n\n  static boundStatic?: () => void;\n  static unboundStatic?(): void;\n}\n\nlet instance = new ContainsMethods();\n\nconst arith = {\n  double(this: void, x: number): number {\n    return x * 2;\n  },\n};\n\nwhile (ContainsMethods.unboundStatic) {}\n    ", options: ""},
		{sourceText: "\nclass ContainsMethods {\n  bound?: () => void;\n  unbound?(): void;\n\n  static boundStatic?: () => void;\n  static unboundStatic?(): void;\n}\n\nlet instance = new ContainsMethods();\n\nconst arith = {\n  double(this: void, x: number): number {\n    return x * 2;\n  },\n};\n\nwhile (ContainsMethods.boundStatic !== undefined) {}\n    ", options: ""},
		{sourceText: "\nclass ContainsMethods {\n  bound?: () => void;\n  unbound?(): void;\n\n  static boundStatic?: () => void;\n  static unboundStatic?(): void;\n}\n\nlet instance = new ContainsMethods();\n\nconst arith = {\n  double(this: void, x: number): number {\n    return x * 2;\n  },\n};\n\nwhile (ContainsMethods.unboundStatic !== undefined) {}\n    ", options: ""},
		{sourceText: "\nclass ContainsMethods {\n  bound?: () => void;\n  unbound?(): void;\n\n  static boundStatic?: () => void;\n  static unboundStatic?(): void;\n}\n\nlet instance = new ContainsMethods();\n\nconst arith = {\n  double(this: void, x: number): number {\n    return x * 2;\n  },\n};\n\ninstance.bound as any;\n    ", options: ""},
		{sourceText: "\nclass ContainsMethods {\n  bound?: () => void;\n  unbound?(): void;\n\n  static boundStatic?: () => void;\n  static unboundStatic?(): void;\n}\n\nlet instance = new ContainsMethods();\n\nconst arith = {\n  double(this: void, x: number): number {\n    return x * 2;\n  },\n};\n\nContainsMethods.boundStatic as any;\n    ", options: ""},
		{sourceText: "\nclass ContainsMethods {\n  bound?: () => void;\n  unbound?(): void;\n\n  static boundStatic?: () => void;\n  static unboundStatic?(): void;\n}\n\nlet instance = new ContainsMethods();\n\nconst arith = {\n  double(this: void, x: number): number {\n    return x * 2;\n  },\n};\n\ninstance.bound++;\n    ", options: ""},
		{sourceText: "\nclass ContainsMethods {\n  bound?: () => void;\n  unbound?(): void;\n\n  static boundStatic?: () => void;\n  static unboundStatic?(): void;\n}\n\nlet instance = new ContainsMethods();\n\nconst arith = {\n  double(this: void, x: number): number {\n    return x * 2;\n  },\n};\n\n+instance.bound;\n    ", options: ""},
		{sourceText: "\nclass ContainsMethods {\n  bound?: () => void;\n  unbound?(): void;\n\n  static boundStatic?: () => void;\n  static unboundStatic?(): void;\n}\n\nlet instance = new ContainsMethods();\n\nconst arith = {\n  double(this: void, x: number): number {\n    return x * 2;\n  },\n};\n\n++instance.bound;\n    ", options: ""},
		{sourceText: "\nclass ContainsMethods {\n  bound?: () => void;\n  unbound?(): void;\n\n  static boundStatic?: () => void;\n  static unboundStatic?(): void;\n}\n\nlet instance = new ContainsMethods();\n\nconst arith = {\n  double(this: void, x: number): number {\n    return x * 2;\n  },\n};\n\ninstance.bound--;\n    ", options: ""},
		{sourceText: "\nclass ContainsMethods {\n  bound?: () => void;\n  unbound?(): void;\n\n  static boundStatic?: () => void;\n  static unboundStatic?(): void;\n}\n\nlet instance = new ContainsMethods();\n\nconst arith = {\n  double(this: void, x: number): number {\n    return x * 2;\n  },\n};\n\n-instance.bound;\n    ", options: ""},
		{sourceText: "\nclass ContainsMethods {\n  bound?: () => void;\n  unbound?(): void;\n\n  static boundStatic?: () => void;\n  static unboundStatic?(): void;\n}\n\nlet instance = new ContainsMethods();\n\nconst arith = {\n  double(this: void, x: number): number {\n    return x * 2;\n  },\n};\n\n--instance.bound;\n    ", options: ""},
		{sourceText: "\nclass ContainsMethods {\n  bound?: () => void;\n  unbound?(): void;\n\n  static boundStatic?: () => void;\n  static unboundStatic?(): void;\n}\n\nlet instance = new ContainsMethods();\n\nconst arith = {\n  double(this: void, x: number): number {\n    return x * 2;\n  },\n};\n\ninstance.bound += 1;\n    ", options: ""},
		{sourceText: "\nclass ContainsMethods {\n  bound?: () => void;\n  unbound?(): void;\n\n  static boundStatic?: () => void;\n  static unboundStatic?(): void;\n}\n\nlet instance = new ContainsMethods();\n\nconst arith = {\n  double(this: void, x: number): number {\n    return x * 2;\n  },\n};\n\ninstance.bound -= 1;\n    ", options: ""},
		{sourceText: "\nclass ContainsMethods {\n  bound?: () => void;\n  unbound?(): void;\n\n  static boundStatic?: () => void;\n  static unboundStatic?(): void;\n}\n\nlet instance = new ContainsMethods();\n\nconst arith = {\n  double(this: void, x: number): number {\n    return x * 2;\n  },\n};\n\ninstance.bound *= 1;\n    ", options: ""},
		{sourceText: "\nclass ContainsMethods {\n  bound?: () => void;\n  unbound?(): void;\n\n  static boundStatic?: () => void;\n  static unboundStatic?(): void;\n}\n\nlet instance = new ContainsMethods();\n\nconst arith = {\n  double(this: void, x: number): number {\n    return x * 2;\n  },\n};\n\ninstance.bound /= 1;\n    ", options: ""},
		{sourceText: "\nclass ContainsMethods {\n  bound?: () => void;\n  unbound?(): void;\n\n  static boundStatic?: () => void;\n  static unboundStatic?(): void;\n}\n\nlet instance = new ContainsMethods();\n\nconst arith = {\n  double(this: void, x: number): number {\n    return x * 2;\n  },\n};\n\ninstance.bound || 0;\n    ", options: ""},
		{sourceText: "\nclass ContainsMethods {\n  bound?: () => void;\n  unbound?(): void;\n\n  static boundStatic?: () => void;\n  static unboundStatic?(): void;\n}\n\nlet instance = new ContainsMethods();\n\nconst arith = {\n  double(this: void, x: number): number {\n    return x * 2;\n  },\n};\n\ninstance.bound && 0;\n    ", options: ""},
		{sourceText: "\nclass ContainsMethods {\n  bound?: () => void;\n  unbound?(): void;\n\n  static boundStatic?: () => void;\n  static unboundStatic?(): void;\n}\n\nlet instance = new ContainsMethods();\n\nconst arith = {\n  double(this: void, x: number): number {\n    return x * 2;\n  },\n};\n\ninstance.bound ? 1 : 0;\n    ", options: ""},
		{sourceText: "\nclass ContainsMethods {\n  bound?: () => void;\n  unbound?(): void;\n\n  static boundStatic?: () => void;\n  static unboundStatic?(): void;\n}\n\nlet instance = new ContainsMethods();\n\nconst arith = {\n  double(this: void, x: number): number {\n    return x * 2;\n  },\n};\n\ninstance.unbound ? 1 : 0;\n    ", options: ""},
		{sourceText: "\nclass ContainsMethods {\n  bound?: () => void;\n  unbound?(): void;\n\n  static boundStatic?: () => void;\n  static unboundStatic?(): void;\n}\n\nlet instance = new ContainsMethods();\n\nconst arith = {\n  double(this: void, x: number): number {\n    return x * 2;\n  },\n};\n\nContainsMethods.boundStatic++;\n    ", options: ""},
		{sourceText: "\nclass ContainsMethods {\n  bound?: () => void;\n  unbound?(): void;\n\n  static boundStatic?: () => void;\n  static unboundStatic?(): void;\n}\n\nlet instance = new ContainsMethods();\n\nconst arith = {\n  double(this: void, x: number): number {\n    return x * 2;\n  },\n};\n\n+ContainsMethods.boundStatic;\n    ", options: ""},
		{sourceText: "\nclass ContainsMethods {\n  bound?: () => void;\n  unbound?(): void;\n\n  static boundStatic?: () => void;\n  static unboundStatic?(): void;\n}\n\nlet instance = new ContainsMethods();\n\nconst arith = {\n  double(this: void, x: number): number {\n    return x * 2;\n  },\n};\n\n++ContainsMethods.boundStatic;\n    ", options: ""},
		{sourceText: "\nclass ContainsMethods {\n  bound?: () => void;\n  unbound?(): void;\n\n  static boundStatic?: () => void;\n  static unboundStatic?(): void;\n}\n\nlet instance = new ContainsMethods();\n\nconst arith = {\n  double(this: void, x: number): number {\n    return x * 2;\n  },\n};\n\nContainsMethods.boundStatic--;\n    ", options: ""},
		{sourceText: "\nclass ContainsMethods {\n  bound?: () => void;\n  unbound?(): void;\n\n  static boundStatic?: () => void;\n  static unboundStatic?(): void;\n}\n\nlet instance = new ContainsMethods();\n\nconst arith = {\n  double(this: void, x: number): number {\n    return x * 2;\n  },\n};\n\n-ContainsMethods.boundStatic;\n    ", options: ""},
		{sourceText: "\nclass ContainsMethods {\n  bound?: () => void;\n  unbound?(): void;\n\n  static boundStatic?: () => void;\n  static unboundStatic?(): void;\n}\n\nlet instance = new ContainsMethods();\n\nconst arith = {\n  double(this: void, x: number): number {\n    return x * 2;\n  },\n};\n\n--ContainsMethods.boundStatic;\n    ", options: ""},
		{sourceText: "\nclass ContainsMethods {\n  bound?: () => void;\n  unbound?(): void;\n\n  static boundStatic?: () => void;\n  static unboundStatic?(): void;\n}\n\nlet instance = new ContainsMethods();\n\nconst arith = {\n  double(this: void, x: number): number {\n    return x * 2;\n  },\n};\n\nContainsMethods.boundStatic += 1;\n    ", options: ""},
		{sourceText: "\nclass ContainsMethods {\n  bound?: () => void;\n  unbound?(): void;\n\n  static boundStatic?: () => void;\n  static unboundStatic?(): void;\n}\n\nlet instance = new ContainsMethods();\n\nconst arith = {\n  double(this: void, x: number): number {\n    return x * 2;\n  },\n};\n\nContainsMethods.boundStatic -= 1;\n    ", options: ""},
		{sourceText: "\nclass ContainsMethods {\n  bound?: () => void;\n  unbound?(): void;\n\n  static boundStatic?: () => void;\n  static unboundStatic?(): void;\n}\n\nlet instance = new ContainsMethods();\n\nconst arith = {\n  double(this: void, x: number): number {\n    return x * 2;\n  },\n};\n\nContainsMethods.boundStatic *= 1;\n    ", options: ""},
		{sourceText: "\nclass ContainsMethods {\n  bound?: () => void;\n  unbound?(): void;\n\n  static boundStatic?: () => void;\n  static unboundStatic?(): void;\n}\n\nlet instance = new ContainsMethods();\n\nconst arith = {\n  double(this: void, x: number): number {\n    return x * 2;\n  },\n};\n\nContainsMethods.boundStatic /= 1;\n    ", options: ""},
		{sourceText: "\nclass ContainsMethods {\n  bound?: () => void;\n  unbound?(): void;\n\n  static boundStatic?: () => void;\n  static unboundStatic?(): void;\n}\n\nlet instance = new ContainsMethods();\n\nconst arith = {\n  double(this: void, x: number): number {\n    return x * 2;\n  },\n};\n\nContainsMethods.boundStatic || 0;\n    ", options: ""},
		{sourceText: "\nclass ContainsMethods {\n  bound?: () => void;\n  unbound?(): void;\n\n  static boundStatic?: () => void;\n  static unboundStatic?(): void;\n}\n\nlet instance = new ContainsMethods();\n\nconst arith = {\n  double(this: void, x: number): number {\n    return x * 2;\n  },\n};\n\ninstane.boundStatic && 0;\n    ", options: ""},
		{sourceText: "\nclass ContainsMethods {\n  bound?: () => void;\n  unbound?(): void;\n\n  static boundStatic?: () => void;\n  static unboundStatic?(): void;\n}\n\nlet instance = new ContainsMethods();\n\nconst arith = {\n  double(this: void, x: number): number {\n    return x * 2;\n  },\n};\n\nContainsMethods.boundStatic ? 1 : 0;\n    ", options: ""},
		{sourceText: "\nclass ContainsMethods {\n  bound?: () => void;\n  unbound?(): void;\n\n  static boundStatic?: () => void;\n  static unboundStatic?(): void;\n}\n\nlet instance = new ContainsMethods();\n\nconst arith = {\n  double(this: void, x: number): number {\n    return x * 2;\n  },\n};\n\nContainsMethods.unboundStatic ? 1 : 0;\n    ", options: ""},
		{sourceText: "\nclass ContainsMethods {\n  bound?: () => void;\n  unbound?(): void;\n\n  static boundStatic?: () => void;\n  static unboundStatic?(): void;\n}\n\nlet instance = new ContainsMethods();\n\nconst arith = {\n  double(this: void, x: number): number {\n    return x * 2;\n  },\n};\n\ntypeof instance.bound === 'function';\n    ", options: ""},
		{sourceText: "\nclass ContainsMethods {\n  bound?: () => void;\n  unbound?(): void;\n\n  static boundStatic?: () => void;\n  static unboundStatic?(): void;\n}\n\nlet instance = new ContainsMethods();\n\nconst arith = {\n  double(this: void, x: number): number {\n    return x * 2;\n  },\n};\n\ntypeof instance.unbound === 'function';\n    ", options: ""},
		{sourceText: "\nclass ContainsMethods {\n  bound?: () => void;\n  unbound?(): void;\n\n  static boundStatic?: () => void;\n  static unboundStatic?(): void;\n}\n\nlet instance = new ContainsMethods();\n\nconst arith = {\n  double(this: void, x: number): number {\n    return x * 2;\n  },\n};\n\ntypeof ContainsMethods.boundStatic === 'function';\n    ", options: ""},
		{sourceText: "\nclass ContainsMethods {\n  bound?: () => void;\n  unbound?(): void;\n\n  static boundStatic?: () => void;\n  static unboundStatic?(): void;\n}\n\nlet instance = new ContainsMethods();\n\nconst arith = {\n  double(this: void, x: number): number {\n    return x * 2;\n  },\n};\n\ntypeof ContainsMethods.unboundStatic === 'function';\n    ", options: ""},
		{sourceText: "\nclass ContainsMethods {\n  bound?: () => void;\n  unbound?(): void;\n\n  static boundStatic?: () => void;\n  static unboundStatic?(): void;\n}\n\nlet instance = new ContainsMethods();\n\nconst arith = {\n  double(this: void, x: number): number {\n    return x * 2;\n  },\n};\n\ninstance.unbound = () => {};\n    ", options: ""},
		{sourceText: "\nclass ContainsMethods {\n  bound?: () => void;\n  unbound?(): void;\n\n  static boundStatic?: () => void;\n  static unboundStatic?(): void;\n}\n\nlet instance = new ContainsMethods();\n\nconst arith = {\n  double(this: void, x: number): number {\n    return x * 2;\n  },\n};\n\ninstance.unbound = instance.unbound.bind(instance);\n    ", options: ""},
		{sourceText: "\nclass ContainsMethods {\n  bound?: () => void;\n  unbound?(): void;\n\n  static boundStatic?: () => void;\n  static unboundStatic?(): void;\n}\n\nlet instance = new ContainsMethods();\n\nconst arith = {\n  double(this: void, x: number): number {\n    return x * 2;\n  },\n};\n\nif (!!instance.unbound) {\n}\n    ", options: ""},
		{sourceText: "\nclass ContainsMethods {\n  bound?: () => void;\n  unbound?(): void;\n\n  static boundStatic?: () => void;\n  static unboundStatic?(): void;\n}\n\nlet instance = new ContainsMethods();\n\nconst arith = {\n  double(this: void, x: number): number {\n    return x * 2;\n  },\n};\n\nvoid instance.unbound;\n    ", options: ""},
		{sourceText: "\nclass ContainsMethods {\n  bound?: () => void;\n  unbound?(): void;\n\n  static boundStatic?: () => void;\n  static unboundStatic?(): void;\n}\n\nlet instance = new ContainsMethods();\n\nconst arith = {\n  double(this: void, x: number): number {\n    return x * 2;\n  },\n};\n\ndelete instance.unbound;\n    ", options: ""},
		{sourceText: "\nclass ContainsMethods {\n  bound?: () => void;\n  unbound?(): void;\n\n  static boundStatic?: () => void;\n  static unboundStatic?(): void;\n}\n\nlet instance = new ContainsMethods();\n\nconst arith = {\n  double(this: void, x: number): number {\n    return x * 2;\n  },\n};\n\nconst { double } = arith;\n    ", options: ""},
		{sourceText: "\ninterface RecordA {\n  readonly type: 'A';\n  readonly a: {};\n}\ninterface RecordB {\n  readonly type: 'B';\n  readonly b: {};\n}\ntype AnyRecord = RecordA | RecordB;\n\nfunction test(obj: AnyRecord) {\n  switch (obj.type) {\n  }\n}\n    ", options: ""},
		{sourceText: "\nclass CommunicationError {\n  constructor() {\n    const x = CommunicationError.prototype;\n  }\n}\n    ", options: ""},
		{sourceText: "\nclass CommunicationError {}\nconst x = CommunicationError.prototype;\n    ", options: ""},
		{sourceText: "\nclass ContainsMethods {\n  bound?: () => void;\n  unbound?(): void;\n\n  static boundStatic?: () => void;\n  static unboundStatic?(): void;\n}\n\nfunction foo(instance: ContainsMethods | null) {\n  instance?.bound();\n  instance?.unbound();\n\n  if (instance?.bound) {\n  }\n  if (instance?.unbound) {\n  }\n\n  typeof instance?.bound === 'function';\n  typeof instance?.unbound === 'function';\n}\n    ", options: ""},
		{sourceText: "\ninterface OptionalMethod {\n  mightBeDefined?(): void;\n}\n\nconst x: OptionalMethod = {};\ndeclare const myCondition: boolean;\nif (myCondition || x.mightBeDefined) {\n  console.log('hello world');\n}\n    ", options: ""},
		{sourceText: "\nclass A {\n  unbound(): void {\n    this.unbound = undefined;\n    this.unbound = this.unbound.bind(this);\n  }\n}\n    ", options: ""},
		{sourceText: "const { parseInt } = Number;", options: ""},
		{sourceText: "const { log } = console;", options: ""},
		{sourceText: "\nlet parseInt;\n({ parseInt } = Number);\n    ", options: ""},
		{sourceText: "\nlet log;\n({ log } = console);\n    ", options: ""},
		{sourceText: "\nconst foo = {\n  bar: 'bar',\n};\nconst { bar } = foo;\n    ", options: ""},
		{sourceText: "\nclass Foo {\n  unbnound() {}\n  bar = 4;\n}\nconst { bar } = new Foo();\n    ", options: ""},
		{sourceText: "\nclass Foo {\n  bound = () => 'foo';\n}\nconst { bound } = new Foo();\n    ", options: ""},
		{sourceText: "\nclass Foo {\n  bound = () => 'foo';\n}\nfunction foo({ bound } = new Foo()) {}\n    ", options: ""},
		{sourceText: "\nclass Foo {\n  bound = () => 'foo';\n}\ndeclare const bar: Foo;\nfunction foo({ bound }: Foo) {}\n    ", options: ""},
		{sourceText: "\nclass Foo {\n  bound = () => 'foo';\n}\nclass Bar {\n  bound = () => 'bar';\n}\nfunction foo({ bound }: Foo | Bar) {}\n    ", options: ""},
		{sourceText: "\nclass Foo {\n  bound = () => 'foo';\n}\ntype foo = ({ bound }: Foo) => void;\n    ", options: ""},
		{sourceText: "\nclass Foo {\n  unbound = function () {};\n}\ntype foo = ({ unbound }: Foo) => void;\n    ", options: ""},
		{sourceText: "\nclass Foo {\n  bound = () => 'foo';\n}\nclass Bar {\n  bound = () => 'bar';\n}\nfunction foo({ bound }: Foo & Bar) {}\n    ", options: ""},
		{sourceText: "\nclass Foo {\n  unbound = function () {};\n}\ndeclare const { unbound }: Foo;\n    ", options: ""},
		{sourceText: "declare const { unbound } = '***';", options: ""},
		{sourceText: "\nclass Foo {\n  unbound = function () {};\n}\ntype foo = (a: (b: (c: ({ unbound }: Foo) => void) => void) => void) => void;\n    ", options: ""},
		{sourceText: "\nclass Foo {\n  unbound = function () {};\n}\nclass Bar {\n  property: ({ unbound }: Foo) => void;\n}\n    ", options: ""},
		{sourceText: "\nclass Foo {\n  unbound = function () {};\n}\nfunction foo<T extends ({ unbound }: Foo) => void>() {}\n    ", options: ""},
		{sourceText: "\nclass Foo {\n  unbound = function () {};\n}\nabstract class Bar {\n  abstract foo({ unbound }: Foo);\n}\n    ", options: ""},
		{sourceText: "\nclass Foo {\n  unbound = function () {};\n}\ndeclare class Bar {\n  foo({ unbound }: Foo);\n}\n    ", options: ""},
		{sourceText: "\nclass Foo {\n  unbound = function () {};\n}\ndeclare function foo({ unbound }: Foo);\n    ", options: ""},
		{sourceText: "\nclass Foo {\n  unbound = function () {};\n}\ninterface Bar {\n  foo: ({ unbound }: Foo) => void;\n}\n    ", options: ""},
		{sourceText: "\nclass Foo {\n  unbound = function () {};\n}\ninterface Bar {\n  foo({ unbound }: Foo): void;\n}\n    ", options: ""},
		{sourceText: "\nclass Foo {\n  unbound = function () {};\n}\ninterface Bar {\n  new ({ unbound }: Foo): Foo;\n}\n    ", options: ""},
		{sourceText: "\nclass Foo {\n  unbound = function () {};\n}\ntype foo = new ({ unbound }: Foo) => void;\n    ", options: ""},
		{sourceText: "const { unbound } = { unbound: () => {} };", options: ""},
		{sourceText: "function foo({ unbound }: { unbound: () => void } = { unbound: () => {} }) {}", options: ""},
		{sourceText: "\nclass BaseClass {\n  x: number = 42;\n  logThis() {}\n}\nclass OtherClass extends BaseClass {\n  superLogThis: any;\n  constructor() {\n    super();\n    this.superLogThis = super.logThis;\n  }\n}\nconst oc = new OtherClass();\noc.superLogThis();\n    ", options: ""},
		{sourceText: "\nclass Foo {\n  bound = () => {};\n}\nclass Bar {\n  bound = 1;\n}\ndeclare const union: Foo | Bar;\nconst bound = union.bound;\n    ", options: ""},
		{sourceText: "\nclass Foo {\n  bazz() {}\n}\ndeclare const foo: Foo;\ndeclare const key: string;\nconst bound = foo[key];\n    ", options: ""},
		{sourceText: "\nclass Foo {\n  bazz() {}\n}\ndeclare const foo: Foo;\ndeclare const bazz: string;\nfoo[bazz];\n    ", options: ""},
	}
	for index, testCase := range cases {
		t.Run(unboundMethodCaseName(index), func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, runUnboundMethod(t, testCase.sourceText, testCase.options))
		})
	}
}

// TestUnboundMethodFiresOnUpstreamFailCases is the imported failing corpus, verbatim, with the span
// of every finding asserted.
//
// Which of the two messages is reported is itself a judgment rather than a detail: a method whose
// first parameter is not named `this` gets the longer message telling the reader that annotating it
// `this: void` would settle the question, and one that already has a `this` parameter of some other
// type gets the short one. Fifty of these sixty-nine findings are the long form and nineteen the
// short, so a port collapsing the two satisfies no fixture here.
//
// The spans come from replaying each case through the installed build, because upstream records a
// start column on only three of its sixty-nine findings and a line number alone cannot see a finding
// anchored on the wrong node on the right line.
func TestUnboundMethodFiresOnUpstreamFailCases(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText string
		options    string
		wantIds    []string
		wantSpans  []string
	}{
		{
			sourceText: "\nclass Console {\n  log(str) {\n    process.stdout.write(str);\n  }\n}\n\nconst console = new Console();\n\nPromise.resolve().then(console.log);\n      ",
			options:    "",
			wantIds:    []string{"unboundWithoutThisAnnotation"},
			wantSpans:  []string{"console.log"},
		},
		{
			sourceText: "\nclass ContainsMethods {\n  bound?: () => void;\n  unbound?(): void;\n\n  static boundStatic?: () => void;\n  static unboundStatic?(): void;\n}\n\nlet instance = new ContainsMethods();\n\nconst arith = {\n  double(this: void, x: number): number {\n    return x * 2;\n  },\n};\n\nfunction foo(arg: ContainsMethods | null) {\n  const unbound = arg?.unbound;\n  arg.unbound += 1;\n  arg?.unbound as any;\n}\n      ",
			options:    "",
			wantIds:    []string{"unboundWithoutThisAnnotation", "unboundWithoutThisAnnotation", "unboundWithoutThisAnnotation"},
			wantSpans:  []string{"arg?.unbound", "arg.unbound", "arg?.unbound"},
		},
		{
			sourceText: "\nclass ContainsMethods {\n  bound?: () => void;\n  unbound?(): void;\n\n  static boundStatic?: () => void;\n  static unboundStatic?(): void;\n}\n\nlet instance = new ContainsMethods();\n\nconst arith = {\n  double(this: void, x: number): number {\n    return x * 2;\n  },\n};\n\nconst unbound = instance.unbound;\n      ",
			options:    "",
			wantIds:    []string{"unboundWithoutThisAnnotation"},
			wantSpans:  []string{"instance.unbound"},
		},
		{
			sourceText: "\nclass ContainsMethods {\n  bound?: () => void;\n  unbound?(): void;\n\n  static boundStatic?: () => void;\n  static unboundStatic?(): void;\n}\n\nlet instance = new ContainsMethods();\n\nconst arith = {\n  double(this: void, x: number): number {\n    return x * 2;\n  },\n};\n\nconst unboundStatic = ContainsMethods.unboundStatic;\n      ",
			options:    "",
			wantIds:    []string{"unboundWithoutThisAnnotation"},
			wantSpans:  []string{"ContainsMethods.unboundStatic"},
		},
		{
			sourceText: "\nclass ContainsMethods {\n  bound?: () => void;\n  unbound?(): void;\n\n  static boundStatic?: () => void;\n  static unboundStatic?(): void;\n}\n\nlet instance = new ContainsMethods();\n\nconst arith = {\n  double(this: void, x: number): number {\n    return x * 2;\n  },\n};\n\nconst { unbound } = instance;\n      ",
			options:    "",
			wantIds:    []string{"unboundWithoutThisAnnotation"},
			wantSpans:  []string{"unbound"},
		},
		{
			sourceText: "\nclass ContainsMethods {\n  bound?: () => void;\n  unbound?(): void;\n\n  static boundStatic?: () => void;\n  static unboundStatic?(): void;\n}\n\nlet instance = new ContainsMethods();\n\nconst arith = {\n  double(this: void, x: number): number {\n    return x * 2;\n  },\n};\n\nconst { unboundStatic } = ContainsMethods;\n      ",
			options:    "",
			wantIds:    []string{"unboundWithoutThisAnnotation"},
			wantSpans:  []string{"unboundStatic"},
		},
		{
			sourceText: "\nclass ContainsMethods {\n  bound?: () => void;\n  unbound?(): void;\n\n  static boundStatic?: () => void;\n  static unboundStatic?(): void;\n}\n\nlet instance = new ContainsMethods();\n\nconst arith = {\n  double(this: void, x: number): number {\n    return x * 2;\n  },\n};\n\n<any>instance.unbound;\n      ",
			options:    "",
			wantIds:    []string{"unboundWithoutThisAnnotation"},
			wantSpans:  []string{"instance.unbound"},
		},
		{
			sourceText: "\nclass ContainsMethods {\n  bound?: () => void;\n  unbound?(): void;\n\n  static boundStatic?: () => void;\n  static unboundStatic?(): void;\n}\n\nlet instance = new ContainsMethods();\n\nconst arith = {\n  double(this: void, x: number): number {\n    return x * 2;\n  },\n};\n\ninstance.unbound as any;\n      ",
			options:    "",
			wantIds:    []string{"unboundWithoutThisAnnotation"},
			wantSpans:  []string{"instance.unbound"},
		},
		{
			sourceText: "\nclass ContainsMethods {\n  bound?: () => void;\n  unbound?(): void;\n\n  static boundStatic?: () => void;\n  static unboundStatic?(): void;\n}\n\nlet instance = new ContainsMethods();\n\nconst arith = {\n  double(this: void, x: number): number {\n    return x * 2;\n  },\n};\n\n<any>ContainsMethods.unboundStatic;\n      ",
			options:    "",
			wantIds:    []string{"unboundWithoutThisAnnotation"},
			wantSpans:  []string{"ContainsMethods.unboundStatic"},
		},
		{
			sourceText: "\nclass ContainsMethods {\n  bound?: () => void;\n  unbound?(): void;\n\n  static boundStatic?: () => void;\n  static unboundStatic?(): void;\n}\n\nlet instance = new ContainsMethods();\n\nconst arith = {\n  double(this: void, x: number): number {\n    return x * 2;\n  },\n};\n\nContainsMethods.unboundStatic as any;\n      ",
			options:    "",
			wantIds:    []string{"unboundWithoutThisAnnotation"},
			wantSpans:  []string{"ContainsMethods.unboundStatic"},
		},
		{
			sourceText: "\nclass ContainsMethods {\n  bound?: () => void;\n  unbound?(): void;\n\n  static boundStatic?: () => void;\n  static unboundStatic?(): void;\n}\n\nlet instance = new ContainsMethods();\n\nconst arith = {\n  double(this: void, x: number): number {\n    return x * 2;\n  },\n};\n\ninstance.unbound || 0;\n      ",
			options:    "",
			wantIds:    []string{"unboundWithoutThisAnnotation"},
			wantSpans:  []string{"instance.unbound"},
		},
		{
			sourceText: "\nclass ContainsMethods {\n  bound?: () => void;\n  unbound?(): void;\n\n  static boundStatic?: () => void;\n  static unboundStatic?(): void;\n}\n\nlet instance = new ContainsMethods();\n\nconst arith = {\n  double(this: void, x: number): number {\n    return x * 2;\n  },\n};\n\nContainsMethods.unboundStatic || 0;\n      ",
			options:    "",
			wantIds:    []string{"unboundWithoutThisAnnotation"},
			wantSpans:  []string{"ContainsMethods.unboundStatic"},
		},
		{
			sourceText: "\nclass ContainsMethods {\n  bound?: () => void;\n  unbound?(): void;\n\n  static boundStatic?: () => void;\n  static unboundStatic?(): void;\n}\n\nlet instance = new ContainsMethods();\n\nconst arith = {\n  double(this: void, x: number): number {\n    return x * 2;\n  },\n};\n\ninstance.unbound ? instance.unbound : null;\n      ",
			options:    "",
			wantIds:    []string{"unboundWithoutThisAnnotation"},
			wantSpans:  []string{"instance.unbound"},
		},
		{
			sourceText: "\nclass ContainsMethods {\n  unbound?(): void;\n\n  static unboundStatic?(): void;\n}\n\nnew ContainsMethods().unbound;\n\nContainsMethods.unboundStatic;\n      ",
			options:    "{\"ignoreStatic\": true}",
			wantIds:    []string{"unboundWithoutThisAnnotation"},
			wantSpans:  []string{"new ContainsMethods().unbound"},
		},
		{
			sourceText: "\nclass CommunicationError {\n  foo() {}\n}\nconst x = CommunicationError.prototype.foo;\n      ",
			options:    "",
			wantIds:    []string{"unboundWithoutThisAnnotation"},
			wantSpans:  []string{"CommunicationError.prototype.foo"},
		},
		{
			sourceText: "const x = Promise.all;",
			options:    "",
			wantIds:    []string{"unboundWithoutThisAnnotation"},
			wantSpans:  []string{"Promise.all"},
		},
		{
			sourceText: "\nclass Foo {\n  unbound() {}\n}\nconst instance = new Foo();\n\nlet x;\n\nx = instance.unbound; // THIS SHOULD ERROR\ninstance.unbound = x; // THIS SHOULD NOT\n      ",
			options:    "",
			wantIds:    []string{"unboundWithoutThisAnnotation"},
			wantSpans:  []string{"instance.unbound"},
		},
		{
			sourceText: "\nclass Foo extends Number {\n  static parseInt = function (string: string, radix?: number): number {};\n}\nconst foo = Foo;\n['1', '2', '3'].map(foo.parseInt);\n      ",
			options:    "",
			wantIds:    []string{"unbound"},
			wantSpans:  []string{"foo.parseInt"},
		},
		{
			sourceText: "\ndeclare const foo: Number;\nconst x = foo.toFixed;\n      ",
			options:    "",
			wantIds:    []string{"unboundWithoutThisAnnotation"},
			wantSpans:  []string{"foo.toFixed"},
		},
		{
			sourceText: "\ndeclare const foo: Object;\nconst x = foo.hasOwnProperty;\n      ",
			options:    "",
			wantIds:    []string{"unboundWithoutThisAnnotation"},
			wantSpans:  []string{"foo.hasOwnProperty"},
		},
		{
			sourceText: "\ndeclare const foo: String;\nconst x = foo.slice;\n      ",
			options:    "",
			wantIds:    []string{"unboundWithoutThisAnnotation"},
			wantSpans:  []string{"foo.slice"},
		},
		{
			sourceText: "\ndeclare const foo: Date;\nconst x = foo.getTime;\n      ",
			options:    "",
			wantIds:    []string{"unboundWithoutThisAnnotation"},
			wantSpans:  []string{"foo.getTime"},
		},
		{
			sourceText: "\nclass Foo extends Number {}\nconst x = Foo.parseInt;\n      ",
			options:    "",
			wantIds:    []string{"unboundWithoutThisAnnotation"},
			wantSpans:  []string{"Foo.parseInt"},
		},
		{
			sourceText: "\nclass Foo extends String {}\nconst x = Foo.fromCharCode;\n      ",
			options:    "",
			wantIds:    []string{"unboundWithoutThisAnnotation"},
			wantSpans:  []string{"Foo.fromCharCode"},
		},
		{
			sourceText: "\nclass Foo extends Object {}\nconst x = Foo.defineProperty;\n      ",
			options:    "",
			wantIds:    []string{"unboundWithoutThisAnnotation"},
			wantSpans:  []string{"Foo.defineProperty"},
		},
		{
			sourceText: "\nclass Foo extends Date {}\nconst x = Foo.parse;\n      ",
			options:    "",
			wantIds:    []string{"unboundWithoutThisAnnotation"},
			wantSpans:  []string{"Foo.parse"},
		},
		{
			sourceText: "\nclass Foo {\n  unbound = function () {};\n}\nconst unbound = new Foo().unbound;\n      ",
			options:    "",
			wantIds:    []string{"unbound"},
			wantSpans:  []string{"new Foo().unbound"},
		},
		{
			sourceText: "\nclass Foo {\n  unbound() {}\n}\nconst { unbound } = new Foo();\n      ",
			options:    "",
			wantIds:    []string{"unboundWithoutThisAnnotation"},
			wantSpans:  []string{"unbound"},
		},
		{
			sourceText: "\nclass Foo {\n  unbound = function () {};\n}\nconst { unbound } = new Foo();\n      ",
			options:    "",
			wantIds:    []string{"unbound"},
			wantSpans:  []string{"unbound"},
		},
		{
			sourceText: "\nclass Foo {\n  unbound() {}\n}\nlet unbound;\n({ unbound } = new Foo());\n      ",
			options:    "",
			wantIds:    []string{"unboundWithoutThisAnnotation"},
			wantSpans:  []string{"unbound"},
		},
		{
			sourceText: "\nclass Foo {\n  unbound = function () {};\n}\nlet unbound;\n({ unbound } = new Foo());\n      ",
			options:    "",
			wantIds:    []string{"unbound"},
			wantSpans:  []string{"unbound"},
		},
		{
			sourceText: "\nclass Foo {\n  unbound = function () {};\n}\nfunction foo({ unbound }: Foo = new Foo()) {}\n      ",
			options:    "",
			wantIds:    []string{"unbound"},
			wantSpans:  []string{"unbound"},
		},
		{
			sourceText: "\nclass Foo {\n  unbound = function () {};\n}\ndeclare const bar: Foo;\nfunction foo({ unbound }: Foo = bar) {}\n      ",
			options:    "",
			wantIds:    []string{"unbound"},
			wantSpans:  []string{"unbound"},
		},
		{
			sourceText: "\nclass Foo {\n  unbound = function () {};\n}\ndeclare const bar: Foo;\nfunction foo({ unbound }: Foo = { unbound: () => {} }) {}\n      ",
			options:    "",
			wantIds:    []string{"unbound"},
			wantSpans:  []string{"unbound"},
		},
		{
			sourceText: "\nclass Foo {\n  unbound = function () {};\n}\ndeclare const bar: Foo;\nfunction foo({ unbound }: Foo = { unbound: function () {} }) {}\n      ",
			options:    "",
			wantIds:    []string{"unboundWithoutThisAnnotation"},
			wantSpans:  []string{"unbound"},
		},
		{
			sourceText: "\nclass Foo {\n  unbound = function () {};\n}\nfunction foo({ unbound }: Foo) {}\n      ",
			options:    "",
			wantIds:    []string{"unbound"},
			wantSpans:  []string{"unbound"},
		},
		{
			sourceText: "\nclass Foo {\n  unbound = function () {};\n}\nfunction bar(cb: (arg: Foo) => void) {}\nbar(({ unbound }) => {});\n      ",
			options:    "",
			wantIds:    []string{"unbound"},
			wantSpans:  []string{"unbound"},
		},
		{
			sourceText: "\nclass Foo {\n  unbound = function () {};\n}\nfunction bar(cb: (arg: { unbound: () => void }) => void) {}\nbar(({ unbound } = new Foo()) => {});\n      ",
			options:    "",
			wantIds:    []string{"unbound"},
			wantSpans:  []string{"unbound"},
		},
		{
			sourceText: "\nclass Foo {\n  unbound = function () {};\n}\nfor (const { unbound } of [new Foo(), new Foo()]) {\n}\n      ",
			options:    "",
			wantIds:    []string{"unbound"},
			wantSpans:  []string{"unbound"},
		},
		{
			sourceText: "\nclass Foo {\n  unbound = function () {};\n\n  foo({ unbound }: Foo) {}\n}\n      ",
			options:    "",
			wantIds:    []string{"unbound"},
			wantSpans:  []string{"unbound"},
		},
		{
			sourceText: "\nclass Foo {\n  unbound = function () {};\n}\nclass Bar {\n  unbound = function () {};\n}\nfunction foo({ unbound }: Foo | Bar) {}\n      ",
			options:    "",
			wantIds:    []string{"unbound"},
			wantSpans:  []string{"unbound"},
		},
		{
			sourceText: "\nclass Foo {\n  unbound = function () {};\n}\nfunction foo({ unbound }: { unbound: () => string } | Foo) {}\n      ",
			options:    "",
			wantIds:    []string{"unbound"},
			wantSpans:  []string{"unbound"},
		},
		{
			sourceText: "\nclass Foo {\n  unbound = function () {};\n}\nclass Bar {\n  unbound = () => {};\n}\nfunction foo({ unbound }: Foo | Bar) {}\n      ",
			options:    "",
			wantIds:    []string{"unbound"},
			wantSpans:  []string{"unbound"},
		},
		{
			sourceText: "\nclass Foo {\n  unbound = function () {};\n}\nconst foo = ({ unbound }: Foo & { foo: () => 'bar' }) => {};\n      ",
			options:    "",
			wantIds:    []string{"unbound"},
			wantSpans:  []string{"unbound"},
		},
		{
			sourceText: "\nclass Foo {\n  unbound = function () {};\n}\nclass Bar {\n  unbound = () => {};\n}\nconst foo = ({ unbound }: (Foo & { foo: () => 'bar' }) | Bar) => {};\n      ",
			options:    "",
			wantIds:    []string{"unbound"},
			wantSpans:  []string{"unbound"},
		},
		{
			sourceText: "\nclass Foo {\n  unbound = function () {};\n}\nclass Bar {\n  unbound = () => {};\n}\nconst foo = ({ unbound }: Foo & Bar) => {};\n      ",
			options:    "",
			wantIds:    []string{"unbound"},
			wantSpans:  []string{"unbound"},
		},
		{
			sourceText: "\nclass Foo {\n  unbound = function () {};\n\n  other = function () {};\n}\nclass Bar {\n  unbound = () => {};\n}\nconst foo = ({ unbound, ...rest }: Foo & Bar) => {};\n      ",
			options:    "",
			wantIds:    []string{"unbound"},
			wantSpans:  []string{"unbound"},
		},
		{
			sourceText: "const { unbound } = { unbound: function () {} };",
			options:    "",
			wantIds:    []string{"unboundWithoutThisAnnotation"},
			wantSpans:  []string{"unbound"},
		},
		{
			sourceText: "\nfunction foo(\n  { unbound }: { unbound: () => void } = { unbound: function () {} },\n) {}\n      ",
			options:    "",
			wantIds:    []string{"unboundWithoutThisAnnotation"},
			wantSpans:  []string{"unbound"},
		},
		{
			sourceText: "\nclass CommunicationError {\n  foo() {}\n}\nconst { foo } = CommunicationError.prototype;\n      ",
			options:    "",
			wantIds:    []string{"unboundWithoutThisAnnotation"},
			wantSpans:  []string{"foo"},
		},
		{
			sourceText: "\nclass CommunicationError {\n  foo() {}\n}\nlet foo;\n({ foo } = CommunicationError.prototype);\n      ",
			options:    "",
			wantIds:    []string{"unboundWithoutThisAnnotation"},
			wantSpans:  []string{"foo"},
		},
		{
			sourceText: "const { all } = Promise;",
			options:    "",
			wantIds:    []string{"unboundWithoutThisAnnotation"},
			wantSpans:  []string{"all"},
		},
		{
			sourceText: "\nclass BaseClass {\n  logThis() {}\n}\nclass OtherClass extends BaseClass {\n  constructor() {\n    super();\n    const x = super.logThis;\n  }\n}\n      ",
			options:    "",
			wantIds:    []string{"unboundWithoutThisAnnotation"},
			wantSpans:  []string{"super.logThis"},
		},
		{
			sourceText: "\nclass BaseClass {\n  logThis() {}\n}\nclass OtherClass extends BaseClass {\n  constructor() {\n    super();\n    let x;\n    x = super.logThis;\n  }\n}\n      ",
			options:    "",
			wantIds:    []string{"unboundWithoutThisAnnotation"},
			wantSpans:  []string{"super.logThis"},
		},
		{
			sourceText: "\nconst values = {\n  a() {},\n  b: () => {},\n};\n\nconst { a, b } = values;\n      ",
			options:    "",
			wantIds:    []string{"unboundWithoutThisAnnotation"},
			wantSpans:  []string{"a"},
		},
		{
			sourceText: "\nconst values = {\n  a() {},\n  b: () => {},\n};\n\nconst { a: c } = values;\n      ",
			options:    "",
			wantIds:    []string{"unboundWithoutThisAnnotation"},
			wantSpans:  []string{"a"},
		},
		{
			sourceText: "\nconst values = {\n  a() {},\n  b: () => {},\n};\n\nconst { b, a } = values;\n      ",
			options:    "",
			wantIds:    []string{"unboundWithoutThisAnnotation"},
			wantSpans:  []string{"a"},
		},
		{
			sourceText: "\nconst objectLiteral = {\n  f: function () {},\n};\nconst f = objectLiteral.f;\n      ",
			options:    "",
			wantIds:    []string{"unboundWithoutThisAnnotation"},
			wantSpans:  []string{"objectLiteral.f"},
		},
		{
			sourceText: "\nclass Foo {\n  bazz() {}\n}\nclass Bar {\n  bazz = 1;\n}\ndeclare const union: Foo | Bar;\nconst bound = union.bazz;\n      ",
			options:    "",
			wantIds:    []string{"unboundWithoutThisAnnotation"},
			wantSpans:  []string{"union.bazz"},
		},
		{
			sourceText: "\nclass Foo {\n  bazz() {}\n}\nclass Bar {\n  bazz = 1;\n}\ndeclare const union: Foo | Bar;\ndeclare const bazz: 'bazz';\nconst bound = union[bazz];\n      ",
			options:    "",
			wantIds:    []string{"unboundWithoutThisAnnotation"},
			wantSpans:  []string{"union[bazz]"},
		},
		{
			sourceText: "\nclass Foo {\n  bazz() {}\n}\ndeclare const foo: Foo;\nfoo['bazz'];\n      ",
			options:    "",
			wantIds:    []string{"unboundWithoutThisAnnotation"},
			wantSpans:  []string{"foo['bazz']"},
		},
		{
			sourceText: "\nclass Foo {\n  bazz() {}\n}\ndeclare const foo: Foo;\ndeclare const bazz: keyof Foo;\nconst bound = foo[bazz];\n      ",
			options:    "",
			wantIds:    []string{"unboundWithoutThisAnnotation"},
			wantSpans:  []string{"foo[bazz]"},
		},
		{
			sourceText: "\nclass Foo {\n  bazz() {}\n}\ndeclare const foo: Foo;\nconst bound = foo[`ba${'zz'}`];\n      ",
			options:    "",
			wantIds:    []string{"unboundWithoutThisAnnotation"},
			wantSpans:  []string{"foo[`ba${'zz'}`]"},
		},
		{
			sourceText: "\nclass Foo {\n  1() {}\n}\ndeclare const foo: Foo;\nfoo[1];\n      ",
			options:    "",
			wantIds:    []string{"unboundWithoutThisAnnotation"},
			wantSpans:  []string{"foo[1]"},
		},
	}
	for index, testCase := range cases {
		t.Run(unboundMethodCaseName(index), func(t *testing.T) {
			t.Parallel()
			result := runUnboundMethod(t, testCase.sourceText, testCase.options)
			rule_testing.ExpectFindings(t, result, testCase.wantIds...)
			if len(result.Diagnostics) != len(testCase.wantSpans) {
				t.Fatalf("wanted %d findings, got %d", len(testCase.wantSpans), len(result.Diagnostics))
			}
			written := asTheTypedUnboundMethodHarnessWroteIt(testCase.sourceText)
			for findingIndex, wantSpan := range testCase.wantSpans {
				reported := result.Diagnostics[findingIndex]
				gotSpan := written[reported.Range.Pos():reported.Range.End()]
				if gotSpan != wantSpan {
					t.Errorf("finding %d points at %q, wanted %q", findingIndex, gotSpan, wantSpan)
				}
			}
		})
	}
}

// TestUnboundMethodResolvesAcrossAModuleBoundary carries the two upstream cases that import.
//
// Both were lifted out of the main failing table rather than deleted or weakened, because neither is
// expressible through the single-file harness: they import `console` from a sibling module that
// upstream ships in its own fixtures directory, and a rule handed one file has nothing to resolve
// that import against. The finding depends entirely on what the import resolves to, so a single-file
// version of these cases would pass while testing nothing.
//
// The sibling module here is the relevant line of upstream's own tests/fixtures/class.ts, copied
// rather than invented. Its comment there says outright that the export exists for this rule's tests.
//
// These are also the two cases that separate the frozen natively-bound name table from the
// type-level check: `console` is declared in @types/node rather than in the default library, so the
// type-level test answers false for it and only the name lookup would exempt it. Here the import
// makes it a local module symbol instead, which is not natively bound at all, so both report.
func TestUnboundMethodResolvesAcrossAModuleBoundary(t *testing.T) {
	t.Parallel()

	const moduleFileName = "/repository/source/class.ts"
	const moduleSource = "export const console = { log() {} };\n"

	cases := []struct {
		name       string
		sourceText string
		wantId     string
		wantSpan   string
	}{
		{
			name:       "member access",
			sourceText: "\nimport { console } from './class';\nconst x = console.log;\n      ",
			wantId:     "unboundWithoutThisAnnotation",
			wantSpan:   "console.log",
		},
		{
			name:       "destructuring",
			sourceText: "\nimport { console } from './class';\nconst { log } = console;\n      ",
			wantId:     "unboundWithoutThisAnnotation",
			wantSpan:   "log",
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTypedFiles(t, UnboundMethod, map[string]string{
				unboundMethodFile: testCase.sourceText,
				moduleFileName:    moduleSource,
			}, unboundMethodFile)
			rule_testing.ExpectFindings(t, result, testCase.wantId)
			if len(result.Diagnostics) != 1 {
				t.Fatalf("wanted one finding, got %d", len(result.Diagnostics))
			}
			written := asTheTypedUnboundMethodHarnessWroteIt(testCase.sourceText)
			reported := result.Diagnostics[0]
			gotSpan := written[reported.Range.Pos():reported.Range.End()]
			if gotSpan != testCase.wantSpan {
				t.Errorf("the finding points at %q, wanted %q", gotSpan, testCase.wantSpan)
			}
		})
	}
}

// TestUnboundMethodOnAUnionWhoseConstituentsDisagree is upstream's one case where our answer differs,
// recorded as a divergence rather than made to pass.
//
// The input destructures `floor` out of `Foo | Math`, where `Foo.floor` is a property holding a
// function expression and `Math.floor` is a method signature. Both are dangerous, so BOTH orderings
// report, on the same node, with the same span. They disagree only about WHICH MESSAGE, because the
// two constituents reach different arms of the danger test: the property arm never reads a parameter
// list and takes the short message, the signature arm reads one, finds no `this`, and takes the long
// one.
//
// Upstream stops at the first constituent that reports, and the constituent order is the type
// checker's own normalization rather than the source order. Measured on both spellings of the
// conditional, on both compilers:
//
//	upstream    Math | Foo    reports through Math, message unboundWithoutThisAnnotation
//	here        Foo | Math    reports through Foo,  message unbound
//
// Neither compiler follows the source: writing `Math` first and writing `Foo` first give the same
// normalized order within each. So this is a stable difference between two type checkers, and there
// is no ordering knob in the rule to turn. Reproducing upstream would mean sorting union constituents
// by something the rule invented, which would be a different rule and would drift the moment either
// checker changed its normalization.
//
// The scope of the divergence is exactly this: a union whose constituents are all dangerous but reach
// different arms. When they agree, or when only one is dangerous, both orderings pick the same one.
// One of upstream's two hundred and eleven cases is in that position.
func TestUnboundMethodOnAUnionWhoseConstituentsDisagree(t *testing.T) {
	t.Parallel()

	const sourceText = "\nclass Foo {\n  floor = function () {};\n}\n\n" +
		"const { floor } = Math.random() > 0.5 ? new Foo() : Math;\n      "

	result := runUnboundMethod(t, sourceText, "")

	// It reports, on the right node. That much matches upstream exactly.
	if len(result.Diagnostics) != 1 {
		t.Fatalf("wanted one finding, got %d", len(result.Diagnostics))
	}
	written := asTheTypedUnboundMethodHarnessWroteIt(sourceText)
	reported := result.Diagnostics[0]
	if gotSpan := written[reported.Range.Pos():reported.Range.End()]; gotSpan != "floor" {
		t.Errorf("the finding points at %q, wanted %q", gotSpan, "floor")
	}

	// And it takes the other message, for the checker-ordering reason above. Asserted rather than
	// skipped so that a future change in either normalization fails here and is looked at, instead
	// of silently agreeing or silently drifting further.
	const upstreamMessageId = "unboundWithoutThisAnnotation"
	const ourMessageId = "unbound"
	if reported.Message.Id != ourMessageId {
		t.Errorf("reported %q, wanted %q; upstream reports %q here and the difference is the "+
			"union constituent order, so a change on either side belongs in the comment above",
			reported.Message.Id, ourMessageId, upstreamMessageId)
	}
}

// TestUnboundMethodSeesThroughWrappersUpstreamsParserRemoves covers the two recursion arms whose
// verdicts upstream's corpus cannot separate.
//
// A safe use is safe no matter how many relabelling wrappers sit between the reference and the thing
// consuming it, so the walk passes through an assertion, a non-null operator, and a parenthesis. The
// last of those has no upstream counterpart: typescript-eslint's parser folds parentheses away, so
// its walk never meets one and no case it writes can tell an arm that recurses from an arm that is
// absent. Deleting either arm survives all two hundred and eleven imported rows.
//
// Both directions are here on purpose. A parenthesis arm that returned true unconditionally would
// pass the two silent rows and quietly exempt every parenthesized detachment, so the reporting rows
// are what make these a measurement rather than a rubber stamp.
//
// Measured against the installed 8.x build, one program per case, with controls in the same run.
func TestUnboundMethodSeesThroughWrappersUpstreamsParserRemoves(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		why        string
		sourceText string
		wantIds    []string
		wantSpans  []string
	}{
		{
			name:       "parenthesized-call",
			why:        "a parenthesized method reference that is immediately CALLED. Our parser keeps KindParenthesizedExpression and upstream's folds it away, so upstream's safe-use walk never meets one and its corpus cannot write this shape at all. Without the parenthesis arm the walk stops at the parenthesis, the call above it is never seen, and this reports",
			sourceText: "declare const o: { m(): void };\n(o.m)();\n",
			wantIds:    []string{},
			wantSpans:  []string{},
		},
		{
			name:       "parenthesized-typeof",
			why:        "the same through a typeof, which is a different safe-use arm reached through the same wrapper",
			sourceText: "declare const o: { m(): void };\nconst t = typeof (o.m);\n",
			wantIds:    []string{},
			wantSpans:  []string{},
		},
		{
			name:       "parenthesized-detach",
			why:        "the half that must still REPORT, so the parenthesis arm is a pass-through rather than a blanket exemption; without this row a port that simply returned true for a parenthesis would look correct",
			sourceText: "declare const o: { m(): void };\nconst f = (o.m);\n",
			wantIds:    []string{"unboundWithoutThisAnnotation"},
			wantSpans:  []string{"o.m"},
		},
		{
			name:       "assertion-then-call",
			why:        "an assertion wrapping a reference that is then called, which is upstream's own recursion arm but which its corpus exercises only in the reporting direction",
			sourceText: "declare const o: { m(): void };\n(o.m as () => void)();\n",
			wantIds:    []string{},
			wantSpans:  []string{},
		},
		{
			name:       "nonnull-then-call",
			why:        "the non-null wrapper form of the row above",
			sourceText: "declare const o: { m?(): void };\no.m!();\n",
			wantIds:    []string{},
			wantSpans:  []string{},
		},
		{
			name:       "assertion-detach",
			why:        "an assertion around a detached reference, which reports; paired with the two rows above so the recursion is measured in both directions rather than only where it stays silent",
			sourceText: "declare const o: { m(): void };\nconst f = o.m as () => void;\n",
			wantIds:    []string{"unboundWithoutThisAnnotation"},
			wantSpans:  []string{"o.m"},
		},
		{
			name:       "control-detach",
			why:        "the control: a bare detached reference, which reports",
			sourceText: "declare const o: { m(): void };\nconst f = o.m;\n",
			wantIds:    []string{"unboundWithoutThisAnnotation"},
			wantSpans:  []string{"o.m"},
		},
		{
			name:       "control-call",
			why:        "the control: a bare call, which does not",
			sourceText: "declare const o: { m(): void };\no.m();\n",
			wantIds:    []string{},
			wantSpans:  []string{},
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := runUnboundMethod(t, testCase.sourceText, "")
			rule_testing.ExpectFindings(t, result, testCase.wantIds...)
			if len(result.Diagnostics) != len(testCase.wantSpans) {
				t.Fatalf("wanted %d findings, got %d, on the case that covers: %s",
					len(testCase.wantSpans), len(result.Diagnostics), testCase.why)
			}
			written := asTheTypedUnboundMethodHarnessWroteIt(testCase.sourceText)
			for findingIndex, wantSpan := range testCase.wantSpans {
				reported := result.Diagnostics[findingIndex]
				gotSpan := written[reported.Range.Pos():reported.Range.End()]
				if gotSpan != wantSpan {
					t.Errorf("finding %d points at %q, wanted %q (%s)",
						findingIndex, gotSpan, wantSpan, testCase.why)
				}
			}
		})
	}
}

// TestUnboundMethodOnThisAnnotationsAndRepeatedConstituents covers two discriminations upstream
// tests only in the extremes.
//
// The first is the `this` parameter. Upstream writes `this: void` and methods with no `this` at all,
// and never a `this` typed as a third thing, so the test for the void KEYWORD specifically can be
// widened to accept any annotation and every imported row stays green. It matters because such a
// method is still dangerous and takes the shorter message: the author has already reasoned about
// `this`, so the advice to annotate it would be wrong.
//
// The second is a union whose constituents all report. Upstream stops at the first, and its corpus
// has no union in that position, so deleting the stop survives everything imported and then reports
// twice on one reference the first time real code has one.
//
// Measured against the installed 8.x build, one program per case, with a control in the same run.
func TestUnboundMethodOnThisAnnotationsAndRepeatedConstituents(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		why        string
		sourceText string
		wantIds    []string
		wantSpans  []string
	}{
		{
			name:       "this-typed-not-void",
			why:        "a `this` parameter typed as something OTHER than void. It is still dangerous, and it takes the SHORT message, because the author has already thought about `this` here and telling them to annotate it would be wrong. Upstream's corpus tests `this: void` and no `this` at all, never a third type, so widening the void test to accept any annotation survives every imported row",
			sourceText: "interface I {\n  m(this: I): void;\n}\ndeclare const i: I;\nconst f = i.m;\n",
			wantIds:    []string{"unbound"},
			wantSpans:  []string{"i.m"},
		},
		{
			name:       "this-typed-unknown",
			why:        "a second spelling of the row above, so a port special-casing one type name is not satisfied by it",
			sourceText: "interface I {\n  m(this: unknown): void;\n}\ndeclare const i: I;\nconst f = i.m;\n",
			wantIds:    []string{"unbound"},
			wantSpans:  []string{"i.m"},
		},
		{
			name:       "this-typed-void",
			why:        "the void half, which is the only annotation that makes the method safe",
			sourceText: "interface I {\n  m(this: void): void;\n}\ndeclare const i: I;\nconst f = i.m;\n",
			wantIds:    []string{},
			wantSpans:  []string{},
		},
		{
			name:       "this-untyped",
			why:        "a `this` parameter with no type at all, which is present but not void, so it reports with the short message like the first two rows",
			sourceText: "interface I {\n  m(this): void;\n}\ndeclare const i: I;\nconst f = i.m;\n",
			wantIds:    []string{"unbound"},
			wantSpans:  []string{"i.m"},
		},
		{
			name:       "union-two-dangerous",
			why:        "a union whose constituents are BOTH dangerous, which upstream reports exactly ONCE. Its corpus has no such union, so removing the stop-at-first-report survives every imported row and then reports twice on one reference in real code",
			sourceText: "interface A {\n  m(): void;\n}\ninterface B {\n  m(): void;\n}\ndeclare const x: A | B;\nconst f = x.m;\n",
			wantIds:    []string{"unboundWithoutThisAnnotation"},
			wantSpans:  []string{"x.m"},
		},
		{
			name:       "control-u13",
			why:        "the control: an ordinary method with no `this` parameter, which takes the long message",
			sourceText: "interface I {\n  m(): void;\n}\ndeclare const i: I;\nconst f = i.m;\n",
			wantIds:    []string{"unboundWithoutThisAnnotation"},
			wantSpans:  []string{"i.m"},
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := runUnboundMethod(t, testCase.sourceText, "")
			rule_testing.ExpectFindings(t, result, testCase.wantIds...)
			if len(result.Diagnostics) != len(testCase.wantSpans) {
				t.Fatalf("wanted %d findings, got %d, on the case that covers: %s",
					len(testCase.wantSpans), len(result.Diagnostics), testCase.why)
			}
			written := asTheTypedUnboundMethodHarnessWroteIt(testCase.sourceText)
			for findingIndex, wantSpan := range testCase.wantSpans {
				reported := result.Diagnostics[findingIndex]
				gotSpan := written[reported.Range.Pos():reported.Range.End()]
				if gotSpan != wantSpan {
					t.Errorf("finding %d points at %q, wanted %q (%s)",
						findingIndex, gotSpan, wantSpan, testCase.why)
				}
			}
		})
	}
}

// TestUnboundMethodStopsAtTheFirstReportingPropertyName pins upstream's stop over property names.
//
// A computed access can name several properties at once, because its subscript type may be a union
// of literals, and upstream asks about each in turn and stops at the first that reports. The stop is
// over NAMES, not over union constituents of the object, which is easy to misread: the first fixture
// written for this mutation used a union OBJECT with two dangerous constituents and could not see
// the line at all, because the checker hands a union one synthetic property and the inner walk
// already stops on its own.
//
// Upstream writes no computed access naming two dangerous properties, so deleting the stop survives
// all two hundred and eleven imported rows and then reports twice on one reference.
//
// Measured against the installed 8.x build, one program per case.
func TestUnboundMethodStopsAtTheFirstReportingPropertyName(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		why        string
		sourceText string
		wantIds    []string
		wantSpans  []string
	}{
		{
			name:       "computed-two-dangerous-names",
			why:        "a COMPUTED access whose subscript type names two properties, BOTH dangerous. Upstream stops at the first that reports, so this is one finding rather than two, and the stop is over property NAMES rather than over union constituents. Nothing in upstream's corpus has a computed access naming two dangerous properties, so removing the stop survives every imported row and then double-reports the first time real code has one",
			sourceText: "interface I {\n  a(): void;\n  b(): void;\n}\ndeclare const i: I;\ndeclare const k: 'a' | 'b';\nconst f = i[k];\n",
			wantIds:    []string{"unboundWithoutThisAnnotation"},
			wantSpans:  []string{"i[k]"},
		},
		{
			name:       "computed-one-dangerous-name",
			why:        "the same access where only one of the two names is dangerous, which reaches the stop without needing it and is the row that keeps the one above from being satisfied by a rule that stopped after the first name unconditionally",
			sourceText: "interface I {\n  a(): void;\n  b: () => void;\n}\ndeclare const i: I;\ndeclare const k: 'a' | 'b';\nconst f = i[k];\n",
			wantIds:    []string{"unboundWithoutThisAnnotation"},
			wantSpans:  []string{"i[k]"},
		},
		{
			name:       "computed-single-name",
			why:        "a computed access naming one property, the baseline the two rows above are read against",
			sourceText: "interface I {\n  a(): void;\n}\ndeclare const i: I;\ndeclare const k: 'a';\nconst f = i[k];\n",
			wantIds:    []string{"unboundWithoutThisAnnotation"},
			wantSpans:  []string{"i[k]"},
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := runUnboundMethod(t, testCase.sourceText, "")
			rule_testing.ExpectFindings(t, result, testCase.wantIds...)
			if len(result.Diagnostics) != len(testCase.wantSpans) {
				t.Fatalf("wanted %d findings, got %d, on the case that covers: %s",
					len(testCase.wantSpans), len(result.Diagnostics), testCase.why)
			}
			written := asTheTypedUnboundMethodHarnessWroteIt(testCase.sourceText)
			for findingIndex, wantSpan := range testCase.wantSpans {
				reported := result.Diagnostics[findingIndex]
				gotSpan := written[reported.Range.Pos():reported.Range.End()]
				if gotSpan != wantSpan {
					t.Errorf("finding %d points at %q, wanted %q (%s)",
						findingIndex, gotSpan, wantSpan, testCase.why)
				}
			}
		})
	}
}

// TestUnboundMethodRequiresTheTypedHarness pins the checker guards.
//
// All three listeners start with a nil check, and the shim answers nil from a type query on a nil
// checker rather than panicking, so a rule missing those guards does not crash. It goes silent, and
// a vacuous green is the more dangerous of the two failures because nothing announces it. This
// asserts the untyped harness produces nothing on an input the typed one reports, so a later revert
// to rule_testing.Run fails loudly rather than quietly.
func TestUnboundMethodRequiresTheTypedHarness(t *testing.T) {
	t.Parallel()

	const sourceText = "declare const o: { m(): void };\nconst f = o.m;\n"
	rule_testing.ExpectClean(t, rule_testing.Run(t, UnboundMethod, unboundMethodFile, sourceText))
	rule_testing.ExpectFindings(t, rule_testing.RunTyped(t, UnboundMethod, unboundMethodFile,
		sourceText), "unboundWithoutThisAnnotation")
}

// TestUnboundMethodRendersUpstreamsMessageText asserts what a reader is told, for both messages.
//
// rule.Message is {Id, Description} with no interpolation, so there is nothing to render; what there
// is to get wrong is the text, and ExpectFindings compares ids and count and nothing else. The
// wanted strings are typed as literals rather than read from the rule's own constants, because a
// comparison against the constant moves with any mutation of it.
func TestUnboundMethodRendersUpstreamsMessageText(t *testing.T) {
	t.Parallel()

	const base = "A method that is not declared with `this: void` may cause unintentional scoping " +
		"of `this` when separated from its object.\n" +
		"Consider using an arrow function or explicitly `.bind()`ing the method to avoid calling " +
		"the method with an unintended `this` value. "

	cases := []struct {
		name        string
		sourceText  string
		wantId      string
		wantMessage string
	}{
		{
			name:        "without a this annotation",
			sourceText:  "declare const o: { m(): void };\nconst f = o.m;\n",
			wantId:      "unboundWithoutThisAnnotation",
			wantMessage: base + "\nIf a function does not access `this`, it can be annotated with `this: void`.",
		},
		{
			name:        "with a this annotation",
			sourceText:  "interface I {\n  m(this: I): void;\n}\ndeclare const i: I;\nconst f = i.m;\n",
			wantId:      "unbound",
			wantMessage: base,
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := runUnboundMethod(t, testCase.sourceText, "")
			if len(result.Diagnostics) != 1 {
				t.Fatalf("wanted one finding, got %d", len(result.Diagnostics))
			}
			reported := result.Diagnostics[0]
			if reported.Message.Id != testCase.wantId {
				t.Errorf("reported id %q, wanted %q", reported.Message.Id, testCase.wantId)
			}
			if reported.Message.Description != testCase.wantMessage {
				t.Errorf("reported message %q, wanted %q",
					reported.Message.Description, testCase.wantMessage)
			}
		})
	}
}
