// #rnbvya6, Lower's should: the narrowed value iterated and handed to a string parameter, at T given
// undefined, where it's never, beside T given string | undefined and string.
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

function letters<T extends string | undefined>(o: Box<T>): string {
	const v = o.f;
	if (v === undefined) {
		return 'none';
	}
	let out = '';
	for (const letter of v) {
		out = out + letter + '.';
	}
	return out;
}

console.log(handed<undefined>({ f: undefined }));
console.log(handed<string | undefined>({ f: 'a' }));
console.log(handed<string>({ f: 'b' }));
console.log(letters<undefined>({ f: undefined }));
console.log(letters<string | undefined>({ f: 'ab' }));
console.log(letters<string>({ f: 'cd' }));
