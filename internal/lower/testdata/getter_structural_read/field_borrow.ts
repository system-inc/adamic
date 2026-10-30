interface Item { readonly name: string; }
interface Reader { readonly extra: string; }
class Owner {
    current: Item;
    count: number;
    constructor(current: Item) { this.current = current; this.count = 0; }
    get extra(): string { this.count = this.count + 1; this.current = { name: "next" + String(this.count) }; return "-" + String(this.count); }
}
function show(owner: Owner, reader: Reader): void { const item = owner.current; const extra = reader.extra; console.log(item.name + extra); }
const owner = new Owner({ name: "start" + String(3) });
show(owner, owner);
show(owner, owner);
