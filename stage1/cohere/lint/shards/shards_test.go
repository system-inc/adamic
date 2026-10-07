package shards

import (
	"strings"
	"testing"
)

func TestMergePutsCasesBackInOrder(t *testing.T) {
	t.Parallel()
	merged, err := Merge([][]byte{
		[]byte("case 0\nfixed\ta\ncase 2\nfixed\tc\n"),
		[]byte("case 1\nfixed\tb\ncase 3\nfixed\td\n"),
	})
	if err != nil {
		t.Fatal(err)
	}
	if want := "case 0\nfixed\ta\ncase 1\nfixed\tb\ncase 2\nfixed\tc\ncase 3\nfixed\td\n"; string(merged) != want {
		t.Fatalf("merged %q, want %q", merged, want)
	}
}

// A line that only starts like a case line stays inside its block.
func TestMergeKeepsLinesThatOnlyLookLikeCases(t *testing.T) {
	t.Parallel()
	merged, err := Merge([][]byte{[]byte("case 0\n  rule  case 12 of the message\ncase 1x\nfixed\ta\n")})
	if err != nil {
		t.Fatal(err)
	}
	if string(merged) != "case 0\n  rule  case 12 of the message\ncase 1x\nfixed\ta\n" {
		t.Fatalf("merged %q", merged)
	}
}

func TestMergeRefusesAMissingOrRepeatedCase(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		name    string
		outputs []string
		want    string
	}{
		{"missing", []string{"case 0\nfixed\ta\n", "case 2\nfixed\tc\n"}, "case 1 missing"},
		{"repeated", []string{"case 0\nfixed\ta\n", "case 0\nfixed\ta\n"}, "case 0 printed twice"},
		{"no case line", []string{"fixed\ta\n"}, "does not start with a case line"},
	} {
		var outputs [][]byte
		for _, output := range test.outputs {
			outputs = append(outputs, []byte(output))
		}
		if _, err := Merge(outputs); err == nil || !strings.Contains(err.Error(), test.want) {
			t.Fatalf("%s: error %v, want %q", test.name, err, test.want)
		}
	}
}
