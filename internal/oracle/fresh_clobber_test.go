package oracle

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/lower"
)

// Node observes the real cycle; refusal or graph ownership must keep it sound.
var freshClobberOutputs = map[string]string{
	"alias_fill.a":          "100\n",
	"alias_mapset.a":        "200\n",
	"alias_reverse.a":       "100\n",
	"alias_setadd.a":        "200\n",
	"alias_sort.a":          "100\n",
	"clobber_link_loop.a":   "100\n",
	"escaped_before.a":      "100\n",
	"map_entries_pattern.a": "100\n",
	"spread_outside.a":      "1 1\n",
	"throw_keeps_old.a":     "caught\n1\n",
}

func freshClobberWitness(test *testing.T, name string) {
	test.Helper()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/fresh_refused", name+".a"))
	if err != nil {
		test.Fatal(err)
	}
	result := onNode(test, path)
	if result.exitCode != 0 || string(result.stdout) != freshClobberOutputs[name+".a"] || len(result.stderr) != 0 {
		test.Fatalf("Node: exit %d stdout %q stderr %q", result.exitCode, result.stdout, result.stderr)
	}
	source, err := os.ReadFile(path)
	if err != nil {
		test.Fatal(err)
	}
	marked := 0
	for index, line := range strings.Split(string(source), "\n") {
		if strings.HasSuffix(line, "// closes the cycle") {
			marked = index + 1
		}
	}
	if marked == 0 {
		test.Fatal("no closing write marked")
	}
	program, err := lowered(test, path)
	if err == nil {
		checkGraphAdmission(test, path, program)
		return
	}
	var refused *lower.Refused
	if !errors.As(err, &refused) ||
		!strings.Contains(err.Error(), "adamic/cycle-capable") ||
		!strings.Contains(err.Error(), fmt.Sprintf("the write at %s:%d:", path, marked)) {
		test.Fatalf("want marked cycle refusal, got %v", err)
	}
	test.Logf("Node stdout %q; %v", result.stdout, err)
}

func TestFreshClobberAliasFill(test *testing.T) {
	test.Parallel()
	freshClobberWitness(test, "alias_fill")
}

func TestFreshClobberAliasMapset(test *testing.T) {
	test.Parallel()
	freshClobberWitness(test, "alias_mapset")
}

func TestFreshClobberAliasReverse(test *testing.T) {
	test.Parallel()
	freshClobberWitness(test, "alias_reverse")
}

func TestFreshClobberAliasSetadd(test *testing.T) {
	test.Parallel()
	freshClobberWitness(test, "alias_setadd")
}

func TestFreshClobberAliasSort(test *testing.T) {
	test.Parallel()
	freshClobberWitness(test, "alias_sort")
}

func TestFreshClobberClobberLinkLoop(test *testing.T) {
	test.Parallel()
	freshClobberWitness(test, "clobber_link_loop")
}

func TestFreshClobberEscapedBefore(test *testing.T) {
	test.Parallel()
	freshClobberWitness(test, "escaped_before")
}

func TestFreshClobberMapEntriesPattern(test *testing.T) {
	test.Parallel()
	freshClobberWitness(test, "map_entries_pattern")
}

func TestFreshClobberSpreadOutside(test *testing.T) {
	test.Parallel()
	freshClobberWitness(test, "spread_outside")
}

func TestFreshClobberThrowKeepsOld(test *testing.T) {
	test.Parallel()
	freshClobberWitness(test, "throw_keeps_old")
}
