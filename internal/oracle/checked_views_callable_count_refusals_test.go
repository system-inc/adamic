package oracle

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/lower"
)

// Judgment 10 retains these never-rest boundaries. They have no executable
// runtime to count; require the same refusal and independent Node observation
// before omitting a count row.
func ruledCallableCountRefusal(t *testing.T, relative string) bool {
	t.Helper()
	const root = "stage3/interface-downcasts/lane5/"
	relative = filepath.ToSlash(relative)
	if !strings.HasPrefix(relative, root) {
		return false
	}
	probe := strings.TrimPrefix(relative, root)
	stdout := ""
	storedCall := false
	switch probe {
	case "marker/parser-cache.a", "marker/stored-call.a":
		stdout = "function view loaded\n"
	case "marker/discarded-required.a", "stored-marker/boolean.a", "stored-marker/number.a", "stored-marker/string.a", "stored-marker/object.a", "stored-marker/array.a", "stored-marker/wrong-arity.a":
		stdout = "producer\ncalled\n"
	case "stored-marker/record.a":
		stdout = "called\n"
	case "stored-marker/observed.a":
		stdout = "observed\n"
	case "stored-marker/void.a", "stored-marker/field-good.a":
		stdout = "producer\ncalled\n"
		storedCall = true
	default:
		return false
	}
	path, err := filepath.Abs(filepath.Join(repository, relative))
	if err != nil {
		t.Fatal(err)
	}
	truth := onNode(t, path)
	if truth.exitCode != 0 || string(truth.stdout) != stdout || len(truth.stderr) != 0 {
		t.Fatalf("Node %s: %#v", relative, truth)
	}
	_, err = lowered(t, path)
	if storedCall {
		var notYet *lower.NotYet
		if !errors.As(err, &notYet) || !strings.Contains(err.Error(), "a call through an erased never-rest callable marker") {
			t.Fatalf("want ruled stored-marker call refusal for %s, got %v", relative, err)
		}
	} else {
		var refused *lower.Refused
		if !errors.As(err, &refused) || !strings.Contains(err.Error(), "adamic/invariant-mutable") || !strings.Contains(err.Error(), "write void where") {
			t.Fatalf("want ruled marker result refusal for %s, got %v", relative, err)
		}
	}
	return true
}
