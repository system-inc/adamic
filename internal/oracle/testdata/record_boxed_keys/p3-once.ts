const record: Record<string, string> = { "1": "one" };
let calls = 0;
function key(): string | number {
    calls += 1;
    return 1;
}
console.log(String(record[key()]));
console.log(String(calls));
