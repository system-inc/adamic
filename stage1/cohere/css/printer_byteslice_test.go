package css

import (
	"bytes"
	"encoding/json"
	"fmt"
	"path/filepath"
	"testing"
	"unicode/utf16"
)

func TestCSSPrinterByteSlicesMatchGo(t *testing.T) {
	t.Parallel()
	var want bytes.Buffer
	encoder := json.NewEncoder(&want)
	encoder.SetEscapeHTML(false)
	inputs := []string{"", "ASCII\t\n", "aé€😀z", "é\u2028中"}
	for _, text := range inputs {
		size := len(text)
		clamp := func(offset int) int {
			if offset < 0 {
				return max(0, size+offset)
			}
			return min(offset, size)
		}
		for start := -size - 2; start <= size+2; start++ {
			a := clamp(start)
			fmt.Fprintln(&want, len(utf16.Encode([]rune(text[:a]))))
			for end := -size - 2; end <= size+2; end++ {
				b := clamp(end)
				value := ""
				// One RuneError per invalid byte is Go's JSON-rendered slice value.
				if a < b {
					value = string([]rune(text[a:b]))
				}
				if err := encoder.Encode(value); err != nil {
					t.Fatal(err)
				}
			}
		}
	}
	path, err := filepath.Abs("testdata/byte_slices.ts")
	if err != nil {
		t.Fatal(err)
	}
	program := lowered(t, path)
	result, binary := natively(t, program)
	for _, side := range []struct {
		name   string
		result run
	}{
		{"sanitized native", result}, {"Node", onNode(t, path)}, {"emitted JavaScript", onJavaScriptBackend(t, program)},
	} {
		if side.result.exitCode != 0 || len(side.result.stderr) != 0 || !bytes.Equal(side.result.stdout, want.Bytes()) {
			t.Fatalf("%s byte slices: exit %d stderr %s difference %s", side.name, side.result.exitCode, side.result.stderr, firstDifference(string(side.result.stdout), want.String()))
		}
	}
	if report := printerLeaks(t, binary); report != "" {
		t.Fatal(report)
	}
	t.Logf("%d byte-identical Go slice/prefix results", bytes.Count(want.Bytes(), []byte{'\n'}))
}
