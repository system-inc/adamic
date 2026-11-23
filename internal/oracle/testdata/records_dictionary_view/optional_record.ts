// A Record<string, string | undefined> holding an undefined value, handed on as object and seen as MapLike<unknown>:
// undefined sits in a reference slot as the record's own absent value, the one value whose read through the view
// could lie, and reads as undefined, as Node reads it (Language's pin on #2v3xtgm).
interface MapLike<T> {
    [index: string]: T;
}
const hasOwnProperty = Object.prototype.hasOwnProperty;

function show(exports: object): void {
    for (const key in exports as MapLike<unknown>) {
        if (hasOwnProperty.call(exports as MapLike<unknown>, key)) {
            const value = (exports as MapLike<unknown>)[key];
            if (typeof value === "string") {
                console.log(key + "=string " + value);
            }
            else {
                console.log(key + "=" + typeof value);
            }
        }
    }
}

const optional: Record<string, string | undefined> = {};
optional["a"] = "x";
optional["b"] = undefined;
show(optional);
console.log("done");
