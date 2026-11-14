const collect = (head: string, ...items: unknown[]): string => head + "|" + items.map((item) => typeof item).join(",");
const fixed: (a: string, b: number, c: boolean) => string = collect;
console.log(fixed("h", 1, true));
