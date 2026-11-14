const collect = (...items: (string | number)[]): string => items.map((item) => typeof item + ":" + String(item)).join(",");
function apply(f: (a: string, b: number) => string): string { return f("x", 2); }
console.log(apply(collect));
