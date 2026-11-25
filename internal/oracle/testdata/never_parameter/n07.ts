abstract class Visitor {
	abstract visit(x: string): string;
	run(v: string): string {
		return this.visit(v);
	}
}
class Dead extends Visitor {
	visit(x: never): string {
		return "dead " + typeof x;
	}
}
console.log(new Dead().run("hi"));
