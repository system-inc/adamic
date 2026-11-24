type A = number | string;
const nums: number[] = [1, 2];
const lead = (first: A, ...rest: number[]): void => { console.log(typeof first + ":" + String(first) + ":" + String(rest.length)); };
lead(7, ...nums);
