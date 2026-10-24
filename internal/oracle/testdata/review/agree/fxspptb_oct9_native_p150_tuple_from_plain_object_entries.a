interface Root { readonly kind: 'Holder'; }
interface View extends Root { readonly value: readonly [string, number]; }
const object = { alpha: 1, beta: 2 };
for (const pair of Object.entries(object)) {
	const raw = { kind: 'Holder' as const, value: pair };
	const base: Root = raw;
	const viewed = base as View;
	console.log(`${viewed.value[0]}:${viewed.value[1]}`);
}
