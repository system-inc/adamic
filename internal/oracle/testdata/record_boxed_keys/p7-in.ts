const record: Record<string, boolean> = { one: true };
function contains(key: string | number): boolean { return key in record; }
console.log(String(contains("one")));
