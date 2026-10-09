interface Root { readonly kind: 'Holder'; }
interface View extends Root { readonly value: readonly [string, number]; }
const record: { [key: string]: number } = { alpha: 1, beta: 2 };
for (const pair of Object.entries(record)) {
	const raw = { kind: 'Holder' as const, value: pair };
	const base: Root = raw;
	const viewed = base as View;
	console.log(`${viewed.value[0]}:${viewed.value[1]}`);
}
