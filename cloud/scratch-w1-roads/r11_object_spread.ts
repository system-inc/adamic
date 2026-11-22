const seen = new WeakSet<object>();
const copy = { ...seen };
console.log(JSON.stringify(copy));
