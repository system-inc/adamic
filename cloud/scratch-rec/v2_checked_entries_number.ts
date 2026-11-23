const r: { value: number; [key: string]: number } = { value: 1 };
r['other'] = 2;
function entries(source: { readonly value: number }): [string, number][] { return Object.entries(source); }
console.log(entries(r).map(([k, v]) => k + '=' + String(v)).join(','));
