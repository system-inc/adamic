const collect = (...items: string[]): string => items.map((item) => typeof item + ":" + item).join(",");
const bound = collect.bind(undefined, "w");
console.log(bound("x", "y"));
