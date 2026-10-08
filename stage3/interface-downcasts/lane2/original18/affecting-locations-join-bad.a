import type { ResolvedModuleWithFailedLookupLocations } from 'original-tsc-types';
interface Base { readonly resolvedModule: undefined; }
const raw = { resolvedModule: undefined, affectingLocations: [7, 8] };
const base: Base = raw;
const items = (base as ResolvedModuleWithFailedLookupLocations).affectingLocations!;
console.log(items.join(';'));
