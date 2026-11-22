// { valid?: boolean } & { valid?: boolean | undefined }: valid stays optional, read as
// boolean | undefined. The intersection's write type is boolean, the first member's, so undefined
// can't be written under exactOptionalPropertyTypes.
// Node prints: false, then undefined, then true.
type Entry = { valid?: boolean } & { valid?: boolean | undefined } & { name: string };
function probe(value: Entry): void {
  console.log(String(value.valid));
}
probe({ name: "entry", valid: false });
const entry: Entry = { name: "entry" };
probe(entry);
entry.valid = true;
probe(entry);
