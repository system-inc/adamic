import "./support/iterator.ts";
const proto = { inherited: 1 };
console.log(String(Object.values({ "__proto__": proto, own: 2 }).length));
