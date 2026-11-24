// A Record<string, string> handed on as object and seen as MapLike<unknown>: a record of references, so every value
// it holds reads as a Union as it is, a raw string pointer included, and agrees with Node.
interface MapLike<T> {
    [index: string]: T;
}
const hasOwnProperty = Object.prototype.hasOwnProperty;

function getOwnKeys<T>(map: MapLike<T>): string[] {
    const keys: string[] = [];
    for (const key in map) {
        if (hasOwnProperty.call(map, key)) {
            keys.push(key);
        }
    }
    return keys;
}
function walk(exports: object): void {
    for (const key of getOwnKeys(exports as MapLike<unknown>)) {
        const value = (exports as MapLike<unknown>)[key];
        if (typeof value === "string") {
            console.log(key + "=" + value);
        }
    }
}

const names: Record<string, string> = {};
names["main"] = "./index.js";
names["types"] = "./index.d.ts";
walk(names);
console.log("done");
