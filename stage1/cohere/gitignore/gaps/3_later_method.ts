// Gap 3, for methods: a call to a method declared later in the class is taken for a void one.
class Counter {
	count = 0;

	twice(): number {
		return this.once() + this.once();
	}

	once(): number {
		this.count = this.count + 1;
		return this.count;
	}
}

console.log(`${new Counter().twice()}`);
