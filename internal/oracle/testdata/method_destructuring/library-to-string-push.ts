interface S { toString(): string; push(...items: number[]): number }
function display({ toString }: S): string { return toString(); }
console.log(display([1, 2]));
