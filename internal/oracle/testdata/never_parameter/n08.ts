class Shape {
	describe(x: string | number): string {
		return "shape " + String(x);
	}
}
class Empty extends Shape {
	override describe(x: never): string {
		return "empty";
	}
}
const s: Shape = new Empty();
console.log(s.describe(5));
