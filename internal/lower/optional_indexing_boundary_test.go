package lower

import (
	"errors"
	"os"
	"testing"
)

func TestOptionalIndexingMapShapeRefused(test *testing.T) {
	test.Parallel()
	source, readError := os.ReadFile("../../review/optional-indexing/map-shape.a")
	if readError != nil {
		test.Fatal(readError)
	}
	_, lowerSourceError := lowerSource(test, string(source))
	// The literal stops where it goes into the ReadonlyMap slot (container_fit.go), before any get on it.
	var notYet *NotYet
	if !errors.As(lowerSourceError, &notYet) || notYet.What != containerFitStop {
		test.Fatalf("got %v, want the literal stopped where a map is held", lowerSourceError)
	}
}
