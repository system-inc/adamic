const count = (...items: (number | string[])[]): number => items.length;
console.log(String(count(0, ["a"])));
