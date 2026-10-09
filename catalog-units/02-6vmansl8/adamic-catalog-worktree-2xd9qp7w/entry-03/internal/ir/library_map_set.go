package ir

// CollectionIterator is a live, resumable Map or Set iterator. The object holds its next function,
// whose state keeps the collection alive until the iterator is released.
type CollectionIterator struct {
	Collection Expression
	Part       string
	Key, Value Type
	Set        bool
}

func (CollectionIterator) Type() Type { return Object }
