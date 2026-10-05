// Gap 4: a declared function used as a value is not lowered.
function isEven(value: number): boolean {
	return value % 2 === 0;
}

const test: (value: number) => boolean = isEven;
console.log(`${test(4)}`);
