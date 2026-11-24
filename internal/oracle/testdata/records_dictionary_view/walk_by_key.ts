// tsc's tryGetModuleNameFromExportsOrImports (moduleSpecifiers.ts:1109-1134): a package.json's exports, a value typed
// object, cast to MapLike<unknown> and walked by getOwnKeys (core.ts), a for...in and an own-key test, then read by
// its key, each value told apart by its typeof as tsc's recursion tells it. The record is built by key, as a JSON
// object is, of values held as a Union: a string, an object, an array and undefined, the kinds an exports map holds.
interface MapLike<T> {
    [index: string]: T;
}
type ExportsValue = string | string[] | { readonly path: string } | undefined;
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
function describe(value: unknown): string {
    if (typeof value === "string") {
        return "string " + value;
    }
    if (value === undefined) {
        return "undefined";
    }
    return typeof value;
}
function walk(exports: object): void {
    for (const key of getOwnKeys(exports as MapLike<unknown>)) {
        const subTarget = (exports as MapLike<unknown>)[key];
        console.log(key + "=" + describe(subTarget));
    }
}

const built: Record<string, ExportsValue> = {};
built["import"] = "./index.mjs";
built["types"] = { path: "./index.d.ts" };
built["files"] = ["a.js", "b.js"];
built["missing"] = undefined;
walk(built);
console.log("done");
