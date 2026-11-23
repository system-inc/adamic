package main

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

// checkHeldStopRow holds one row of a held-stop list to the tree at root: its program is there, the row's hatch is
// that program's line as the row quotes it, a language hatch's ruling is said once in docs/rulings.md by the row's
// opening words, wherever its row now sits, and the test the row names is in its file and the file names the program, by its file name, by its
// directory's path from the test's own, or by that directory's last segment as a path piece or a string, as a test
// that walks the directory does. readHeldStops reads a row's form; this reads what the form points at, so a row can't
// name a hatch, a ruling or a test the tree doesn't hold.
func checkHeldStopRow(root string, stop heldStop) error {
	source, err := os.ReadFile(filepath.Join(root, stop.Path))
	if err != nil {
		return fmt.Errorf("%s: its program: %w", stop.Path, err)
	}
	hatch := heldStopHatch.FindStringSubmatch(stop.Hatch)
	if hatch == nil {
		return fmt.Errorf("%s: hatch %q isn't <kind>[ (<detail>)], line <n>: <text>", stop.Path, stop.Hatch)
	}
	if err = quotesLine(source, hatch[3], hatch[4]); err != nil {
		return fmt.Errorf("%s: its hatch: %w", stop.Path, err)
	}
	if hatch[1] == "language" {
		ruling := heldStopRuling.FindStringSubmatch(hatch[2])
		if ruling == nil {
			return fmt.Errorf("%s: its language hatch cites no ruling", stop.Path)
		}
		rulings, err := os.ReadFile(filepath.Join(root, "docs/rulings.md"))
		if err != nil {
			return fmt.Errorf("%s: its ruling: %w", stop.Path, err)
		}
		if err = saysOnce(rulings, ruling[2]); err != nil {
			return fmt.Errorf("%s: its ruling, docs/rulings.md: %w", stop.Path, err)
		}
	}
	test := heldStopTest.FindStringSubmatch(stop.Test)
	if test == nil {
		return fmt.Errorf("%s: test %q isn't <repository path>_test.go: <TestName>", stop.Path, stop.Test)
	}
	data, err := os.ReadFile(filepath.Join(root, test[1]))
	if err != nil {
		return fmt.Errorf("%s: its test: %w", stop.Path, err)
	}
	text := string(data)
	if !strings.Contains(text, "func "+test[2]+"(") {
		if err = holdsFamily(root, test[1], test[2], text); err != nil {
			return fmt.Errorf("%s: %s holds no func %s, %w", stop.Path, test[1], test[2], err)
		}
	}
	directory := path.Dir(stop.Path)
	segment := path.Base(directory)
	relative, err := filepath.Rel(path.Dir(test[1]), directory)
	if err != nil {
		return err
	}
	for _, name := range []string{path.Base(stop.Path), filepath.ToSlash(relative), `"` + segment + `"`, segment + "/", "/" + segment + `"`} {
		if strings.Contains(text, name) {
			return nil
		}
	}
	return fmt.Errorf("%s: %s names neither the program nor its directory", stop.Path, test[1])
}

// familyRunner is a row of a package's families table (families_test.go): a family's test name and the runner each
// of its generated shards calls.
var familyRunner = regexp.MustCompile(`\{name: "(Test[A-Za-z0-9_]*)",[^\n]*\brunner: "([A-Za-z0-9_]+)"`)

// holdsFamily says a test that isn't a func of its own is a family of the test file's package: the families table
// names it and its runner, the generated wrappers (families_generated_test.go) hold its first shard, and the test
// file, text, defines the runner. So a row names the family whose runner holds its program, as a shard's name can't.
func holdsFamily(root, testFile, name, text string) error {
	directory := filepath.Join(root, path.Dir(testFile))
	table, err := os.ReadFile(filepath.Join(directory, "families_test.go"))
	if err != nil {
		return fmt.Errorf("and its package has no families table: %w", err)
	}
	runner := ""
	for _, family := range familyRunner.FindAllStringSubmatch(string(table), -1) {
		if family[1] == name {
			runner = family[2]
		}
	}
	if runner == "" {
		return fmt.Errorf("and no family of its package is named so")
	}
	generated, err := os.ReadFile(filepath.Join(directory, "families_generated_test.go"))
	if err != nil {
		return fmt.Errorf("and its family has no generated wrappers: %w", err)
	}
	if !strings.Contains(string(generated), "func "+name+"_000(") {
		return fmt.Errorf("and its family's first shard isn't generated")
	}
	if !strings.Contains(text, "func "+runner+"(") {
		return fmt.Errorf("and it doesn't define the family's runner %s", runner)
	}
	return nil
}

// quotesLine says the 1-based line number of source, trimmed, is text.
func quotesLine(source []byte, number, text string) error {
	at, err := strconv.Atoi(number)
	if err != nil {
		return err
	}
	lines := strings.Split(string(source), "\n")
	if at > len(lines) {
		return fmt.Errorf("line %d is past the end, %d lines", at, len(lines))
	}
	if got := strings.TrimSpace(lines[at-1]); got != text {
		return fmt.Errorf("line %d is %q, not %q", at, got, text)
	}
	return nil
}

// saysOnce says text says words exactly once. A language hatch cites its ruling by line and by words, and only the
// words are held: rulings.md grows above every row as rulings land, so a guard on the line would red each change that
// adds a ruling, and the words still name exactly one row wherever it now sits (Language on #d358hc1).
func saysOnce(text []byte, words string) error {
	if count := strings.Count(string(text), words); count != 1 {
		return fmt.Errorf("%q is said %d times, not once", words, count)
	}
	return nil
}

// The repository's own held-stop list is read, and every row's program, hatch line, ruling and test are the tree's.
// This is #d358hc1's guard: a row whose test stops naming its program, or whose program's hatch line moves, reds here
// before the daily run reads the list.
func TestHeldStopListHoldsItsRows(t *testing.T) {
	t.Parallel()
	root := filepath.Join("..", "..")
	list, err := readHeldStops(filepath.Join(root, "cloud/admission-corpus/held-stops.json"))
	if err != nil {
		t.Fatal(err)
	}
	for _, stop := range list.Stops {
		if err := checkHeldStopRow(root, stop); err != nil {
			t.Error(err)
		}
	}
}

// heldStopReads are the places the tools line's executors.txt keys cmd/adamic-admission-delta on for what
// TestHeldStopListHoldsItsRows reads, each a prefix and a suffix as the reads line's glob reads it (its * crossing
// slashes): a row's program, a row's test file and that package's families tables. A row outside them would be read
// by the test and held by no key, so an edit to it alone would reuse the unit's pass (Adamic Loom's must on #z0h4bpd).
var heldStopReads = []struct{ prefix, suffix string }{
	// Reads: internal/oracle/testdata/*.a
	{"internal/oracle/testdata/", ".a"},
	// Reads: internal/oracle/testdata/*.ts
	{"internal/oracle/testdata/", ".ts"},
	// Reads: internal/lower/testdata/*.a
	{"internal/lower/testdata/", ".a"},
	// Reads: internal/lower/testdata/*.ts
	{"internal/lower/testdata/", ".ts"},
	// Reads: stage3/interface-downcasts/*.ts
	{"stage3/interface-downcasts/", ".ts"},
	// Reads: stage3/fixtures/*.a
	{"stage3/fixtures/", ".a"},
	// Reads: stage1/cohere/markdownblocks/gaps/*.ts
	{"stage1/cohere/markdownblocks/gaps/", ".ts"},
	// Reads: internal/oracle/*_test.go
	{"internal/oracle/", "_test.go"},
	// Reads: internal/lower/*_test.go
	{"internal/lower/", "_test.go"},
	// Reads: stage1/cohere/markdownblocks/*_test.go
	{"stage1/cohere/markdownblocks/", "_test.go"},
}

// heldStopKeyed says file sits in one of heldStopReads' places.
func heldStopKeyed(file string) bool {
	for _, place := range heldStopReads {
		if strings.HasPrefix(file, place.prefix) && strings.HasSuffix(file, place.suffix) {
			return true
		}
	}
	return false
}

// Every row's program and test file, and so its package's families tables, sit where the unit's key reads, so an edit to
// any of them moves the key. The list and docs/rulings.md are read whole, keyed by name. A row elsewhere reds here until
// the tools line names its place and heldStopReads mirrors it. Mutant: a row's program outside every place, red.
func TestHeldStopRowsSitWhereTheKeyReads(t *testing.T) {
	t.Parallel()
	// Reads: cloud/admission-corpus/held-stops.json
	list, err := readHeldStops(filepath.Join("..", "..", "cloud/admission-corpus/held-stops.json"))
	if err != nil {
		t.Fatal(err)
	}
	// Reads: docs/rulings.md
	for _, stop := range list.Stops {
		if !heldStopKeyed(stop.Path) {
			t.Errorf("%s: its program sits outside every place cmd/adamic-admission-delta's key reads (heldStopReads): add its place to the tools line's executors.txt and here", stop.Path)
		}
		test := heldStopTest.FindStringSubmatch(stop.Test)
		if test == nil {
			t.Errorf("%s: test %q isn't <repository path>_test.go: <TestName>", stop.Path, stop.Test)
			continue
		}
		if !heldStopKeyed(test[1]) {
			t.Errorf("%s: its test %s sits outside every place cmd/adamic-admission-delta's key reads (heldStopReads): add its place to the tools line's executors.txt and here", stop.Path, test[1])
		}
	}
}

// The guard refuses a row whose hatch line isn't the program's, whose ruling words aren't said once, whose test file
// lacks its func and holds no family of that name whose runner it defines, or names neither the program nor its
// directory, and one whose program is gone. It reads a row whose test names the program, one whose test walks its
// directory, and one whose family's runner, in the named file, walks it. Mutant (agreement.mutants.md): quotesLine's
// comparison dropped, so the moved hatch line reads.
func TestHeldStopRowCheckRefuses(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	for name, text := range map[string]string{
		"corpus/walked/viewed.a":                     "interface View { x: string }\nconst base = {};\nconst viewed = base as View;\n",
		"docs/rulings.md":                            "# Rulings\n\n- Array writes only at an existing index.\n- Twice said.\n- Twice said.\n",
		"internal/oracle/a_test.go":                  "package oracle\n\nfunc TestNames(t *testing.T) { run(\"viewed.a\") }\n",
		"internal/oracle/b_test.go":                  "package oracle\n\nfunc TestWalks(t *testing.T) { walk(filepath.Join(\"corpus\", \"walked\")) }\n",
		"internal/oracle/families_test.go":           "package oracle\n\nvar families = []family{\n\t{name: \"TestAgrees\", shards: 2, runner: \"runAgreesShard\", keys: paths},\n}\n",
		"internal/oracle/families_generated_test.go": "package oracle\n\nfunc TestAgrees_000(t *testing.T) { runAgreesShard(t, 0) }\n",
		"internal/oracle/e_test.go":                  "package oracle\n\nfunc runAgreesShard(t *testing.T, shard int) { walk(\"walked/\") }\n",
		"internal/oracle/c_test.go":                  "package oracle\n\nfunc TestOther(t *testing.T) { run(\"other.a\") }\n",
	} {
		if err := os.MkdirAll(filepath.Dir(filepath.Join(root, name)), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(root, name), []byte(text), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	row := func(hatch, test string) heldStop {
		return heldStop{Path: "corpus/walked/viewed.a", Stderr: "adamic: panic: held\n", Hatch: hatch, Test: test}
	}
	const cast = "cast, line 3: const viewed = base as View;"
	for name, stop := range map[string]heldStop{
		"a test naming the program":                 row(cast, "internal/oracle/a_test.go: TestNames"),
		"a family whose runner walks its directory": row(cast, "internal/oracle/e_test.go: TestAgrees"),
		"a test walking its directory":              row(cast, "internal/oracle/b_test.go: TestWalks"),
		"a language hatch whose ruling has moved":   row(`language (docs/rulings.md:1, "Array writes only at"), line 3: const viewed = base as View;`, "internal/oracle/a_test.go: TestNames"),
		"a language hatch":                          row(`language (docs/rulings.md:3, "Array writes only at"), line 3: const viewed = base as View;`, "internal/oracle/a_test.go: TestNames"),
	} {
		if err := checkHeldStopRow(root, stop); err != nil {
			t.Errorf("%s: %v, want it read", name, err)
		}
	}
	gone := row(cast, "internal/oracle/a_test.go: TestNames")
	gone.Path = "corpus/walked/gone.a"
	for name, stop := range map[string]heldStop{
		"a hatch line moved":                  row("cast, line 2: const viewed = base as View;", "internal/oracle/a_test.go: TestNames"),
		"a hatch line past the end":           row("cast, line 9: const viewed = base as View;", "internal/oracle/a_test.go: TestNames"),
		"a ruling whose words aren't said":    row(`language (docs/rulings.md:3, "Array writes past the end"), line 3: const viewed = base as View;`, "internal/oracle/a_test.go: TestNames"),
		"a ruling whose words are said twice": row(`language (docs/rulings.md:4, "Twice said"), line 3: const viewed = base as View;`, "internal/oracle/a_test.go: TestNames"),
		"a test func the file lacks":          row(cast, "internal/oracle/a_test.go: TestWalks"),
		"a family whose runner is elsewhere":  row(cast, "internal/oracle/a_test.go: TestAgrees"),
		"a family its table lacks":            row(cast, "internal/oracle/e_test.go: TestOthers"),
		"a test naming another program":       row(cast, "internal/oracle/c_test.go: TestOther"),
		"a test file that isn't there":        row(cast, "internal/oracle/d_test.go: TestNames"),
		"a program that's gone":               gone,
	} {
		if err := checkHeldStopRow(root, stop); err == nil {
			t.Errorf("%s: read, want it refused", name)
		}
	}
}
