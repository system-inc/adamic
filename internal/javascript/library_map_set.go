package javascript

// Native iterator next is represented as a closure, so the JavaScript backend wraps Node's next
// with the same closure calling convention. Node's actual iterator still decides its behavior.
const collectionIteratorRuntime = `const adamicIteratorPrototype = {
 __adamic_symbol_iterator(self) { return self; },
 [Symbol.iterator]() { return this; }
};
const adamicIteratorFamilies = Array.from({ length: 4 }, () => Object.create(adamicIteratorPrototype, {
 next: { value: function(self = this) { const state = self.__adamic_iterator_state; return state.code(state, []); } }
}));
const adamicCollectionIterator = (collection, part) => {
 const iterator = part === 'iterator' ? collection[Symbol.iterator]() : collection[part]();
 const family = typeof collection === 'string' ? 3 : Array.isArray(collection) ? 2 : collection instanceof Set ? 1 : 0;
 const object = Object.create(adamicIteratorFamilies[family]);
 Object.defineProperty(object, '__adamic_iterator_state', { value: new AdamicClosure(() => iterator.next(), []) });
 return object;
};
`
