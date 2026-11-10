function two(a: number, b: number): number { return a + b; }
function defaults(a: number, b: number = 1, ...rest: number[]): number { return a + b + rest.length; }
const arrow = (x: string, y: string): string => x + y;
function arity(f: Function): number { return f.length; }
console.log(`${arity(two)} ${arity(defaults)} ${arity(arrow)}`);
