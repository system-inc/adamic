package javascript

import "testing"

const viewNumberCheck = `const numberCheck = (value, expression) => {
    if (typeof value !== "number") panic("element read failed: " + expression + " expected number, found " + typeof value);
    return value;
};
`

func TestViewArraysNode(t *testing.T) {
	t.Parallel()
	runtime := viewTestRuntime + fieldReadinessRuntime + viewArraysRuntime + viewNumberCheck
	for _, test := range []struct {
		name, source, out, err string
		exit                   int
	}{
		{"lazy", `const a = [7, "bad"]; const v = adamicViewArray(a, "node.items", "number[]"); console.log(v === a, v.length); console.log(adamicViewArrayRead(v, 0, "node.items", numberCheck));`, "true 2\n7\n", "", 0},
		{"second", `const v = adamicViewArray([7, "bad"], "node.items", "number[]"); console.log(adamicViewArrayRead(v, 0, "node.items", numberCheck)); adamicViewArrayRead(v, 1, "node.items", numberCheck);`, "7\n", "adamic: panic: element read failed: node.items[1] expected number, found string\n", 70},
		{"kind", `adamicViewArray({length: 1, 0: 7}, "node.items", "number[]");`, "", "adamic: panic: field read failed: node.items expected number[], found object\n", 70},
		{"alias", `const a = [7]; const v = adamicViewArray(a, "node.items", "readonly number[]"); a[0] = "bad"; adamicViewArrayRead(v, 0, "node.items", numberCheck);`, "", "adamic: panic: element read failed: node.items[0] expected number, found string\n", 70},
		{"object", `const child = {name: "ok"}; const a = [child]; const v = adamicViewArray(a, "node.items", "Child[]"); const read = (value, expression) => { if (value === null || typeof value !== "object") panic(expression); return value; }; const element = adamicViewArrayRead(v, 0, "node.items", read); console.log(element === child, adamicViewField(element, "name", "node.items[0].name", 3));`, "true ok\n", "", 0},
		{"object payload", `const a = [{name: 7}]; const element = adamicViewArrayRead(adamicViewArray(a, "node.items", "Child[]"), 0, "node.items", value => value); adamicViewField(element, "name", "node.items[0].name", 3);`, "", "adamic: panic: field read failed: node.items[0].name is not a string; expected string, found number\n", 70},
		{"hole", `const a = new Array(1); adamicViewArrayRead(adamicViewArray(a, "node.items", "number[]"), 0, "node.items", numberCheck);`, "", "adamic: panic: element read failed: node.items[0] expected number, found undefined\n", 70},
		{"optional index", `const a = []; console.log(adamicViewArrayRead(adamicViewArray(a, "node.items", "number[]"), 1, "node.items", (value, expression) => value === undefined ? value : numberCheck(value, expression)));`, "undefined\n", "", 0},
		{"once", `let receivers = 0, indices = 0, checks = 0; const read = value => { checks++; return value; }; console.log(adamicViewArrayRead((receivers++, [7]), (indices++, 0), "node.items", read), receivers, indices, checks);`, "7 1 1 1\n", "", 0},
	} {
		t.Run(test.name, func(t *testing.T) { runViewNode(t, runtime+test.source, test.out, test.err, test.exit) })
	}
	// Positive behavior is also held against Node without the helpers.
	runViewNode(t, `const a = [7, "bad"]; console.log(a === a, a.length); console.log(a[0]);`, "true 2\n7\n", "", 0)
	runViewNode(t, `const child = {name: "ok"}; const a = [child]; console.log(a[0] === child, a[0].name);`, "true ok\n", "", 0)
}

func TestViewArrayFieldPresence(t *testing.T) {
	t.Parallel()
	runtime := viewTestRuntime + fieldReadinessRuntime + viewArraysRuntime
	read := emitViewArrayFieldRead("object", "values", "node.values", "readonly number[]")
	runViewNode(t, runtime+`const object = {values: [7]}; console.log(`+read+`.length);`, "1\n", "", 0)
	runViewNode(t, runtime+`const object = {}; `+read+`;`, "", "adamic: panic: field read failed: node.values is not initialized; expected readonly number[], found missing\n", 70)
	runViewNode(t, runtime+`const object = adamicUninitializedFields({values: [7]}, ["values"]); `+read+`;`, "", "adamic: panic: field read failed: node.values is not initialized; expected readonly number[], found uninitialized\n", 70)
}
