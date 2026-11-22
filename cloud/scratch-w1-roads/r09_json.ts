const seen = new WeakSet<object>();
console.log(JSON.stringify(seen));
console.log(JSON.stringify({ seen }));
