// Spec: Writing past the end (or a negative or fractional index) panics, in both backends
// Expect: panics index -1 is outside an array of length 3
const values: number[] = [1, 2, 3];
const offsets: number[] = [-1];
console.log(String(values.length));
values[offsets[0] ?? 0] = 9;
console.log(values.join(','));
