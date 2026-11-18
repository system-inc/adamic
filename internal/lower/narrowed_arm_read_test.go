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
			// Named over-refusal: the lift waits on proving the closure is never called.
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
