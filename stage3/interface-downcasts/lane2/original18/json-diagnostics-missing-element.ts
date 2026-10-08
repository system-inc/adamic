import type { JsonSourceFile } from 'original-tsc-types';
interface Base { readonly kind: number; }
const raw = { kind: 308, parseDiagnostics: [{  }] };
const base: Base = raw;
const items = (base as JsonSourceFile).parseDiagnostics;
console.log(`${items[0]!.start}`);
