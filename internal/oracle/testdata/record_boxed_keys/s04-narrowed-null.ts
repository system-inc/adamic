const record: Record<string, string> = { own: "own", null: "via null" };
function pick(): string | number { return String(1).length > 0 ? "own" : 1; }
let key: string | number | null = pick();
function restore(): void { key = null; }
if (key !== null) {
    restore();
    console.log(String(record[key]));
}
