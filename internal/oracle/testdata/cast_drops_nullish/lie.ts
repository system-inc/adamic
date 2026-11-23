// A lying cast that only removes undefined: the value is undefined at the cast, so the presence check a non-null
// assertion makes fails there, naming the cast as written. Node runs on without a check and prints undefined.
interface Expression { readonly kind: number; readonly text: string }
const identifier: Expression = { kind: 1, text: 'identifier' };
function argument(index: number): Expression | undefined { return index === 0 ? identifier : undefined; }
console.log('casting');
const missing = argument(1) as Expression;
console.log(typeof missing);
