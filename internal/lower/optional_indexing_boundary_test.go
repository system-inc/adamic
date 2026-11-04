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
	// The literal stops where it goes into the ReadonlyMap slot (container_fit.go), before any get on it.
	var notYet *NotYet
	if !errors.As(err, &notYet) || notYet.What != containerFitStop {
		t.Fatalf("got %v, want the literal stopped where a map is held", err)
	}
}
