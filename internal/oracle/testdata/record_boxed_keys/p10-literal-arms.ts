const record: Record<string, string> = { a: "letter", "1": "number" };
function read(key: "a" | 1): string { return String(record[key]); }
console.log(read("a"));
console.log(read(1));
