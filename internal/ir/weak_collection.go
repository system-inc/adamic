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
