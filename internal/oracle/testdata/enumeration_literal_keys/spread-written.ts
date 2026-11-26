import "./support/iterator.ts";
const source = { one: 1 };
console.log(String(Object.entries({ ...source }).length));
source.one = 2;
