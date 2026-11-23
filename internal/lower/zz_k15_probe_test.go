package lower

import "testing"

// A probe, never merged: WebAssembly's k15 on #t2p4wyv, a tag rewritten through a writable wider type.
func TestZZProbeK15(t *testing.T) {
	for _, probe := range []struct{ name, source string }{
		{"retag before the test", kindUnionOwnModel + kindUnionOwnMake + `function retag(w: { kind: string }): void { w.kind = "a"; }
const n = make(1) as A | B | C;
retag(n);
if (n.kind !== "c") {
  n.f = 5;
  console.log(String(n.f));
}
`},
		{"retag between test and write", kindUnionOwnModel + kindUnionOwnMake + `function retag(w: { kind: string }): void { w.kind = "a"; }
const n = make(1) as A | B | C;
if (n.kind !== "a") {
  retag(n);
  n.f = 5;
  console.log(String(n.f));
}
`},
		{"retag through an alias, no call", kindUnionOwnModel + kindUnionOwnMake + `const n = make(1) as A | B | C;
const w: { kind: string } = n;
if (n.kind !== "a") {
  w.kind = "a";
  n.f = 5;
  console.log(String(n.f));
}
`},
	} {
		_, err := lowerSourceAs(t, probe.source, ".ts")
		t.Logf("PROBE %s => %v", probe.name, err)
	}
}
