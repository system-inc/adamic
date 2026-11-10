const pick: (a: number, b: number) => number = (a) => a;
function arity(f: Function): number { return f.length; }
console.log(`${arity(pick)} ${arity(pick.bind(null, 1))}`);
