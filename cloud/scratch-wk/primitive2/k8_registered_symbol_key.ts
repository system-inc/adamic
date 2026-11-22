const m = new WeakMap<WeakKey, number>();
const registered: symbol = Symbol.for('registered');
console.log(`${m.has(registered)}`);
m.set(registered, 1);
console.log('registered symbol did not throw');
