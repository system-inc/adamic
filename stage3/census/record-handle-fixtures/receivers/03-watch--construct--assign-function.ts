// Self-contained excerpts from microsoft/TypeScript tag v6.0.3.
// Only required Type/Symbol fields are retained. Dependency leaves are opaque:
// their contents do not affect rejection of a property-free function.
export {};
interface MapLike<T> { [index: string]: T; }
declare const enum WatchDirectoryFlags {
    None = 0,
    Recursive = 1 << 0,
}


// The value whose static receiver type is being tested is the record itself.
type Receiver = new (kind: number, pos: number, end: number) => object;
type Holding = MapLike<WatchDirectoryFlags>;
const h: Holding = {};
const source: Receiver = class { constructor(kind: number, pos: number, end: number) {} };
const assigned = Object.assign(h, source);
const receiver: Receiver = assigned;
console.log("same record:", Object.is(receiver, h));
console.log("typeof receiver:", typeof receiver);
try {
 console.log("construct result:", new receiver(1, 2, 3));
} catch (error) {
 if (!(error instanceof TypeError)) throw error;
 console.log("construct threw: TypeError");
}
