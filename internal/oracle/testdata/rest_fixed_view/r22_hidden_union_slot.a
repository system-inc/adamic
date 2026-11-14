const collect = (...items: string[]): string => items.map((item) => typeof item + ":" + item).join(",");
const slot: ((a: string) => string) | ((a: number) => number) = collect;
const two: ((a: string, b: number) => string) | ((a: number) => number) = slot;
console.log(String((two as (a: string, b: number) => string)("x", 2)));
