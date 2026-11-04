package javascript

// Native iterator next is represented as a closure, so the JavaScript backend wraps Node's next
// with the same closure calling convention. Node's actual iterator still decides its behavior,
// and each [key, value] it gives is a tuple to a view, as a literal's is.
//
// The wrapper stands for Node's Map, Set or String Iterator, whose prototype isn't Object's, so it
// is an AdamicIterator, never a plain object: a run-time key read of an Object.prototype name stops
// on it with the runtime value stop (dynamic_keyed.go), as native's does on its iterator, whose
// shape is marked host (map_set.c), where a plain object would hand out Object as its constructor.
const iteratorClassRuntime = "class AdamicIterator { constructor(next) { this.next = next; } }\n"

const collectionIteratorRuntime = "const adamicCollectionIterator = (collection, part) => { const iterator = collection[part](); return new AdamicIterator(new AdamicClosure(() => { const step = iterator.next(); if (part === 'entries' && !step.done) adamicTuple(step.value); return step; }, [])); };\n"

const stringIteratorRuntime = "const adamicStringIterator = string => { const iterator = string[Symbol.iterator](); return new AdamicIterator(new AdamicClosure(() => iterator.next(), [])); };\n"
