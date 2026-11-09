// Self-contained excerpts from microsoft/TypeScript tag v6.0.3.
// Only required Type/Symbol fields are retained. Dependency leaves are opaque:
// their contents do not affect rejection of a property-free function.
export {};
interface MapLike<T> { [index: string]: T; }

// The value whose static receiver type is being tested is the record itself.
type Receiver = (path: string) => string;
type Holding = MapLike<(() => string) | undefined>;
function extend<T>(first: MapLike<T | undefined>, second: MapLike<T | undefined>): MapLike<T | undefined> {
 const result: MapLike<T | undefined> = {};
 for (const id in second) result[id] = second[id];
 for (const id in first) result[id] = first[id];
 return result;
}
const h: Holding = extend<() => string>({}, {});
function isReceiver(x: unknown): x is Receiver { return typeof x === "function"; }
const erased: unknown = h;
if (isReceiver(erased)) {
 const receiver: Receiver = erased;
 console.log("true branch");
 console.log(receiver("called"));
} else {
 console.log("record is not a function; false branch");
}
