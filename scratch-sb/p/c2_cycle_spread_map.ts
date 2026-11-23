interface Named { readonly name: string }
const table = new Map<string, Named>();
const extra = { back: table };
table.set("k", { ...extra, name: "m" });
console.log(String(table.size), table.get("k")?.name ?? "none");
