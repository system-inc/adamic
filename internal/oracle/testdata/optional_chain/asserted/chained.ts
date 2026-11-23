// A ! inside a guarded chain keeps its own check (#cdpxjmb, Lower's must): outer?.b!.c is undefined where
// outer is, and where outer is there but b isn't, the assertion fails, where Node throws reading c of
// undefined.
interface Inner {
	readonly c: string;
}
interface Outer {
	readonly b?: Inner | undefined;
}
function chained(outer: Outer | undefined): string {
	return String(outer?.b!.c);
}
console.log(chained(undefined));
console.log(chained({ b: { c: 'here' } }));
console.log(chained({}));
