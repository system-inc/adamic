interface Tree {
    readonly name: string;
    readonly children: Tree[];
    readonly parent: Tree | undefined;
}
const root: Tree = { name: 'root', children: [], parent: undefined };
const child: Tree = { name: 'child', children: [], parent: root };
root.children.push(child);
console.log(child.parent?.name ?? 'missing');
