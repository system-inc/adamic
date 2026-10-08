import type { InterfaceDeclaration } from 'original-tsc-types';
interface Base { readonly kind: number; }
const raw = { kind: 265, modifiers: 'bad' };
const base: Base = raw;
const items = (base as InterfaceDeclaration).modifiers!;
console.log(`${items.length}`);
