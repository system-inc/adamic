import type { JsonSourceFile } from 'original-tsc-types';
interface Base { readonly kind: number; }
const raw = { kind: 308, parseDiagnostics: [{ start: 7 }, { start: 'bad' }] };
const base: Base = raw;
const items = (base as JsonSourceFile).parseDiagnostics;
console.log(`${items.length}:${items[0]!.start}`);
