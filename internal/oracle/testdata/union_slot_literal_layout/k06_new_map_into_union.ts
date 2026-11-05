const v: string | Map<string, number | string> = new Map([["a", 1]]);
if (typeof v !== "string") {
    v.set("b", "two");
    for (const [key, value] of v) { console.log(key + " = " + String(value)); }
}
