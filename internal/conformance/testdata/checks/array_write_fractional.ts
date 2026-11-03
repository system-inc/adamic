// Spec: Writing past the end (or a negative or fractional index) panics, in both backends
// Expect: panics index 1.5 is outside an array of length 3
const values: number[] = [1, 2, 3];
const at: number[] = [1.5];
console.log(String(values.length));
values[at[0] ?? 0] = 9;
console.log(values.join(','));
