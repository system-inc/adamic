const record: Record<string, boolean> = { one: true };
function contains(key: string | number): boolean { return Object.hasOwn(record, key); }
console.log(String(contains("one")));
