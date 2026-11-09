// Logging-only adaptation for Adamic's console.log(message: string).
// Self-contained excerpts from microsoft/TypeScript tag v6.0.3.
// Only required Type/Symbol fields are retained. Dependency leaves are opaque:
// their contents do not affect rejection of a property-free function.
export {};
interface MapLike<T> { [index: string]: T; }


// The value whose static receiver type is being tested is the record itself.
type Receiver = (path: string) => string;
type Holding = { [index: string]: string | { value: string; pos: number; end: number } };
const h: Holding = {};
const source: Receiver = path => path;
const assigned = Object.assign(h, source);
const receiver: Receiver = assigned;
console.log(Object.is(receiver, h) ? "same record: true" : "same record: false");
console.log("typeof receiver: " + typeof receiver);
try {
 receiver("called");
 console.log("call returned");
} catch (error) {
 if (!(error instanceof TypeError)) throw error;
 console.log("call threw: TypeError");
}
