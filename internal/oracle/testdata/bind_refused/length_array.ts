function three(a: number, b: number, c: number): number { return a + b + c; }
const list: Function[] = [three.bind(null, 1), three];
for (const f of list) console.log(String(f.length));
