package oracle

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/lower"
)

// Represented methods have no mutable data slot. Keep the refusal until runtime
// replacement has a language ruling and an implementation of instance dispatch.
func TestClassSetPropertyMethodRefusal(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/class_set_property_method.a"))
	if err != nil {
		t.Fatal(err)
	}
	source := onNode(t, path)
	if source.exitCode != 0 || string(source.stdout) != "1 1\n2 1\n" || len(source.stderr) != 0 {
		t.Fatalf("source Node: %+v", source)
	}
	_, err = lowered(t, path)
	var stopped *lower.Refused
	if !errors.As(err, &stopped) || !strings.HasPrefix(stopped.What, "a method read as a value") {
		t.Fatalf("want represented method refusal, got %v", err)
	}
}
