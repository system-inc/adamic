const values = new Map<string, number[]>();
values.set('key', [1]);
console.log(`${values.get('key')?.[0] ?? 0}`);
