interface Node { readonly kind: string; }
interface Base { readonly kind: string; }
interface Box extends Base { readonly kind: 'Box'; readonly externalModuleIndicator: true | Node | undefined; }
function read(value: Box): string {
 const member = value.externalModuleIndicator;
 if (member === undefined) return 'absent';
 if (typeof member === 'boolean') return `${member}`;
 return member.kind;
}
const concrete0 = { kind: 'Box', externalModuleIndicator: true };
const source0: Base = concrete0;
console.log(read(source0 as Box));
const concrete1 = { kind: 'Box', externalModuleIndicator: { kind: 'Identifier' } };
const source1: Base = concrete1;
console.log(read(source1 as Box));
const concrete2 = { kind: 'Box', externalModuleIndicator: undefined };
const source2: Base = concrete2;
console.log(read(source2 as Box));
