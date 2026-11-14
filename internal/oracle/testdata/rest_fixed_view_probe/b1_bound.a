const collect = (...items: string[]): string => items.map((item) => typeof item + ":" + item).join(",");
const bound = collect.bind(undefined, "w");
const one: (a: string) => string = bound;
const two: (a: string, b: number) => string = one;
console.log(two("x", 2));
