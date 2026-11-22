// The scout's program: an intersection with an optional boolean field, read present and absent.
// Node prints: false, then undefined.
type Entry = { valid?: boolean } & { name: string };
function probe(value: Entry): void {
  console.log(String(value.valid));
}
probe({ name: "entry", valid: false });
probe({ name: "entry" });
