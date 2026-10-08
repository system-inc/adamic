import type { ResolvedModuleWithFailedLookupLocations } from 'original-tsc-types';
interface Base { readonly resolvedModule: undefined; }
const raw = { resolvedModule: undefined, failedLookupLocations: ['ok', undefined] };
const base: Base = raw;
const items = (base as ResolvedModuleWithFailedLookupLocations).failedLookupLocations!;
console.log(`${items.length}:${items[0]!}`);
