const own: MapIterator<number> = {
	next(): IteratorReturnResult<undefined> {
		return { done: true, value: undefined };
	},
	[Symbol.iterator](): MapIterator<number> {
		return own;
	},
};
console.log(own.toString());
console.log(Object.prototype.toString.call(own));
