package oracle

import (
	"errors"
	"github.com/system-inc/adamic/internal/lower"
	"path/filepath"
	"strings"
	"testing"
)

func init() {
	for _, name := range []string{"02_scanner_digits", "03_scanner_hoisting", "04_scanner_escaped_text", "05_scanner_reassigned_text", "06_parser_token_state", "07_binder_symbol_count", "08_checker_symbol_recursion", "10_checker_arrow_cleanup", "11_scanner_method_wrapper"} {
		fixtures = append(fixtures, struct {
			path    string
			lowers  bool
			checked bool
		}{"stage3/fixtures/nested-functions/" + name + ".a", true, false})
	}
}

func TestRealNestedBlockers(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct {
		name, want string
		refused    bool
	}{
		{"01_scanner_frame", "comma operator", true},
		{"09_checker_constituent_recursion", "PrefixUnaryExpression on a number", false},
	} {
		t.Run(probe.name, func(t *testing.T) {
			path := filepath.Join(repository, "stage3/fixtures/nested-functions", probe.name+".a")
			_, err := lowered(t, path)
			var refused *lower.Refused
			var notYet *lower.NotYet
			if !strings.Contains(errorText(err), probe.want) || (probe.refused && !errors.As(err, &refused)) || (!probe.refused && !errors.As(err, &notYet)) {
				t.Fatalf("want blocker %s, got %v", probe.want, err)
			}
		})
	}
}
func errorText(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}
