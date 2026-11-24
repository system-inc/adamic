// The lane's prototype-inline.ts (#pyak7te): the shorthand made inline, as an argument, inside another literal.
const hash = { "#": 1 };
const __proto__ = { inherited: 1 };
console.log(Object.values({
	row: Object.keys({ __proto__, own: 2 }).join(","),
}).join(","));
