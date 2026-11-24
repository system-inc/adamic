// k03 (#pyak7te, Seats' lane build22-44 round 5): a shorthand { __proto__ } makes an own key named __proto__,
// never the prototype, so Object.keys lists it.
const __proto__ = { inherited: 1 };
const shorthand = { __proto__, own: 2 };
console.log(Object.keys(shorthand).join(","));
