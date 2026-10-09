// Gap 3: a call to a function declared later in the module is taken for a void one.
function first(value: number): number {
	return second(value) + 1;
}

function second(value: number): number {
	return value * 2;
}

console.log(`${first(3)}`);
