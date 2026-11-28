const pairs = new Map<string, [string, number]>();
pairs.set("k", ["a", 1]);
const [specifier, count] = pairs.get("k")!;
console.log(`${specifier} ${String(count)}`);
