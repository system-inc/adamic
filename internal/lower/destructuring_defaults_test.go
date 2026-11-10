package lower

import (
	"errors"
	"os"
	"testing"
)

func TestDestructuringMaybeBooleanStops(t *testing.T) {
	t.Parallel()
	source, err := os.ReadFile("../oracle/testdata/destructure_union_refused/maybe_boolean.a")
	if err != nil {
		t.Fatal(err)
	}
	_, err = lowerSource(t, string(source))
	var stop *NotYet
	if !errors.As(err, &stop) || stop.What != "a destructured name held otherwise than its field" {
		t.Fatalf("want the optional boolean field refusal, got %v", err)
	}
}
