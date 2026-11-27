import "./support/iterator.ts";
function table(row: { one: number }): number { return Object.entries({ ...row }).length; }
console.log(String(table({ one: 1 })));
