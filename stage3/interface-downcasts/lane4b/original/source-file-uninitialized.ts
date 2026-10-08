import type { SourceFile } from 'original-tsc-types';
interface Base { readonly kind: number; }
function read(file: SourceFile): string {
 const member = file.externalModuleIndicator;
 if (member === undefined) return 'absent';
 if (member === null) return 'null';
 if (typeof member === 'boolean') return `${member}`;
 return `${member.kind}`;
}
const concrete: { readonly kind: number; externalModuleIndicator: SourceFile['externalModuleIndicator']; } = { kind: 308, externalModuleIndicator: null! };
const base: Base = concrete;
console.log(read(base as SourceFile));
