interface A { name: string; x: unknown }
interface B { label: string; y: unknown }
const a: A = { name: "a", x: undefined };
const holder = { tag: 1, n: a };
const v: { tag: number } = holder;
const b: B = { label: "b", y: v };
a.x = b;
console.log(a.name);
