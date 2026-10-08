import type { ResolvedModuleWithFailedLookupLocations } from 'original-tsc-types';
interface Base { readonly resolvedModule: undefined; }
const raw = { resolvedModule: undefined, affectingLocations: 'bad' };
const base: Base = raw;
const items = (base as ResolvedModuleWithFailedLookupLocations).affectingLocations!;
console.log(`${items.length}`);
