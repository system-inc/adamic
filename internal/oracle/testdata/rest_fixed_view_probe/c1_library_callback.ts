const collect = (...items: string[]): string => items.join("+");
console.log(["a", "b"].map(collect).join(","));
