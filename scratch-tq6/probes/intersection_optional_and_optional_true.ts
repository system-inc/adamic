// { valid?: boolean } & { valid?: true }: valid stays optional, read as true | undefined, written as
// true (exactOptionalPropertyTypes is on, so undefined can't be written).
// Node prints: true, then undefined, then true.
type Entry = { valid?: boolean } & { valid?: true } & { name: string };
function probe(value: Entry): void {
  console.log(String(value.valid));
}
probe({ name: "entry", valid: true });
const entry: Entry = { name: "entry" };
probe(entry);
entry.valid = true;
probe(entry);
