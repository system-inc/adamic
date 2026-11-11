// A function expression written into a literal's method member, then called through the member.
interface Counter {
	step(by: number): number;
}
const counter: Counter = {
	step(by: number): number {
		return by + 1;
	},
};
console.log(String(counter.step(1)));
counter.step = function (by: number): number {
	return by * 10;
};
console.log(String(counter.step(2)));
