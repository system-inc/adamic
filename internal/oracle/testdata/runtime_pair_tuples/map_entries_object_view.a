interface Root { readonly kind: 'Holder'; }
interface View extends Root { readonly value: { readonly 0: string; readonly 1: number }; }
const map = new Map<string, number>();
map.set('a'.toUpperCase(), 1);
for (const pair of map.entries()) {
	const raw = { kind: 'Holder' as const, value: pair };
	const base: Root = raw;
	const viewed = base as View;
	console.log(`${viewed.value[0]}:${viewed.value[1]}`);
}
