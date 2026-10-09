type Mutable<T> = { -readonly [K in keyof T]: T[K] };
interface Parent { readonly name: string }
interface Node { readonly parent: Parent | undefined }
function clearParent<T extends Node>(visited: T): T {
    const copy = { ...visited };
    (copy as Mutable<T>).parent = undefined;
    return copy;
}
const parent: Parent = { name: 'parent'.repeat(2) };
const node: { readonly parent: Parent } = { parent };
clearParent(node);
console.log('stored');
