// A literal method called with a non-null assertion on its callee, o.m!() and counter.add!(4), which
// JavaScript runs as o.m(): the call keeps its object (#0bxc4jh), so this is the object, where Node
// prints 1 and 7. The first is a method of the literal's own type, the second one through an optional
// function-typed field, which the assertion's check guards.
interface Counter {
	x: number;
	add?: (amount: number) => number;
}
const o = {
	x: 1,
	m(): number {
		return this.x;
	},
};
console.log(String(o.m!()));
function show(counter: Counter): void {
	console.log(String(counter.add!(4)));
}
show({
	x: 3,
	add(amount: number): number {
		return this.x + amount;
	},
});
