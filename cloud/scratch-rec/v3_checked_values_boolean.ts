const r: { value: boolean; [key: string]: boolean } = { value: true };
r['other'] = false;
function values(source: { readonly value: boolean }): boolean[] { return Object.values(source); }
console.log(values(r).map((v) => String(v)).join(','));
