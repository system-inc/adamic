// #rnbvya6: the never read through a parameter typed T itself, not a field, returned bare.
function read<T extends string | undefined>(x: T): string {
	if (x === undefined) {
		return 'u';
	}
	return x;
}

console.log(read<undefined>(undefined));
console.log(read<string | undefined>(undefined));
console.log(read<string | undefined>('p'));
console.log(read<string>('q'));
