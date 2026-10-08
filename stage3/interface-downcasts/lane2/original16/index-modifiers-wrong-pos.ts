import type { IndexSignatureDeclaration } from 'original-tsc-types';
interface Base { readonly kind: number; }
const raw = { kind: 182, modifiers: [{ kind: 95, pos: 'bad' }] };
const base: Base = raw;
const items = (base as IndexSignatureDeclaration).modifiers!;
console.log(`${items[0]!.pos}`);
