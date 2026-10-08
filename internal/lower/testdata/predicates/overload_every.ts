const SyntaxKind = { Identifier: 80 } as const;
interface Node { readonly kind: number }
interface Identifier extends Node { readonly kind: typeof SyntaxKind.Identifier; readonly escapedText: string }
function isIdentifier(node: Node): node is Identifier { return node.kind === SyntaxKind.Identifier; }
function every<T, U extends T>(array: readonly T[] | undefined, callback: (element: T, index: number) => element is U): array is readonly U[] | undefined;
function every<T>(array: readonly T[] | undefined, callback: (element: T, index: number) => boolean): boolean {
    if (array !== undefined) {
        for (let i = 0; i < array.length; i++) {
            if (!callback(array[i]!, i)) return false;
        }
    }
    return true;
}
const identifier: Identifier = { kind: SyntaxKind.Identifier, escapedText: "name" };
const nodes: readonly Node[] = [identifier];
if (every(nodes, isIdentifier)) console.log(nodes[0]!.escapedText);
const missing: readonly Node[] | undefined = undefined;
console.log(`${every(missing, isIdentifier)}`);
