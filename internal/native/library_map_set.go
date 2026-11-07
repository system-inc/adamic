package native

import (
	"fmt"

	"github.com/system-inc/adamic/internal/ir"
)

func (nativeEmitter *emitter) collectionIterator(iterator ir.CollectionIterator) string {
	part := 1
	if iterator.Part == "values" {
		part = 2
	}
	if iterator.Part == "entries" {
		part = 3
	}
	return nativeEmitter.own(ir.Object, fmt.Sprintf(
		"adamic_collection_iterator(%s, %d, %d, %d, %t)",
		nativeEmitter.value(iterator.Collection),
		part,
		iterator.Key,
		iterator.Value,
		iterator.Set,
	))
}
