const collect = (...items: string[]): string => items.join(";");
const through = (callback: (item: string) => string, items: string[]): string => items.map(callback).join(";");
console.log(through((item) => item, ["a", "b"]) + " " + collect("a", "b"));
