// Gap 4: a #private method is refused, as "a method with a computed name"; a `private` method and a
// #private field lower.
class Counter {
	#count = 0;
	#step(): void {
		this.#count++;
	}
	twice(): number {
		this.#step();
		this.#step();
		return this.#count;
	}
}
console.log(`${new Counter().twice()}`);
