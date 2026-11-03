// Spec: over an array that shrinks mid-search pass `undefined` to the callback
// Expect: panics find: index 1 is undefined; element type number does not admit undefined
const values: number[] = [1, 2, 3];
const found = values.find((value) => {
	if (value === 1) {
		values.pop();
		values.pop();
	}
	return value > 5;
});
console.log(String(found));
