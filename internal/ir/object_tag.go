package ir

// ObjectTag is Object.prototype.toString of Value, an object, a weak collection or a Promise: "[object <Tag>]" for a
// library object with a tag of its own, one of LibraryObjectTags, and "[object Object]" for anything else. It's read
// from the value itself, never its static type, since a type proves no tag (#eksbz4c, head T): a value typed as an
// ArrayIterator can hold a Map's iterator, and a program's own object can carry MapIterator's type. Natively it's
// adamic_object_to_string, which knows an iterator by its shape's address and a weak collection or a Promise by its
// kind; on the JavaScript backend it's Object.prototype.toString.call of the value, which reads Node's own objects
// and the iterator wrappers' [Symbol.toStringTag] (library_map_set.go). The string is immortal natively.
type ObjectTag struct{ Value Expression }

func (ObjectTag) Type() Type { return String }

// LibraryObjectTags is each library object that answers a tag of its own and that a program can hold today, by the
// library interface that declares it and its [Symbol.toStringTag], as Node answers it (J-e35t6n9). It's the one Go
// spelling of the tags: lowering asks it which receivers read their tag at run time (prototype.go), the JavaScript
// backend's iterator wrappers are written from it (library_map_set.go), and adamic_object_tag, the one native spelling,
// is held to it by TestRuntimeObjectTagsAreIRs. An array iterator, a generator, an Intl object and the rest aren't
// values a program holds yet, so they aren't here, and a toString on one stops at compile time.
var LibraryObjectTags = []struct{ Interface, Tag string }{
	{"MapIterator", "Map Iterator"},
	{"SetIterator", "Set Iterator"},
	{"StringIterator", "String Iterator"},
	{"RegExpStringIterator", "RegExp String Iterator"},
	{"WeakMap", "WeakMap"},
	{"WeakSet", "WeakSet"},
	{"Promise", "Promise"},
}

// LibraryObjectTag is the tag LibraryObjectTags gives the library interface named, or empty for any other.
func LibraryObjectTag(name string) string {
	for _, row := range LibraryObjectTags {
		if row.Interface == name {
			return row.Tag
		}
	}
	return ""
}
