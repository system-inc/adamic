type A = number | string;
const nums: number[] = [1, 2];
const strs: string[] = ["a", "b"];
function show(...rest: A[]): void { for (const value of rest) console.log(typeof value + ":" + String(value)); }
show(...strs);
function lead(first: A, ...rest: number[]): void { console.log(typeof first + ":" + String(first) + ":" + String(rest.length)); }
lead("x", ...nums);
