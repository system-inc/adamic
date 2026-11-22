interface Holder { readonly marks: WeakSet<object> }
const h: Holder = { marks: new WeakSet<object>() };
const u: unknown = h;
console.log(JSON.stringify(u) ?? 'undefined');
