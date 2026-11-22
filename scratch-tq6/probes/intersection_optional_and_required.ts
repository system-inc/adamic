// { valid?: boolean } & { valid: boolean }: tsc makes valid required, typed boolean, since an
// intersection's property is optional only when every member's is.
// Node prints: false, then true.
type Entry = { valid?: boolean } & { valid: boolean } & { name: string };
function probe(value: Entry): void {
  console.log(String(value.valid));
}
const entry: Entry = { name: "entry", valid: false };
probe(entry);
entry.valid = true;
probe(entry);
