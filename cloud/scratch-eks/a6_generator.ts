function* numbers(): Generator<number> {
	yield 1;
}
console.log(numbers().toString());
