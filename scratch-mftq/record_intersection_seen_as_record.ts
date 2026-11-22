// Lower's question on #mft3zc6: an intersection holding a record keeps sameRecordStorage's nil answer.
// Seen as the record. Node prints: 2, 1.
type Mixed = Record<string, number> & { a: number };
const mixed: Mixed = { a: 1, b: 2 };
const table: Record<string, number> = mixed;
console.log(String(table["b"]));
console.log(String(table.a));
