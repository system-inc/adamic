type A = number | string;
const nums: number[] = [1, 2];
function show(first: A, ...rest: number[]): void { console.log(typeof first + ":" + String(first) + ":" + String(rest.length)); }
show(7, ...nums);
