const record: Record<string, string> = { own: "own", undefined: "via undefined" };
function pick(): string | number { return String(1).length > 0 ? "own" : 1; }
const box: { key: string | number | undefined } = { key: pick() };
function restore(): void { box.key = undefined; }
if (box.key !== undefined) {
    restore();
    console.log(String(record[box.key]));
}
