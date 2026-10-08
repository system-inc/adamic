import type { ResolvedModuleWithFailedLookupLocations } from 'original-tsc-types';
interface Base { readonly resolvedModule: undefined; }
function read(base: Base): string {
 const items = (base as ResolvedModuleWithFailedLookupLocations).affectingLocations;
 if (items === undefined) return 'absent';
 return `${items.length}:${items[0]!}`;
}
const present = { resolvedModule: undefined, affectingLocations: ['ok'] };
console.log(read(present));
const absent = { resolvedModule: undefined };
console.log(read(absent));
const explicit = { resolvedModule: undefined, affectingLocations: undefined };
console.log(read(explicit));
