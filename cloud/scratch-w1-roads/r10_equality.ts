const seen = new WeakSet<object>();
const same = seen;
const other = new WeakSet<object>();
console.log(`${seen === same} ${seen === other} ${seen !== other}`);
