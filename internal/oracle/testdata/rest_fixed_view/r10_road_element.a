const collect = (...items: (string | number)[]): string => items.map((item) => typeof item + ":" + String(item)).join(",");
const collects: ((a: string, b: number) => string)[] = [collect];
for (const f of collects) { console.log(f("x", 2)); }
