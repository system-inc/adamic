package native

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
)

func TestNoReaderCallingConvention(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct {
		file    string
		counted bool
	}{
		{"closure_convention_plain.a", false},
		{"arguments_length_no_reader.a", true},
		{"arguments_length_unrelated_type.a", true},
	} {
		t.Run(probe.file, func(t *testing.T) {
			t.Parallel()
			checked, err := load.Load([]string{filepath.Join("..", "oracle", "testdata", probe.file)})
			if err != nil {
				t.Fatal(err)
			}
			program, err := lower.Lower(context.Background(), checked)
			if err != nil {
				t.Fatal(err)
			}
			code := C(program)
			if strings.Contains(code, "/* actual argument count */") {
				t.Fatal("count must use its typed parameter, not a buffer slot")
			}
			if got := strings.Contains(code, "#define ADAMIC_CLOSURE_CONVENTION 1"); got != probe.counted {
				t.Fatalf("counted convention: got %t, want %t", got, probe.counted)
			}
			if probe.file == "arguments_length_unrelated_type.a" && strings.Count(code, "= adamic_closure_call(") != 1 {
				t.Fatal("count reached an unrelated function type")
			}
			if !probe.counted && (strings.Contains(code, "argument_count") || strings.Contains(code, "adamic_closure_call(")) {
				t.Fatal("unobserved program contains a count binding")
			}

		})
	}
}

// The third canary of tools 26226fda selects this package (developer tools, Oct 9).
