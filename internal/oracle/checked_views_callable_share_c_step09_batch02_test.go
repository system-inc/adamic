package oracle

import (
	"strings"
	"testing"
)

// These observations are unresolved boundaries, excluded from certification.
func TestCheckedViewCallableShareCStep09Batch02Boundaries(t *testing.T) {
	t.Run("rank-125", func(t *testing.T) {
		p, path := interfaceFixture(t, "lane5/share-c/families/rank-125/good")
		truth := onNode(t, path)
		if truth.exitCode != 0 || string(truth.stdout) != "1\n" {
			t.Fatalf("Node positive: %#v", truth)
		}
		sanitized, binary := nativelyUncached(t, p)
		for _, got := range []run{releasedUncached(t, p), sanitized, onJavaScriptBackend(t, p)} {
			if diff := disagreement(truth, got); diff != "" {
				t.Fatalf("positive: %s", diff)
			}
		}
		if report := leaksUncached(t, p, binary); report != "" {
			t.Fatal(report)
		}
		p, path = interfaceFixture(t, "lane5/share-c/families/rank-125/wrong-element")
		truth = onNode(t, path)
		if truth.exitCode != 0 || string(truth.stdout) != "undefined\n" {
			t.Fatalf("Node negative control: %#v", truth)
		}
		javascript := onJavaScriptBackend(t, p)
		if javascript.exitCode != 70 || string(javascript.stderr) != "adamic: panic: element read failed: copied[0] expected Type, found string\n" {
			t.Fatalf("JavaScript boundary changed: %#v", javascript)
		}
		sanitized, _ = nativelyUncached(t, p)
		if sanitized.exitCode == 0 || sanitized.exitCode == 70 || !strings.Contains(string(sanitized.stderr), "runtime error: member access within misaligned address") || !strings.Contains(string(sanitized.stderr), "UndefinedBehaviorSanitizer") {
			t.Fatalf("native boundary changed: %#v", sanitized)
		}
		t.Logf("Uncertified rank 125: slice result loses transitive element/field checks; JavaScript rejects the wrong element at exit 70; native sanitizer exit %d, stderr %s", sanitized.exitCode, sanitized.stderr)
	})
}
