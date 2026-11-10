package ir

// CollectionIterator is a live, resumable Map or Set iterator. The object holds its next function,
// whose state keeps the collection alive until the iterator is released.
type CollectionIterator struct {
	Collection Expression
	Part       string
	Key, Value Type
	Set        bool
	// PairGraphTypes is an entries() pair's own tuple type, which plans each pair next() makes: a
	// pair can reach a graph slot from a collection no graph slot holds (#dps0kp6).
	PairGraphTypes []int
}

func (CollectionIterator) Type() Type { return Object }

// StringIterator retains its string and advances by JavaScript code points.
type StringIterator struct{ Value Expression }

func (StringIterator) Type() Type { return Object }
