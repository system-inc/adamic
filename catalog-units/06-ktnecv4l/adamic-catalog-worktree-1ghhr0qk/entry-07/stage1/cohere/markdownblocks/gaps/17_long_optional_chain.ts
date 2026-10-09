interface Item { readonly values: readonly number[]; }
const items: readonly Item[] = [{ values: [1] }];
console.log(`${items[0]?.values.length ?? -1}`);
