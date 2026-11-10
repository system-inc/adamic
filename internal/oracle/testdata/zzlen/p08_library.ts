function arity(f: Function): number { return f.length; }
console.log(`${arity(Math.max)} ${arity(parseInt)} ${arity(JSON.stringify)} ${arity(String)}`);
