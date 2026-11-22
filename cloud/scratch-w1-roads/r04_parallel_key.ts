import { parallelMap } from 'adamic';
const seen = new WeakSet<object>();
const keys = [{ k: 1 }, { k: 2 }];
seen.add(keys[0]);
const results = parallelMap(keys, (item: { k: number }): number => item.k);
console.log(`${results.join(',')} ${seen.has(keys[0])}`);
