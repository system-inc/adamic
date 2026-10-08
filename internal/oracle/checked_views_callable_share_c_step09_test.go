package oracle

import (
	"testing"
)

// This records an unresolved producer-contract boundary, not a certificate.
// It stays outside certification and allocation-count manifests.
func TestCheckedViewCallableShareCStep09LiteralArrayBoundary(t *testing.T) {
	p, path := interfaceFixture(t, "lane5/share-c/families/rank-56/wrong-element")
	truth := onNode(t, path)
	if truth.exitCode != 0 || string(truth.stdout) != "2\n" {
		t.Fatalf("Node control: %#v", truth)
	}
	for _, got := range []run{releasedUncached(t, p), onJavaScriptBackend(t, p)} {
		if diff := disagreement(truth, got); diff != "" {
			t.Fatalf("boundary changed; reassess rank 56: %s", diff)
		}
		t.Log("Uncertified rank 56: producer 1[] accepts push(3) through number[] view; needs literal producer write-contract enforcement")
	}
}
