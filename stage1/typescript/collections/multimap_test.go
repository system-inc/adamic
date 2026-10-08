package collections

import (
	"os"
	"strings"
	"testing"
)

// The tracked adaptation reads these exact input spans. Keep both the original
// protocol and the custom collection body available for the next compiler proof.
func TestSlice2BodiesMatchPinnedTSC(t *testing.T) {
	t.Parallel()
	read := func(path string) string {
		t.Helper()
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		return string(data)
	}
	normalize := func(text string) string { return strings.Join(strings.Fields(text), " ") }
	upstream := read("../../../cohere/TypeScript/tsc/testdata/fixtures/compiler/core.ts")
	span := func(text, start, end string) string {
		t.Helper()
		first := strings.Index(text, start)
		if first < 0 {
			t.Fatalf("missing start %q", start)
		}
		last := strings.Index(text[first:], end)
		if last < 0 {
			t.Fatalf("missing end %q", end)
		}
		return normalize(text[first : first+last])
	}
	original := read("../../../stage3/scout/19-slice2/multimap-composition/original.a")
	if normalize(original) != span(upstream, "export interface MultiMap<", "/** @internal */\r\nexport function createQueue") {
		t.Fatal("original MultiMap adaptation input drifted")
	}
	custom := read("../../../internal/oracle/testdata/scout19_slice2_create_set.a")
	expected := span(upstream, "export function createSet<", "\r\n/**\r\n * Tests whether a value is an array.")
	actual := span(custom, "export function createSet<", "\n// Harness dependencies")
	if expected != actual {
		t.Fatal("createSet must remain exactly tsc's own collection body")
	}
	lookup := read("../../../internal/oracle/testdata/scout19_slice2_presence.a")
	if span(upstream, "export function getOrUpdate<", "\r\n/**") != span(lookup, "export function getOrUpdate<", "\nlet calls") {
		t.Fatal("presence proof no longer uses tsc's getOrUpdate body")
	}
}
