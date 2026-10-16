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

func freshClobberWitness(t *testing.T, name string) {
	t.Helper()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/fresh_refused", name+".a"))
	if err != nil {
		t.Fatal(err)
	}
	result := onNode(t, path)
	if result.exitCode != 0 || string(result.stdout) != freshClobberOutputs[name+".a"] || len(result.stderr) != 0 {
		t.Fatalf("Node: exit %d stdout %q stderr %q", result.exitCode, result.stdout, result.stderr)
	}
	source, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	marked := 0
	for index, line := range strings.Split(string(source), "\n") {
		if strings.HasSuffix(line, "// closes the cycle") {
			marked = index + 1
		}
	}
	if marked == 0 {
		t.Fatal("no closing write marked")
	}
	program, err := lowered(t, path)
	if err == nil {
		checkGraphAdmission(t, path, program)
		return
	}
	var refused *lower.Refused
	if !errors.As(err, &refused) || !strings.Contains(err.Error(), "adamic/cycle-capable") || !strings.Contains(err.Error(), fmt.Sprintf("the write at %s:%d:", path, marked)) {
		t.Fatalf("want marked cycle refusal, got %v", err)
	}
	t.Logf("Node stdout %q; %v", result.stdout, err)
}

func TestFreshClobberAliasFill(t *testing.T) {
	t.Parallel()
	freshClobberWitness(t, "alias_fill")
}

func TestFreshClobberAliasMapset(t *testing.T) {
	t.Parallel()
	freshClobberWitness(t, "alias_mapset")
}

func TestFreshClobberAliasReverse(t *testing.T) {
	t.Parallel()
	freshClobberWitness(t, "alias_reverse")
}

func TestFreshClobberAliasSetadd(t *testing.T) {
	t.Parallel()
	freshClobberWitness(t, "alias_setadd")
}

func TestFreshClobberAliasSort(t *testing.T) {
	t.Parallel()
	freshClobberWitness(t, "alias_sort")
}

func TestFreshClobberClobberLinkLoop(t *testing.T) {
	t.Parallel()
	freshClobberWitness(t, "clobber_link_loop")
}

func TestFreshClobberEscapedBefore(t *testing.T) {
	t.Parallel()
	freshClobberWitness(t, "escaped_before")
}

func TestFreshClobberMapEntriesPattern(t *testing.T) {
	t.Parallel()
	freshClobberWitness(t, "map_entries_pattern")
}

func TestFreshClobberSpreadOutside(t *testing.T) {
	t.Parallel()
	freshClobberWitness(t, "spread_outside")
}

func TestFreshClobberThrowKeepsOld(t *testing.T) {
	t.Parallel()
	freshClobberWitness(t, "throw_keeps_old")
}
