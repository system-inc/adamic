const record: Record<string, string> = { "1": "one" };
function remove(key: string | number): void {
    delete record[key];
}
remove(1);
console.log(String(record["1"]));
