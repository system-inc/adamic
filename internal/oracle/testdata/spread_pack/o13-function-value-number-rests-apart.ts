const nums: number[] = [1, 2];
const sum = (...rest: number[]): number => { let total = 0; for (const value of rest) total += value; return total; };
const count = (first: string, ...rest: number[]): string => first + ":" + String(rest.length);
console.log(String(sum(...nums)));
console.log(count("n", ...nums));
