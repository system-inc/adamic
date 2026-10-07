// CSS printer predicates and path reads from cohere/utilities.go and print_helpers.go.
import { Tree, ObjectNode } from './tree.ts';
import { byteSlice } from './convert.ts';
import { utf8Length } from 'adamic';
export class Node {
    readonly tree: Tree;
    readonly index: number;
    constructor(tree: Tree, index: number) { this.tree = tree; this.index = index; }
    data(): ObjectNode { return this.tree.maybe(this.index); }
    type(): string { return this.data().type(); }
    string(key: string): string { return this.data().string(key); }
    has(key: string): boolean { return this.data().keys.includes(key); }
    child(key: string): Node { return new Node(this.tree, this.data().object(key)); }
    list(key: string): number[] { return this.data().list(key); }
    truth(key: string): boolean {
        const n = this.data();
        return n.booleans.get(key) === true || n.string(key) !== '' || n.objects.has(key) || n.lists.has(key) || n.number(key) !== 0;
    }
    raw(key: string): string { return this.child('raws').string(key); }
    start(): number { return this.child('source').data().number('startOffset'); }
    end(): number { return this.child('source').data().number('endOffset'); }
    line(): number { return this.child('source').child('start').data().number('line'); }
}
export class Path {
    readonly tree: Tree;
    readonly nodes: number[];
    readonly keys: string[];
    readonly positions: number[];
    constructor(tree: Tree, nodes: number[], keys: string[], positions: number[]) { this.tree = tree; this.nodes = nodes; this.keys = keys; this.positions = positions; }
    node(): Node { return this.up(0); }
    up(depth: number): Node { return new Node(this.tree, this.nodes[this.nodes.length - depth - 1] ?? -1); }
    key(depth: number = 0): string { return this.keys[this.keys.length - depth - 1] ?? ''; }
    child(key: string, position: number = -1): Path {
        const n = this.node(); const index = position < 0 ? n.child(key).index : n.list(key)[position] ?? -1;
        return new Path(this.tree, [...this.nodes, index], [...this.keys, key], [...this.positions, position]);
    }
    parent(): Path { return new Path(this.tree, this.nodes.slice(0, -1), this.keys.slice(0, -1), this.positions.slice(0, -1)); }
    ancestor(type: string): Node {
        for(let i = this.nodes.length - 2; i >= 0; i--) { const n = new Node(this.tree, this.nodes[i] ?? -1); if(n.type() === type) return n; }
        return new Node(this.tree, -1);
    }
    sibling(delta: number): Node {
        const position = this.positions[this.positions.length - 1] ?? -1;
        return new Node(this.tree, position < 0 ? -1 : this.up(1).list(this.key())[position + delta] ?? -1);
    }
    last(): boolean { return (this.positions[this.positions.length - 1] ?? -1) === this.up(1).list(this.key()).length - 1; }
    insideRule(names: readonly string[]): boolean { return names.includes(this.ancestor('css-atrule').string('name').toLowerCase()); }
    insideFunc(name: string): boolean { return this.ancestor('value-func').string('value').toLowerCase() === name; }
    icss(): boolean {
        for(let i = this.nodes.length - 2; i >= 0; i--) { const n = new Node(this.tree, this.nodes[i] ?? -1); if(n.type() === 'css-rule' && (n.raw('selector').startsWith(':import') || n.raw('selector').startsWith(':export'))) return true; }
        return false;
    }
}
export function lower(value: string): string { return value.replace(/Σ/g, 'σ').toLowerCase(); }
export function maybeLower(value: string): string {
    return value.includes('$') || value.includes('@') || value.includes('#') || value.startsWith('%') || value.startsWith('--') || value.startsWith(':--') || (value.includes('(') && value.includes(')')) ? value : lower(value);
}
export function inlineLast(text: string): boolean { return text.slice(Math.max(text.lastIndexOf('\n'), text.lastIndexOf('\r')) + 1).includes('//'); }
export function emptyBefore(n: Node): boolean { return n.child('raws').data().strings.has('before') && n.raw('before') === ''; }
export function control(n: Node, scss: boolean): boolean { return scss && n.type() === 'css-atrule' && ['if','else','for','each','while'].includes(n.string('name')); }
export function placeholder(n: Node): boolean { return n.type() === 'value-atword' && n.string('value').startsWith('prettier-placeholder-'); }
export function operator(n: Node, values: string = '*/+-%'): boolean { return n.type() === 'value-operator' && values.includes(n.string('value')) && n.string('value') !== ''; }
export function word(n: Node): boolean { return n.type() === 'value-word' || n.type() === 'value-atword'; }
export function valueWord(n: Node, value: string): boolean { return n.type() === 'value-word' && n.string('value') === value; }
export function inlineValue(n: Node): boolean { return n.type() === 'value-comment' && n.truth('inline'); }
export function pair(n: Node): boolean { return n.type() === 'value-comma_group' && n.list('groups').length > 1 && new Node(n.tree, n.list('groups')[1] ?? -1).type() === 'value-colon'; }
export function parenPair(n: Node): boolean { return n.type() === 'value-paren_group' && pair(new Node(n.tree, n.list('groups')[0] ?? -1)); }
export function parens(n: Node): boolean { return n.type() === 'value-paren_group' && n.child('open').string('value') === '(' && n.child('close').string('value') === ')'; }
export function mapItem(path: Path, scss: boolean): boolean {
    if(!scss) return false;
    const n = path.node(); const groups = n.list('groups'); const parent = path.up(1); const grand = path.up(2);
    if(groups.length === 0) return false;
    if(n.type() === 'value-paren_group' && n.child('open').index >= 0 && n.child('close').index >= 0 && groups.length === 1 && new Node(n.tree, groups[0] ?? -1).type() !== 'value-comma_group') return false;
    if(parent.type() === 'value-func' && parent.string('value') === 'if') return false;
    if(!parenPair(n) && !parenPair(grand)) return false;
    if(path.ancestor('css-decl').string('prop').startsWith('$')) return true;
    if(parenPair(grand)) return !parent.list('groups').some(i => operator(new Node(n.tree, i)));
    return grand.type() === 'value-func';
}
export function breakList(path: Path): boolean {
    const n = path.node();
    return n.type() === 'value-paren_group' && n.child('open').index < 0 && n.list('groups').some(i => new Node(n.tree, i).type() === 'value-comma_group') &&
        path.up(1).type() === 'value-value' && path.key() === 'group' && path.up(2).type() === 'value-root' && path.key(1) === 'group' && path.key(2) === 'value' &&
        ((path.up(3).type() === 'css-decl' && !path.up(3).string('prop').startsWith('--')) || (path.up(3).type() === 'css-atrule' && path.up(3).truth('variable')));
}
export function precedeSoft(path: Path): boolean {
    return path.node().type() === 'value-paren_group' && path.node().child('open').index < 0 && path.up(1).type() === 'value-value' && path.up(2).type() === 'value-root' && path.up(3).type() === 'css-decl' && path.key() === 'group' && path.key(1) === 'group' && path.key(2) === 'value';
}
export function hasNewline(text: string, offset: number, backwards: boolean): boolean {
    const before = byteSlice(text, 0, offset); let index = backwards ? before.length - 1 : before.length;
    while(index >= 0 && index < text.length && (text[index] === ' ' || text[index] === '\t')) index += backwards ? -1 : 1;
    return /[\n\r\u2028\u2029]/.test(text.slice(index, index + 1));
}
export function nextEmpty(text: string, offset: number): boolean {
    let index = byteSlice(text, 0, offset).length; let old = -1;
    while(old !== index) {
        old = index;
        while(index < text.length && ',; \t'.includes(text[index] ?? '')) index++;
        if(text.slice(index, index + 2) === '/*') { const end = text.indexOf('*/', index + 2); if(end >= 0) index = end + 2; }
        while(index < text.length && ' \t'.includes(text[index] ?? '')) index++;
    }
    if(text.slice(index, index + 2) === '//') { while(index < text.length && text[index] !== '\n' && text[index] !== '\r') index++; }
    if(text.slice(index, index + 2) === '\r\n') index += 2;
    else if(/[\n\r\u2028\u2029]/.test(text.slice(index, index + 1))) index++;
    else return false;
    return hasNewline(text, utf8Length(text.slice(0, index)), false);
}
