import { panic, utf8Length } from 'adamic';
import { Tree } from './tree.ts';
import type { Input, Position } from './input.ts';
import type { CssNode } from './nodes.ts';
import type { SelectorNode } from '../selector/nodes.ts';
import type { MediaNode } from '../mediaquery/nodes.ts';
import { keysOf, hasNodes, type ValueTree } from '../values/nodes.ts';

function position(tree: Tree, pos: Position, input: Input): number {
    const result = tree.make();
    const node = tree.at(result);
    node.setNumber('column', pos.column);
    node.setNumber('line', pos.line);
    node.setNumber('offset', pos.offset);
    return result;
}
export function fromCSS(tree: Tree, nodes: readonly CssNode[], index: number, input: Input): number {
    const original = nodes[index] ?? panic('missing raw CSS node');
    const result = tree.make(`css-${original.type}`);
    const node = tree.at(result);
    const rawIndex = tree.make();
    const raws = tree.at(rawIndex);
    for(const [key, value] of original.rawStrings) {
        raws.setString(key, value);
    }
    for(const [key, value] of original.rawBooleans) {
        raws.setBoolean(key, value);
    }
    for(const [key, value] of original.rawValues) {
        const index = tree.make();
        const object = tree.at(index);
        object.setString('raw', value.raw);
        object.setString('value', value.value);
        if(value.scss !== undefined) {
            object.setString('scss', value.scss);
        }
        raws.setObject(key, index);
    }
    node.setObject('raws', rawIndex);
    for(const key of original.keys) {
        if(key === 'nodes') {
            const children: number[] = [];
            for(const child of original.children) {
                children.push(fromCSS(tree, nodes, child, input));
            }
            node.setList(key, children);
        }
        else if(key === 'source') {
            const sourceIndex = tree.make();
            const source = tree.at(sourceIndex);
            source.setObject('start', position(tree, original.start ?? panic('missing CSS source'), input));
            if(original.end !== undefined) {
                source.setObject('end', position(tree, original.end, input));
            }
            node.setObject(key, sourceIndex);
        }
        else if(original.strings.has(key)) {
            node.setString(key, original.get(key));
        }
        else if(original.booleans.has(key)) {
            node.setBoolean(key, original.booleans.get(key) === true);
        }
    }
    return result;
}
function lineColumn(tree: Tree, line: number | undefined, column: number | undefined): number {
    const index = tree.make();
    const node = tree.at(index);
    if(column === undefined) {
        node.setObject('column', -1);
    }
    else {
        node.setNumber('column', column);
    }
    if(line === undefined) {
        node.setObject('line', -1);
    }
    else {
        node.setNumber('line', line);
    }
    return index;
}
export function fromSelector(tree: Tree, nodes: readonly SelectorNode[], index: number): number {
    const original = nodes[index] ?? panic('missing nested selector');
    const result = tree.make(`selector-${original.type}`);
    const node = tree.at(result);
    const spacesIndex = tree.make();
    const spaces = tree.at(spacesIndex);
    spaces.setString('before', original.before);
    spaces.setString('after', original.after);
    node.setObject('spaces', spacesIndex);
    if(original.type === 'root' || original.type === 'selector' || original.type === 'pseudo') {
        const children: number[] = [];
        for(const child of original.children) {
            children.push(fromSelector(tree, nodes, child));
        }
        node.setList('nodes', children);
    }
    if(original.type !== 'root' && original.type !== 'selector') {
        if(original.value === undefined) {
            node.setObject('value', -1);
        }
        else {
            node.setString('value', original.value);
        }
        const source = original.source ?? panic('missing selector source');
        const sourceIndex = tree.make();
        const object = tree.at(sourceIndex);
        object.setObject('start', lineColumn(tree, source.startLine, source.startColumn));
        object.setObject('end', lineColumn(tree, source.endLine, source.endColumn));
        node.setObject('source', sourceIndex);
        if(original.sourceIndex === undefined) {
            node.setObject('sourceIndex', -1);
        }
        else {
            node.setNumber('sourceIndex', original.sourceIndex);
        }
    }
    if(original.namespace !== undefined) {
        node.setString('namespace', original.namespace);
    }
    if(original.namespaceTrue) {
        node.setBoolean('namespace', true);
    }
    if(original.type === 'attribute') {
        node.setString('attribute', original.attribute ?? '');
        if(original.operator === undefined) {
            node.setObject('operator', -1);
        }
        else {
            node.setString('operator', original.operator);
        }
        if(original.insensitive) {
            node.setBoolean('insensitive', true);
        }
        if(original.quotedPresent) {
            node.setBoolean('quoted', original.quoted);
        }
        const rawsIndex = tree.make();
        const raws = tree.at(rawsIndex);
        if(original.rawInsensitive !== undefined) {
            raws.setString('insensitive', original.rawInsensitive);
        }
        if(original.unquoted !== undefined) {
            raws.setString('unquoted', original.unquoted);
        }
        node.setObject('raws', rawsIndex);
    }
    if(original.trailingComma) {
        node.setBoolean('trailingComma', true);
    }
    return result;
}
export function fromValue(tree: Tree, original: ValueTree, input: Input): number {
    const value = original.node;
    const result = tree.make(value.type);
    const node = tree.at(result);
    for(const key of keysOf(value.shape)) {
        switch(key) {
            case 'raws': {
                const index = tree.make();
                const raws = tree.at(index);
                raws.setString('before', value.raws.before);
                raws.setString('after', value.raws.after);
                if(value.raws.quote !== undefined) {
                    raws.setString('quote', value.raws.quote);
                }
                node.setObject(key, index);
                break;
            }
            case 'source': {
                const source = value.source ?? panic('missing value source');
                const index = tree.make();
                const object = tree.at(index);
                object.setObject('start', lineColumn(tree, source.start.line, source.start.column));
                object.setObject('end', lineColumn(tree, source.end.line, source.end.column));
                node.setObject(key, index);
                break;
            }
            case 'sourceIndex': {
                const index = value.sourceIndex;
                const low = input.css.charCodeAt(index);
                const high = input.css.charCodeAt(index - 1);
                node.setNumber(
                    key,
                    input.byteOffset(index) -
                        (low >= 0xdc00 && low <= 0xdfff && high >= 0xd800 && high <= 0xdbff ? 2 : 0),
                );
                break;
            }
            case 'value':
                node.setString(key, value.value);
                break;
            case 'unit':
                node.setString(key, value.unit);
                break;
            case 'unbalanced':
                node.setNumber(key, value.unbalanced);
                break;
            case 'isHex':
                node.setBoolean(key, value.isHex);
                break;
            case 'isColor':
                node.setBoolean(key, value.isColor);
                break;
            case 'inline':
            case 'quoted':
                node.setBoolean(key, value.flag);
                break;
            case 'parenType':
                node.setString(key, '');
                break;
        }
    }
    if(hasNodes(value.shape)) {
        const children: number[] = [];
        for(const child of original.nodes) {
            children.push(fromValue(tree, child, input));
        }
        node.setList('nodes', children);
    }
    return result;
}
export function fromMedia(tree: Tree, original: MediaNode, input: Input): number {
    const result = tree.make(
        original.type === '' && original.value === ''
            ? ''
            : original.type.startsWith('media-')
              ? original.type
              : `media-${original.type === '' ? 'unknown' : original.type}`,
        true,
    );
    const node = tree.at(result);
    node.setString('after', original.after);
    node.setString('before', original.before);
    node.setString('value', original.value);
    node.setNumber('sourceIndex', input.byteOffset(original.sourceIndex));
    const children = original.nodes;
    if(children !== undefined) {
        const nodes: number[] = [];
        for(const child of children) {
            nodes.push(fromMedia(tree, child, input));
        }
        node.setList('nodes', nodes);
    }
    return result;
}
// Byte slices used by Go's composition. A cut inside a code point is represented
// by replacement characters when rendered, just as Go's JSON encoder does.
export function byteSlice(text: string, start: number, end: number = utf8Length(text)): string {
    const length = utf8Length(text);
    start = start < 0 ? Math.max(0, length + start) : Math.min(start, length);
    end = end < 0 ? Math.max(0, length + end) : Math.min(end, length);
    if(end <= start) return '';
    if(length === text.length) return text.slice(start, end);
    let offset = 0;
    let result = '';
    for(const character of text) {
        if(offset >= end) break;
        const next = offset + utf8Length(character);
        if(offset >= start && next <= end) {
            result += character;
        }
        else if(next > start && offset < end) {
            result += repeatText('�', Math.min(next, end) - Math.max(offset, start));
        }
        offset = next;
    }
    return result;
}

export function repeatText(text: string, count: number): string {
    let result = '';
    for(let i = 0; i < count; i++) {
        result += text;
    }
    return result;
}
