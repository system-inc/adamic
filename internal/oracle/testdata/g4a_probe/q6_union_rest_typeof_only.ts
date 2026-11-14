const collect = (...items: (string | number | string[])[]): string => items.map((item) => typeof item).join(";");
console.log(collect("a", 0));
