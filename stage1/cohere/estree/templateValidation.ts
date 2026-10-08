import type { ParseNode } from '../../typescript/parser/nodes.ts';
export function validTemplate(id: number, nodes: readonly ParseNode[], parents: readonly number[]): boolean {
    let parent = parents[id] ?? -1;
    if((nodes[parent]?.kind ?? '') === 'TemplateSpan') {
        parent = parents[parent] ?? -1;
    }
    if((nodes[parent]?.kind ?? '') === 'TemplateExpression') {
        parent = parents[parent] ?? -1;
    }
    if((nodes[parent]?.kind ?? '') !== 'TaggedTemplateExpression') {
        return true;
    }
    const raw = nodes[id]?.raw ?? '';
    for(let index = 0; index + 1 < raw.length; index++) {
        if(raw.slice(index, index + 1) !== '\\') {
            continue;
        }
        const marker = raw.slice(index + 1, index + 2);
        const rest = raw.slice(index + 2);
        const width = marker === 'u' ? (rest.startsWith('{') ? 0 : 4) : marker === 'x' ? 2 : 0;
        if(width > 0) {
            if(rest.length < width) {
                return false;
            }
            for(let digit = 0; digit < width; digit++) {
                if(!'0123456789abcdefABCDEF'.includes(rest.slice(digit, digit + 1))) {
                    return false;
                }
            }
        }
    }
    return true;
}
