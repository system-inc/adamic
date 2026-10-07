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
class DumpFrame {
    readonly id: number;
    readonly depth: number;
    readonly property: number;
    constructor(id: number, depth: number, property: number) {
        this.id = id;
        this.depth = depth;
        this.property = property;
    }
}
export function dump(arena: Arena, root: number, depth = 0): string {
    const parts: string[] = [];
    const stack = [new DumpFrame(root, depth, -1)];
    while(stack.length > 0) {
        const frame = stack.pop();
        if(frame === undefined) {
            break;
        }
        const level = frame.depth;
        if(frame.id < 0) {
            parts.push(`${level} null\n`);
            continue;
        }
        const node = arena.node(frame.id);
        if(frame.property < 0) {
            parts.push(
                `${level} ${node.type} ${node.start} ${node.end} ${node.parenthesized ? 1 : 0} ${node.hasContentEnd ? 1 : 0} ${node.contentEnd} ${locStart(arena, frame.id)} ${locEnd(arena, frame.id)} ${ignoredSemicolon(arena, frame.id) ? 1 : 0}\n`,
            );
            for(let index = node.properties.length - 1; index >= 0; index--) {
                stack.push(new DumpFrame(frame.id, level, index));
            }
            continue;
        }
        const property = node.properties[frame.property];
        if(property === undefined) {
            continue;
        }
        parts.push(`${level} .${property.key} ${property.value.kind}`);
        if(property.value.kind === 'node') {
            parts.push('\n');
            stack.push(new DumpFrame(property.value.node, level + 1, -1));
        }
        else if(property.value.kind === 'list') {
            parts.push(` ${property.value.list.length}\n`);
            for(let index = property.value.list.length - 1; index >= 0; index--) {
                stack.push(new DumpFrame(property.value.list[index] ?? -1, level + 1, -1));
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
