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
type Receiver = () => string | undefined;
type Holding = MapLike<WatchDirectoryFlags>;
const h: Holding = {};
const erased: {} = h;
const receiver: Receiver = erased;
console.log("same record:", Object.is(receiver, h));
console.log("typeof receiver:", typeof receiver);
try {
 console.log("call result:", receiver());
} catch (error) {
 if (!(error instanceof TypeError)) throw error;
 console.log("call threw: TypeError");
}
