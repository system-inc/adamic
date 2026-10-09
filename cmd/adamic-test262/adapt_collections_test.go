package main

import (
	"strings"
	"testing"
)

func TestAdaptCollectionTypes(t *testing.T) {
	t.Parallel()
	for _, test := range []struct{ source, want string }{
		{"var s = new Set(); s.add(-0); s.has(+0); s.delete(NaN);", "new Set<number>()"},
		{"var s = new Set(); s.add('a'); s.has('b');", "new Set<string>()"},
		{"var s = new Set(); s.has(false);", "new Set<boolean>()"},
		{"var m = new Map(); m.set(-0, 42); m.has(+0); m.clear(); m.set(NaN, 43);", "new Map<number, number>()"},
	} {
		got := adaptSource(test.source)
		if !strings.Contains(got.Source, test.want) || got.Counts[adaptCollectionType] != 1 {
			t.Fatalf("%s: %#v", test.source, got)
		}
	}
	for _, source := range []string{
		"var s = new Set(); s.add(1); s.add('a');",
		"var s = new Set(); var alias = s; alias.add('a'); s.add(1);",
		"var s = new Set(); use(s); s.add(1);",
		"var s = new Set(); s.add(undefined);",
		"var s = new Set(); s.add(null);",
		"var s = new Set(); s.add({x: 1});",
		"var s = new Set(); s.has; s.add(1);",
		"var s = new Set(); s = new Set(); s.add(1);",
		"var Set = factory; var s = new Set(); s.add(1);",
		"var s = new Set(); s.add(1); var Set = factory;",
		"var s = new Set();",
		"var s = new (Set)(); s.add(1);",
		"var s = new Set(); console.log(`${s.size}`); s.add(1);",
		"var m = new Map(); m.has(1);",
	} {
		if got := adaptSource(source); got.Counts[adaptCollectionType] != 0 {
			t.Fatalf("unproven type adapted: %#v", got)
		}
	}
}
