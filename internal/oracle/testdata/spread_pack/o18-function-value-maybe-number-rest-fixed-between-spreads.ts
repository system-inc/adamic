const nums: number[] = [1.5, 2];
const more: number[] = [3];
const show = (...rest: (number | undefined)[]): void => { let text = ""; for (const value of rest) text += String(value) + ","; console.log(text); };
show(...nums, undefined, ...more);
show(...more, 4, ...nums, ...more);
