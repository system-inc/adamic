interface Item { readonly name: string; }
interface Reader { readonly extra: string; }
class Sneaky {
    readonly items: Item[];
    count: number;
    constructor(items: Item[]) { this.items = items; this.count = 0; }
    get extra(): string { this.count = this.count + 1; this.items[0] = { name: "replaced" + String(this.count) }; return "-" + String(this.count); }
}
function show(items: Item[], reader: Reader): void { for (const item of items) { const extra = reader.extra; console.log(item.name + extra); } }
const items: Item[] = [{ name: "first" + String(items0()) }, { name: "second" + String(items0()) }];
function items0(): number { return 7; }
show(items, new Sneaky(items));
