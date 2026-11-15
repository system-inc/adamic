const count = (...items: string[]): number => items.map((item) => item).length;
console.log(String(count("a", "b")));
