// Writes through the intersection, then reads back. The member says valid?: boolean | undefined,
// since exactOptionalPropertyTypes refuses writing undefined to valid?: boolean.
// Node prints: undefined, then true, then undefined, then false.
type Entry = { valid?: boolean | undefined } & { name: string };
function write(value: Entry, valid: boolean | undefined): void {
  value.valid = valid;
}
const entry: Entry = { name: "entry" };
console.log(String(entry.valid));
write(entry, true);
console.log(String(entry.valid));
write(entry, undefined);
console.log(String(entry.valid));
entry.valid = false;
console.log(String(entry.valid));
