const collect = (...items: string[]): string => items.map((item) => typeof item + ":" + item).join(",");
const one: (a: string) => string = collect;
const two: (a: string, b: number) => string = one;
console.log(two("x", 2));
