package oracle

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/native"
)

func TestDateISOStringRuntimeMutantsAreCaught(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/date_iso_string.a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	truth := onNode(t, path)
	if truth.exitCode != 0 || len(truth.stderr) != 0 {
		t.Fatalf("Node fixture: %+v", truth)
	}
	original, err := os.ReadFile(filepath.Join(repository, "internal/native_runtime/c/date.c"))
	if err != nil {
		t.Fatal(err)
	}
	for _, mutant := range []struct {
		name, before, after string
		lines               []int
	}{
		{"expanded_year", "if (year >= 0 && year <= 9999) {", "if (1) {", []int{6, 7}},
		{"floored_day", "if (within < 0) { days--; within += 86400000; }", "(void)0;", []int{3, 4, 11}},
		{"milliseconds", "int fraction = (int)(within % 1000);", "int fraction = 0;", []int{12, 13}},
		{"invalid_date", `adamic_throw_kind("RangeError", sizeof "RangeError" - 1, message, sizeof message - 1);`, `(void)message;`, []int{15}},
	} {
		t.Run(mutant.name, func(t *testing.T) {
			source := string(original)
			if strings.Count(source, mutant.before) != 1 {
				t.Fatal("mutant seam must be unique")
			}
			source = strings.Replace(source, mutant.before, mutant.after, 1)
			if mutant.name == "invalid_date" {
				source = strings.Replace(source, "return NULL;", `return adamic_decode_utf8((const unsigned char *)"invalid", 7);`, 1)
			}
			source = strings.ReplaceAll(source, "adamic_date_to_iso_string(", "mutant_date_to_iso_string(")
			code := strings.ReplaceAll(native.C(program), "adamic_date_to_iso_string(", "mutant_date_to_iso_string(")
			code = strings.Replace(code, `#include "adamic.h"`, `#include "adamic.h"`+"\n"+source, 1)
			binary := filepath.Join(t.TempDir(), "mutant")
			if err := native.Build(code, binary, native.Options{Sanitize: true}); err != nil {
				t.Fatal(err)
			}
			got := execute(t, binary)
			if got.exitCode != 0 || len(got.stderr) != 0 {
				t.Fatalf("mutant must reach Node comparison cleanly: %+v", got)
			}
			if bytes.Equal(got.stdout, truth.stdout) {
				t.Fatal("Node stdout did not catch Date mutant")
			}
			wantLines, gotLines := strings.Split(string(truth.stdout), "\n"), strings.Split(string(got.stdout), "\n")
			for _, line := range mutant.lines {
				if len(gotLines) < line || gotLines[line-1] == wantLines[line-1] {
					t.Fatalf("Date mutant must change stdout line %d", line)
				}
				t.Logf("caught stdout line %d: Node %q; mutant %q", line, wantLines[line-1], gotLines[line-1])
			}
		})
	}
}
