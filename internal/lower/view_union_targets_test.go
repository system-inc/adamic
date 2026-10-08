package lower

import (
	"errors"
	"github.com/system-inc/adamic/internal/ir"
	"testing"
)

func TestViewUnionTargetAdmission(t *testing.T) {
	for _, probe := range []struct {
		name, source string
		admitted     bool
	}{
		{"broad readonly source", "interface Base { readonly kind: string }; type A={readonly kind:'a';readonly n:number};type B={readonly kind:'b';readonly s:string};function f(x:Base):A|B{return x as A|B;}", true},
		{"duplicate target tags", "interface Base { readonly kind: string }; type A={readonly kind:'a';readonly n:number};type B={readonly kind:'a';readonly s:string};function f(x:Base):A|B{return x as A|B;}", false},
		{"mutable source tag", "interface Base { kind: string }; type A={kind:'a';readonly n:number};type B={kind:'b';readonly s:string};function f(x:Base):A|B{return x as A|B;}", false},
		{"mutable target tag", "interface Base { readonly kind: string }; type A={kind:'a';readonly n:number};type B={kind:'b';readonly s:string};function f(x:Base):A|B{return x as A|B;}", false},
		{"mixed target representations", "interface Base { readonly kind: string }; type A={readonly kind:'a';readonly n:number};function f(x:Base):A|string{return x as A|string;}", false},
		{"incompatible source slot", "interface Base { readonly kind: string;readonly n:number }; type A={readonly kind:'a';readonly n:number};type B={readonly kind:'b';readonly n:string};function f(x:Base):A|B{return x as A|B;}", false},
	} {
		t.Run(probe.name, func(t *testing.T) {
			program, err := lowerSource(t, probe.source)
			if !probe.admitted {
				var refused *Refused
				if !errors.As(err, &refused) {
					t.Fatalf("unsafe union target admitted: %v", err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			count := 0
			for _, f := range program.Functions {
				walk(f.Body, func(node any) bool {
					if cast, ok := node.(ir.CheckedCast); ok {
						count++
						if len(cast.Allowed) != 2 || !cast.CheckedFields {
							t.Fatalf("incomplete union guard: %#v", cast)
						}
					}
					return true
				})
			}
			if count != 1 {
				t.Fatalf("want one complete tag guard, got %d", count)
			}
		})
	}
}
