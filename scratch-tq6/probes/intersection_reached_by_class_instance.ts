// The object reaching the intersection is a class instance whose valid field is a plain boolean.
// Node prints: false, then true.
type Entry = { valid?: boolean } & { name: string };
class Flagged {
  name = "entry";
  valid = false;
}
function probe(value: Entry): void {
  console.log(String(value.valid));
}
const flagged = new Flagged();
probe(flagged);
flagged.valid = true;
probe(flagged);
