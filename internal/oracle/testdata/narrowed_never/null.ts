// #rnbvya6: the same never read where T holds only null, T extends string | null narrowed past === null.
interface Box<T> {
	f: T;
}

function read<T extends string | null>(o: Box<T>): string {
	const v = o.f;
	if (v === null) {
		return 'n';
	}
	return v;
}

console.log(read<null>({ f: null }));
console.log(read<string | null>({ f: null }));
console.log(read<string | null>({ f: 'x' }));
console.log(read<string>({ f: 'y' }));
