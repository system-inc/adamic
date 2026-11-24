// Ruling 2's direct read (#bbssdcy): the real object behind arrow, made as a Base by a literal, has no
// name slot, and the read the checker types never is a :76 read, a loud stop at exit 70 where Node
// prints undefined and goes on.
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
const arrow = makeArrow();
console.log("kind " + String(arrow.kind));
console.log(String(arrow.name));
console.log("after");
