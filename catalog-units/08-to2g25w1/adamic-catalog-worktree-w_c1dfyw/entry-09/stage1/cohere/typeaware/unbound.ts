import type { Rules } from './rules.ts';
import type { Shadow } from './shadow.ts';
import type { Types } from './types.ts';
import type { TypeFact } from './type_fact.ts';
import { types } from './facts.ts';
import { Frames, header } from './frames.ts';
import { boundMembers } from './bound.ts';

export class Unbound {
    readonly rules: Rules;
    readonly shadow: Shadow;
    constructor(rules: Rules, shadow: Shadow) {
        this.rules = rules;
        this.shadow = shadow;
    }
    safe(index: number): boolean {
        const parentIndex = this.rules.parents[index] ?? -1;
        if(parentIndex < 0) {
            return false;
        }
        const parent = this.rules.parser.node(parentIndex);
        if(
            [
                'IfStatement',
                'ForStatement',
                'SwitchStatement',
                'WhileStatement',
                'PropertyAccessExpression',
                'ElementAccessExpression',
                'PostfixUnaryExpression',
                'DeleteExpression',
                'TypeOfExpression',
                'VoidExpression',
            ].includes(parent.kind)
        ) {
            return true;
        }
        if(['CallExpression', 'ConditionalExpression', 'TaggedTemplateExpression'].includes(parent.kind)) {
            return parent.children[0] === index;
        }
        if(parent.kind === 'PrefixUnaryExpression') {
            return ['ExclamationToken', 'PlusPlusToken', 'MinusMinusToken'].includes(parent.operator);
        }
        if(parent.kind === 'BinaryExpression') {
            const operator = this.rules.parser.node(parent.children[1] ?? -1).kind;
            if(
                [
                    'ExclamationEqualsToken',
                    'ExclamationEqualsEqualsToken',
                    'EqualsEqualsToken',
                    'EqualsEqualsEqualsToken',
                    'InstanceOfKeyword',
                ].includes(operator)
            ) {
                return true;
            }
            if(operator === 'EqualsToken') {
                if(parent.children[0] === index) {
                    return true;
                }
                const member = this.rules.parser.node(index);
                const left = this.rules.parser.node(parent.children[0] ?? -1);
                return (
                    member.kind === 'PropertyAccessExpression' &&
                    this.rules.parser.node(member.children[0] ?? -1).kind === 'SuperKeyword' &&
                    left.kind === 'PropertyAccessExpression' &&
                    this.rules.parser.node(left.children[0] ?? -1).kind === 'ThisKeyword'
                );
            }
            if(operator === 'AmpersandAmpersandToken' && parent.children[0] === index) {
                return true;
            }
            if(['AmpersandAmpersandToken', 'BarBarToken', 'QuestionQuestionToken'].includes(operator)) {
                return this.safe(parentIndex);
            }
        }
        if(
            [
                'NonNullExpression',
                'AsExpression',
                'TypeAssertionExpression',
                'SatisfiesExpression',
                'ParenthesizedExpression',
            ].includes(parent.kind)
        ) {
            return this.safe(parentIndex);
        }
        return false;
    }
    builtin(index: number, facts: Types, subject: TypeFact, anyName: boolean, depth = 0): boolean {
        if(depth >= 32) {
            return false;
        }
        if((subject.flags & 268435456) !== 0) {
            return subject.parts.some((part) => this.builtin(index, facts, facts.type(part), anyName, depth + 1));
        }
        if((subject.flags & 134217728) !== 0) {
            return subject.parts.every((part) => this.builtin(index, facts, facts.type(part), anyName, depth + 1));
        }
        if((subject.flags & 524288) !== 0) {
            return (
                subject.constraint !== 0 &&
                this.builtin(index, facts, facts.type(subject.constraint), anyName, depth + 1)
            );
        }
        const frames = new Frames(this.rules.ask(index, `type-origin\n${subject.id}`));
        header(frames, 'type-origin');
        if(!frames.yes()) {
            frames.end();
            return false;
        }
        const name = frames.field();
        const count = frames.natural();
        let library = false;
        for(let at = 0; at < count; at++) {
            frames.field();
            const declaration = frames.yes();
            const defaultLib = frames.yes();
            if(declaration && defaultLib) {
                library = true;
            }
        }
        frames.end();
        if(
            library &&
            (anyName ||
                [
                    'NumberConstructor',
                    'ObjectConstructor',
                    'StringConstructor',
                    'SymbolConstructor',
                    'ArrayConstructor',
                    'Array',
                    'ProxyConstructor',
                    'Console',
                    'DateConstructor',
                    'Atomics',
                    'Math',
                    'JSON',
                ].includes(name))
        ) {
            return true;
        }
        const bases = types(this.rules.ask(index, `base-shapes\n${subject.id}`), 'base-shapes');
        return bases.roots.some((id) => this.builtin(index, bases, bases.type(id), anyName, depth + 1));
    }
    native(object: number, key: number): boolean {
        const node = this.rules.parser.node(object);
        const property = this.rules.parser.node(key);
        if(
            node.kind === 'Identifier' &&
            property.kind === 'Identifier' &&
            boundMembers.includes(`${node.text}.${property.text}`)
        ) {
            const frames = new Frames(this.rules.ask(object, 'symbol-origin'));
            header(frames, 'symbol-origin');
            const file = frames.field();
            frames.end();
            if(file !== '' && file !== this.rules.path) {
                return true;
            }
        }
        const objectType = types(this.rules.ask(object, 'raw-shape'), 'raw-shape');
        if(!this.builtin(object, objectType, objectType.root(), false)) {
            return false;
        }
        const propertyType = types(this.rules.ask(key, 'raw-shape'), 'raw-shape');
        // Unlike the object test, Go checks only this type's own symbol here.
        const frames = new Frames(this.rules.ask(key, `type-origin\n${propertyType.root().id}`));
        header(frames, 'type-origin');
        if(!frames.yes()) {
            frames.end();
            return false;
        }
        frames.field();
        const count = frames.natural();
        let library = false;
        for(let at = 0; at < count; at++) {
            frames.field();
            const declaration = frames.yes();
            const defaultLib = frames.yes();
            if(declaration && defaultLib) {
                library = true;
            }
        }
        frames.end();
        return library;
    }
    report(index: number, id: number, name: string): boolean {
        const frames = new Frames(this.rules.ask(index, `property-info\n${id}\n${name}`));
        header(frames, 'property-info');
        if(!frames.yes()) {
            frames.end();
            return false;
        }
        const kind = frames.field();
        const initializer = frames.field();
        const firstName = frames.field();
        const annotation = frames.field();
        frames.end();
        let state = -1;
        if(kind === 'PropertyDeclaration') {
            if(initializer !== 'FunctionExpression') {
                return false;
            }
            state = 0;
        }
        else if(
            kind === 'MethodDeclaration' ||
            kind === 'MethodSignature' ||
            (kind === 'PropertyAssignment' && initializer === 'FunctionExpression')
        ) {
            if(firstName === 'this' && annotation === 'VoidKeyword') {
                return false;
            }
            state = firstName === 'this' ? 1 : 2;
        }
        if(state < 0) {
            return false;
        }
        const base =
            'A method that is not declared with `this: void` may cause unintentional scoping of `this` when separated from its object.\nConsider using an arrow function or explicitly `.bind()`ing the method to avoid calling the method with an unintended `this` value. ';
        this.rules.add(
            'unbound-method',
            state === 2 ? 'unboundWithoutThisAnnotation' : 'unbound',
            base +
                (state === 2 ? '\nIf a function does not access `this`, it can be annotated with `this: void`.' : ''),
            index,
        );
        return true;
    }
    constituent(index: number, facts: Types, name: string): boolean {
        for(const union of facts.parts(facts.root())) {
            for(const intersection of facts.parts(facts.type(union), 268435456)) {
                if(this.report(index, intersection, name)) {
                    return true;
                }
            }
        }
        return false;
    }
    member(index: number): void {
        const node = this.rules.parser.node(index);
        const object = node.children[0] ?? -1;
        const key = node.children[node.children.length - 1] ?? -1;
        if(this.safe(index) || this.native(object, key)) {
            return;
        }
        const names: string[] = [];
        if(node.kind === 'PropertyAccessExpression') {
            if(this.rules.parser.node(key).kind === 'Identifier') {
                names.push(this.rules.parser.node(key).text);
            }
        }
        else {
            const facts = types(this.rules.ask(key, 'raw-shape'), 'raw-shape');
            for(const part of facts.parts(facts.root())) {
                if((facts.type(part).flags & 3072) !== 0) {
                    let name = this.rules.name(facts.type(part), key);
                    if(
                        name.length >= 2 &&
                        ["'", '"'].includes(name.slice(0, 1)) &&
                        name.slice(-1) === name.slice(0, 1)
                    ) {
                        name = name.slice(1, -1);
                    }
                    names.push(name);
                }
            }
        }
        if(names.length === 0) {
            return;
        }
        const facts = types(this.rules.ask(object, 'raw-shape'), 'raw-shape');
        for(const name of names) {
            if(this.constituent(index, facts, name)) {
                break;
            }
        }
    }
    insideType(index: number): boolean {
        let parent = this.rules.parents[index] ?? -1;
        while(parent >= 0) {
            const node = this.rules.parser.node(parent);
            if(
                ['InterfaceDeclaration', 'TypeAliasDeclaration', 'FunctionType', 'MethodSignature'].includes(
                    node.kind,
                ) ||
                (['ClassDeclaration', 'VariableStatement', 'FunctionDeclaration'].includes(node.kind) &&
                    node.children.some((child) => this.rules.parser.node(child).kind === 'DeclareKeyword')) ||
                (node.kind === 'MethodDeclaration' &&
                    node.children.some((child) => this.rules.parser.node(child).kind === 'AbstractKeyword'))
            ) {
                return true;
            }
            parent = this.rules.parents[parent] ?? -1;
        }
        return false;
    }
    pattern(index: number): void {
        const node = this.rules.parser.node(index);
        const parentIndex = this.rules.parents[index] ?? -1;
        if(parentIndex < 0 || this.insideType(index)) {
            return;
        }
        const parent = this.rules.parser.node(parentIndex);
        let initializer = this.shadow.initializer(parentIndex);
        if(node.kind === 'ObjectLiteralExpression') {
            if(
                parent.kind !== 'BinaryExpression' ||
                parent.children[0] !== index ||
                this.rules.parser.node(parent.children[1] ?? -1).kind !== 'EqualsToken'
            ) {
                return;
            }
            initializer = parent.children[2] ?? -1;
        }
        const defaultValue = ['Parameter', 'BindingElement'].includes(parent.kind) && initializer >= 0;
        const facts = types(this.rules.ask(index, 'raw-shape'), 'raw-shape');
        for(const element of node.children) {
            const item = this.rules.parser.node(element);
            if(!['BindingElement', 'ShorthandPropertyAssignment', 'PropertyAssignment'].includes(item.kind)) {
                continue;
            }
            if(item.children.some((child) => this.rules.parser.node(child).kind === 'DotDotDotToken')) {
                continue;
            }
            const key = item.children[0] ?? -1;
            if(key < 0 || this.rules.parser.node(key).kind !== 'Identifier') {
                continue;
            }
            const name = this.rules.parser.node(key).text;
            if(initializer >= 0) {
                if(!this.native(initializer, key)) {
                    const initial = types(this.rules.ask(initializer, 'raw-shape'), 'raw-shape');
                    if(this.report(key, initial.root().id, name)) {
                        continue;
                    }
                }
                else if(!defaultValue) {
                    continue;
                }
            }
            this.constituent(key, facts, name);
        }
    }
    run(): void {
        for(let index = 0; index < this.rules.parser.nodes.length; index++) {
            const kind = this.rules.parser.node(index).kind;
            if(['PropertyAccessExpression', 'ElementAccessExpression'].includes(kind)) {
                this.member(index);
            }
            if(['ObjectBindingPattern', 'ObjectLiteralExpression'].includes(kind)) {
                this.pattern(index);
            }
        }
    }
}
