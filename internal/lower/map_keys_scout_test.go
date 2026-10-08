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
		{"reused_pair_mutation", "assigning an element of a value", false},
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

func TestScoutMapIteratorPairsLower(t *testing.T) {
	source, err := os.ReadFile(filepath.Join("..", "..", "stage3", "map-keys", "iterator_pairs.a"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err = lowerSource(t, string(source)); err != nil {
		t.Fatal(err)
	}
}

func TestScoutMapIteratorPairRepresentationGap(t *testing.T) {
	source, err := os.ReadFile(filepath.Join("..", "..", "stage3", "map-keys", "iterator_pairs.a"))
	if err != nil {
		t.Fatal(err)
	}
	widened := strings.Replace(string(source), "new Map(pairs())", "new Map<string, string | number>(pairs())", 1)
	_, err = lowerSource(t, widened)
	var notYet *NotYet
	if !errors.As(err, &notYet) || !strings.Contains(err.Error(), "new Map from pairs held otherwise than the Map's keys and values") {
		t.Fatalf("want representation-conversion gap, got %v", err)
	}
}
