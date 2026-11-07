package css

import (
	"bytes"
	"encoding/json"
	"fmt"
	"path/filepath"
	"testing"
	"unicode/utf16"
)

func TestCSSPrinterByteSlicesMatchGo(test *testing.T) {
	test.Parallel()
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
				if encodeError := encoder.Encode(value); encodeError != nil {
					test.Fatal(encodeError)
				}
			}
		}
	}
	path, absolutePathError := filepath.Abs("testdata/byte_slices.ts")
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	program := lowered(test, path)
	binary := filepath.Join(programSanitized(test, "testdata/byte_slices.ts"), "port")
	result := sanitizedRun(test, binary)
	for _, side := range []struct {
		name   string
		result run
	}{
		{"sanitized native", result},
		{"Node", onNode(test, path)},
		{"emitted JavaScript", onJavaScriptBackend(test, program)},
	} {
		if side.result.exitCode != 0 || len(side.result.stderr) != 0 || !bytes.Equal(side.result.stdout, want.Bytes()) {
			test.Fatalf(
				"%s byte slices: exit %d stderr %s difference %s",
				side.name,
				side.result.exitCode,
				side.result.stderr,
				firstDifference(string(side.result.stdout), want.String()),
			)
		}
	}
	if report := printerLeaks(test, binary); report != "" {
		test.Fatal(report)
	}
	test.Logf("%d byte-identical Go slice/prefix results", bytes.Count(want.Bytes(), []byte{'\n'}))
}
