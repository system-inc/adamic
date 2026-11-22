// The source side #mft3zc6 leaves open: a fixed intersection value seen as a finite Partial record,
// then read through the record. Node prints: 3, undefined.
type Entry = { count?: number } & { label?: number };
const entry: Entry = { count: 3 };
const table: Partial<Record<"count" | "label", number>> = entry;
console.log(String(table.count));
console.log(String(table.label));
