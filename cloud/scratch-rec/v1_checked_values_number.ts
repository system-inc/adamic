const r: { value: number; [key: string]: number } = { value: 1 };
r['other'] = 2;
function values(source: { readonly value: number }): number[] { return Object.values(source); }
console.log(values(r).join(','));
