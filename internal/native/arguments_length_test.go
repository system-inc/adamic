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
	for _, probe := range []struct {
		file  string
		slots int
	}{
		{"arguments_length_no_reader.a", 0},
		{"arguments_length_unrelated_type.a", 1},
	} {
		t.Run(probe.file, func(t *testing.T) {
			checked, err := load.Load([]string{filepath.Join("..", "oracle", "testdata", probe.file)})
			if err != nil {
				t.Fatal(err)
			}
			program, err := lower.Lower(context.Background(), checked)
			if err != nil {
				t.Fatal(err)
			}
			code := C(program)
			if got := strings.Count(code, "/* actual argument count */"); got != probe.slots {
				t.Fatalf("hidden count slots: got %d, want %d", got, probe.slots)
			}
			if strings.Contains(code, "size_t argument_count") {
				t.Fatal("packed convention must retain main's two-argument ABI")
			}
			if probe.slots == 0 && strings.Contains(code, "argument_count") {
				t.Fatal("readerless program contains a count binding")
			}
		})
	}
}
