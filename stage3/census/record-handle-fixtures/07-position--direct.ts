// Self-contained excerpts from microsoft/TypeScript tag v6.0.3.
// Only required Type/Symbol fields are retained. Dependency leaves are opaque:
// their contents do not affect rejection of a property-free function.
export {};
interface MapLike<T> { [index: string]: T; }
type Fn = (path: string) => string;
const fn: Fn = path => path;
const k = "k";

type Holding = { [index: string]: string | { value: string; pos: number; end: number } };
const h: Holding = {};
h[k] = fn;
const receiver = h;
const value = receiver.k;
if (value === undefined) throw new Error("missing k");
const read: Fn = value;
console.log(read("called"));
