import { absent, boolValue, listValue, stringValue } from './values.ts';
export function accessibility(nodes, id) {
    const owner = nodes[id];
    if (owner === undefined) {
        return absent();
    }
    for (const child of owner.children){
        const kind = nodes[child]?.kind ?? '';
        if ([
            'PublicKeyword',
            'PrivateKeyword',
            'ProtectedKeyword'
        ].includes(kind)) {
            return stringValue(kind.slice(0, -7).toLowerCase());
        }
        if (!kind.endsWith('Keyword') && kind !== 'Decorator') {
            break;
        }
    }
    return absent();
}
export function emptyIdentifier(arena, position) {
    const result = arena.newNode('Identifier', position, position);
    const node = arena.node(result);
    node.set('decorators', listValue([]));
    node.set('name', stringValue(''));
    node.set('optional', boolValue(false));
    node.set('typeAnnotation', absent());
    return result;
}
export function dataChildren(nodes, id) {
    const result = [];
    const owner = nodes[id];
    if (owner === undefined) {
        return result;
    }
    for (const child of owner.children){
        if ([
            'ExportKeyword',
            'DefaultKeyword',
            'DeclareKeyword',
            'AsyncKeyword',
            'AbstractKeyword',
            'ConstKeyword',
            'PublicKeyword',
            'PrivateKeyword',
            'ProtectedKeyword',
            'ReadonlyKeyword',
            'StaticKeyword',
            'OverrideKeyword',
            'AccessorKeyword',
            'InKeyword',
            'OutKeyword',
            'Decorator',
            'AsteriskToken'
        ].includes(nodes[child]?.kind ?? '')) {
            continue;
        }
        result.push(child);
    }
    return result;
}
