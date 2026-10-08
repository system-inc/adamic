package lower

import (
	"strings"
	"testing"
)

func TestOrdinaryPrimitiveAdmission(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		"const point = {x:1}; console.log(`${point}`);",
		"console.log(String({value:1}));",
		"console.log(`${null} ${undefined}`);",
		"const o = {valueOf() {return {};}, toString() {return {};}}; try { console.log(''+o); } catch(e) { if(e instanceof Error) console.log(e.message); }",
	} {
		t.Run(source, func(t *testing.T) {
			if _, err := lowerSource(t, source); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestOrdinaryPrimitiveBoundaries(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ source, reason string }{
		{"console.log(`${new Error('host')}`);", "host object with intrinsic conversion behavior"},
		{"const pair:[number,number] = [1,2]; console.log(`${pair}`);", "tuple held as an object"},
		{"function show(o:{x:number;toString?:()=>string}) {console.log(`${o}`);} show({x:1,toString:()=> 'yes'});", "possibly absent conversion method"},
		{"const o = {get toString() {return ()=> 'getter';}}; console.log(`${o}`);", "conversion accessor"},
	} {
		t.Run(probe.reason, func(t *testing.T) {
			_, err := lowerSource(t, probe.source)
			if err == nil || !strings.Contains(err.Error(), probe.reason) {
				t.Fatalf("want %s, got %v", probe.reason, err)
			}
		})
	}
}
