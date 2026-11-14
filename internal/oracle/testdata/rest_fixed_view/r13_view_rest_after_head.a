const collect = (...items: (string | number)[]): string => items.map((item) => typeof item + ":" + String(item)).join(",");
const viewed: (a: string, ...values: number[]) => string = collect;
console.log(viewed("x", 2));
