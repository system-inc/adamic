const enum NumberKind { A = 2, B = 7, C = 12 }
const enum StringKind { A = 'a', B = 'b', C = 'c' }
type NA = { readonly kind: NumberKind.A; readonly text: string };
type NB = { readonly kind: NumberKind.B; readonly text: string };
type NC = { readonly kind: NumberKind.C; readonly text: string };
type SA = { readonly kind: StringKind.A; readonly text: string };
type SB = { readonly kind: StringKind.B; readonly text: string };
type SC = { readonly kind: StringKind.C; readonly text: string };
function numberOne(node: NA | NB | NC): NA { return node as NA; }
function numberSeveral(node: NA | NB | NC): NA | NC { return node as NA | NC; }
function stringOne(node: SA | SB | SC): SA { return node as SA; }
function stringSeveral(node: SA | SB | SC): SA | SC { return node as SA | SC; }
console.log(numberOne({ kind: NumberKind.A, text: 'number'.repeat(2) }).text);
console.log(numberSeveral({ kind: NumberKind.A, text: 'first'.repeat(2) }).text);
console.log(numberSeveral({ kind: NumberKind.C, text: 'last'.repeat(2) }).text);
console.log(stringOne({ kind: StringKind.A, text: 'string'.repeat(2) }).text);
console.log(stringSeveral({ kind: StringKind.A, text: 'first'.repeat(2) }).text);
console.log(stringSeveral({ kind: StringKind.C, text: 'last'.repeat(2) }).text);
