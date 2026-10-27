interface Node { readonly kind: string; }
interface Base { readonly kind: string; }
interface Box extends Base { readonly kind: 'Box'; readonly externalModuleIndicator: true | Node | undefined; }
function read(value: Box): string {
 const member = value.externalModuleIndicator;
 if (member === undefined) return 'absent';
 if (typeof member === 'boolean') return `${member}`;
 return member.kind;
}
const concrete0: { readonly kind: string; readonly externalModuleIndicator: boolean | Node } = { kind: 'Box', externalModuleIndicator: false };
const source0: Base = concrete0;
console.log(read(source0 as Box));
