// Gap 1: two functions that call each other. Each call to a function declared later is lowered before
// that function's body and taken for a void call (gitignore's GAPS.md, gap 3), and with two functions
// that call each other one of the calls is always to the later one, so no order lowers.
function isEven(value: number): boolean {
	return value === 0 ? true : isOdd(value - 1);
}
function isOdd(value: number): boolean {
	return value === 0 ? false : isEven(value - 1);
}
console.log(`${isEven(4)} ${isOdd(4)}`);
