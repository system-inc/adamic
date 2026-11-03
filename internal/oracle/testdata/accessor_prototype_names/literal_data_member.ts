// An object literal with a getter, and so a class of its own, whose data member is named after an
// Object.prototype member: the member is a data field on both backends, as Node reads it (#76y1mth).
// Node prints 1, 2.
const literal = {
	valueOf: 1,
	get x(): number {
		return 2;
	},
};
console.log(String(literal.valueOf));
console.log(String(literal.x));
