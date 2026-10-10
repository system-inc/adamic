package collections

import (
	"context"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
)

// The implementation bodies remain faithful excerpts; the interface adaptations
// (a present map and an explicit receiver) are intentionally outside these spans.
func TestCollectionBodiesMatchPinnedTSC(t *testing.T) {
	t.Parallel()
	upstream, err := os.ReadFile("../../../cohere/TypeScript/tsc/testdata/fixtures/compiler/core.ts")
	if err != nil {
		t.Fatal(err)
	}
	port, err := os.ReadFile("core.a")
	if err != nil {
		t.Fatal(err)
	}
	normalize := func(s string) string { return strings.Join(strings.Fields(s), " ") }
	entries := regexp.MustCompile(`(?s)const result = new Map<K2, V2>\(\);.*?return result;`)
	if normalize(entries.FindString(string(upstream))) != normalize(entries.FindString(string(port))) {
		t.Fatal("mapEntries body drifted from pinned tsc")
	}
	add := regexp.MustCompile(`(?s)let values = (?:this|map)\.get\(key\);.*?return values;`)
	original := strings.ReplaceAll(add.FindString(string(upstream)), "this.", "map.")
	original = strings.Replace(original, "map.set(key, values = [value]);", "values = [value]; map.set(key, values);", 1)
	if original == "" || normalize(original) != normalize(add.FindString(string(port))) {
		t.Fatal("multiMapAdd body drifted beyond the explicit receiver and assignment desugaring")
	}
}

func TestCollectionInstantiationsLower(t *testing.T) {
	t.Parallel()
	for _, name := range []string{"map_entries", "multimap"} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs("../../../internal/oracle/testdata/scout19_" + name + ".a")
			if err != nil {
				t.Fatal(err)
			}
			program, err := load.Load([]string{path})
			if err != nil {
				t.Fatal(err)
			}
			if _, err := lower.Lower(context.Background(), program); err != nil {
				t.Fatal(err)
			}
		})
	}
}
