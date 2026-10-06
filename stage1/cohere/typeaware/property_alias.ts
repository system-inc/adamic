import type { Rules } from './rules.ts';
import { Bindings } from './bindings.ts';
import { Reassign } from './reassign.ts';
import { Diagnostic } from './diagnostic.ts';
import { types } from './facts.ts';
import { Frames, header } from './frames.ts';

export class PropertyAlias {
    readonly rules: Rules;
    readonly bindings: Bindings;
    readonly writes: Reassign;
    constructor(bindings: Bindings) {
        this.rules = bindings.rules; this.bindings = bindings; this.writes = new Reassign(this.rules);
    }
    enclosing(index: number): number {
        let current = this.bindings.parent(index);
        while(current >= 0) {
            if(this.bindings.functionKind(this.rules.parser.node(current).kind)) { return current; }
            current = this.bindings.parent(current);
        }
        return -1;
    }
    descendants(index: number, output: number[]): void {
        output.push(index);
        for(const child of this.rules.parser.node(index).children) { this.descendants(child, output); }
    }
    unwrap(index: number): number {
        let current = index;
        while(current >= 0) {
            const node = this.rules.parser.node(current);
            if(node.kind === 'TypeAssertionExpression') { current = node.children[1] ?? -1; }
            else if(['ParenthesizedExpression', 'NonNullExpression', 'AsExpression', 'SatisfiesExpression'].includes(node.kind)) {
                current = node.children[0] ?? -1;
            }
            else { return current; }
        }
        return -1;
    }
    chain(index: number, question: string): boolean {
        let current = index;
        while(current >= 0) {
            const node = this.rules.parser.node(current);
            if(question === 'call' && node.kind === 'CallExpression') { return true; }
            if(question === 'cast' && ['AsExpression', 'TypeAssertionExpression'].includes(node.kind)) { return true; }
            if(question === 'optional' && ['PropertyAccessExpression', 'ElementAccessExpression', 'CallExpression'].includes(node.kind) &&
                node.children.some((child) => this.rules.parser.node(child).kind === 'QuestionDotToken')) { return true; }
            if(['PropertyAccessExpression', 'ElementAccessExpression', 'ParenthesizedExpression'].includes(node.kind) ||
                (question === 'optional' && node.kind === 'CallExpression') ||
                (question === 'cast' && node.kind === 'NonNullExpression')) { current = node.children[0] ?? -1; }
            else { return false; }
        }
        return false;
    }
    reference(index: number, name: number): boolean {
        const node = this.rules.parser.node(index);
        if(node.kind !== 'Identifier' || index === name || node.text !== this.rules.parser.node(name).text) { return false; }
        const parentIndex = this.bindings.parent(index);
        if(parentIndex < 0) { return true; }
        const parent = this.rules.parser.node(parentIndex);
        if(parent.kind === 'ShorthandPropertyAssignment') { return true; }
        if(parent.kind === 'PropertyAccessExpression' || parent.kind === 'QualifiedName') {
            return parent.children[parent.children.length - 1] !== index;
        }
        if(this.bindings.declaring(index)) { return false; }
        return parent.kind !== 'BindingElement' || parent.children[0] !== index || this.bindings.shape.name(parentIndex) === index;
    }
    hook(index: number, name: number): boolean {
        const node = this.rules.parser.node(index);
        if(node.kind !== 'CallExpression') { return false; }
        const callee = this.rules.parser.node(node.children[0] ?? -1);
        let hook = '';
        if(callee.kind === 'Identifier') { hook = callee.text; }
        else if(callee.kind === 'PropertyAccessExpression' &&
            this.rules.parser.node(callee.children[0] ?? -1).kind === 'Identifier' &&
            this.rules.parser.node(callee.children[0] ?? -1).text === 'React') {
            hook = this.rules.parser.node(callee.children[callee.children.length - 1] ?? -1).text;
        }
        if(hook !== 'use' && !(hook.startsWith('use') && hook.charCodeAt(3) >= 65 && hook.charCodeAt(3) <= 90)) { return false; }
        let args: number[] = [];
        if(node.list > 0) { args = node.children.slice(node.children.length - node.list); }
        for(const argument of args.slice(1)) {
            if(this.rules.parser.node(argument).kind !== 'ArrayLiteralExpression') { continue; }
            const descendants: number[] = []; this.descendants(argument, descendants);
            if(descendants.some((child) => this.reference(child, name))) { return true; }
        }
        return false;
    }
    target(index: number): number {
        const node = this.rules.parser.node(index);
        if(node.kind === 'BinaryExpression' && this.writes.assignment(this.rules.parser.node(node.children[1] ?? -1).kind)) {
            return node.children[0] ?? -1;
        }
        if(node.kind === 'DeleteExpression' || (['PrefixUnaryExpression', 'PostfixUnaryExpression'].includes(node.kind) &&
            ['PlusPlusToken', 'MinusMinusToken'].includes(node.operator))) { return node.children[0] ?? -1; }
        return -1;
    }
    same(first: number, second: number): boolean {
        const a = this.unwrap(first); const b = this.unwrap(second);
        if(a < 0 || b < 0) { return false; }
        const left = this.rules.parser.node(a); const right = this.rules.parser.node(b);
        if(left.kind !== right.kind) { return false; }
        if(left.kind === 'Identifier') { return left.text === right.text; }
        if(left.kind === 'ThisKeyword') { return true; }
        return left.kind === 'PropertyAccessExpression' &&
            this.rules.parser.node(left.children[left.children.length - 1] ?? -1).text ===
            this.rules.parser.node(right.children[right.children.length - 1] ?? -1).text &&
            this.same(left.children[0] ?? -1, right.children[0] ?? -1);
    }
    links(index: number): number[] {
        const result: number[] = [];
        let current = this.unwrap(index);
        while(current >= 0 && this.rules.parser.node(current).kind === 'PropertyAccessExpression') {
            result.push(current); current = this.unwrap(this.rules.parser.node(current).children[0] ?? -1);
        }
        return result;
    }
    getter(index: number): boolean {
        const node = this.rules.parser.node(index);
        const name = node.children[node.children.length - 1] ?? -1;
        const frames = new Frames(this.rules.ask(name, 'binding-declarations')); header(frames, 'binding-declarations');
        if(!frames.yes()) { frames.end(); return false; }
        const flags = frames.natural(); let getter = (flags & 32768) !== 0;
        const count = frames.natural();
        for(let at = 0; at < count; at++) {
            frames.field(); const kind = frames.field(); frames.natural(); frames.natural();
            if(kind === 'GetAccessor') { getter = true; }
        }
        frames.end(); return getter;
    }
    narrowed(index: number): boolean {
        const node = this.rules.parser.node(index);
        const declared = types(this.rules.ask(node.children[node.children.length - 1] ?? -1, 'symbol-shape'), 'symbol-shape');
        return declared.present && declared.root().id !== types(this.rules.ask(index, 'raw-shape'), 'raw-shape').root().id;
    }
    run(): void {
        for(let index = 0; index < this.rules.parser.nodes.length; index++) {
            if(this.bindings.parent(index) < 0 || this.rules.parser.node(index).kind !== 'VariableDeclaration') { continue; }
            const node = this.rules.parser.node(index);
            const name = this.bindings.shape.name(index);
            if(name < 0 || this.rules.parser.node(name).kind !== 'Identifier') { continue; }
            const initializer = this.bindings.shape.initializer(index);
            if(initializer < 0) { continue; }
            const access = this.rules.parser.node(initializer);
            if(access.kind !== 'PropertyAccessExpression' || this.chain(initializer, 'optional') ||
                this.rules.parser.node(access.children[access.children.length - 1] ?? -1).text !== this.rules.parser.node(name).text ||
                this.chain(access.children[0] ?? -1, 'call')) { continue; }
            const enclosing = this.enclosing(index); if(enclosing < 0) { continue; }
            const all: number[] = []; this.descendants(enclosing, all);
            if(all.some((child) => this.hook(child, name)) ||
                all.some((child) => this.reference(child, name) && this.writes.writes(child))) { continue; }
            let lastRead = -1;
            for(const child of all) {
                if(this.reference(child, name) && !this.writes.writes(child)) { lastRead = Math.max(lastRead, this.rules.parser.node(child).end); }
            }
            const links: number[] = []; let current = this.unwrap(initializer);
            while(current >= 0) {
                links.push(current);
                if(this.rules.parser.node(current).kind !== 'PropertyAccessExpression') { break; }
                current = this.unwrap(this.rules.parser.node(current).children[0] ?? -1);
            }
            if(lastRead >= 0 && all.some((child) => {
                const target = this.target(child); const expression = this.rules.parser.node(child);
                return target >= 0 && expression.pos >= node.end && expression.end <= lastRead && links.some((link) => this.same(target, link));
            })) { continue; }
            if(this.chain(access.children[0] ?? -1, 'cast')) { continue; }
            const propertyLinks = this.links(initializer);
            if(propertyLinks.some((link) => this.getter(link))) { continue; }
            const closureReads = all.filter((child) => this.reference(child, name) && !this.writes.writes(child) && this.enclosing(child) !== enclosing);
            if(closureReads.length > 0 && propertyLinks.some((link) => this.narrowed(link))) { continue; }
            const declared = types(this.rules.ask(name, 'symbol-shape'), 'symbol-shape');
            if(declared.present && closureReads.some((child) => types(this.rules.ask(child, 'raw-shape'), 'raw-shape').root().id !== declared.root().id)) { continue; }
            const annotation = node.children.filter((child) => child !== name && child !== initializer &&
                !['ExclamationToken', 'QuestionToken'].includes(this.rules.parser.node(child).kind));
            if(annotation.length > 0 && types(this.rules.ask(annotation[0] ?? -1, 'annotation-shape'), 'annotation-shape').root().id !==
                types(this.rules.ask(initializer, 'raw-shape'), 'raw-shape').root().id) { continue; }
            const local = this.rules.parser.node(name).text;
            const object = this.rules.text(this.rules.parser.node(access.children[0] ?? -1));
            this.rules.findings.push(new Diagnostic('nexus/consistency-no-property-alias', 'noPropertyAlias',
                `\`${local}\` is a pure alias for \`${object}.${local}\`. Reach for the property directly, so the value keeps the object it came from and a naked local still means this scope created it. Allowed only when the local is read inside a hook dependency array.`,
                this.rules.byte(this.rules.start(node)), this.rules.byte(node.end), ''));
        }
    }
}
