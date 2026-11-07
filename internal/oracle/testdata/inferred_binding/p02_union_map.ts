interface Tagged {
	kind: "x";
	n: number;
}
function addDefault<T extends { kind: string }>(to: Map<string, T | { kind: string }>): void {
	to.set("b", { kind: "other" });
}
const xs: Map<string, Tagged> = new Map([["a", { kind: "x", n: 1 }]]);
addDefault(xs);
const last = xs.get("b");
console.log(last === undefined ? "none" : `${last.kind} ${String(last.n)}`);
