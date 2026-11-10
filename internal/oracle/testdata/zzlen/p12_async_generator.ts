async function later(a: number, b: number): Promise<number> { return a + b; }
function* steps(limit: number): Generator<number> { for (let i = 0; i < limit; i++) yield i; }
function arity(f: Function): number { return f.length; }
console.log(`${arity(later)} ${arity(steps)}`);
