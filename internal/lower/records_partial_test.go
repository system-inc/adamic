package lower

import (
	"errors"
	"strings"
	"testing"
)

func TestPartialRecordStorageRefusals(t *testing.T) {
	t.Parallel()
	// extension is .a unless a probe names .ts: in .a, Dog's first seen through Record<string, Animal>
	// is a declared property under a wider index signature (records_declared_keys.go).
	for _, probe := range []struct{ source, reason, extension string }{
		{`const fixed={first:1}; const cache:Partial<Record<'first'|'second',number>>=fixed;`, "seen as", ""},
		{`const cache:Partial<Record<'first'|'second',number>>={}; const fixed:{first?:number}=cache;`, "seen as", ""},
		{`interface Animal {name:string} interface Dog extends Animal {bark:string} const cache:Partial<Record<'first'|'second',Dog>>={}; const view:Record<string,Animal>=cache;`, "invariant-mutable", ".ts"},
		{`interface Animal {name:string} interface Dog extends Animal {bark:string} const cache:Partial<Record<'first'|'second',Dog>>={}; const view:Record<string,Animal>=cache;`, "declared properties seen through a wider index signature", ""},
		{`const cache:Partial<Record<'first'|'second',number>>={}; const view:Record<string,number|undefined>=cache;`, "invariant-mutable", ""},
	} {
		if probe.extension == "" {
			probe.extension = ".a"
		}
		t.Run(probe.extension+" "+probe.source, func(t *testing.T) {
			_, err := lowerSourceAs(t, probe.source, probe.extension)
			var ny *NotYet
			var refused *Refused
			if err == nil || (!errors.As(err, &ny) && !errors.As(err, &refused)) || !strings.Contains(err.Error(), probe.reason) {
				t.Fatalf("got %v, want %s refusal", err, probe.reason)
			}
		})
	}
}
