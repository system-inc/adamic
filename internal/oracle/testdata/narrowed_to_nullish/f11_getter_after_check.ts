let n = 0;
class B {
	get x(): number | undefined {
		return n > 0 ? 4 : undefined;
	}
}
const b = new B();
function bump(): void {
	n = 1;
}
if (b.x === undefined) {
	bump();
	console.log(String(b.x));
}
