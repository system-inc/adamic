package shards

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMergePutsCasesBackInOrder(t *testing.T) {
	t.Parallel()
	merged, err := Merge([][]byte{
		[]byte("case 0\nfixed\ta\ncase 2\nfixed\tc\n"),
		[]byte("case 1\nfixed\tb\ncase 3\nfixed\td\n"),
	}, 4)
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
	merged, err := Merge([][]byte{[]byte("case 0\n  rule  case 12 of the message\ncase 1x\nfixed\ta\n")}, 1)
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
		rows    int
		want    string
	}{
		{"missing", []string{"case 0\nfixed\ta\n", "case 2\nfixed\tc\n"}, 3, "case 1 missing"},
		{"repeated", []string{"case 0\nfixed\ta\n", "case 0\nfixed\ta\n"}, 1, "case 0 printed twice"},
		{"no case line", []string{"fixed\ta\n"}, 1, "does not start with a case line"},
		// A shard that printed nothing, so the manifest's last case never arrived: no gap shows it.
		{"empty shard", []string{"case 0\nfixed\ta\n", ""}, 2, "case 1 missing"},
		{"extra case", []string{"case 0\nfixed\ta\ncase 1\nfixed\tb\n"}, 1, "2 cases printed, want the manifest's 1"},
	} {
		var outputs [][]byte
		for _, output := range test.outputs {
			outputs = append(outputs, []byte(output))
		}
		if _, err := Merge(outputs, test.rows); err == nil || !strings.Contains(err.Error(), test.want) {
			t.Fatalf("%s: error %v, want %q", test.name, err, test.want)
		}
	}
}

// A manifest that opens a checker program, as every typed run's does, merges across shards (#3mecm8z). Its
// `program <tsconfig>` line is no case: main.ts skips it when numbering, so a count that included it
// expected one case more than any shard printed, and every sharded typed run failed Merge. The driver here
// is a stand-in for main.ts, dealing out and numbering rows exactly as main.ts does, so the test needs no build.
func TestRunMergesAManifestThatOpensAProgram(t *testing.T) {
	t.Parallel()
	directory := t.TempDir()
	driver := filepath.Join(directory, "driver")
	script := "#!/bin/sh\n" +
		"shard=$4\n" +
		"awk -v shard_index=\"${shard%/*}\" -v shard_count=\"${shard#*/}\" 'BEGIN { cases = 0 } $0 == \"\" { next } /^program / { next } { if (cases % shard_count == shard_index) print \"case \" cases; cases++ }' \"$2\"\n"
	if err := os.WriteFile(driver, []byte(script), 0755); err != nil {
		t.Fatal(err)
	}
	manifest := filepath.Join(directory, "manifest.txt")
	if err := os.WriteFile(manifest, []byte("program /repository/tsconfig.json\na.ts\tall\n\nb.ts\tall\nc.ts\tall\n"), 0644); err != nil {
		t.Fatal(err)
	}
	for _, count := range []int{1, 2, 3} {
		merged, err := Run(driver, manifest, count, false)
		if err != nil {
			t.Fatalf("%d shards: %v", count, err)
		}
		if want := "case 0\ncase 1\ncase 2\n"; string(merged) != want {
			t.Fatalf("%d shards merged %q, want %q", count, merged, want)
		}
	}
}

// Cases counts as main.ts numbers: every non-empty line is a case except the program line.
func TestCasesCountsAsMainDoes(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		manifest string
		want     int
	}{
		{"", 0},
		{"a.ts\tall\n", 1},
		{"a.ts\tall\n\n\nb.ts\tall", 2},
		{"program /repository/tsconfig.json\na.ts\tall\nb.ts\tall\n", 2},
		{"program /repository/tsconfig.json\n", 0},
	} {
		if got := Cases([]byte(test.manifest)); got != test.want {
			t.Errorf("Cases(%q) = %d, want %d", test.manifest, got, test.want)
		}
	}
}
