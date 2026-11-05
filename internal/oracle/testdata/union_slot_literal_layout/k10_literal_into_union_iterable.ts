function show(v: "s" | Iterable<number | string>): void {
    if (v !== "s") { for (const x of v) { console.log(String(x)); } }
}
show([0, 1]);
