interface Node { readonly kind: number; }
interface Base { readonly tag: 1; }
interface SyntaxList extends Base { readonly _children: readonly (Node | undefined)[]; readonly counts: readonly number[]; readonly maybes: readonly (number | undefined)[]; }
const raw={tag:1 as const,_children:[{kind:1},undefined],counts:[7],maybes:[undefined,5]};
const base: Base = raw;
const view = base as SyntaxList;
const present=view._children[0];
console.log(present === undefined ? "missing" : String(present.kind));
const hole=view._children[1];
console.log(hole === undefined ? "missing" : String(hole.kind));
const past=view._children[2];
console.log(past === undefined ? "missing" : String(past.kind));
const count=view.counts[0];
console.log(count === undefined ? "missing" : String(count));
const none=view.counts[1];
console.log(none === undefined ? "missing" : String(none));
const absent=view.maybes[0];
console.log(absent === undefined ? "missing" : String(absent));
const five=view.maybes[1];
console.log(five === undefined ? "missing" : String(five));
