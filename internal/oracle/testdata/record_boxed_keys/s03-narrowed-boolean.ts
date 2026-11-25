const record: Record<string, string> = { own: "own", true: "via true" };
function pick(): string | number { return String(1).length > 0 ? "own" : 1; }
let key: string | number | boolean = pick();
function restore(): void { key = true; }
if (typeof key !== "boolean") {
    restore();
    console.log(String(record[key]));
}
