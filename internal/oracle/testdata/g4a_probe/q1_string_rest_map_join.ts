const collect = (...items: string[]): string => items.map((item) => item).join(";");
console.log(collect("a", "b"));
