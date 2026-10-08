package oracle

import (
	"path/filepath"
	"testing"
)

func init() {
	for _, path := range []string{"internal/oracle/testdata/object_property_function_length.a", "internal/oracle/testdata/object_property_function_length_missing.a"} {
		fixtures = append(fixtures, struct {
			path            string
			lowers, checked bool
		}{path, true, false})
	}
}

// Runtime arity is not the number of parameters in a callback's static signature.
func TestObjectPropertyFunctionLengthNodeWitness(t *testing.T) {
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/object_property_function_length.a"))
	if err != nil {
		t.Fatal(err)
	}
	result := onNode(t, path)
	if result.exitCode != 0 || string(result.stdout) != "1/2/1/0\n1/2/2\n1/2\n0/0/1\n-1/1\n-1/1\n2\n" {
		t.Fatalf("Node arities: exit %d, stdout %q, stderr %q", result.exitCode, result.stdout, result.stderr)
	}
}
