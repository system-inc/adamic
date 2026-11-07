package oracle

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/lower"
)

func init() {
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{"internal/oracle/testdata/scout_nullable_objects.a", true, false})
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{"internal/oracle/testdata/scout_nullable_object_stale_field.a", true, true})
}

func init() {
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{"internal/oracle/testdata/scout_nullable_object_null_field.a", true, false})
}

func init() {
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{"internal/oracle/testdata/scout_nullable_object_stale_local.a", true, true})
}

func TestNullableObjectMapKeepsArrayStorageViewNotYet(test *testing.T) {
	test.Parallel()
	path, err := filepath.Abs(
		filepath.Join(repository, "internal/oracle/testdata/scout_nullable_storage_refused/array_callback.a"),
	)
	if err != nil {
		test.Fatal(err)
	}
	node := onNode(test, path)
	if node.exitCode != 0 || string(node.stdout) != "null\n" || len(node.stderr) != 0 {
		test.Fatalf("Node: exit=%d stdout=%q stderr=%q", node.exitCode, node.stdout, node.stderr)
	}
	_, err = lowered(test, path)
	var pending *lower.NotYet
	if !errors.As(err, &pending) ||
		!strings.Contains(err.Error(), "sharing nullable object storage with different null/undefined tags") {
		test.Fatalf("want shared-storage NotYet, got %v", err)
	}
}

// A search adapts only its callback's value parameter, so an array parameter
// viewing the receiver's storage with different nullable tags still stops.
func TestNullableObjectSearchKeepsArrayStorageViewNotYet(test *testing.T) {
	test.Parallel()
	path, err := filepath.Abs(
		filepath.Join(repository, "internal/oracle/testdata/scout_nullable_storage_refused/search_callback.a"),
	)
	if err != nil {
		test.Fatal(err)
	}
	node := onNode(test, path)
	if node.exitCode != 0 || string(node.stdout) != "0\n" || len(node.stderr) != 0 {
		test.Fatalf("Node: exit=%d stdout=%q stderr=%q", node.exitCode, node.stdout, node.stderr)
	}
	_, err = lowered(test, path)
	var pending *lower.NotYet
	if !errors.As(err, &pending) ||
		!strings.Contains(err.Error(), "sharing nullable object storage with different null/undefined tags") {
		test.Fatalf("want shared-storage NotYet, got %v", err)
	}
}
