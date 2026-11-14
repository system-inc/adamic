const collect = (...items: string[]): string => items.map((item) => typeof item + ":" + item).join(",");
const one: (a: string) => string = collect;
function run(f: (a: string, b: number) => string): string { return f("x", 2); }
console.log(run(one));
