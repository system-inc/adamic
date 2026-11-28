const modes = new Map<string, [specifier: string, mode: number | undefined]>();
modes.set("a", ["b", undefined]);
modes.set("c", ["d", 99]);
for (const key of ["a", "c"]) {
	const [specifier, mode] = modes.get(key)!;
	console.log(`${specifier} ${String(mode)}`);
}
const labels = new Map<string, [string | undefined, string]>();
labels.set("x", [undefined, "y"]);
const [label, text] = (labels.get("x"))!;
console.log(`${String(label)} ${text}`);
