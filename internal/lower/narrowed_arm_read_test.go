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

const narrowedArmTagged = `interface U { kind: 1; v: number }
interface L { kind: 2; v: string }
type N = U | L;
function makeU(): U { return { kind: 1, v: 7 }; }
function makeL(): L { return { kind: 2, v: "abc" }; }
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
			"three-arms-to-one",
			narrowedArmTagged + "let flow: U | L | undefined = makeL();\nfunction replace(): void { flow = makeU(); }\nreplace();\n" + narrowedArmRead,
			":9:20",
			"a read of flow narrowed to L where a call may have rewritten it (flow is assigned inside replace)",
		},
		{
			"loop-back-edge",
			narrowedArmTagged + `function run(start: N): number {
	let flow: N = start;
	const replace = (): void => { flow = makeU(); };
	if (flow.kind !== 2) return 0;
	let result = 0;
	for (let i = 0; i < 2; i++) {
		result = flow.v.length;
		replace();
	}
	return result;
}
console.log(String(run(makeL())));
`,
			":12:12",
			"a read of flow narrowed to L where a call may have rewritten it (flow is assigned inside a function)",
		},
		{
			"call-after-assignment",
			narrowedArmTagged + "let flow: N;\nfunction replace(): void { flow = makeU(); }\nflow = makeL();\nreplace();\n" + narrowedArmRead,
			":10:20",
			"a read of flow narrowed to L where a call may have rewritten it (flow is assigned inside replace)",
		},
		{
			"conditional-assignment-may-skip",
			narrowedArmTagged + `function run(c: boolean): number {
	let flow: N = makeL();
	const replace = (): void => { flow = makeU(); };
	replace();
	if (c) { flow = makeL(); }
	return flow.v.length;
}
console.log(String(run(false)));
`,
			":11:9",
			"a read of flow narrowed to L where a call may have rewritten it (flow is assigned inside a function)",
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
	const source = narrowedArmTagged + `function run(start: N): number {
	let flow: N = start;
	const replace = (): void => { flow = makeU(); };
	if (flow.kind !== 2) return 0;
	return flow.v.length;
}
console.log(String(run(makeL())));
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

func TestNarrowedArmReadThatOnlyStripsUndefinedAgrees(t *testing.T) {
	t.Parallel()
	const source = `class Box { n: number = 1; }
let x: Box | undefined = new Box();
function clear(): void { x = undefined; }
clear();
console.log(String(x.n));
`
	const typeError = "TypeError: Cannot read properties of undefined (reading 'n')"
	for _, extension := range []string{".a", ".ts"} {
		t.Run(extension[1:], func(t *testing.T) {
			t.Parallel()
			path := writeAgreementSource(t, "main"+extension, source)
			truth := runAgreementNode(t, path)
			if truth.code == 0 || len(truth.stdout) != 0 || !strings.Contains(string(truth.stderr), typeError) {
				t.Fatalf("source Node: exit %d, stdout %q, stderr %q; want %s", truth.code, truth.stdout, truth.stderr, typeError)
			}
			brandedStopsProgram(t, lowerAgreementEntry(t, path), "adamic: panic: "+typeError+"\n")
		})
	}
}
