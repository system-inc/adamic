// A ! inside a guarded chain is part of the chain, as tsc reads a?.b!.c (#cdpxjmb, Language's o01, o04
// and o08): where the receiver is missing the whole chain is undefined and the assertion never runs, as in
// Node, and where it's present the asserted step reads through. A field, a call and a deeper read, each
// with a missing receiver and then a present one.
interface Inner {
	readonly c: number;
	readonly d: { readonly e: string };
}
interface Outer {
	readonly b: Inner | undefined;
}
interface Box {
	readonly run: (() => number) | undefined;
}
function make(present: boolean): Outer | undefined {
	if (!present) {
		return undefined;
	}
	return { b: { c: 7, d: { e: 'deep' } } };
}
function box(present: boolean): Box | undefined {
	return present ? { run: (): number => 5 } : undefined;
}
console.log(String(make(false)?.b!.c));
console.log(String(make(true)?.b!.c));
console.log(String(box(false)?.run!()));
console.log(String(box(true)?.run!()));
console.log(String(make(false)?.b!.d.e));
console.log(String(make(true)?.b!.d.e));
