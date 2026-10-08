import type { ResolvedModuleWithFailedLookupLocations } from 'original-tsc-types';
interface Base { readonly resolvedModule: undefined; }
const raw = { resolvedModule: undefined, failedLookupLocations: ['a', 'b'] };
const base: Base = raw;
const items = (base as ResolvedModuleWithFailedLookupLocations).failedLookupLocations!;
console.log(items.join(';'));
