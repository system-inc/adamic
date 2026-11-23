package lower

import "testing"

// A probe, never merged: Lower's k16 and k17 on #t2p4wyv, a tag rewritten with no view, by Object.assign or
// Object.defineProperty, between the tag test and the write, and before the test.
func TestZZProbeK16K17(t *testing.T) {
	for _, probe := range []struct{ name, source string }{
		{"k16 assign between", kindUnionOwnModel + kindUnionOwnMake + `const n = make(0) as A | B | C;
if (n.kind !== "c") {
  Object.assign(n, { kind: "c" });
  n.f = 5;
  console.log(String(n.f));
}
`},
		{"k16 assign before", kindUnionOwnModel + kindUnionOwnMake + `const n = make(1) as A | B | C;
Object.assign(n, { kind: "a" });
if (n.kind !== "c") {
  n.f = 5;
  console.log(String(n.f));
}
`},
		{"k17 define between", kindUnionOwnModel + kindUnionOwnMake + `const n = make(0) as A | B | C;
if (n.kind !== "c") {
  Object.defineProperty(n, "kind", { value: "c" });
  n.f = 5;
  console.log(String(n.f));
}
`},
		{"k17 define before", kindUnionOwnModel + kindUnionOwnMake + `const n = make(1) as A | B | C;
Object.defineProperty(n, "kind", { value: "a" });
if (n.kind !== "c") {
  n.f = 5;
  console.log(String(n.f));
}
`},
	} {
		_, err := lowerSourceAs(t, probe.source, ".ts")
		t.Logf("PROBE %s => %v", probe.name, err)
	}
}
