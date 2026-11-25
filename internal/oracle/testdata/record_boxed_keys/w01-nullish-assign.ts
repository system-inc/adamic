const record: Record<string, string> = {};
function update(key: string | number): void {
    record[key] ??= "set";
}
update(1);
console.log(String(record["1"]));
