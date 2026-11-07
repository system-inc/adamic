package lower

import (
	"errors"
	"os"
	"testing"
)

func TestOptionalIndexingMapShapeRefused(test *testing.T) {
	test.Parallel()
	source, err := os.ReadFile("../../review/optional-indexing/map-shape.a")
	if err != nil {
		test.Fatal(err)
	}
	_, err = lowerSource(test, string(source))
	// The literal stops where it goes into the ReadonlyMap slot (container_fit.go), before any get on it.
	var notYet *NotYet
	if !errors.As(err, &notYet) || notYet.What != containerFitStop {
		test.Fatalf("got %v, want the literal stopped where a map is held", err)
	}
}
