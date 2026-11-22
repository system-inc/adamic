interface Holder { readonly names: Set<string> }
const h: Holder = { names: new Set<string>(['a']) };
const u: unknown = h;
if (typeof u === 'object' && u !== null && 'names' in u) {
	console.log(String(u.names));
}
