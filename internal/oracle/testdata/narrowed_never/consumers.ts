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

console.log(length<undefined>({ f: undefined }));
console.log(length<string | undefined>({ f: 'abc' }));
console.log(length<string>({ f: 'ab' }));
console.log(joined<undefined>({ f: undefined }));
console.log(joined<string | undefined>({ f: 'a' }));
console.log(joined<string>({ f: 'b' }));
console.log(upper<undefined>({ f: undefined }));
console.log(upper<string | undefined>({ f: 'a' }));
console.log(upper<string>({ f: 'b' }));
console.log(held<undefined>({ f: undefined }));
console.log(held<string | undefined>({ f: 'a' }));
console.log(held<string>({ f: 'b' }));
logged<undefined>({ f: undefined });
logged<string | undefined>({ f: 'a' });
logged<string>({ f: 'b' });
