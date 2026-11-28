function load(): any { return ["a", 1]; }
const [first, second] = load()!;
console.log(`${String(first)} ${String(second)}`);
