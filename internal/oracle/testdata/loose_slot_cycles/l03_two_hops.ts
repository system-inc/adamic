interface A { name: string; x: unknown }
interface B { tag: number; y: object | undefined }
const a: A = { name: "a", x: undefined };
const b: B = { tag: 1, y: undefined };
a.x = b;
b.y = a;
console.log(a.name);
