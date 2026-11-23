// Parentheses end an optional chain (#cdpxjmb, Lower's must): (outer?.b.c)! asserts what the chain gives,
// outside its guard, so where outer is missing the assertion fails, where Node throws reading length of
// undefined. Taking the parentheses into the guard would give undefined, a silent wrong answer.
interface Inner {
	readonly c: string;
}
interface Outer {
	readonly b: Inner;
}
function grouped(outer: Outer | undefined): string {
	return String((outer?.b.c)!.length);
}
console.log(grouped({ b: { c: 'here' } }));
console.log(grouped(undefined));
