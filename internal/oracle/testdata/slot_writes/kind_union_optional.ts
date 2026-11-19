// flowNode as tsc declares it, optional on both kinds, written through a kind-checked cast's union
// (#b5w3ycg head 2): odd was built with flowNode undefined, as tsc's factory writes it, so it holds the slot
// and the write is admitted (#w8nanzw), and agrees with Node.
enum SyntaxKind { Identifier = 80, PropertyAccessExpression = 211, ElementAccessExpression = 212 }
interface Node { readonly kind: SyntaxKind; readonly pos: number }
interface PropertyAccessExpression extends Node { readonly kind: SyntaxKind.PropertyAccessExpression; flowNode?: number | undefined }
interface ElementAccessExpression extends Node { readonly kind: SyntaxKind.ElementAccessExpression; flowNode?: number | undefined }
function mark(node: Node): void { (node as PropertyAccessExpression | ElementAccessExpression).flowNode = 5; }
const odd: ElementAccessExpression = { kind: SyntaxKind.ElementAccessExpression, pos: 0, flowNode: undefined };
mark(odd);
console.log(String(odd.flowNode));
