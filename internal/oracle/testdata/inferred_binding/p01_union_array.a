interface Tagged {
	kind: "x";
	n: number;
}
function addDefault<T extends { kind: string }>(to: (T | { kind: string })[]): void {
	to.push({ kind: "other" });
}
const xs: Tagged[] = [{ kind: "x", n: 1 }];
addDefault(xs);
const last = xs[1];
console.log(last === undefined ? "none" : `${last.kind} ${String(last.n)}`);
