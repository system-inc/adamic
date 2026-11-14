const collect = (...items: (string | number)[]): string => items.map((item) => typeof item + ":" + String(item)).join(",");
const fixed = collect as (a: string, b: number) => string;
console.log(fixed("x", 2));
