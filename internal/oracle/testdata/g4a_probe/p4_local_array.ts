const items: (string | number | string[])[] = ["a", 0, ["a", "b"]];
console.log(items.map((item) => typeof item + ":" + (Array.isArray(item) ? item.join(",") : String(item))).join(";"));
