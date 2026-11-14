const collect = (...items: (string | number)[]): string => items.map((item) => typeof item + ":" + String(item)).join(",");
const holder: { f: (a: string, b: number) => string } = { f: collect };
console.log(holder.f("x", 2));
