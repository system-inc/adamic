// The plain twin on main: a fixed object of a plain type seen as the same Partial record.
// Node prints: 3, undefined.
type Entry = { count?: number; label?: number };
const entry: Entry = { count: 3 };
const table: Partial<Record<"count" | "label", number>> = entry;
console.log(String(table.count));
console.log(String(table.label));
