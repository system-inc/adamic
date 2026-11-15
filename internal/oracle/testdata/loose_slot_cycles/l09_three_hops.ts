interface A { name: string; x: unknown }
interface B { tag: number; y: unknown }
interface C { flag: boolean; z: object | undefined }
const a: A = { name: "a", x: undefined };
const b: B = { tag: 1, y: undefined };
const c: C = { flag: true, z: undefined };
a.x = b;
b.y = c;
c.z = a;
console.log(a.name);
