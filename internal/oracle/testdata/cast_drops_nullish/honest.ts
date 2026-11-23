// A cast that only removes undefined or null (#b5w3ycg, presentCast): tsc's `args[0] as Expression`, with the value
// there at the cast. It is the check a non-null assertion makes, so it passes, and the program prints what Node prints.
interface Expression { readonly kind: number; readonly text: string }
const identifier: Expression = { kind: 1, text: 'identifier' };
function argument(index: number): Expression | undefined {
	return index === 0 ? identifier : undefined;
}
function parent(index: number): Expression | null {
	return index === 0 ? identifier : null;
}
const first = argument(0) as Expression;
console.log(first.text + ' ' + first.kind);
console.log((parent(0) as Expression).text);
