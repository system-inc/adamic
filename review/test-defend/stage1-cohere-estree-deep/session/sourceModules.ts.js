export function externalModule(nodes, root) {
    const source = nodes[root];
    if (source === undefined) {
        return false;
    }
    for (const id of source.children){
        const node = nodes[id];
        if (node === undefined) {
            continue;
        }
        if ([
            'ImportDeclaration',
            'ExportDeclaration',
            'ExportAssignment',
            'NamespaceExportDeclaration'
        ].includes(node.kind)) {
            return true;
        }
        if (node.kind === 'ImportEqualsDeclaration') {
            for (const child of node.children){
                if (nodes[child]?.kind === 'ExternalModuleReference') {
                    return true;
                }
            }
        }
        for (const child of node.children){
            if (nodes[child]?.kind === 'ExportKeyword') {
                return true;
            }
        }
    }
    for (const node of nodes){
        if (node.kind === 'MetaProperty' && node.operator === 'ImportKeyword') {
            return true;
        }
    }
    return false;
}
