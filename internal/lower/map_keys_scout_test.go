package lower

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestScoutMapKeyOutcomes(t *testing.T) {
	for _, probe := range []struct {
		name, reason string
		refused      bool
	}{
		{"branded_map", "a value of type Path", false},
		{"branded_set", "a value of type Path", false},
		{"mixed_map", "a Map whose keys aren't", false},
		{"nullish_set", "a Set of null | undefined", false},
		{"maplike", "an index signature", true},
		{"iterator_pairs", "new Map from something that isn't [key, value] pairs", false},
	} {
		t.Run(probe.name, func(t *testing.T) {
			source, err := os.ReadFile(filepath.Join("..", "..", "stage3", "map-keys", probe.name+".a"))
			if err != nil {
				t.Fatal(err)
			}
			_, err = lowerSource(t, string(source))
			var refused *Refused
			var notYet *NotYet
			if (probe.refused && !errors.As(err, &refused)) || (!probe.refused && !errors.As(err, &notYet)) || err == nil || !strings.Contains(err.Error(), probe.reason) {
				t.Fatalf("want %q with refused=%v, got %v", probe.reason, probe.refused, err)
			}
			t.Log(err)
		})
	}
}
