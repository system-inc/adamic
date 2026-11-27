type A = number | string;
const nums: number[] = [1.5, 2];
const more: number[] = [3];
const show = (...rest: A[]): void => { let text = ""; for (const value of rest) text += typeof value + ":" + String(value) + ","; console.log(text); };
const sum = (...rest: number[]): void => { let total = 0; for (const value of rest) total += value; console.log(String(total)); };
let pick: (...rest: number[]) => void = sum;
pick(...nums, 7, ...more);
pick = show;
pick(...nums, 7, ...more);
