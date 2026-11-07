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

func TestGenericArrowObjectShapesAgreeWithNode(test *testing.T) {
	test.Parallel()
	path, absolutePathError := filepath.Abs(filepath.Join(repository, genericArrowObjectShapesFixture))
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	program, loweredError := lowered(test, path)
	if loweredError != nil {
		test.Fatal(loweredError)
	}
	want := onNode(test, path)
	if want.exitCode != 0 || string(want.stdout) != "11\n22\n" || len(want.stderr) != 0 {
		test.Fatalf("Node probe: %+v", want)
	}
	native, leaked, comparable := nativelyCheckingLeaks(test, program, true)
	if leaked != "" {
		test.Fatalf("probe leaks: %s", leaked)
	}
	if !comparable {
		test.Fatal("native probe was not comparable")
	}
	for name, got := range map[string]run{"native": native, "JavaScript": onJavaScriptBackend(test, program)} {
		if difference := disagreement(want, got); difference != "" {
			test.Fatalf("%s: %s: %+v", name, difference, got)
		}
	}
	test.Log("different object layouts match Node in both backends, with no native leaks")
}
