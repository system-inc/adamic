class Shape {
	describe(x: unknown): string {
		return "shape";
	}
}
class Empty extends Shape {
	override describe(x: never): string {
		return "empty";
	}
}
const s: Shape = new Empty();
console.log(s.describe({ a: 1 }));
