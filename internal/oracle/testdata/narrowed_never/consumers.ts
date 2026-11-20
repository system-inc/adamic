// #rnbvya6: the narrowed value taken by other consumers than a bare return, each at T given undefined,
// where it's never, beside T given string | undefined and string, where it's a string.
interface Box<T> {
	f: T;
}

function length<T extends string | undefined>(o: Box<T>): number {
	const v = o.f;
	if (v === undefined) {
		return -1;
	}
	return v.length;
}

function joined<T extends string | undefined>(o: Box<T>): string {
	const v = o.f;
	if (v === undefined) {
		return 'none';
	}
	return v + '!';
}

function upper<T extends string | undefined>(o: Box<T>): string {
	const v = o.f;
	if (v === undefined) {
		return 'none';
	}
	return v.toUpperCase();
}

function held<T extends string | undefined>(o: Box<T>): string {
	const v = o.f;
	if (v === undefined) {
		return 'none';
	}
	const s: string = v;
	return s;
}

function logged<T extends string | undefined>(o: Box<T>): void {
	const v = o.f;
	if (v === undefined) {
		console.log('none');
		return;
	}
	console.log(v);
}

console.log(length<undefined>({ f: undefined }), length<string | undefined>({ f: 'abc' }), length<string>({ f: 'ab' }));
console.log(joined<undefined>({ f: undefined }), joined<string | undefined>({ f: 'a' }), joined<string>({ f: 'b' }));
console.log(upper<undefined>({ f: undefined }), upper<string | undefined>({ f: 'a' }), upper<string>({ f: 'b' }));
console.log(held<undefined>({ f: undefined }), held<string | undefined>({ f: 'a' }), held<string>({ f: 'b' }));
logged<undefined>({ f: undefined });
logged<string | undefined>({ f: 'a' });
logged<string>({ f: 'b' });
