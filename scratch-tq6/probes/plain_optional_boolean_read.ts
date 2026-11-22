// The scout's program's non-intersection twin: one object type holds both fields.
// Node prints: false, then undefined.
type Entry = { valid?: boolean; name: string };
function probe(value: Entry): void {
  console.log(String(value.valid));
}
probe({ name: "entry", valid: false });
probe({ name: "entry" });
