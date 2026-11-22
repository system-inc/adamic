const m = new WeakMap<Object, number>();
console.log(`${m.has(1)} ${m.get(1)} ${m.delete(1)}`);
m.set(1, 1);
console.log('map set did not throw');
