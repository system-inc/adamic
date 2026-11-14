const collect = (...items: string[]): string => items.join(",");
const viewed: (...items: string[]) => string = collect;
console.log(viewed("x", "y"));
