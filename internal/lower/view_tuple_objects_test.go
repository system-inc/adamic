package lower

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func tupleObjectSource(t *testing.T) string {
	t.Helper()
	source, err := os.ReadFile("testdata/tuple_object_views/p70.a")
	if err != nil {
		t.Fatal(err)
	}
	return string(source)
}

func TestTupleObjectViewRefused(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs("testdata/tuple_object_views/p70.a")
	if err != nil {
		t.Fatal(err)
	}
	want := runAgreementNode(t, path)
	if want.code != 0 || string(want.stdout) != "2\n" {
		t.Fatalf("Node control: %+v", want)
	}
	_, err = lowerSource(t, tupleObjectSource(t))
	var refusal *NotYet
	if !errors.As(err, &refusal) {
		t.Fatalf("want located tuple object refusal, got %v", err)
	}
	if !strings.Contains(refusal.Where, "main.a:9:") || !strings.Contains(refusal.What, "copy pair.length") {
		t.Fatalf("missing read path or fix: %v", err)
	}
	t.Log(err)
}

func TestTupleObjectViewCopiedLength(t *testing.T) {
	t.Parallel()
	source, err := os.ReadFile("testdata/tuple_object_views/p70_copy.a")
	if err != nil {
		t.Fatal(err)
	}
	// Native distinguishes the tuple storage from the copied scalar slot.
	lowersAndAgreesWithNodeNative(t, string(source))
}

func TestTupleObjectViewOrdinaryTuple(t *testing.T) {
	t.Parallel()
	lowersAndAgreesWithNode(t, "const pair: [string, number] = ['first', 1]; console.log(`${pair.length}`);")
}
