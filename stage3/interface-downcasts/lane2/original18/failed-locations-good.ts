import type { ResolvedModuleWithFailedLookupLocations } from 'original-tsc-types';
interface Base { readonly resolvedModule: undefined; }
function read(base: Base): string {
 const items = (base as ResolvedModuleWithFailedLookupLocations).failedLookupLocations;
 if (items === undefined) return 'absent';
 return `${items.length}:${items[0]!}`;
}
const present = { resolvedModule: undefined, failedLookupLocations: ['ok'] };
console.log(read(present));
const absent = { resolvedModule: undefined };
console.log(read(absent));
const explicit = { resolvedModule: undefined, failedLookupLocations: undefined };
console.log(read(explicit));
