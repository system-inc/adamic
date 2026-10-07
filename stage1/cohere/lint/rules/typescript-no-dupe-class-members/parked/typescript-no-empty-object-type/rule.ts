import { panic } from 'adamic';
import type { RuleContext } from '../../context.ts';
import { messageEmptyInterface, messageEmptyInterfaceWithSuper, messageEmptyObject } from './messages.ts';
import { Edit, Suggestion } from './repairs.ts';

export class Rule {
    readonly context: RuleContext;
    constructor(context: RuleContext) { this.context = context; }
    name(index: number): number {
        for(const child of this.context.node(index).children) {
            if(this.context.node(child).kind === 'Identifier') { return child; }
        }
        return -1;
    }
    bases(index: number): number[] {
        const result: number[] = [];
        for(const child of this.context.node(index).children) {
            if(this.context.node(child).kind === 'HeritageClause') {
                for(const base of this.context.node(child).children) { result.push(base); }
            }
        }
        return result;
    }
    matchesName(index: number): boolean {
        const pattern = this.context.settings.read('allowwithname', '');
        const name = this.name(index);
        return pattern !== '' && name >= 0 && new RegExp(pattern, 'u').test(this.context.node(name).text);
    }
    reports(index: number): boolean {
        const node = this.context.node(index);
        if(node.kind === 'InterfaceDeclaration') {
            const option = this.context.settings.read('allowinterfaces', 'never');
            if(option === 'always' || this.name(index) < 0 || this.matchesName(index)) { return false; }
            for(const child of node.children) {
                if(['PropertySignature', 'MethodSignature', 'CallSignature', 'ConstructSignature', 'IndexSignature'].includes(this.context.node(child).kind)) { return false; }
            }
            const bases = this.bases(index);
            return bases.length <= 1 && !(bases.length === 1 && option === 'with-single-extends');
        }
        if(node.kind !== 'TypeLiteral' || node.children.length !== 0 || this.context.settings.read('allowobjecttypes', 'never') === 'always') { return false; }
        let parent = this.context.parents[index] ?? -1;
        while(parent >= 0 && this.context.node(parent).kind === 'ParenthesizedType') { parent = this.context.parents[parent] ?? -1; }
        if(parent >= 0 && this.context.node(parent).kind === 'IntersectionType') { return false; }
        return !(parent >= 0 && this.context.node(parent).kind === 'TypeAliasDeclaration' && this.matchesName(parent));
    }
    merges(index: number): boolean {
        const declaration = this.context.node(index);
        for(const child of declaration.children) {
            if(this.context.node(child).kind === 'TypeParameter') { return false; }
        }
        const name = this.name(index);
        if(name < 0) { return false; }
        let scope = this.context.parents[index] ?? -1;
        if(scope < 0) { return false; }
        if(['CaseClause', 'DefaultClause'].includes(this.context.node(scope).kind)) {
            scope = this.context.parents[scope] ?? -1;
        }
        if(scope < 0) { return false; }
        const statements = this.context.node(scope).children.slice();
        if(this.context.node(scope).kind === 'CaseBlock') {
            statements.splice(0);
            for(const clause of this.context.node(scope).children) {
                for(const statement of this.context.node(clause).children) { statements.push(statement); }
            }
        }
        for(const statement of statements) {
            if(statement === index || !['ClassDeclaration', 'InterfaceDeclaration'].includes(this.context.node(statement).kind)) { continue; }
            const other = this.name(statement);
            if(other >= 0 && this.context.node(other).text === this.context.node(name).text) { return true; }
        }
        return false;
    }
    rewriteStart(index: number): number {
        const node = this.context.node(index);
        let after = -1;
        for(const child of node.children) {
            const kind = this.context.node(child).kind;
            if(kind === 'ExportKeyword' || kind === 'DefaultKeyword') { after = this.context.node(child).end; }
            else if(kind === 'DeclareKeyword' || kind === 'AbstractKeyword') { return this.context.start(child); }
        }
        if(after < 0) { return this.context.start(index); }
        this.context.scanner.pos = after;
        this.context.scanner.scan();
        return this.context.scanner.start;
    }
    interfaceEdit(index: number, target: string): Edit {
        const name = this.name(index);
        if(name < 0) { panic('interface without name'); }
        let parameters = '';
        let last = -1;
        for(const child of this.context.node(index).children) {
            if(this.context.node(child).kind === 'TypeParameter') { last = child; }
        }
        if(last >= 0) {
            this.context.scanner.pos = this.context.node(name).end;
            this.context.scanner.scan();
            const opening = this.context.scanner.start;
            this.context.scanner.pos = this.context.node(last).end;
            this.context.scanner.scan();
            parameters = this.context.source.slice(opening, this.context.scanner.start + 1);
        }
        return new Edit(this.rewriteStart(index), this.context.node(index).end, `type ${this.context.node(name).text}${parameters} = ${target}`);
    }
    suggestions(index: number): Suggestion[] {
        if(!this.reports(index)) { return []; }
        const node = this.context.node(index);
        if(node.kind === 'TypeLiteral') {
            const start = this.context.start(index);
            return [new Suggestion('replaceEmptyObjectType', 'Replace `{}` with `object`, which means any non-primitive.', [new Edit(start, node.end, 'object')]), new Suggestion('replaceEmptyObjectType', 'Replace `{}` with `unknown`, which means any value at all.', [new Edit(start, node.end, 'unknown')])];
        }
        for(const child of node.children) { if(this.context.node(child).kind === 'DefaultKeyword') { return []; } }
        if(this.merges(index)) { return []; }
        const bases = this.bases(index);
        if(bases.length === 1) {
            const base = bases[0] ?? panic('missing base');
            const text = this.context.source.slice(this.context.start(base), this.context.node(base).end);
            return [new Suggestion('replaceEmptyInterfaceWithSuper', 'Replace the empty interface with a type alias of the type it extends.', [this.interfaceEdit(index, text)])];
        }
        return [new Suggestion('replaceEmptyInterface', 'Replace the empty interface with `object`, which means any non-primitive.', [this.interfaceEdit(index, 'object')]), new Suggestion('replaceEmptyInterface', 'Replace the empty interface with `unknown`, which means any value at all.', [this.interfaceEdit(index, 'unknown')])];
    }
    visit(index: number, parent: number): void {
        if(!this.reports(index)) { return; }
        if(this.context.node(index).kind === 'TypeLiteral') {
            this.context.report(index, '@typescript-eslint/no-empty-object-type', 'noEmptyObject', messageEmptyObject, '', '', '');
        }
        else {
            const extended = this.bases(index).length === 1;
            this.context.report(this.name(index), '@typescript-eslint/no-empty-object-type', extended ? 'noEmptyInterfaceWithSuper' : 'noEmptyInterface', extended ? messageEmptyInterfaceWithSuper : messageEmptyInterface, '', '', '');
        }
    }
}
export function create(context: RuleContext): Rule { return new Rule(context); }
