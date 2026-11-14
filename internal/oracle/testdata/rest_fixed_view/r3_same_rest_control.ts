const collect = (...items: string[]): string => items.join(",");
const fixed: (a: string, b: string) => string = collect;
console.log(fixed("x", "y"));
