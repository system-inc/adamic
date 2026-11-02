// An array literal where a checked view says an object stops at the view's read on both backends: natively an
// array is never an object, and the JavaScript backend marks only tuples as objects (#nazvcep). Node prints on.
interface Thing { readonly label: string; }
interface Node { readonly kind: 'leaf' | 'branch'; }
interface Leaf extends Node { readonly kind: 'leaf'; readonly child: Thing; }
const raw = { kind: 'leaf' as const, child: [1, 2] };
const node: Node = raw;
console.log('before');
const child = (node as Leaf).child;
console.log(typeof child);
console.log(String(child.label));
