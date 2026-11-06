package oracle

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestFunctionBindLengthFunctionSlot(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/bind_refused/length_function_slot.ts"))
	if err != nil {
		t.Fatal(err)
	}
	truth := onNode(t, path)
	if truth.exitCode != 0 || string(truth.stdout) != "1\n" {
		t.Fatalf("source Node: %+v", truth)
	}
	_, err = lowered(t, path)
	if err == nil || !strings.Contains(err.Error(), "a dynamic-shape object in a graph program") || !strings.Contains(err.Error(), "length_function_slot.ts:1:") {
		t.Fatalf("want existing graph/dynamic stop at Function.length, got %v", err)
	}
}
