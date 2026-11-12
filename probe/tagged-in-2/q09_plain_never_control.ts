interface Plain { readonly kind: number; readonly tag: number }
const p: Plain = { kind: 1, tag: 2 };
if (!('tag' in p)) {
	const lie: { n: number } = p;
	console.log(String(lie.n + 1));
} else {
	console.log('present');
}
