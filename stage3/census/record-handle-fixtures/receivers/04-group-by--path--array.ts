// Self-contained excerpts from microsoft/TypeScript tag v6.0.3.
// Only required Type/Symbol fields are retained. Dependency leaves are opaque:
// their contents do not affect rejection of a property-free function.
export {};
interface MapLike<T> { [index: string]: T; }


// The value whose static receiver type is being tested is the record itself.
type Receiver = (path: string) => string;
type Holding <T> = Record<string, T[]>;
function attempt<T>(): void {
const h: Holding<T> = {};
const values: Receiver[] = [h];
const receiver = values[0];
if (receiver === undefined) throw new Error("missing receiver");
console.log("same record:", Object.is(receiver, h));
console.log("typeof receiver:", typeof receiver);
try {
 console.log("call result:", receiver("called"));
} catch (error) {
 if (!(error instanceof TypeError)) throw error;
 console.log("call threw: TypeError");
}

}
attempt<string>();
