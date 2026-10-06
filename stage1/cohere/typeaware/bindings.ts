import { panic } from 'adamic';
import type { Rules } from './rules.ts';
import { Shadow } from './shadow.ts';
import { Frames, header } from './frames.ts';

export class Bindings {
    readonly rules: Rules;
    readonly shape: Shadow;
    readonly spans = new Map<string, number>();
    readonly resolved = new Map<number, number[]>();
    constructor(rules: Rules) {
        this.rules = rules;
        this.shape = new Shadow(rules);
        for(let index = 0; index < rules.parser.nodes.length; index++) {
            const node = rules.parser.node(index);
            this.spans.set(`${node.kind}:${rules.byte(node.pos)}:${rules.byte(node.end)}`, index);
        }
    }
    parent(index: number): number {
        return this.rules.parents[index] ?? -1;
    }
    functionKind(kind: string): boolean {
        return ['FunctionDeclaration', 'FunctionExpression', 'ArrowFunction', 'MethodDeclaration',
            'GetAccessor', 'SetAccessor', 'Constructor'].includes(kind);
    }
    contains(outer: number, inner: number): boolean {
        const a = this.rules.parser.node(outer);
        const b = this.rules.parser.node(inner);
        return a.pos <= b.pos && a.end >= b.end;
    }
    list(index: number, question: string, followAlias = true): number[] {
        const frames = new Frames(this.rules.ask(index, question));
        header(frames, question);
        const result: number[] = [];
        if(frames.yes()) {
            const flags = frames.natural();
            const length = frames.natural();
            for(let at = 0; at < length; at++) {
                const path = frames.field();
                const kind = frames.field();
                const start = frames.natural();
                const end = frames.natural();
                if(path === this.rules.path) {
                    result.push(this.spans.get(`${kind}:${start}:${end}`) ?? panic('binding absent from Adamic tree'));
                }
            }
            frames.end();
            if(followAlias && question === 'binding-declarations' && (flags & 2097152) !== 0) {
                const aliases = this.list(index, 'alias-declarations');
                if(aliases.length > 0) {
                    return aliases;
                }
            }
            return result;
        }
        frames.end();
        return result;
    }
    declarations(index: number): number[] {
        if(!this.resolved.has(index)) {
            this.resolved.set(index, this.list(index, 'binding-declarations'));
        }
        return this.resolved.get(index) ?? panic('missing resolved binding');
    }
    earliest(index: number): number {
        let first = -1;
        for(const declaration of this.declarations(index)) {
            const name = this.shape.name(declaration);
            if(name >= 0 && (first < 0 || this.rules.parser.node(name).end <
                this.rules.parser.node(this.shape.name(first)).end)) {
                first = declaration;
            }
        }
        return first;
    }
    declaring(index: number): boolean {
        const parentIndex = this.parent(index);
        if(parentIndex < 0) {
            return false;
        }
        const parent = this.rules.parser.node(parentIndex);
        if(parent.kind === 'BindingElement' && parent.children.length > 1 &&
            parent.children[0] === index && this.shape.name(parentIndex) !== index) {
            return true;
        }
        if(['VariableDeclaration', 'Parameter', 'BindingElement', 'FunctionDeclaration',
            'FunctionExpression', 'ClassDeclaration', 'ClassExpression', 'InterfaceDeclaration',
            'TypeAliasDeclaration', 'EnumDeclaration', 'EnumMember', 'ModuleDeclaration',
            'TypeParameter', 'ImportClause', 'ImportEqualsDeclaration', 'NamespaceImport',
            'PropertyDeclaration', 'PropertySignature', 'MethodDeclaration', 'MethodSignature',
            'GetAccessor', 'SetAccessor'].includes(parent.kind)) {
            return this.shape.name(parentIndex) === index;
        }
        if(parent.kind === 'ImportSpecifier' || parent.kind === 'JsxAttribute') {
            return true;
        }
        if(parent.kind === 'PropertyAccessExpression' || parent.kind === 'QualifiedName') {
            return parent.children[parent.children.length - 1] === index;
        }
        if(parent.kind === 'PropertyAssignment') {
            return parent.children[0] === index;
        }
        if(parent.kind === 'ExportSpecifier') {
            let current = parentIndex;
            while(current >= 0) {
                const node = this.rules.parser.node(current);
                if(node.kind === 'ExportDeclaration') {
                    if(node.children.some((child) => this.rules.parser.node(child).kind === 'StringLiteral')) {
                        return true;
                    }
                    break;
                }
                current = this.parent(current);
            }
            const names = parent.children.filter((child) => this.rules.parser.node(child).kind === 'Identifier');
            return names.length > 1 && names[0] !== index;
        }
        if(parent.kind === 'TypePredicate') {
            return this.shape.name(parentIndex) === index;
        }
        if(parent.kind === 'TypeQuery') {
            let current = parentIndex;
            while(current >= 0) {
                const kind = this.rules.parser.node(current).kind;
                const name = this.shape.name(current);
                if(['Parameter', 'VariableDeclaration'].includes(kind) && name >= 0 &&
                    this.rules.parser.node(name).text === this.rules.parser.node(index).text) {
                    return true;
                }
                current = this.parent(current);
            }
        }
        return false;
    }
}
