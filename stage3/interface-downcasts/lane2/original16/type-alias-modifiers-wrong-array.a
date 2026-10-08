import type { TypeAliasDeclaration } from 'original-tsc-types';
interface Base { readonly kind: number; }
const raw = { kind: 266, modifiers: 'bad' };
const base: Base = raw;
const items = (base as TypeAliasDeclaration).modifiers!;
console.log(`${items.length}`);
