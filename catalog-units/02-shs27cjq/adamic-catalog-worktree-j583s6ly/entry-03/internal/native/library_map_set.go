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
	return e.own(ir.Object, fmt.Sprintf("adamic_collection_iterator(%s, %d, %d, %d, %t)", e.value(iterator.Collection), part, iterator.Key, iterator.Value, iterator.Set))
}
