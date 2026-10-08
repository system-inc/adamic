package lower

import (
	"errors"
	"github.com/system-inc/adamic/internal/ir"
	"strings"
	"testing"
)

func TestViewFrameFamilyDescriptors(t *testing.T) {
	for _, probe := range []struct{ declaration, family string }{
		{"number[]", "array"}, {"(value:number)=>number", "callable"},
		{"{readonly [key:string]:number}", "dictionary"},
		{"{readonly a:number}&{readonly b:number}", "intersection"},
		{"string|number", "union"},
		{"Uint8Array", "dictionary"},
		{"{}", "representation conversion"},
	} {
		t.Run(probe.family, func(t *testing.T) {
			program, err := lowerSource(t, `interface Base {readonly kind:'box'|'other'}
interface Box extends Base {readonly kind:'box';readonly opaque:`+probe.declaration+`}
function make():Base {const kind:'box'='box';const raw={kind};return raw}
function view(value:Base):Box{return value as Box}
console.log(view(make()).kind);`)
			if probe.declaration == "{readonly [key:string]:number}" {
				var refused *Refused
				if !errors.As(err, &refused) || !strings.Contains(err.Error(), "an index signature") {
					t.Fatalf("V1 must retain the base index-signature refusal: %v", err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			found := false
			for _, descriptor := range program.ViewContracts {
				if descriptor.Unsupported == probe.family {
					found = true
					if descriptor.Kind != ir.ViewUnknown {
						t.Fatalf("unsupported descriptor certifies a kind: %#v", descriptor)
					}
				}
			}
			if !found {
				t.Fatalf("missing unavailable %s descriptor", probe.family)
			}
			if viewErasureProof(program, ir.Property{}) {
				t.Fatal("V1 has no erasure proof")
			}
		})
	}
}
