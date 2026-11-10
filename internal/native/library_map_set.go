package native

import (
	"fmt"

	"github.com/system-inc/adamic/internal/ir"
)

func (e *emitter) collectionIterator(iterator ir.CollectionIterator) string {
	part := 1
	if iterator.Part == "values" {
		part = 2
	}
	if iterator.Part == "entries" {
		part = 3
	}
	// An entries() pair whose tuple type is a graph type is adopted as next() makes it (#dps0kp6).
	constructor := "adamic_collection_iterator"
	if e.graphTypes(iterator.PairGraphTypes) {
		constructor = "adamic_collection_iterator_planned"
	}
	return e.own(ir.Object, fmt.Sprintf("%s(%s, %d, %d, %d, %t)", constructor, e.value(iterator.Collection), part, iterator.Key, iterator.Value, iterator.Set))
}
