function three(a: number, b: number, c: number): number { return a + b + c; }
const holder: { readonly f: Function } = { f: three.bind(null, 1) };
console.log(String(holder.f.length));
