// spectral-norm: the Computer Language Benchmarks Game's, the largest eigenvalue of an infinite
// matrix by the power method. Arithmetic over number arrays, read and written by index.

function entry(row: number, column: number): number {
    return 1 / (((row + column) * (row + column + 1)) / 2 + row + 1);
}

function times(into: number[], vector: readonly number[]): void {
    const size = vector.length;
    for(let row = 0; row < size; row += 1) {
        let sum = 0;
        for(let column = 0; column < size; column += 1) {
            sum += entry(row, column) * (vector[column] ?? 0);
        }
        into[row] = sum;
    }
}

function timesTransposed(into: number[], vector: readonly number[]): void {
    const size = vector.length;
    for(let row = 0; row < size; row += 1) {
        let sum = 0;
        for(let column = 0; column < size; column += 1) {
            sum += entry(column, row) * (vector[column] ?? 0);
        }
        into[row] = sum;
    }
}

function timesBoth(into: number[], vector: readonly number[], scratch: number[]): void {
    times(scratch, vector);
    timesTransposed(into, scratch);
}

const size = 1500;
const estimate: number[] = new Array<number>(size).fill(1);
const product: number[] = new Array<number>(size).fill(0);
const scratch: number[] = new Array<number>(size).fill(0);
for(let round = 0; round < 10; round += 1) {
    timesBoth(product, estimate, scratch);
    timesBoth(estimate, product, scratch);
}
let vBv = 0;
let vv = 0;
for(let index = 0; index < size; index += 1) {
    const estimateEntry = estimate[index] ?? 0;
    const productEntry = product[index] ?? 0;
    vBv += estimateEntry * productEntry;
    vv += productEntry * productEntry;
}
console.log(Math.sqrt(vBv / vv).toFixed(9));
