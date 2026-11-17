// A class keyed by Symbol.iterator beside reads of unknown values (#ppgk12h): in on an unknown and a read of the
// narrowed unknown each walk every class's members by name, and a computed name spells none, so the walk passes over
// it rather than asking its node for text, which panicked.
class Counter {
    [Symbol.iterator]() {
        return [1, 2][Symbol.iterator]();
    }
}

function show(u: unknown): string {
    if (typeof u === "object" && u !== null && "a" in u) {
        return String(u.a);
    }
    return "none";
}

interface Holder {
    value: unknown;
}

function read(holder: Holder): string {
    const value = holder.value;
    if (typeof value === "object" && value !== null && "a" in value) {
        return String(value.a);
    }
    return "none";
}

console.log(show({ a: 1 }));
console.log(show(3));
console.log(read({ value: { a: "two" } }));
console.log(read({ value: null }));
console.log(String(new Counter() !== undefined));
