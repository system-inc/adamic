import "./support/iterator.ts";
console.log(Object.values(({ first: "one", second: "two" } as const)).join(","));
