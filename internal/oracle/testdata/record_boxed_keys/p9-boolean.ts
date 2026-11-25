const record: Record<string, string> = { true: "yes" };
function read(key: string | number | boolean): string {
    return String(record[key]);
}
console.log(read(true));
