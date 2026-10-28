interface Node { readonly kind: 'container' | 'other'; }
interface Container extends Node { readonly kind: 'container'; readonly child: { readonly ready: boolean; readonly label: string; }; }
function narrow(node: Node): Container { return node as Container; }
class Child { ready!: boolean; readonly label='ok'; }
const raw={kind:'container' as const,child:new Child()};
const held:Node=raw; const view=narrow(held); console.log(`${view.child.ready}`);
