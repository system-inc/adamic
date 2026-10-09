package lower

import (
	"errors"
	"strings"
	"testing"
)

func TestPartialRecordStorageRefusals(t *testing.T) {
	for _, probe := range []struct{ source, reason string }{
		{`const fixed={first:1}; const cache:Partial<Record<'first'|'second',number>>=fixed;`, "seen as"},
		{`const cache:Partial<Record<'first'|'second',number>>={}; const fixed:{first?:number}=cache;`, "seen as"},
		{`interface Animal {name:string} interface Dog extends Animal {bark:string} const cache:Partial<Record<'first'|'second',Dog>>={}; const view:Record<string,Animal>=cache;`, "invariant-mutable"},
		{`const cache:Partial<Record<'first'|'second',number>>={}; const view:Record<string,number|undefined>=cache;`, "invariant-mutable"},
	} {
		t.Run(probe.source, func(t *testing.T) {
			_, err := lowerSource(t, probe.source)
			var ny *NotYet
			var refused *Refused
			if err == nil || (!errors.As(err, &ny) && !errors.As(err, &refused)) || !strings.Contains(err.Error(), probe.reason) {
				t.Fatalf("got %v, want %s refusal", err, probe.reason)
			}
		})
	}
}
