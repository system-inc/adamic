package oracle

import (
	"os"
	"testing"
)

func TestCheckedViewTupleLegacyMapAdmissionMutant(t *testing.T) {
	if os.Getenv("ADAMIC_TUPLE_LEGACY_MAP_MUTANT") != "certificate" {
		t.Skip("opt-in source certificate mutant")
	}
	for _, name := range []string{"optional-tuple", "rest-tuple"} {
		t.Run(name, func(t *testing.T) {
			p, _ := interfaceFixture(t, "nullish/maps/entry-convert-gap-"+name)
			if len(p.MapCertificates) == 0 {
				t.Fatal("source certificate absent")
			}
			p.MapCertificates = nil
			actual := onJavaScriptBackend(t, p)
			if diff := disagreement(run{stdout: []byte("object\n")}, actual); diff != "" {
				t.Fatalf("source certificate mutant caught: %s; %#v", diff, actual)
			}
			// An escaped mutant must leave this runner green, exposing the missing kill.
			return
		})
	}
}
