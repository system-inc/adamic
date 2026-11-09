// Self-contained excerpts from microsoft/TypeScript tag v6.0.3.
// Only required Type/Symbol fields are retained. Dependency leaves are opaque:
// their contents do not affect rejection of a property-free function.
export {};
interface MapLike<T> { [index: string]: T; }
type Fn = () => string;
const fn: Fn = () => "called";
const k = "k";
function extend<T>(first: MapLike<T | undefined>, second: MapLike<T | undefined>): MapLike<T | undefined> {
 const result: MapLike<T | undefined> = {};
 for (const id in second) result[id] = second[id];
 for (const id in first) result[id] = first[id];
 return result;
}
const h = extend<() => string>({ k: fn }, {});
const value = h.k;
if (value === undefined) throw new Error("missing k");
const read: Fn = value;
console.log(read());
