interface Root { readonly kind: 'Holder'; }
interface View extends Root { readonly value: readonly [string, number]; }
const triple: [string, number, boolean] = ['first'.repeat(1), 1, true];
const raw = { kind: 'Holder' as const, value: triple };
const base: Root = raw;
const viewed = base as View;
console.log(`${viewed.value[0]}:${viewed.value[1]}`);
