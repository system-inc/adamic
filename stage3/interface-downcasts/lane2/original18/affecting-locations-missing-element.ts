import type { ResolvedModuleWithFailedLookupLocations } from 'original-tsc-types';
interface Base { readonly resolvedModule: undefined; }
const raw = { resolvedModule: undefined, affectingLocations: [undefined] };
const base: Base = raw;
const items = (base as ResolvedModuleWithFailedLookupLocations).affectingLocations!;
console.log(`${items[0]!}`);
