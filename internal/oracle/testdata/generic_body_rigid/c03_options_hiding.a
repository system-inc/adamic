interface Options {
	readonly overrideImportMode?: number;
	readonly other?: string;
}
function readFirst<T extends Options>(items: T[]): string {
	for (const item of items) return String(item.overrideImportMode);
	return "none";
}
const wide = { other: "o", overrideImportMode: "lie" };
const narrow: { readonly other: string } = wide;
const items: { readonly other: string }[] = [narrow];
console.log(readFirst(items));
