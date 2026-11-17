package lower

import (
	"strings"
	"testing"
)

// An overload whose implementation writes its parameter through a shorthand destructuring, ({ x } = other),
// before returning it: censusReturnProof found writes through GetSymbolAtLocation, which on a shorthand name
// answers the property, so the reassigned parameter was taken for the argument the overload typed, the narrow
// result proven and the call's check erased (a 1 holding 7, and a Narrow read past a Wide's fields). The
// write now disables the proof: in .ts the call keeps its check and stops, as a plain write already did, and
// in .a the set is refused. Mutant: the shorthand arm dropped, and both rows lower unchecked.
const overloadShorthandWriteShape = "const other: { x: number } = { x: 7 };\nfunction pick(x: 1): 1;\nfunction pick(x: number): number {\n    ({ x } = other);\n    return x;\n}\nconst one: 1 = pick(1);\nconsole.log(`${one === 1 ? \"one\" : \"other\"} ${one + 1}`);\n"

func TestOverloadShorthandWriteKeepsTheCheck(t *testing.T) {
	t.Parallel()
	program, err := lowerSourceAs(t, overloadShorthandWriteShape, ".ts")
	if err != nil {
		t.Fatalf("want it lowered with the call's result check, got %v", err)
	}
	code, stdout, stderr := brandedNative(t, program)
	if code != 70 || stdout != "" || !strings.Contains(stderr, "overload 1 of pick result at ") || !strings.Contains(stderr, ": expected 1") {
		t.Fatalf("want the check to stop the 7 at the call, got exit %d, stdout %q, stderr %q", code, stdout, stderr)
	}
}

func TestOverloadShorthandWriteIsRefusedInA(t *testing.T) {
	t.Parallel()
	_, err := lowerSourceAs(t, overloadShorthandWriteShape, ".a")
	if err == nil || !strings.Contains(err.Error(), "overload 1 of pick result") {
		t.Fatalf("want overload 1 of pick refused at its declaration for its result, got %v", err)
	}
}
