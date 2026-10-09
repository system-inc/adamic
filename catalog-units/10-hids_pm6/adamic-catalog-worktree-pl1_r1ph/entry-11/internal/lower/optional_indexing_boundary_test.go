package lower

import (
	"errors"
	"os"
	"testing"
)

func TestOptionalIndexingMapShapeRefused(t *testing.T) {
	t.Parallel()
	source, err := os.ReadFile("../../review/optional-indexing/map-shape.a")
	if err != nil {
		t.Fatal(err)
	}
	_, err = lowerSource(t, string(source))
	var notYet *NotYet
	if !errors.As(err, &notYet) || notYet.What != "an optional Map get through a structural receiver" {
		t.Fatalf("got %v, want structural Map receiver boundary", err)
	}
}
