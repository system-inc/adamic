import type { ResolvedModuleWithFailedLookupLocations } from 'original-tsc-types';
interface Base { readonly resolvedModule: undefined; }
const raw = { resolvedModule: undefined, affectingLocations: ['ok', undefined] };
const base: Base = raw;
const items = (base as ResolvedModuleWithFailedLookupLocations).affectingLocations!;
console.log(`${items.length}:${items[0]!}`);
