interface Root { readonly kind: 'Holder'; }
interface View extends Root { readonly value: readonly [string, string]; }
const set = new Set<string>();
set.add('a'.toUpperCase());
for (const pair of set.entries()) {
	const raw = { kind: 'Holder' as const, value: pair };
	const base: Root = raw;
	const viewed = base as View;
	console.log(`${viewed.value[0]}:${viewed.value[1]}`);
}
