// Cohere's byte-based port of Prettier's loc.js, including empty group locations.
import { utf8Length } from 'adamic';
import type { ObjectNode, Tree } from './tree.ts';

function fixWord(node: ObjectNode, index: number): number {
    const value = node.string('value');
    if(value === '-' || value === '--' || !value.startsWith('-')) {
        return index;
    }
    return index - (value[1] === '-' ? 2 : 1);
}
function lineColumnToIndex(position: ObjectNode, text: string): number {
    let index = 0;
    const line = position.number('line');
    let column = position.number('column');
    for(let i = 0; i < line - 1; i++) {
        const found = text.slice(index).indexOf('\n');
        index = found < 0 ? 0 : index + found + 1;
    }
    const prefix = utf8Length(text.slice(0, index));
    let bytes = prefix;
    for(const character of text.slice(index)) {
        if(column <= 0 || character.length > column) {
            break;
        }
        column -= character.length;
        bytes += utf8Length(character);
    }
    return bytes + column;
}
function source(tree: Tree, index: number): ObjectNode {
    return tree.maybe(tree.at(index).object('source'));
}
function raws(tree: Tree, index: number): ObjectNode {
    return tree.maybe(tree.at(index).object('raws'));
}
function start(tree: Tree, index: number, text: string): number {
    const node = tree.at(index);
    const position = tree.maybe(source(tree, index).object('start'));
    if(position.numbers.has('offset')) {
        return position.number('offset');
    }
    if(node.numbers.has('sourceIndex')) {
        return node.type() === 'value-word' ? fixWord(node, node.number('sourceIndex')) : node.number('sourceIndex');
    }
    if(source(tree, index).object('start') >= 0) {
        return lineColumnToIndex(position, text);
    }
    throw new Error('Can not locate node.');
}
function end(tree: Tree, index: number, text: string): number | undefined {
    const node = tree.at(index);
    if(node.type() === 'value-paren' && node.numbers.has('sourceIndex')) {
        return node.number('sourceIndex') + (node.string('value') === ')' ? utf8Length(node.string('value')) : 0);
    }
    const originalSource = node.object('source');
    const object = source(tree, index);
    const position = tree.maybe(object.object('end'));
    if(position.numbers.has('offset')) {
        return position.number('offset');
    }
    if(originalSource >= 0) {
        if(object.object('end') >= 0) {
            const index = lineColumnToIndex(position, text);
            return node.type() === 'value-word' ? fixWord(node, index) : index;
        }
        const children = node.list('nodes');
        if(children.length > 0) {
            return end(tree, children[children.length - 1] ?? 0, text);
        }
        if(node.type() === 'css-atrule' && node.strings.has('name')) {
            const raw = raws(tree, index);
            return start(tree, index, text) + 1 + utf8Length(node.string('name')) + utf8Length(raw.string('afterName')) + utf8Length(raw.string('params'));
        }
    }
    if(node.numbers.has('sourceIndex') && node.strings.has('value')) {
        return node.number('sourceIndex') + utf8Length(node.string('value'));
    }
    return undefined;
}
function leadingColonAndWhitespace(text: string): string {
    const match = /^\s*:?\s*/.exec(text);
    return match === null ? '' : (match[0] ?? '');
}
function valueOffset(tree: Tree, index: number): number {
    const node = tree.at(index);
    const raw = raws(tree, index);
    let result = source(tree, index).number('startOffset');
    if(node.strings.has('prop')) {
        result += utf8Length(node.string('prop'));
    }
    if(node.type() === 'css-atrule' && node.strings.has('name')) {
        result += 1 + utf8Length(node.string('name')) + utf8Length(leadingColonAndWhitespace(raw.string('afterName')));
    }
    if(node.type() !== 'css-atrule' && raw.strings.has('between')) {
        result += utf8Length(raw.string('between'));
    }
    return result;
}
function paramsOffset(tree: Tree, index: number): number {
    const node = tree.at(index);
    let result = source(tree, index).number('startOffset');
    if(node.type() === 'css-atrule' && node.strings.has('name')) {
        result += 1 + utf8Length(node.string('name')) + utf8Length(leadingColonAndWhitespace(raws(tree, index).string('afterName')));
    }
    return result;
}
function selectorText(tree: Tree, parentIndex: number, index: number): string | undefined {
    const raw = raws(tree, index);
    if(raw.strings.has('selector')) {
        return raw.string('selector');
    }
    const parent = tree.at(parentIndex);
    if(parent.object('selector') !== index) {
        return undefined;
    }
    const parentRaw = raws(tree, parentIndex);
    if(parentRaw.strings.has('selector')) {
        return parentRaw.string('selector');
    }
    if(parent.string('customSelector') !== '' && parentRaw.strings.has('params')) {
        // The custom name is bytes long in cohere; names here are valid UTF-8.
        return parentRaw.string('params').slice(parent.string('customSelector').length).trim();
    }
    if(parentRaw.strings.has('params')) {
        return parentRaw.string('params');
    }
    if(parent.type() === 'css-decl' && parentRaw.strings.has('value')) {
        return parentRaw.string('value');
    }
    return undefined;
}
function stringOffset(text: string, search: string): number {
    const index = text.indexOf(search);
    return index < 0 ? 0 : utf8Length(text.slice(0, index));
}
function selectorOffset(tree: Tree, parentIndex: number, index: number, text: string, rootOffset: number): number {
    const node = tree.at(index);
    if(node.numbers.has('sourceIndex') && raws(tree, index).strings.has('selector')) {
        return node.number('sourceIndex') + rootOffset;
    }
    const parent = tree.at(parentIndex);
    const parentRaw = raws(tree, parentIndex);
    if(parentRaw.strings.has('selector')) {
        return source(tree, parentIndex).number('startOffset') + stringOffset(parentRaw.string('selector'), text);
    }
    if(parentRaw.strings.has('params')) {
        return paramsOffset(tree, parentIndex) + stringOffset(parentRaw.string('params'), text);
    }
    if(parent.type() === 'css-decl' && parentRaw.strings.has('value')) {
        return valueOffset(tree, parentIndex) + stringOffset(parentRaw.string('value'), text);
    }
    return rootOffset;
}
function emptyLocFromParent(tree: Tree, index: number, parent: number): void {
    const node = tree.at(index);
    const parentSource = source(tree, parent);
    if(node.type() === '' || node.object('source') >= 0 || !parentSource.numbers.has('startOffset') || !parentSource.numbers.has('endOffset')) {
        return;
    }
    if(!(node.lists.has('nodes') && node.list('nodes').length === 0 || node.lists.has('groups') && node.list('groups').length === 0)) {
        return;
    }
    const object = tree.at(tree.ensure(index, 'source'));
    object.setNumber('startOffset', parentSource.number('startOffset'));
    object.setNumber('endOffset', parentSource.number('startOffset'));
    object.setObject('start', parentSource.object('start'));
    object.setObject('end', parentSource.object('start'));
}
function locFromChildren(tree: Tree, index: number): void {
    const node = tree.at(index);
    if(node.type() === '') {
        return;
    }
    let object = source(tree, index);
    const offsets = object.numbers.has('startOffset') && object.numbers.has('endOffset');
    if(offsets && object.object('start') >= 0 && object.object('end') >= 0) {
        return;
    }
    let found = false;
    let startOffset = 0;
    let endOffset = 0;
    let start = -1;
    let end = -1;
    for(const key of node.keys) {
        if(key === 'source' || key === 'raws' || key === 'spaces') {
            continue;
        }
        for(const child of tree.children(index, key)) {
            const object = source(tree, child);
            if(!object.numbers.has('startOffset') || !object.numbers.has('endOffset')) {
                continue;
            }
            if(!found || object.number('startOffset') < startOffset) {
                startOffset = object.number('startOffset');
                start = object.object('start');
            }
            if(!found || object.number('endOffset') > endOffset) {
                endOffset = object.number('endOffset');
                end = object.object('end');
            }
            found = true;
        }
    }
    if(found) {
        object = tree.at(tree.ensure(index, 'source'));
        if(!offsets) {
            object.setNumber('startOffset', startOffset);
            object.setNumber('endOffset', endOffset);
        }
        if(object.object('start') < 0) {
            object.setObject('start', start);
        }
        if(object.object('end') < 0) {
            object.setObject('end', end);
        }
    }
}
function emptyChildLocs(tree: Tree, index: number): void {
    const node = tree.at(index);
    const object = source(tree, index);
    if(!object.numbers.has('startOffset') || !object.numbers.has('endOffset')) {
        return;
    }
    for(const key of node.keys) {
        if(key !== 'source' && key !== 'raws' && key !== 'spaces') {
            for(const child of tree.children(index, key)) {
                emptyLocFromParent(tree, child, index);
            }
        }
    }
}
export function calculateLoc(tree: Tree, index: number, text: string, rootOffset: number = 0, isRoot: boolean = false): void {
    const node = tree.at(index);
    let object = source(tree, index);
    if(isRoot && node.type() !== '') {
        object = tree.at(tree.ensure(index, 'source'));
        object.setNumber('startOffset', rootOffset);
        object.setNumber('endOffset', rootOffset + utf8Length(text));
    }
    else if(object.numbers.has('startOffset') && object.numbers.has('endOffset')) {
        // A nested custom property may already have been walked.
    }
    else if(node.object('source') >= 0 || node.numbers.has('sourceIndex')) {
        object = tree.at(tree.ensure(index, 'source'));
        const maximum = rootOffset + utf8Length(text);
        object.setNumber('startOffset', Math.min(start(tree, index, text) + rootOffset, maximum));
        const last = end(tree, index, text);
        if(last === undefined) {
            object.setObject('endOffset', -1);
        }
        else {
            object.setNumber('endOffset', Math.min(last + rootOffset, maximum));
        }
    }
    for(const key of node.keys.slice()) {
        if(key === 'source') {
            continue;
        }
        for(const child of tree.children(index, key)) {
            const childNode = tree.at(child);
            const type = childNode.type();
            if(type === 'value-root' || type === 'value-unknown') {
                calculateLoc(tree, child, childNode.string('text') !== '' ? childNode.string('text') : childNode.string('value'), valueOffset(tree, index), true);
            }
            else if(type === 'media-query-list' || node.object('params') === child) {
                calculateLoc(tree, child, raws(tree, index).string('params') !== '' ? raws(tree, index).string('params') : childNode.string('value'), paramsOffset(tree, index), true);
            }
            else if(type.startsWith('selector-')) {
                const nestedText = selectorText(tree, index, child);
                if(nestedText !== undefined) {
                    calculateLoc(tree, child, nestedText, selectorOffset(tree, index, child, nestedText, rootOffset), true);
                }
                else {
                    calculateLoc(tree, child, text, rootOffset);
                }
            }
            else {
                calculateLoc(tree, child, text, rootOffset);
            }
            emptyLocFromParent(tree, child, index);
        }
    }
    locFromChildren(tree, index);
    emptyChildLocs(tree, index);
}

export function setRanges(tree: Tree, index: number): void {
    const node = tree.at(index);
    const object = source(tree, index);
    node.range[0] = object.number('startOffset');
    node.range[1] = object.number('endOffset');
    for(const key of node.keys) {
        if(key !== 'source') {
            for(const child of tree.children(index, key)) {
                setRanges(tree, child);
            }
        }
    }
}
