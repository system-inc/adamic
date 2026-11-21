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

console.log(String(length<undefined>({ f: undefined })));
console.log(String(length<string | undefined>({ f: 'abc' })));
console.log(String(length<string>({ f: 'ab' })));
// joined<undefined> is left out: main stops it at lowering, "a BinaryExpression with a value and a string",
// before any C (J-mc644tv, at main and at #rnbvya6's head alike), an older NotYet this fix doesn't reach.
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
