function arity(f: Function): number { return f.length; }
const entries = new Map<string, number>([["a", 1]]).entries();
console.log(String(arity(entries.next)));
