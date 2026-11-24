// A program's own object held where a MapIterator goes.
const own: MapIterator<number> = {
	next() {
		return { done: true, value: undefined };
	},
	[Symbol.iterator]() {
		return own;
	},
};
console.log(own.toString());
