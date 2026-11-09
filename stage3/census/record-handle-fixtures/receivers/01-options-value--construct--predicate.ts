// Self-contained excerpts from microsoft/TypeScript tag v6.0.3.
// Only required Type/Symbol fields are retained. Dependency leaves are opaque:
// their contents do not affect rejection of a property-free function.
export {};
interface MapLike<T> { [index: string]: T; }
interface PluginImport {
    name: string;
}
interface ProjectReference {
    /** A normalized path on disk */
    path: string;
    /** The path as the user originally wrote it */
    originalPath?: string;
    /** @deprecated */
    prepend?: boolean;
    /** True if it is intended that this reference form a circularity */
    circular?: boolean;
}
type CompilerOptionsValue = string | number | boolean | (string | number)[] | string[] | MapLike<string[]> | PluginImport[] | ProjectReference[] | null | undefined; // eslint-disable-line no-restricted-syntax


// The value whose static receiver type is being tested is the record itself.
type Receiver = new (kind: number, pos: number, end: number) => object;
type Holding = Record<string, CompilerOptionsValue>;
const h: Holding = {};
function isReceiver(x: unknown): x is Receiver { return typeof x === "function"; }
const erased: unknown = h;
if (isReceiver(erased)) {
 const receiver: Receiver = erased;
 console.log("true branch");
 console.log(new receiver(1, 2, 3));
} else {
 console.log("record is not a function; false branch");
}
