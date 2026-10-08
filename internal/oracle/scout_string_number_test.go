package oracle

import (
	"path/filepath"
	"strings"
	"testing"
)

func init() {
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{"internal/oracle/testdata/scout_string_number_fields_locals.a", true, false})
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{"internal/oracle/testdata/scout_string_number_stale_field.a", true, true})
}

func TestNullableObjectStorageViewOutcomes(t *testing.T) {
	for _, name := range []string{"array", "field", "callable", "array-spread", "object-spread", "lookup"} {
		t.Run(name, func(t *testing.T) {
			path, err := filepath.Abs(filepath.Join(repository, "docs/step-17-unions/probes/storage", name+".a"))
			if err != nil {
				t.Fatal(err)
			}
			if difference := disagreement(run{stdout: []byte("true\n")}, onNode(t, path)); difference != "" {
				t.Fatal(difference)
			}
			_, err = lowered(t, path)
			if err == nil || !(strings.Contains(err.Error(), "sharing nullable object storage") || name == "array-spread" && strings.Contains(err.Error(), "spreading an array of other elements") || name == "lookup" && strings.Contains(err.Error(), "a nullable object lookup without separate null/undefined tags")) {
				t.Fatalf("want a storage view stop, got %v", err)
			}
			t.Log(err)
		})
	}
}
