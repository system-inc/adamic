abstract class Visitor {
	abstract visit(x: string | number): string;
	run(v: string | number): string {
		return this.visit(v);
	}
}
class Dead extends Visitor {
	visit(x: never): string {
		return "dead";
	}
}
console.log(new Dead().run(3));
