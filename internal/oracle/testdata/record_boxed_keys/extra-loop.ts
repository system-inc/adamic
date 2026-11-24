const record: Record<string, string> = { a: "letter", "2": "number" };
const keys: (string | number)[] = ["a", 2, "missing"];
for (const key of keys) console.log(String(record[key]));
