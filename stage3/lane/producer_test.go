package lane

import (
	"path/filepath"
	"testing"

	"github.com/system-inc/adamic/stage3/applyproducts"
)

func TestStage3LaneAdaptedTreeMatchesProducer(t *testing.T) {
	t.Parallel()
	consumer := adaptedTree(t)
	producer, err := applyproducts.Get()
	if err != nil {
		t.Fatal(err)
	}
	consumer, err = filepath.EvalSymlinks(consumer)
	if err != nil {
		t.Fatal(err)
	}
	producer, err = filepath.EvalSymlinks(producer)
	if err != nil {
		t.Fatal(err)
	}
	if consumer != producer {
		t.Fatalf("adapted product key differs from apply producer: consumer %s, producer %s", consumer, producer)
	}
}
