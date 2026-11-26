import "./support/iterator.ts";
console.log(Object.keys(({ first: 1, second: 2 } as const)).join(","));
