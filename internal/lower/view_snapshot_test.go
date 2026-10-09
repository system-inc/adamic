package lower

import (
	"os"
	"testing"
)

func TestViewSnapshotMaybeBoolean(t *testing.T) {
	t.Parallel()
	source, err := os.ReadFile("../../stage3/interface-downcasts/v2/snapshot-maybe-boolean.a")
	if err != nil {
		t.Fatal(err)
	}
	lowersAndAgreesWithNode(t, string(source))
}

func TestViewSnapshotMaybeNumber(t *testing.T) {
	t.Parallel()
	source, err := os.ReadFile("../../stage3/interface-downcasts/v2/snapshot-maybe-number.a")
	if err != nil {
		t.Fatal(err)
	}
	lowersAndAgreesWithNode(t, string(source))
}
