const held: unknown = new WeakSet<object>();
console.log(typeof held);
console.log(`${held instanceof WeakSet}`);
