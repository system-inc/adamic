const count = (...items: (string | number | string[])[]): number => items.length;
console.log(String(count("a", 0, ["a", "b"])));
