import "./support/iterator.ts";
function source(): { one: number } { return { one: 1 }; }
console.log(String(Object.entries({ ...source() }).length));
