package javascript

import "testing"

func TestViewCallablesNode(t *testing.T) {
	t.Parallel()
	runtime := viewTestRuntime + fieldReadinessRuntime + viewCallablesRuntime
	for _, test := range []struct {
		name, source, out, err string
		exit                   int
	}{
		{"field", `const object = {run: new AdamicClosure((self, arguments_) => arguments_[0] + 1)}; const f = adamicViewCallableRead(object, "run", "node.run"); console.log(adamicViewCallableCall(f, object, [7], false));`, "8\n", "", 0},
		{"method", `const prototype = {run: (self, n) => self.base + n}; const object = Object.assign(Object.create(prototype), {base: 3}); const f = adamicViewCallableRead(object, "run", "node.run", true); console.log(adamicViewCallableCall(f, object, [7], true));`, "10\n", "", 0},
		{"missing", `adamicViewCallableRead({}, "run", "node.run");`, "", "adamic: panic: cast failed: field read failed: node.run is not initialized; expected function, found missing\n", 70},
		{"uninitialized", `const object = adamicUninitializedFields({run: new AdamicClosure(() => 7)}, ["run"]); adamicViewCallableRead(object, "run", "node.run");`, "", "adamic: panic: cast failed: field read failed: node.run is not initialized; expected function, found uninitialized\n", 70},
		{"kind", `adamicViewCallableRead({run: 7}, "run", "node.run");`, "", "adamic: panic: cast failed: field read failed: node.run is not a function; expected function, found number\n", 70},
		{"once", `let reads = 0; const receiver = () => { reads++; return {run: new AdamicClosure(() => 7)}; }; console.log(adamicViewCallableCall(adamicViewCallableRead(receiver(), "run", "node.run"), undefined, [], false), reads);`, "7 1\n", "", 0},
	} {
		t.Run(test.name, func(t *testing.T) { runViewNode(t, runtime+test.source, test.out, test.err, test.exit) })
	}
	runViewNode(t, `const object = {run: n => n + 1}; console.log(object.run(7));`, "8\n", "", 0)
	runViewNode(t, `class Example { base = 3; run(n) { return this.base + n; } } console.log(new Example().run(7));`, "10\n", "", 0)
}
