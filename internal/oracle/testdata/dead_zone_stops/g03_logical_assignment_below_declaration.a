interface Box { readonly flags: number; }
let deferred: Box | undefined;
const empty: Box = { flags: 0 };
let calls = 0;
function make(): Box | undefined { calls++; return undefined; }
function get(): Box { return deferred ||= make() || empty; }
console.log(String(get() === empty));
console.log(String(get() === empty));
console.log(String(calls));
