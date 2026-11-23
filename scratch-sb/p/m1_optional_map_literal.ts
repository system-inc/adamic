const real = new Map<string, number>([["a", 1]]);
const fake = {
    extra: "not a map",
    size: 1,
    get: (_key: string): number | undefined => 42,
    has: (key: string): boolean => real.has(key),
    forEach: (callback: (value: number, key: string, map: ReadonlyMap<string, number>) => void): void => real.forEach(callback),
    entries: () => real.entries(),
    keys: () => real.keys(),
    values: () => real.values(),
    [Symbol.iterator]: () => real.entries(),
};
function read(map: ReadonlyMap<string, number> | undefined): string {
    return String(map?.get("a"));
}
console.log(read(fake));
