package adaptedtree

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

// TestPerformanceCoreYewmv46PinsEqualAdaptedProductBytes holds Lower's two
// landed-in-line pins to the keyed adapted-tree product used by stage3. Product
// preparation belongs to TestProduct_Stage3AdaptedTree, ahead of this unit.
// SplitAfter retains every CR and LF; there is no normalization or trimming.
func TestPerformanceCoreYewmv46PinsEqualAdaptedProductBytes(t *testing.T) {
	t.Parallel()
	product := Product(t)
	t.Logf("adapted product: %s", product)
	source, err := os.ReadFile(filepath.Join(product, "tree", "src/compiler/performanceCore.ts"))
	if err != nil {
		t.Fatal(err)
	}
	lines := bytes.SplitAfter(source, []byte("\n"))
	pins := []struct {
		file        string
		first, last int
	}{
		{"performance_core_adapted.txt", 7, 29},
		{"performance_core_hooks.txt", 62, 88},
	}
	for _, pin := range pins {
		t.Run(pin.file, func(t *testing.T) {
			t.Parallel()
			if len(lines) < pin.last {
				t.Fatalf("performanceCore.ts has fewer than %d lines", pin.last)
			}
			actual := bytes.Join(lines[pin.first-1:pin.last], nil)
			// Reads: internal/lower/testdata/host_binding/performance_core_adapted.txt
			// Reads: internal/lower/testdata/host_binding/performance_core_hooks.txt
			expected, err := os.ReadFile(filepath.Join(repositoryFromLane, "internal/lower/testdata/host_binding", pin.file))
			if err != nil {
				t.Fatal(err)
			}
			if bytes.Equal(actual, expected) {
				return
			}
			offset := 0
			for offset < len(actual) && offset < len(expected) && actual[offset] == expected[offset] {
				offset++
			}
			sourceLine := pin.first + bytes.Count(actual[:offset], []byte("\n"))
			sourceColumn := offset - bytes.LastIndexByte(actual[:offset], '\n')
			pinLine := 1 + bytes.Count(expected[:offset], []byte("\n"))
			pinColumn := offset - bytes.LastIndexByte(expected[:offset], '\n')
			t.Errorf("%s differs from performanceCore.ts lines %d-%d at byte offset %d (zero-based): source line %d byte-column %d is %s, pin line %d byte-column %d is %s; lengths %d/%d",
				pin.file, pin.first, pin.last, offset, sourceLine, sourceColumn, performancePinByte(actual, offset), pinLine, pinColumn, performancePinByte(expected, offset), len(actual), len(expected))
		})
	}
}

func performancePinByte(data []byte, offset int) string {
	if offset == len(data) {
		return "EOF"
	}
	return fmt.Sprintf("0x%02x", data[offset])
}
