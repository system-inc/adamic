type A = number | string;
const nums: number[] = [1.5, 2];
const show = (...rest: A[]): void => { for (const value of rest) console.log(typeof value + ":" + String(value)); };
show(...nums);
