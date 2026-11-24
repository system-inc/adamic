interface Base {
	readonly name?: string;
	readonly kind: number;
}
interface Arrow extends Base {
	readonly name: never;
}
function makeArrow(): Arrow {
	const made: Base = { kind: 1 };
	return made as Arrow;
}
function nameOf(b: Base): string {
	return String(b.name);
}
console.log(nameOf(makeArrow()));
