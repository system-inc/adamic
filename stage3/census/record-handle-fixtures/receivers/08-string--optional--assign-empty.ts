// Self-contained excerpts from microsoft/TypeScript tag v6.0.3.
// Only required Type/Symbol fields are retained. Dependency leaves are opaque:
// their contents do not affect rejection of a property-free function.
export {};
interface MapLike<T> { [index: string]: T; }


// The value whose static receiver type is being tested is the record itself.
type Receiver = () => string | undefined;
type Holding = { [index: string]: string | undefined };
const h: Holding = {};
const assigned = Object.assign(h, {});
const receiver: Receiver = assigned;
console.log("same record:", Object.is(receiver, h));
console.log("typeof receiver:", typeof receiver);
try {
 console.log("call result:", receiver());
} catch (error) {
 if (!(error instanceof TypeError)) throw error;
 console.log("call threw: TypeError");
}
