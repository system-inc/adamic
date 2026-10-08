import type { TypeAliasDeclaration } from 'original-tsc-types';
interface Base { readonly kind: number; }
const raw = { kind: 266, modifiers: [{ kind: 95, pos: 7 }, { kind: 95, pos: 'bad' }] };
const base: Base = raw;
const items = (base as TypeAliasDeclaration).modifiers!;
console.log(`${items.length}:${items[0]!.pos}`);
