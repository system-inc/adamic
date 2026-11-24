const set = new WeakSet<object>();
const view: { has(value: object): boolean } = set;
console.log(view.toString());
