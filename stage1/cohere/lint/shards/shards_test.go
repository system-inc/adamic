package shards

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMergePutsCasesBackInOrder(test *testing.T) {
	test.Parallel()
	merged, mergeError := Merge([][]byte{
		[]byte("case 0\nfixed\ta\ncase 2\nfixed\tc\n"),
		[]byte("case 1\nfixed\tb\ncase 3\nfixed\td\n"),
	}, 4)
	if mergeError != nil {
		test.Fatal(mergeError)
	}
	if want := "case 0\nfixed\ta\ncase 1\nfixed\tb\ncase 2\nfixed\tc\ncase 3\nfixed\td\n"; string(merged) != want {
		test.Fatalf("merged %q, want %q", merged, want)
	}
}

// A line that only starts like a case line stays inside its block.
func TestMergeKeepsLinesThatOnlyLookLikeCases(test *testing.T) {
	test.Parallel()
	merged, mergeError := Merge([][]byte{[]byte("case 0\n  rule  case 12 of the message\ncase 1x\nfixed\ta\n")}, 1)
	if mergeError != nil {
		test.Fatal(mergeError)
	}
	if string(merged) != "case 0\n  rule  case 12 of the message\ncase 1x\nfixed\ta\n" {
		test.Fatalf("merged %q", merged)
	}
}

func TestMergeRefusesAMissingOrRepeatedCase(test *testing.T) {
	test.Parallel()
	for _, testCase := range []struct {
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
		for _, output := range testCase.outputs {
			outputs = append(outputs, []byte(output))
		}
		if _, mergeError := Merge(
			outputs,
			testCase.rows,
		); mergeError == nil || !strings.Contains(mergeError.Error(), testCase.want) {
			test.Fatalf("%s: error %v, want %q", testCase.name, mergeError, testCase.want)
		}
	}
}

// A manifest that opens a checker program, as every typed run's does, merges across shards (#3mecm8z). Its
// `program <tsconfig>` line is no case: main.ts skips it when numbering, so a count that included it
// expected one case more than any shard printed, and every sharded typed run failed Merge. The driver here
// is a stand-in for main.ts, dealing out and numbering rows exactly as main.ts does, so the test needs no build.
func TestRunMergesAManifestThatOpensAProgram(test *testing.T) {
	test.Parallel()
	directory := test.TempDir()
	driver := filepath.Join(directory, "driver")
	script := "#!/bin/sh\n" +
		"shard=$4\n" +
		"awk -v shard_index=\"${shard%/*}\" -v shard_count=\"${shard#*/}\" 'BEGIN { cases = 0 } $0 == \"\" { next } /^program / { next } { if (cases % shard_count == shard_index) print \"case \" cases; cases++ }' \"$2\"\n"
	if writeError := os.WriteFile(driver, []byte(script), 0755); writeError != nil {
		test.Fatal(writeError)
	}
	manifest := filepath.Join(directory, "manifest.txt")
	if writeError := os.WriteFile(
		manifest,
		[]byte("program /repository/tsconfig.json\na.ts\tall\n\nb.ts\tall\nc.ts\tall\n"),
		0644,
	); writeError != nil {
		test.Fatal(writeError)
	}
	for _, count := range []int{1, 2, 3} {
		merged, runError := Run(driver, manifest, count, false)
		if runError != nil {
			test.Fatalf("%d shards: %v", count, runError)
		}
		if want := "case 0\ncase 1\ncase 2\n"; string(merged) != want {
			test.Fatalf("%d shards merged %q, want %q", count, merged, want)
		}
	}
}

// Cases counts as main.ts numbers: every non-empty line is a case except the program line.
func TestCasesCountsAsMainDoes(test *testing.T) {
	test.Parallel()
	for _, testCase := range []struct {
		manifest string
		want     int
	}{
		{"", 0},
		{"a.ts\tall\n", 1},
		{"a.ts\tall\n\n\nb.ts\tall", 2},
		{"program /repository/tsconfig.json\na.ts\tall\nb.ts\tall\n", 2},
		{"program /repository/tsconfig.json\n", 0},
	} {
		if got := Cases([]byte(testCase.manifest)); got != testCase.want {
			test.Errorf("Cases(%q) = %d, want %d", testCase.manifest, got, testCase.want)
		}
	}
}
