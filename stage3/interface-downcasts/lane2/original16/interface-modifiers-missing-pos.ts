import type { InterfaceDeclaration } from 'original-tsc-types';
interface Base { readonly kind: number; }
const raw = { kind: 265, modifiers: [{ kind: 95 }] };
const base: Base = raw;
const items = (base as InterfaceDeclaration).modifiers!;
console.log(`${items[0]!.pos}`);
