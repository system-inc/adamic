// The same intersection seen as a plain shape. Node prints: 1.
type Mixed = Record<string, number> & { a: number };
const mixed: Mixed = { a: 1, b: 2 };
const plain: { a: number } = mixed;
console.log(String(plain.a));
