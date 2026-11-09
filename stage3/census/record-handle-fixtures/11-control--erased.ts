// Self-contained excerpts from microsoft/TypeScript tag v6.0.3.
// Only required Type/Symbol fields are retained. Dependency leaves are opaque:
// their contents do not affect rejection of a property-free function.
export {};
interface MapLike<T> { [index: string]: T; }
type Fn = () => string;
const fn: Fn = () => "called";
const k = "k";

type Holding = MapLike<Fn>;
const h: Holding = {};
const erased: {} = h;
const receiver: MapLike<Fn> = erased;
receiver[k] = fn;
const value = receiver.k;
if (value === undefined) throw new Error("missing k");
const read: Fn = value;
console.log(read());
