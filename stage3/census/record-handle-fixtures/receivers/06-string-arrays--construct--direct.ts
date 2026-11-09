// Self-contained excerpts from microsoft/TypeScript tag v6.0.3.
// Only required Type/Symbol fields are retained. Dependency leaves are opaque:
// their contents do not affect rejection of a property-free function.
export {};
interface MapLike<T> { [index: string]: T; }


// The value whose static receiver type is being tested is the record itself.
type Receiver = new (kind: number, pos: number, end: number) => object;
type Holding = { [value: string]: string[] };
const h: Holding = {};
const receiver: Receiver = h;
console.log("same record:", Object.is(receiver, h));
console.log("typeof receiver:", typeof receiver);
try {
 console.log("construct result:", new receiver(1, 2, 3));
} catch (error) {
 if (!(error instanceof TypeError)) throw error;
 console.log("construct threw: TypeError");
}
