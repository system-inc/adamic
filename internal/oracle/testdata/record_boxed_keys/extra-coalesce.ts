const record: Record<string, string> = { a: "letter" };
function show(key: string | number): void {
    console.log(record[key] ?? "missing");
}
show("a");
show(2);
