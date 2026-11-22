let seen: WeakSet<object> | undefined;
const key = {};
seen?.add(key);
console.log(`${seen?.has(key)}`);
seen = new WeakSet<object>();
seen?.add(key);
console.log(`${seen?.has(key)}`);
