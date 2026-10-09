package oracle

import (
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/native"
)

// The fractional cases must reach number.c rather than a folded Go constant.
func TestNumbersFractionalPowersReachRuntime(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/numbers.a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	code := native.C(program)
	quarter := regexp.MustCompile(`adamic_power\([^\n]+, \(0x1p-0?2\)\)`)
	if calls := quarter.FindAllString(code, -1); len(calls) != 2 {
		t.Fatalf("want two quarter-power runtime calls, got %v", calls)
	}
	for _, line := range strings.Split(code, "\n") {
		if strings.Contains(line, "adamic_power(") {
			t.Log(line)
		}
	}
}
