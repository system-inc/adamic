const hash = { "#": 1 };
const __proto__ = { inherited: 1 };
console.log(Object.values({
    row: Object.keys({ __proto__, own: 2 }).join(","),
}).join(","));
