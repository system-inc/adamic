package ir

// A WeakMap or a WeakSet is a WeakCollection (runtime/weak_collection.c, adamic.h's adamic_weak_map):
// keyed by each key's weak handle, so an entry holds a count on the handle and none on the key, and
// when the key is freed every entry for it goes. A key is an object, an array, a Map or Set, a
// function or a typed array, never a value a primitive can stand for (lower/library_weak_collections.go).

// WeakCollectionNew is new WeakMap() or new WeakSet(), with no entries. Set says it's a WeakSet, which
// keeps no values; Key and Value are the representations its keys and values are held as, Value 0
// for a WeakSet.
type WeakCollectionNew struct {
	Key, Value Type
	Set        bool
}

func (WeakCollectionNew) Type() Type { return WeakCollection }

// WeakSetAdd is set.add(Key), which is the set. The set holds the key weakly: an entry is no strong
// write of the key, so it has no write site.
type WeakSetAdd struct {
	Set, Key Expression
	KeyType  Type
}

func (WeakSetAdd) Type() Type { return WeakCollection }

// WeakCollectionHas is has(Key) on a WeakMap or a WeakSet, and WeakCollectionDelete delete(Key), each
// false for a key the collection holds no entry for, as one it never saw is.
type WeakCollectionHas struct {
	Collection, Key Expression
	KeyType         Type
}

func (WeakCollectionHas) Type() Type { return Boolean }

type WeakCollectionDelete struct {
	Collection, Key Expression
	KeyType         Type
}

func (WeakCollectionDelete) Type() Type { return Boolean }

// WeakMapGet is map.get(Key): the value stored under the key, or undefined for a key the map holds no
// entry for (a null string, or a Maybe pair), as Map's get is. Its values are numbers, booleans or
// strings so far (lower/library_weak_collections.go's weakCollectionAdmitted); a string comes back as the
// map's own, borrowed, and the reader retains it.
type WeakMapGet struct {
	Map, Key  Expression
	KeyType   Type
	ValueType Type
}

func (g WeakMapGet) Type() Type { return Maybe(g.ValueType) }

// WeakMapSet is map.set(Key, Value), which is the map. The entry holds the key's weak handle, never the
// key, and holds the value: Site is the value's write, for the cycle finder, and the key is no strong
// write at all.
type WeakMapSet struct {
	Map, Key, Value Expression
	KeyType         Type
	ValueType       Type
	// Site is which write of the program this is, for the cycle finder (lowering keeps the type of what it
	// writes into), or 0 when nothing recorded one.
	Site int
}

func (WeakMapSet) Type() Type { return WeakCollection }
