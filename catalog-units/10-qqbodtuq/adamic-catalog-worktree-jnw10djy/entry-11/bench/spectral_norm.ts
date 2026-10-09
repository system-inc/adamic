// spectral-norm: the Computer Language Benchmarks Game's, the largest eigenvalue of an infinite
// matrix by the power method. Arithmetic over number arrays, read and written by index.

function entry(i: number, j: number): number {
	return 1 / (((i + j) * (i + j + 1)) / 2 + i + 1);
}

function times(into: number[], vector: readonly number[]): void {
	const size = vector.length;
	for (let i = 0; i < size; i += 1) {
		let sum = 0;
		for (let j = 0; j < size; j += 1) {
			sum += entry(i, j) * (vector[j] ?? 0);
		}
		into[i] = sum;
	}
}

function timesTransposed(into: number[], vector: readonly number[]): void {
	const size = vector.length;
	for (let i = 0; i < size; i += 1) {
		let sum = 0;
		for (let j = 0; j < size; j += 1) {
			sum += entry(j, i) * (vector[j] ?? 0);
		}
		into[i] = sum;
	}
}

function timesBoth(into: number[], vector: readonly number[], scratch: number[]): void {
	times(scratch, vector);
	timesTransposed(into, scratch);
}

const size = 1500;
const u: number[] = new Array<number>(size).fill(1);
const v: number[] = new Array<number>(size).fill(0);
const scratch: number[] = new Array<number>(size).fill(0);
for (let round = 0; round < 10; round += 1) {
	timesBoth(v, u, scratch);
	timesBoth(u, v, scratch);
}
let vBv = 0;
let vv = 0;
for (let i = 0; i < size; i += 1) {
	const ui = u[i] ?? 0;
	const vi = v[i] ?? 0;
	vBv += ui * vi;
	vv += vi * vi;
}
console.log(Math.sqrt(vBv / vv).toFixed(9));
