package oracle

import (
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/native"
)

func init() {
	fixtures = append(fixtures, struct {
		path    string
		lowers  bool
		checked bool
	}{"internal/oracle/testdata/pow_fractional.a", true, false})
}

// The fractional cases must reach number.c rather than a folded Go constant.
func TestFractionalPowersReachRuntime(test *testing.T) {
	test.Parallel()
	path, absolutePathError := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/pow_fractional.a"))
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	program, loweredError := lowered(test, path)
	if loweredError != nil {
		test.Fatal(loweredError)
	}
	code := native.C(program)
	quarter := regexp.MustCompile(`adamic_power\([^\n]+, \(0x1p-0?2\)\)`)
	if calls := quarter.FindAllString(code, -1); len(calls) != 2 {
		test.Fatalf("want two quarter-power runtime calls, got %v", calls)
	}
	for _, line := range strings.Split(code, "\n") {
		if strings.Contains(line, "adamic_power(") {
			test.Log(line)
		}
	}
}
