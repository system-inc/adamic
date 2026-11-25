const record: Record<string, string> = { a: "letter", "1": "number" };
function read(key: (string & {}) | number): void { console.log(String(record[key])); }
read("a");
read(1);
