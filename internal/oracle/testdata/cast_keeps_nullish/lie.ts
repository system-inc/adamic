// A lying cast that keeps undefined: the value at the cast is present, an Expression of another kind, so the tag
// test the cast from Expression to Literal makes fails there, with that cast's own words. Node runs on without a
// check and prints the value's type.
interface Expression { readonly kind: number }
interface Literal extends Expression { readonly kind: 7; readonly text: string }
const eight: Expression = { kind: 8 };
function find(index: number): Expression | undefined { return index === 0 ? eight : undefined; }
console.log('casting');
const wrong = find(0) as Literal | undefined;
console.log(typeof wrong);
