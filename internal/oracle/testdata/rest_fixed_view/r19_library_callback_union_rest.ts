const collect = (...items: (string | number | string[])[]): string => items.map((item) => typeof item + ":" + String(item)).join(";");
console.log(["a", "b"].map(collect).join(" | "));
