// @ts-check
function observed(value: number, ...tail: number[]): number { return value + tail.length; }
const fn: (value: number, ...tail: number[]) => number = observed;
const bound = fn.bind(undefined);
console.log(String(bound(3)));
