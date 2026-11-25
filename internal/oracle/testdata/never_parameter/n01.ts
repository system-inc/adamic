class Shape {
	describe(x: string): string {
		return "shape " + x;
	}
}
class Empty extends Shape {
	override describe(x: never): string {
		return "empty " + typeof x;
	}
}
const s: Shape = new Empty();
console.log(s.describe("hi"));
