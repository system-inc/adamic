const first = (a: string): string => "string:" + a;
const two: (a: string, b: number) => string = first;
console.log(two("x", 2));
