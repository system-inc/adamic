// #rnbvya6: a field typed T, T extends string | undefined, narrowed past === undefined and returned bare.
// Given undefined, T holds nothing else, so the return after the test is never, yet the checker narrows T
// by its constraint and calls it a string: the function returns a string, and native C once returned the
// missing object the field is held as, which clang refused.
interface Box<T> {
	f: T;
}

function read<T extends string | undefined>(o: Box<T>): string {
	const v = o.f;
	if (v === undefined) {
		return 'u';
	}
	return v;
}

console.log(read<string | undefined>({ f: undefined }));
console.log(read<undefined>({ f: undefined }));
console.log(read<string>({ f: 's' }));
console.log(read<string | undefined>({ f: 'cd' }));
