function show(value: unknown): string { return String(value); }
console.log(show(new WeakSet<object>()));
