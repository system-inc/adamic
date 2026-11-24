class Steps implements MapIterator<number> {
	next(): IteratorReturnResult<undefined> {
		return { done: true, value: undefined };
	}
	[Symbol.iterator](): MapIterator<number> {
		return this;
	}
}
const own: MapIterator<number> = new Steps();
console.log(own.toString());
console.log(Object.prototype.toString.call(own));
