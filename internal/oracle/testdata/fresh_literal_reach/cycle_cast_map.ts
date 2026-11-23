interface Named { readonly name: string }
const table = new Map<string, Named>();
table.set("k", { name: "m", back: table } as Named);
console.log(`${table.size} ${table.get("k")?.name ?? "none"}`);
