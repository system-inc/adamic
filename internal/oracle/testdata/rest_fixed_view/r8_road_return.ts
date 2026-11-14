const collect = (...items: (string | number)[]): string => items.map((item) => typeof item + ":" + String(item)).join(",");
function make(): (a: string, b: number) => string { return collect; }
console.log(make()("x", 2));
