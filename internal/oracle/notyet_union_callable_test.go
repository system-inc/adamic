package oracle

import (
	"path/filepath"
	"testing"
)

func init() {
	for _, path := range []string{
		"internal/oracle/testdata/notyet_union_callable_name.a",
		"internal/oracle/testdata/notyet_union_callable_location.a",
	} {
		fixtures = append(fixtures, struct {
			path            string
			lowers, checked bool
		}{path, false, false})
	}
}

func init() {
	for _, path := range []string{
		"internal/oracle/testdata/notyet_union_callable_adapter.a",
		"internal/oracle/testdata/notyet_union_callable_adapter_location.a",
	} {
		fixtures = append(fixtures, struct {
			path            string
			lowers, checked bool
		}{path, false, false})
	}
}

// Native method tables currently exclude union parameter adapters. Keep the
// caller stop until closure signatures and method adapters share a boxed ABI.
func TestUnionCallableSourceObservations(t *testing.T) {
	t.Parallel()
	for fixture, expected := range map[string]string{
		"notyet_union_callable_name.a":             "name1\nnode2\n",
		"notyet_union_callable_location.a":         "1:3\n2:3\nmissing:3\n",
		"notyet_union_callable_adapter.a":          "name1\nnode2\nmissing\ndirect3\nmissing\n",
		"notyet_union_callable_adapter_location.a": "1:3\n2:3\nmissing:3\nno resolver\n",
	} {
		t.Run(fixture, func(t *testing.T) {
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata", fixture))
			if err != nil {
				t.Fatal(err)
			}
			observed := onNode(t, path)
			if observed.exitCode != 0 || string(observed.stdout) != expected || len(observed.stderr) != 0 {
				t.Fatalf("Node observation: %#v", observed)
			}
			t.Logf("Node: %q", observed.stdout)
		})
	}
}
