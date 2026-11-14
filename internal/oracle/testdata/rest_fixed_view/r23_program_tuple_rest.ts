const collect = (...items: [string]): string => String(items.length) + ":" + items.join(",");
const one: (a: string) => string = collect;
const two: (a: string, b: number) => string = one;
console.log(two("x", 2));
