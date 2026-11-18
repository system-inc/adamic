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
			// Named over-refusal: the source-order scan counts makeL(), the call in the initializer.
			"uncalled-closure",
			narrowedArmString + "const replace = (): void => { flow = makeU(); };\n" + narrowedArmRead,
			":8:20",
			"a read of flow narrowed to L where a call may have rewritten it (flow is assigned inside a function)",
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
			if _, err := lowerSourceAs(t, narrowedArmString+narrowedArmRead, extension); err != nil {
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
