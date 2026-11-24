function* numbers(): Generator<number> {
	yield 1;
}
console.log(Object.prototype.toString.call(numbers()));
