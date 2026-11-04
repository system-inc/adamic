// A reference left of && is falsy only when it's missing, so the result is the right where the left
// is present and undefined where it isn't, held as the result's own representation: a string, an
// array, a Map or an object, each read through an optional field (#v5k6t3k, tsc's
// `holder.item && holder.item.name`). Every value printed is built, so a lost release shows.
interface Item {
	readonly name: string;
	readonly items: number[];
	readonly table: Map<string, number>;
	readonly next?: Item;
}
interface Holder {
	readonly item?: Item;
	readonly list?: number[];
	readonly map?: Map<string, number>;
}
function make(label: string): Item {
	const table = new Map<string, number>();
	table.set(label + "!", label.length);
	return { name: label + "-" + String(label.length), items: [label.length, 2], table: table };
}
function nameOf(holder: Holder): string | undefined {
	return holder.item && holder.item.name;
}
function itemsOf(holder: Holder): number[] | undefined {
	return holder.item && holder.item.items;
}
function tableOf(holder: Holder): Map<string, number> | undefined {
	return holder.item && holder.item.table;
}
function nextOf(holder: Holder): Item | undefined {
	return holder.item && holder.item.next;
}
function firstOf(holder: Holder): Map<string, number> | undefined {
	return holder.list && holder.map;
}
function show(label: string, holder: Holder): void {
	const name = nameOf(holder);
	const items = itemsOf(holder);
	const table = tableOf(holder);
	const next = nextOf(holder);
	const first = firstOf(holder);
	console.log(label + " " + (name ?? "none") + " " + (items === undefined ? "none" : items.join("+")) + " " + (table === undefined ? "none" : String(table.size)) + " " + (next === undefined ? "none" : next.name) + " " + (first === undefined ? "none" : String(first.size)));
}
const inner = make("in");
const outer: Item = { name: "out-" + String(3), items: [7], table: new Map<string, number>(), next: inner };
show("full", { item: outer, list: [1], map: inner.table });
show("bare", { item: inner });
show("empty", {});
