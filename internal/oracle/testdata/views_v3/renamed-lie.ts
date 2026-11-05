interface Token { readonly kind: number; }
interface Base { readonly tag: 1; }
interface SyntaxList extends Base { readonly _children: ReadonlyArray<Token>; }
const raw = {tag:1 as const, _children:[7]};
const base: Base = raw;
const view = base as SyntaxList;
console.log('before');
const child=view._children[0];
console.log('after');
