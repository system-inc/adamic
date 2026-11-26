import "./support/iterator.ts";
const source = { first: "a", last: "z" };
console.log(Object.values({ ...source }).join(","));
