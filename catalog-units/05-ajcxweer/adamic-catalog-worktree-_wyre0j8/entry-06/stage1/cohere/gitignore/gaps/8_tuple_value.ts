// Gap 8: a tuple as a value, other than an entry of new Map([...]), is not lowered.
function cut(text: string): readonly [string, number] {
	return [text.slice(1), 1];
}

const [rest, count] = cut('abc');
console.log(`${rest} ${count}`);
