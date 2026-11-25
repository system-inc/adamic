interface Visit {
	visit(x: string | number): string;
}
class Dead implements Visit {
	visit(x: never): string {
		return "dead";
	}
}
const v: Visit = new Dead();
console.log(v.visit("hi"));
