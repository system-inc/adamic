// @ts-check
function observed(value: number): number { return arguments.length + value; }
const fn: (value: number) => number = observed;
const bound = fn.bind(undefined);
console.log(String(bound(3)));
