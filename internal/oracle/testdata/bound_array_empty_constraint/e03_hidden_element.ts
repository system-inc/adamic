interface Options {
	readonly overrideImportMode?: number;
	readonly other?: string;
}
function read<T extends { readonly 0: Options }>(holder: T): string {
	return String(holder[0].overrideImportMode);
}
const wide = { other: "o", overrideImportMode: "lie" };
const narrow: { readonly other: string } = wide;
console.log(read([narrow]));
