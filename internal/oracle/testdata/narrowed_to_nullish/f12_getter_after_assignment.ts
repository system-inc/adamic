class B {
	get x(): number | undefined {
		return 4;
	}
	set x(value: number | undefined) {}
}
const b = new B();
b.x = undefined;
console.log(String(b.x));
