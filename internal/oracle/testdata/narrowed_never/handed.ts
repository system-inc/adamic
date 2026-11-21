// #rnbvya6, Lower's should: the narrowed value handed to a string parameter, at T given undefined, where
// it's never, beside T given string | undefined and string. Her second row, a for-of over v, can't be
// written: tsc refuses `for (const letter of v)` with TS2488, since past === undefined v is T & {}, which has
// no [Symbol.iterator]() (J-r5vf3kf), so no program brings this read to iterable_views' for-of.
interface Box<T> {
	f: T;
}

function shout(s: string): string {
	return s.toUpperCase();
}

function handed<T extends string | undefined>(o: Box<T>): string {
	const v = o.f;
	if (v === undefined) {
		return 'none';
	}
	return shout(v);
}

console.log(handed<undefined>({ f: undefined }));
console.log(handed<string | undefined>({ f: 'a' }));
console.log(handed<string>({ f: 'b' }));
