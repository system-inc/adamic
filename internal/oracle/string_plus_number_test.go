package oracle

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/native"
)

func init() {
	for _, name := range []string{"scanner_string_number_696", "scanner_string_number_1221", "string_plus_number"} {
		fixtures = append(fixtures, struct {
			path    string
			lowers  bool
			checked bool
		}{"internal/oracle/testdata/" + name + ".a", true, false})
	}
}

// The mutant completes under sanitizers; only Node's number spelling catches it.
func TestStringPlusNumberOracleCatchesFormatterMutant(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/string_plus_number.a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	source := native.C(program)
	const original = "adamic_string_from_number("
	if !strings.Contains(source, original) {
		t.Fatal("number conversion absent")
	}
	source = strings.ReplaceAll(source, original, "wrong_number_format(")
	const include = `#include "adamic.h"`
	const helper = `
#include <stdio.h>
static adamic_string *wrong_number_format(double value) {
    char buffer[64];
    int length = snprintf(buffer, sizeof(buffer), "%g", value);
    return adamic_decode_utf8((const unsigned char *)buffer, (size_t)length);
}
`
	if !strings.Contains(source, include) {
		t.Fatal("runtime include absent")
	}
	source = strings.Replace(source, include, include+helper, 1)
	binary := filepath.Join(t.TempDir(), "mutant")
	if err := native.Build(source, binary, native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	result := execute(t, binary)
	if result.exitCode != 0 || len(result.stderr) != 0 {
		t.Fatalf("mutant must finish cleanly: exit %d, stderr %q", result.exitCode, result.stderr)
	}
	oracle := onNode(t, path)
	if difference := disagreement(oracle, result); difference != "stdout differs" {
		t.Fatalf("got %q, want stdout differs", difference)
	}
	// Pin the failure to each new operation, independently of the template control.
	for _, wrong := range []string{"left:-0\n", "-0:right\n", "slot:-0\n"} {
		if !strings.Contains(string(result.stdout), wrong) || strings.Contains(string(oracle.stdout), wrong) {
			t.Fatalf("want the formatter mutant alone to print %q", wrong)
		}
	}
	t.Log("Node caught the general-format mutant: stdout differs; sanitizer clean")
}
