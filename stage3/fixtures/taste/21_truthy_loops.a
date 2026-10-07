// From TypeScript 6.0.3, src/compiler/utilities.ts:5255
// From TypeScript 6.0.3, src/compiler/binder.ts:1607
// From TypeScript 6.0.3, src/compiler/checker.ts:25300
// Census reason: non-boolean control condition
interface Node {readonly kind: number; readonly parent: Node | undefined;}
type SyntaxKind = number;
interface Label {readonly name: string; readonly next: Label | undefined;}
type __String = string;
let activeLabelList: Label | undefined;
type Type = {readonly flags: 0} | IndexedAccessType;
interface IndexedAccessType {readonly flags: 8388608; readonly objectType: Type;}
const TypeFlags = {IndexedAccess: 8388608};
export function getAncestor(node: Node | undefined, kind: SyntaxKind): Node | undefined {
    while (node) {
        if (node.kind === kind) {
            return node;
        }
        node = node.parent;
    }
    return undefined;
}
function findActiveLabel(name: __String) {
        for (let label = activeLabelList; label; label = label.next) {
            if (label.name === name) {
                return label;
            }
        }
        return undefined;
    }
function getRecursionIdentity(type: Type): Type {
        if (type.flags & TypeFlags.IndexedAccess) {
            // Identity is the leftmost object type in a chain of indexed accesses, eg, in A[P1][P2][P3] it is A.
            do {
                type = (type as IndexedAccessType).objectType;
            }
            while (type.flags & TypeFlags.IndexedAccess);
            return type;
        }
        return type;
    }
const root: Node = {kind: 1, parent: undefined};
const child: Node = {kind: 2, parent: root};
console.log(String(getAncestor(child, 1) === root));
console.log(String(getAncestor(child, 3) === undefined));
activeLabelList = {name: "inner", next: {name: "outer", next: undefined}};
console.log(findActiveLabel("outer")?.name ?? "missing");
console.log(findActiveLabel("missing")?.name ?? "missing");
const leaf: Type = {flags: 0};
const access: IndexedAccessType = {flags: 8388608, objectType: {flags: 8388608, objectType: leaf}};
console.log(String(getRecursionIdentity(access) === leaf));
