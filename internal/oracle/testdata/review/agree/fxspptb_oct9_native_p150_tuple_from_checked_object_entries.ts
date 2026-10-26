interface Root { readonly kind: 'Holder'; }
interface View extends Root { readonly value: readonly [string, number]; }
function enumerate(view: { readonly z: number }): void {
	for (const pair of Object.entries(view)) {
		const raw = { kind: 'Holder' as const, value: pair };
		const base: Root = raw;
		const viewed = base as View;
		console.log(`${viewed.value[0]}:${viewed.value[1]}`);
	}
}
const source = { z: 9, end: 5 };
enumerate(source);
