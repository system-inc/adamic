const nums: number[] = [1, 2];
const loose = (first?: number, ...rest: number[]): string => String(first) + ":" + String(rest.length);
const strict = (first: number, ...rest: number[]): string => String(first + 1) + ":" + String(rest.length);
console.log(loose(1, ...nums));
console.log(strict(2, ...nums));
console.log(loose(...nums));
