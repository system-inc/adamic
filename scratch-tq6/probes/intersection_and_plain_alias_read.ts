// One object read through the intersection and through a plain alias of it, each write seen by both.
// Node prints: true true, then undefined undefined, then false false.
type Plain = { valid?: boolean | undefined; name: string };
type Entry = { valid?: boolean | undefined } & { name: string };
const plain: Plain = { name: "entry", valid: true };
const entry: Entry = plain;
console.log(`${String(entry.valid)} ${String(plain.valid)}`);
entry.valid = undefined;
console.log(`${String(entry.valid)} ${String(plain.valid)}`);
plain.valid = false;
console.log(`${String(entry.valid)} ${String(plain.valid)}`);
