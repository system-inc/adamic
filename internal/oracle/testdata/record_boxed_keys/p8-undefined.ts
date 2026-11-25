const record: Record<string, string> = { undefined: "unset" };
function read(key: string | number | undefined): string {
    return String(record[key]);
}
console.log(read(undefined));
