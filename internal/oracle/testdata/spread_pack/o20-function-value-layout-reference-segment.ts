type A = number | string;
const nums: number[] = [1, 2];
const label = "s" + String(nums.length);
const show = (...rest: A[]): void => { for (const value of rest) console.log(typeof value + ":" + String(value)); };
show("a", "b");
const lead = (first: string, ...rest: number[]): void => { console.log(first + ":" + String(rest.length)); };
lead(label, ...nums);
console.log(label);
