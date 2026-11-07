package native

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
)

func TestNoReaderCallingConvention(test *testing.T) {
	test.Parallel()
	for _, probe := range []struct {
		file    string
		counted bool
	}{
		{"closure_convention_plain.a", false},
		{"arguments_length_no_reader.a", true},
		{"arguments_length_unrelated_type.a", true},
	} {
		test.Run(probe.file, func(test *testing.T) {
			test.Parallel()
			checked, loadError := load.Load([]string{filepath.Join("..", "oracle", "testdata", probe.file)})
			if loadError != nil {
				test.Fatal(loadError)
			}
			program, lowerError := lower.Lower(context.Background(), checked)
			if lowerError != nil {
				test.Fatal(lowerError)
			}
			code := C(program)
			if strings.Contains(code, "/* actual argument count */") {
				test.Fatal("count must use its typed parameter, not a buffer slot")
			}
			if got := strings.Contains(code, "#define ADAMIC_CLOSURE_CONVENTION 1"); got != probe.counted {
				test.Fatalf("counted convention: got %t, want %t", got, probe.counted)
			}
			if probe.file == "arguments_length_unrelated_type.a" && strings.Count(code, "= adamic_closure_call(") != 1 {
				test.Fatal("count reached an unrelated function type")
			}
			if !probe.counted &&
				(strings.Contains(code, "argument_count") || strings.Contains(code, "adamic_closure_call(")) {
				test.Fatal("unobserved program contains a count binding")
			}

		})
	}
}
