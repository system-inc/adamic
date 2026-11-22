interface Holder { readonly table: Map<string, number> }
const h: Holder = { table: new Map<string, number>([['a', 1]]) };
const u: unknown = h;
console.log(JSON.stringify(u) ?? 'undefined');
