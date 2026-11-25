const record: Record<string, boolean> = {};
function write(key: string | number): void {
    record[key] = true;
}
write(1);
console.log(String(record["1"]));
