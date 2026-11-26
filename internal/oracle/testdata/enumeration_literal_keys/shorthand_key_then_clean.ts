const hash = { "#": 1 };
const __proto__ = { inherited: 1 };
const shorthand = { __proto__, own: 2 };
console.log(Object.keys(shorthand).join(","));
console.log(Object.values({ first: "a", last: "z" }).join(","));
