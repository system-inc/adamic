// The object reaching the intersection was built under another static type: a const literal whose
// own type says valid: boolean, required, so its slot was made as a plain boolean.
// Node prints: false, then true.
type Entry = { valid?: boolean } & { name: string };
function probe(value: Entry): void {
  console.log(String(value.valid));
}
const built = { name: "entry", valid: false };
probe(built);
built.valid = true;
probe(built);
