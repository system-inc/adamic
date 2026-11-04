interface Options { readonly overrideImportMode?: number; readonly other?: string; }
type Narrow = { readonly other: string };
class Box<T extends Options> {
	constructor(readonly held: T) {}
	mode(): string { return String(this.held.overrideImportMode); }
}
class NarrowBox extends Box<Narrow> {}
const wide = { other: "o", overrideImportMode: "lie" };
const narrow: Narrow = wide;
console.log(new NarrowBox(narrow).mode());
