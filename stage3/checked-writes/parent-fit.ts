type Mutable<T> = { -readonly [K in keyof T]: T[K] };
interface Parent { readonly name: string }
interface Node { readonly parent: Parent | undefined }
function replaceParent<T extends Node>(visited: T, parent: Parent): T {
    const copy = { ...visited };
    (copy as Mutable<T>).parent = parent;
    return copy;
}
const parent: Parent = { name: 'parent'.repeat(2) };
const node: { readonly parent: Parent } = { parent };
console.log(replaceParent(node, parent).parent.name);
