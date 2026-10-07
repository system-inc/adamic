package javascript

// Native iterator next is represented as a closure, so the JavaScript backend wraps Node's next
// with the same closure calling convention. Node's actual iterator still decides its behavior.
const collectionIteratorRuntime = "const adamicCollectionIterator = (collection, part) => { const iterator = collection[part](); return { next: new AdamicClosure(() => iterator.next(), []) }; };\n"
