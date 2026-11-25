const record: Record<string, string> = { own: "own", undefined: "via undefined" };
function pick(): string | number { return String(1).length > 0 ? "own" : 1; }
let key: string | number | undefined = pick();
function restore(): void { key = undefined; }
if (key !== undefined) {
    restore();
    console.log(String(record[key]));
}
