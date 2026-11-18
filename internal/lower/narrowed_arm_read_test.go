package lower

import (
	"errors"
	"strings"
	"testing"
)

const narrowedArmString = `interface U { kind: number; v: number }
interface L { kind: number; v: string }
type N = U | L;
function makeU(): N { return { kind: 1, v: 7 }; }
function makeL(): L { return { kind: 2, v: "abc" }; }
let flow: N = makeL();
`

const narrowedArmRead = "console.log(String(flow.v.length));\n"

func TestNarrowedArmReadOfACallWrittenLetIsNotYet(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, source, where, what string }{
		{
			"number-as-string",
			narrowedArmString + "function replace(): void { flow = makeU(); }\nreplace();\n" + narrowedArmRead,
			":9:20",
			"a read of flow narrowed to L where a call may have rewritten it (flow is assigned inside replace)",
		},
		{
			"map-as-array",
			`interface U { kind: number; v: Map<string, number> }
interface L { kind: number; v: string[] }
type N = U | L;
function makeU(): N { return { kind: 1, v: new Map([["a", 1]]) }; }
function makeL(): L { return { kind: 2, v: ["x", "y"] }; }
let flow: N = makeL();
function replace(): void { flow = makeU(); }
replace();
` + narrowedArmRead,
			":9:20",
			"a read of flow narrowed to L where a call may have rewritten it (flow is assigned inside replace)",
		},
		{
			"strips-undefined",
			`class Box { n: number = 1; }
let x: Box | undefined = new Box();
function clear(): void { x = undefined; }
clear();
console.log(String(x.n));
`,
			":5:20",
			"a read of x narrowed to Box where a call may have rewritten it (x is assigned inside clear)",
		},
		{
			"loop-back-edge",
			`class Box { n: number = 1; }
function run(start: Box | undefined): number {
	let x: Box | undefined = start;
	const clear = (): void => { x = undefined; };
	if (!x) return 0;
	let result = 0;
	for (let i = 0; i < 2; i++) {
		result = x.n;
		clear();
	}
	return result;
}
console.log(String(run(new Box())));
`,
			":8:12",
			"a read of x narrowed to Box where a call may have rewritten it (x is assigned inside a function)",
		},
		{
			"call-after-assignment",
			`class Box { n: number = 1; }
let x: Box | undefined;
function clear(): void { x = undefined; }
x = new Box();
clear();
console.log(String(x.n));
`,
			":6:20",
			"a read of x narrowed to Box where a call may have rewritten it (x is assigned inside clear)",
		},
		{
			"conditional-assignment-may-skip",
			`class Box { n: number = 1; }
function run(c: boolean): number {
	let x: Box | undefined = new Box();
	const clear = (): void => { x = undefined; };
	clear();
	if (c) { x = new Box(); }
	return x.n;
}
console.log(String(run(false)));
`,
			":7:9",
			"a read of x narrowed to Box where a call may have rewritten it (x is assigned inside a function)",
		},
		{
			"enum-switch-literal",
			`enum E { A, B }
let value: E = E.A;
function replace(): void { value = E.B; }
replace();
switch (value) {
case E.A: console.log("a"); break;
default: console.log("b"); break;
}
`,
			":5:9",
			"a read of value narrowed to E.A where a call may have rewritten it (value is assigned inside replace)",
		},
		{
			"computed-field-literal",
			`const values: { 0: string; 1: string } = { 0: "a", 1: "b" };
let key: 0 | 1 = 0;
function replace(): void { key = 1; }
replace();
console.log(values[key]);
`,
			":5:20",
			"a read of key narrowed to 0 where a call may have rewritten it (key is assigned inside replace)",
		},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			for _, extension := range []string{".a", ".ts"} {
				t.Run(extension[1:], func(t *testing.T) {
					t.Parallel()
					_, err := lowerSourceAs(t, probe.source, extension)
					var notYet *NotYet
					if !errors.As(err, &notYet) || !strings.HasSuffix(notYet.Where, "main"+extension+probe.where) || notYet.What != probe.what {
						t.Fatalf("got %v, want main%s%s: stage 0 can't lower %s yet", err, extension, probe.where, probe.what)
					}
				})
			}
		})
	}
}

func TestNarrowedArmReadOfAnUnwrittenLetLowers(t *testing.T) {
	t.Parallel()
	for _, extension := range []string{".a", ".ts"} {
		t.Run(extension[1:], func(t *testing.T) {
			t.Parallel()
			if _, err := lowerSourceAs(t, narrowedArmString+"makeU();\n"+narrowedArmRead, extension); err != nil {
				t.Fatalf("want the narrowed read to lower, got %v", err)
			}
		})
	}
}

func TestNarrowedArmReadOfCallWrittenScalarAgrees(t *testing.T) {
	t.Parallel()
	lowersAndAgreesWithNode(t, `enum E { A, B }
let value: E = E.A;
function replace(): void { value = E.B; }
replace();
console.log(String(value));
`)
}

func TestNarrowedArmReadWithNoInterveningCallLowers(t *testing.T) {
	t.Parallel()
	const source = `class Box { n: number = 1; }
function absent(): undefined { return undefined; }
function run(start: Box | undefined): number {
	let x: Box | undefined = start;
	const clear = (): void => { x = absent(); };
	if (!x) return 0;
	return x.n;
}
console.log(String(run(new Box())));
`
	for _, extension := range []string{".a", ".ts"} {
		t.Run(extension[1:], func(t *testing.T) {
			t.Parallel()
			lowersAndAgreesWithNodeAs(t, source, extension)
		})
	}
}

// The former over-refusal lowers because makeL completes before it writes flow.
// The writer's call to makeU stays in the uncalled closure.
func TestNarrowedArmReadWithUncalledClosureLowers(t *testing.T) {
	t.Parallel()
	const source = narrowedArmString + "const replace = (): void => { flow = makeU(); };\n" + narrowedArmRead
	for _, extension := range []string{".a", ".ts"} {
		t.Run(extension[1:], func(t *testing.T) {
			t.Parallel()
			lowersAndAgreesWithNodeAs(t, source, extension)
		})
	}
}

func TestNarrowedArmScalarBooleanBranchAgrees(t *testing.T) {
	t.Parallel()
	program := lowersAndAgreesWithNode(t, `let seen: boolean = true;
function clear(): void { seen = false; }
clear();
if (!seen) console.log("no"); else console.log("yes");
`)
	if code, stdout, stderr := brandedNative(t, program); code != 0 || stdout != "no\n" || stderr != "" {
		t.Fatalf("native exit %d, stdout %q, stderr %q; want stdout no", code, stdout, stderr)
	}
}
