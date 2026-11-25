enum Kind { A, B, C }
const names = {
	[Kind.A]: "a",
	[Kind.B]: "b",
	[7]: "seven",
};
function nameOf(kind: Kind.A | Kind.B): string {
	return names[kind];
}
console.log(nameOf(Kind.A) + " " + nameOf(Kind.B) + " " + names[7]);
