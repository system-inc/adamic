package countproducts

import (
	"bytes"
	"path/filepath"
	"reflect"
	"slices"
	"testing"
)

// standIns are rows of each way one runs, as Rows takes them: what each holds is a stand-in, never built.
func standIns(root string) []Row {
	return []Row{
		{Key: "internal/oracle/testdata/plain.a", C: "#define ADAMIC_CLOSURE_CONVENTION 1\nint main(void) { return 0; }\n", Run: Plain, ParallelMap: true},
		{Key: "internal/oracle/testdata/arguments.a", C: "const char *at = \"" + filepath.Join(root, "internal", "oracle", "testdata", "arguments.a") + "\";\n", Run: Input,
			Directory: "internal/oracle/testdata", Arguments: []string{"plain", "", "a\xffb", "\xe2\x82"}, Inputs: "trees", Checked: true},
		{Key: "internal/oracle/testdata/node_fs_file_close.a", C: "int main(void) { return 1; }\n", Run: Input, Directory: "internal/oracle/testdata", FSFile: true, Checked: true},
		{Key: "internal/oracle/testdata/process_env/reads.a", C: "#define ADAMIC_NODE_HOST 1\n", Run: Environment, Variables: []string{"ADAMIC_ENV_READ_UNICODE=héllo 世界 🌍", "EMPTY="}, Checked: true},
		{Key: "internal/oracle/testdata/wrong-kind.ts", C: "int main(void) { return 70; }\n", Run: Plain, Checked: true, Exit: 70},
	}
}

// A shard's rows file is the same bytes whatever order its rows come in, and wherever the checkout is, so a shard's key
// moves with what its rows run and nothing else. Mutants: Rows writing its rows in the order it's handed them, or ranging
// over a map of them, unsorted; the checkout's root kept in a C as it's spelled here.
func TestARowsFileIsTheSameBytesInAnyOrderAndCheckout(t *testing.T) {
	t.Parallel()
	here, _, err := Rows(standIns("/checkouts/one/adamic"), "/checkouts/one/adamic")
	if err != nil {
		t.Fatal(err)
	}
	reversed := standIns("/checkouts/one/adamic")
	slices.Reverse(reversed)
	again, _, err := Rows(reversed, "/checkouts/one/adamic")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(here, again) {
		t.Errorf("one shard's rows, handed in another order, wrote other bytes:\n%s\n%s", here, again)
	}
	elsewhere, _, err := Rows(standIns("/elsewhere/adamic"), "/elsewhere/adamic")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(here, elsewhere) {
		t.Errorf("one shard's rows, in another checkout, wrote other bytes:\n%s\n%s", here, elsewhere)
	}
}

// A row comes back from its rows file whole: arguments that aren't valid UTF-8, as arguments.a's, byte for byte, and
// its C naming the checkout it's read in; the feature sets come back in the order a build is handed their archives.
// Mutants: a fileRow holding Arguments as strings, which encoding/json writes with U+FFFD for each bad byte; readRows
// leaving repositoryMark in a C.
func TestARowComesBackFromItsRowsFileWhole(t *testing.T) {
	t.Parallel()
	rows := standIns("/checkouts/one/adamic")
	encoded, sets, err := Rows(rows, "/checkouts/one/adamic")
	if err != nil {
		t.Fatal(err)
	}
	read, readSets, err := readRows(encoded, "/elsewhere/adamic")
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(sets, readSets) {
		t.Errorf("Rows gave the feature sets %q, and its file holds %q", sets, readSets)
	}
	want := standIns("/elsewhere/adamic")
	slices.SortFunc(want, func(first, second Row) int {
		switch {
		case first.Key < second.Key:
			return -1
		case first.Key > second.Key:
			return 1
		}
		return 0
	})
	if !reflect.DeepEqual(read, want) {
		t.Errorf("rows came back from their file as\n%+v\nwant\n%+v", read, want)
	}
	if want := [][]string{nil, {"ADAMIC_CLOSURE_CONVENTION"}, {"ADAMIC_NODE_HOST"}}; !reflect.DeepEqual(sets, want) {
		t.Errorf("the rows' feature sets are %q, want %q", sets, want)
	}
}

// Rows refuses what would make a rows file read back as another: two rows of one key, a C holding the mark the file
// names the checkout by, and a checkout that isn't an absolute path. Mutant: Rows taking a relative checkout, which
// marks every place it's a substring of a C.
func TestRowsRefusesWhatWouldNotReadBack(t *testing.T) {
	t.Parallel()
	for name, each := range map[string]struct {
		rows       []Row
		repository string
	}{
		"two rows of one key":  {[]Row{{Key: "a.a", Run: Plain}, {Key: "a.a", Run: Plain}}, "/adamic"},
		"a C holding the mark": {[]Row{{Key: "a.a", C: "x" + repositoryMark, Run: Plain}}, "/adamic"},
		"a relative checkout":  {[]Row{{Key: "a.a", Run: Plain}}, "adamic"},
	} {
		if _, _, err := Rows(each.rows, each.repository); err == nil {
			t.Errorf("%s: Rows wrote it", name)
		}
	}
}
