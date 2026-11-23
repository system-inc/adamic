const r: { value: boolean; [key: string]: boolean } = { value: true };
r['other'] = false;
function entries(source: { readonly value: boolean }): [string, boolean][] { return Object.entries(source); }
console.log(entries(r).map(([k, v]) => k + '=' + String(v)).join(','));
