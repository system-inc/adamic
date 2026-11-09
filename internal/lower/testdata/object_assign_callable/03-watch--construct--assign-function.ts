// Logging-only adaptation for Adamic's console.log(message: string).
// Self-contained excerpts from microsoft/TypeScript tag v6.0.3.
// Only required Type/Symbol fields are retained. Dependency leaves are opaque:
// their contents do not affect rejection of a property-free function.
export {};
interface MapLike<T> { [index: string]: T; }
// Cleaned for #k8aak04: the ambient enum is used only as a type here, and an ambient enum is refused
// ahead of Object.assign, so it stands as its value type.
type WatchDirectoryFlags = number;


// The value whose static receiver type is being tested is the record itself.
type Receiver = new (kind: number, pos: number, end: number) => object;
type Holding = MapLike<WatchDirectoryFlags>;
const h: Holding = {};
const source: Receiver = class { constructor(kind: number, pos: number, end: number) {} };
const assigned = Object.assign(h, source);
const receiver: Receiver = assigned;
console.log(Object.is(receiver, h) ? "same record: true" : "same record: false");
console.log("typeof receiver: " + typeof receiver);
try {
 new receiver(1, 2, 3);
 console.log("construct returned");
} catch (error) {
 if (!(error instanceof TypeError)) throw error;
 console.log("construct threw: TypeError");
}
