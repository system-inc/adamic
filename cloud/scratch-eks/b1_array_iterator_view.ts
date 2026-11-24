// A Map's iterator held where an ArrayIterator goes: the checker relates them structurally.
const map = new Map<number, number>([[1, 1]]);
const iterator: ArrayIterator<number> = map.keys();
console.log(iterator.toString());
