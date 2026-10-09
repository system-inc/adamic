package oracle

import (
	"context"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"path/filepath"
	"testing"
)

func checkCallbackWideningRefusal(t *testing.T, name, parameter, method, want string) {
	t.Helper()
	path, _ := filepath.Abs("../../review/compiler/callback-widening/fixtures/number-" + name + ".a")
	node := onNode(t, path)
	if diff := disagreement(run{stdout: []byte(want)}, node); diff != "" {
		t.Fatal(diff)
	}
	loaded, err := load.Load([]string{path})
	if err != nil {
		t.Fatal(err)
	}
	_, err = lower.Lower(context.Background(), loaded)
	refused, ok := err.(*lower.Refused)
	if !ok {
		t.Fatalf("expected narrow callback refusal, got %v", err)
	}
	expected := "a " + method + " callback parameter " + parameter + " of type string | number receiving array element type number without a representation adapter"
	if refused.What != expected || refused.Fix != "annotate parameter "+parameter+" as the element type number" {
		t.Fatalf("wrong refusal: %#v", refused)
	}
}
func TestCallbackWideningRefuseMap(t *testing.T) {
	t.Parallel()
	checkCallbackWideningRefusal(t, "map", "x", "map", "1,2,3\n")
}
func TestCallbackWideningRefuseFilter(t *testing.T) {
	t.Parallel()
	checkCallbackWideningRefusal(t, "filter", "x", "filter", "1,3\n")
}
func TestCallbackWideningRefuseForEach(t *testing.T) {
	t.Parallel()
	checkCallbackWideningRefusal(t, "forEach", "x", "forEach", "1\n2\n3\n")
}
func TestCallbackWideningRefuseFind(t *testing.T) {
	t.Parallel()
	checkCallbackWideningRefusal(t, "find", "x", "find", "2\n")
}
func TestCallbackWideningRefuseSome(t *testing.T) {
	t.Parallel()
	checkCallbackWideningRefusal(t, "some", "x", "some", "true\n")
}
func TestCallbackWideningRefuseEvery(t *testing.T) {
	t.Parallel()
	checkCallbackWideningRefusal(t, "every", "x", "every", "true\n")
}
func TestCallbackWideningRefuseReduce(t *testing.T) {
	t.Parallel()
	checkCallbackWideningRefusal(t, "reduce", "x", "reduce", "123\n")
}
func TestCallbackWideningRefuseSort(t *testing.T) {
	t.Parallel()
	checkCallbackWideningRefusal(t, "sort", "a", "sort", "1,2,3\n")
}
