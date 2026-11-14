const collect = (...items: (string | number | string[])[]): string => items.map((item) => typeof item + ":" + (Array.isArray(item) ? item.join(",") : String(item))).join(";");
console.log(collect("a", 0));
