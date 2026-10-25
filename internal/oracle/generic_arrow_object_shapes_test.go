package oracle

import (
	"path/filepath"
	"testing"
)

const genericArrowObjectShapesFixture = "internal/oracle/testdata/generic_arrow_object_shapes.a"

func init() {
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{genericArrowObjectShapesFixture, true, false})
}

func TestGenericArrowObjectShapesAgreeWithNode(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, genericArrowObjectShapesFixture))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	want := onNode(t, path)
	if want.exitCode != 0 || string(want.stdout) != "11\n22\n" || len(want.stderr) != 0 {
		t.Fatalf("Node probe: %+v", want)
	}
	native, leaked, comparable := nativelyCheckingLeaks(t, program, true)
	if leaked != "" {
		t.Fatalf("probe leaks: %s", leaked)
	}
	if !comparable {
		t.Fatal("native probe was not comparable")
	}
	for name, got := range map[string]run{"native": native, "JavaScript": onJavaScriptBackend(t, program)} {
		if difference := disagreement(want, got); difference != "" {
			t.Fatalf("%s: %s: %+v", name, difference, got)
		}
	}
	t.Log("different object layouts match Node in both backends, with no native leaks")
}
