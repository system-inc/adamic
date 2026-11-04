interface Options { readonly overrideImportMode?: number; readonly other?: string; }
type Narrow = { readonly other: string };
class Box<T extends Options = Narrow> {
	readonly held: T[] = [];
	mode(): string {
		for (const held of this.held) return String(held.overrideImportMode);
		return "none";
	}
}
const wide = { other: "o", overrideImportMode: "lie" };
const narrow: Narrow = wide;
const box = new Box();
box.held.push(narrow);
console.log(box.mode());
