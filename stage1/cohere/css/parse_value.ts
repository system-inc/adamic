// Prettier's value grouping over the existing postcss-values-parser slice.
import { panic } from 'adamic';
import { parse as parseValues } from '../values/values.ts';
import { parse as parseSelectors } from '../selector/parser.ts';
import { Input } from './input.ts';
import { fromValue, fromSelector, byteSlice } from './convert.ts';
import type { Tree } from './tree.ts';

export function parseSelector(tree: Tree, text: string): number {
    const comments = /\/[/*]/.test(text.replace(/"[^"]+"|'[^']+'/g, ''));
    if(!comments) {
        const parsed = parseSelectors(text);
        if(parsed.kind === 'Parsed') {
            return fromSelector(tree, parsed.nodes, 0);
        }
    }
    const index = tree.make('selector-unknown');
    tree.at(index).setString('value', comments ? text.trim() : text);
    return index;
}
function parenGroup(tree: Tree, open = -1): number {
    const index = tree.make('paren_group');
    const node = tree.at(index);
    node.setObject('open', open);
    node.setObject('close', -1);
    node.setList('groups', []);
    return index;
}
function commaGroup(tree: Tree): number {
    const index = tree.make('comma_group');
    tree.at(index).setList('groups', []);
    return index;
}
function pushGroup(tree: Tree, group: number, child: number): void {
    tree.at(group).list('groups').push(child);
}
function flatten(tree: Tree, index: number): number {
    const node = tree.at(index);
    const groups = node.list('groups');
    if(
        (node.type() === 'comma_group' ||
            (node.type() === 'paren_group' && node.object('open') < 0 && node.object('close') < 0)) &&
        groups.length === 1
    ) {
        return flatten(tree, groups[0] ?? panic('missing group'));
    }
    if(node.type() === 'comma_group' || node.type() === 'paren_group') {
        const children: number[] = [];
        for(const child of groups) {
            children.push(flatten(tree, child));
        }
        node.setList('groups', children);
    }
    return index;
}
function groupNodes(tree: Tree, index: number, text: string, scss: boolean): number {
    let paren = parenGroup(tree);
    const root = paren;
    const parens: number[] = [paren];
    let comma = commaGroup(tree);
    const commas: number[] = [comma];
    const nodes = tree.at(index).list('nodes');
    for(let nodePosition = 0; nodePosition < nodes.length; nodePosition++) {
        const child = nodes[nodePosition] ?? panic('missing value node');
        const node = tree.at(child);
        if(scss && node.type() === 'number' && node.string('unit') === '..' && node.string('value').endsWith('.')) {
            node.setString('value', node.string('value').slice(0, -1));
            node.setString('unit', '...');
        }
        if(node.type() === 'func' && (node.string('value') === 'selector' || node.string('value') === 'url')) {
            const group = tree.at(node.object('group'));
            const open = tree.maybe(group.object('open')).number('sourceIndex');
            const close = tree.maybe(group.object('close')).number('sourceIndex');
            if(node.string('value') === 'selector') {
                const selectorText = byteSlice(text, open + 1, close);
                const selector = parseSelector(tree, selectorText);
                tree.at(selector).setNumber('sourceIndex', open + 1);
                tree.at(tree.ensure(selector, 'raws')).setString('selector', selectorText);
                group.setList('groups', [selector]);
            }
            else {
                const groups: number[] = [];
                for(const memberIndex of group.list('groups')) {
                    const groupNode = tree.at(memberIndex);
                    if(groupNode.type() === 'comma_group') {
                        for(const groupChild of groupNode.list('groups')) {
                            groups.push(groupChild);
                        }
                    }
                    else {
                        groups.push(memberIndex);
                    }
                }
                let interpolation = false;
                let hasString = false;
                for(let groupPosition = 0; groupPosition < groups.length; groupPosition++) {
                    const current = tree.at(groups[groupPosition] ?? panic('missing url argument'));
                    if(
                        current.type() === 'string' ||
                        (current.type() === 'func' && !current.string('value').endsWith('\\'))
                    ) {
                        hasString = true;
                    }
                    if(groupPosition > 0 && current.type() === 'word' && current.string('value') === '{') {
                        const previous = tree.at(groups[groupPosition - 1] ?? panic('missing url previous argument'));
                        if(previous.type() === 'word' && previous.string('value').endsWith('#')) {
                            interpolation = true;
                        }
                    }
                }
                const first = groups.length === 0 ? -1 : (groups[0] ?? panic('missing first url argument'));
                const variable =
                    scss && tree.maybe(first).type() === 'word' && tree.maybe(first).string('value').startsWith('$');
                if(interpolation || (!hasString && !variable)) {
                    group.setList('groups', [tree.literal(byteSlice(text, open + 1, close).trim())]);
                }
            }
        }
        if(node.type() === 'paren' && node.string('value') === '(') {
            paren = parenGroup(tree, child);
            parens.push(paren);
            comma = commaGroup(tree);
            commas.push(comma);
        }
        else if(node.type() === 'paren' && node.string('value') === ')') {
            if(tree.at(comma).list('groups').length > 0) {
                pushGroup(tree, paren, comma);
            }
            tree.at(paren).setObject('close', child);
            if(commas.length === 1) {
                throw new Error('Unbalanced parenthesis');
            }
            commas.pop();
            comma = commas[commas.length - 1] ?? panic('missing enclosing comma');
            pushGroup(tree, comma, paren);
            parens.pop();
            paren = parens[parens.length - 1] ?? panic('missing enclosing paren');
        }
        else if(node.type() === 'comma') {
            if(
                nodePosition === nodes.length - 3 &&
                tree.at(nodes[nodePosition + 1] ?? panic('missing trailing comment')).type() === 'comment'
            ) {
                const closing = tree.at(nodes[nodePosition + 2] ?? panic('missing trailing paren'));
                if(closing.type() === 'paren' && closing.string('value') === ')') {
                    continue;
                }
            }
            pushGroup(tree, paren, comma);
            comma = commaGroup(tree);
            commas[commas.length - 1] = comma;
        }
        else {
            pushGroup(tree, comma, child);
        }
    }
    if(tree.at(comma).list('groups').length > 0) {
        pushGroup(tree, paren, comma);
    }
    return root;
}
function nestedValue(tree: Tree, index: number, text: string, scss: boolean): void {
    const node = tree.at(index);
    const keys = node.keys.slice();
    for(const key of keys) {
        for(const child of tree.children(index, key)) {
            nestedValue(tree, child, text, scss);
        }
        if(key === 'nodes') {
            if(!(node.type() === 'atword' && node.list(key).length === 0)) {
                node.setObject('group', flatten(tree, groupNodes(tree, index, text, scss)));
            }
            node.remove(key);
        }
    }
}
function prefix(tree: Tree, index: number): void {
    const node = tree.at(index);
    for(const key of node.keys) {
        for(const child of tree.children(index, key)) {
            prefix(tree, child);
        }
    }
    if(node.type() !== '' && !/^selector-/.test(node.type()) && !node.type().startsWith('value-')) {
        node.setString('type', `value-${node.type()}`);
    }
}
export function parseValue(tree: Tree, value: string, scss: boolean): number {
    const parsed = parseValues(value, true);
    if(parsed.kind === 'Error') {
        const unknown = tree.make('value-unknown');
        tree.at(unknown).setString('value', value);
        return unknown;
    }
    const index = fromValue(tree, parsed.root, new Input(value, false));
    tree.at(index).setString('text', value);
    nestedValue(tree, index, value, scss);
    prefix(tree, index);
    return index;
}
