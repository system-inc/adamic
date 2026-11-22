// The source side (#mft3zc6, J-rmbk4bb's probe): a fixed intersection value seen as a finite Partial
// record. On main it compiled and its read through the record was an AddressSanitizer SEGV natively, where
// Node prints: 3, undefined. It stops at the relation, as its plain twin does.
type Entry = { count?: number } & { label?: number };
const entry: Entry = { count: 3 };
const table: Partial<Record<"count" | "label", number>> = entry;
console.log(String(table.count));
console.log(String(table.label));
