interface Options { readonly overrideImportMode?: number; readonly other?: string; }
class Box<T extends Options> {
	constructor(readonly held: T) {}
	mode(): string { return String(this.held.overrideImportMode); }
}
const wide = { other: "o", overrideImportMode: "lie" };
const narrow: { readonly other: string } = wide;
console.log(new Box(narrow).mode());
