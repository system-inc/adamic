const s = new WeakSet<String>();
console.log(`${s.has('abc')} ${s.delete('abc')}`);
s.add('abc');
console.log('set add did not throw');
