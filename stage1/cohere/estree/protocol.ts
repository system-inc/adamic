// Canonical traversal includes every own field, ordered keys and printer-visible locations.
import type { Arena } from './arena.ts';
import { ignoredSemicolon, locStart, locEnd } from './location.ts';
export function written(text: string): string {
    let result = '';
    for(let index = 0; index < text.length; index++) {
        const unit = text.charCodeAt(index);
        if(
            (unit >= 0xd800 &&
                unit <= 0xdbff &&
                !(text.charCodeAt(index + 1) >= 0xdc00 && text.charCodeAt(index + 1) <= 0xdfff)) ||
            (unit >= 0xdc00 &&
                unit <= 0xdfff &&
                !(text.charCodeAt(index - 1) >= 0xd800 && text.charCodeAt(index - 1) <= 0xdbff))
        ) {
            result += '\\ufffd\\ufffd\\ufffd';
            continue;
        }
        result +=
            unit >= 32 && unit <= 126 && unit !== 92
                ? text.slice(index, index + 1)
                : `\\u${unit.toString(16).padStart(4, '0')}`;
    }
    return result;
}
export function dump(arena: Arena, id: number, depth = 0): string {
    if(id < 0) {
        return `${depth} null\n`;
    }
    const node = arena.node(id);
    const parts: string[] = [
        `${depth} ${node.type} ${node.start} ${node.end} ${node.parenthesized ? 1 : 0} ${node.hasContentEnd ? 1 : 0} ${node.contentEnd} ${locStart(arena, id)} ${locEnd(arena, id)} ${ignoredSemicolon(arena, id) ? 1 : 0}\n`,
    ];
    for(const property of node.properties) {
        parts.push(`${depth} .${property.key} ${property.value.kind}`);
        if(property.value.kind === 'node') {
            parts.push('\n');
            parts.push(dump(arena, property.value.node, depth + 1));
        }
        else if(property.value.kind === 'list') {
            parts.push(` ${property.value.list.length}\n`);
            for(const child of property.value.list) {
                parts.push(dump(arena, child, depth + 1));
            }
        }
        else {
            parts.push(
                property.value.kind === 'string'
                    ? ` ${written(property.value.text)}`
                    : property.value.kind === 'bool' || property.value.kind === 'number'
                      ? ` ${property.value.number}`
                      : property.value.kind === 'regex'
                        ? ` ${written(property.value.text)}\t${written(property.value.cooked)}`
                        : property.value.kind === 'template'
                          ? ` ${written(property.value.text)}\t${property.value.number === 1 ? written(property.value.cooked) : '<null>'}`
                          : '',
            );
            parts.push('\n');
        }
    }
    return parts.join('');
}
