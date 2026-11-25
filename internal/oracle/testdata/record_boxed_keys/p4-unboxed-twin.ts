const record: Record<string, string> = { own: "value" };
function read(key: string): void {
    console.log(typeof record[key]);
}
function key(): string { return "toString"; }
read(key());
