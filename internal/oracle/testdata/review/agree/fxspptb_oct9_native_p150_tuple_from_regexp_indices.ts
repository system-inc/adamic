interface Root { readonly kind: 'Holder'; }
interface View extends Root { readonly value: readonly [number, number]; }
const match = /a(b)/d.exec('xab');
if (match !== null && match.indices !== undefined) {
	const pair = match.indices[1];
	const raw = { kind: 'Holder' as const, value: pair };
	const base: Root = raw;
	const viewed = base as View;
	console.log(`${viewed.value[0]}:${viewed.value[1]}`);
}
