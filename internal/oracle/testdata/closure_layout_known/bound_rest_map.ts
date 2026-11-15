const collect = (...items: (string | number | string[])[]): string => items.map((item) => typeof item).join(";");
const bound = collect.bind(null);
console.log(["a", "b"].map(bound).join("|"));
