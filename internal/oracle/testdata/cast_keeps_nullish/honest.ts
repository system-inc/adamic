// A cast that keeps undefined or null (#b5w3ycg, keepsNullishCast): tsc's `node as L | undefined` of an
// `N | undefined`. A nullish value passes through the cast as itself, and a present one takes the checked view the
// cast from Expression to Literal makes, so each prints what Node prints.
interface Expression { readonly kind: number }
interface Literal extends Expression { readonly kind: 7; readonly text: string }
const seven: Literal = { kind: 7, text: 'seven' };
function find(index: number): Expression | undefined {
	return index === 0 ? seven : undefined;
}
function parent(index: number): Expression | null {
	return index === 0 ? seven : null;
}
const found = find(0) as Literal | undefined;
console.log(found === undefined ? 'absent' : found.text + ' ' + found.kind);
const missing = find(1) as Literal | undefined;
console.log(missing === undefined ? 'absent' : missing.text);
console.log(typeof missing);
const held = parent(0) as Literal | null;
console.log(held === null ? 'null' : held.text);
const none = parent(1) as Literal | null;
console.log(none === null ? 'null' : none.text);
