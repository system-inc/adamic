const v: string | Set<number | string> = new Set([0, 1]);
if (typeof v !== "string") {
    v.add("c");
    for (const x of v) { console.log(String(x)); }
}
