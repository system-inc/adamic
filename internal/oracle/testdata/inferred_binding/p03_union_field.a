interface Tagged {
	kind: "x";
	n: number;
}
function addDefault<T extends { kind: string }>(to: { items: (T | { kind: string })[] }): void {
	to.items.push({ kind: "other" });
}
const holder: { items: Tagged[] } = { items: [{ kind: "x", n: 1 }] };
addDefault(holder);
const last = holder.items[1];
console.log(last === undefined ? "none" : `${last.kind} ${String(last.n)}`);
