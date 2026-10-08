package flow

import (
	"path/filepath"
	"testing"
)

// These source controls pin compile-time stops and therefore have no IR graph.
func classSetPropertyRefusalFixture(path string) bool {
	switch filepath.Base(path) {
	case "class_set_property_array_dynamic.a", "class_set_property_any.a", "class_set_property_method.a", "class_set_property_intersection_boundary.a":
		return true
	}
	return false
}

func TestClassSetPropertyFlow(t *testing.T) {
	paths, err := filepath.Glob("../oracle/testdata/class_set_property_*.a")
	if err != nil {
		t.Fatal(err)
	}
	if len(paths) < 8 {
		t.Fatalf("missing field controls: %d", len(paths))
	}
	for _, path := range paths {
		if classSetPropertyRefusalFixture(path) {
			continue
		}
		program := lowered(t, path)
		for function := -1; function < len(program.Functions); function++ {
			checkReads(t, path, program, function)
			graph := Build(program, function)
			reaching := reachingDefinitions(graph)
			Construct(graph)
			for _, violation := range VerifySSA(graph) {
				t.Errorf("%s: %s", path, violation)
			}
			checkReaching(t, path, graph, reaching)
		}
	}
}
