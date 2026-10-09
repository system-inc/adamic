package ir

// CollectionIterator is a live, resumable built-in iterator. The object holds private counted state and inherits shared protocol methods,
// whose state keeps the collection alive until the iterator is released.
type CollectionIterator struct {
	Collection Expression
	Part       string
	Key, Value Type
	Set        bool
}

func (CollectionIterator) Type() Type { return Object }
