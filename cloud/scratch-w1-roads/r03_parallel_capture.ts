import { parallelMap } from 'adamic';
const seen = new WeakSet<object>();
const key = { k: 1 };
seen.add(key);
const results = parallelMap([1, 2, 3], (item: number): number => (seen.has(key) ? item : 0));
console.log(results.join(','));
