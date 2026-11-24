type A = number | string;
const nums: number[] = [1, 2];
const show = (...rest: A[]): void => { for (const value of rest) console.log(typeof value + ":" + String(value)); };
show("a", "b");
const lead = (first: A, ...rest: number[]): void => { console.log(typeof first + ":" + String(first) + ":" + String(rest.length)); };
lead(7, ...nums);
