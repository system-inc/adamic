import type { RuleContext } from '../../context.ts';
import { message } from './messages.ts';
import { Edit, type Suggestion } from './repairs.ts';

export class Rule {
    readonly context: RuleContext;
    constructor(context: RuleContext) { this.context = context; }
    specifiers(index: number): number[] {
        if(this.context.node(index).kind !== 'ImportDeclaration') { return []; }
        for(const child of this.context.node(index).children) {
            const clause = this.context.node(child);
            if(clause.kind !== 'ImportClause') { continue; }
            if(clause.semantic === 'TypeKeyword') { return []; }
            let named = -1;
            for(const binding of clause.children) {
                const kind = this.context.node(binding).kind;
                if(kind === 'Identifier' || kind === 'NamespaceImport') { return []; }
                if(kind === 'NamedImports') { named = binding; }
            }
            if(named < 0) { return []; }
            const elements = this.context.node(named).children;
            if(elements.length === 0) { return []; }
            for(const element of elements) {
                const specifier = this.context.node(element);
                if(specifier.kind !== 'ImportSpecifier' || specifier.semantic !== '1') { return []; }
            }
            return elements;
        }
        return [];
    }
    fixes(index: number): Edit[] {
        const specifiers = this.specifiers(index);
        if(specifiers.length === 0) { return []; }
        const edits: Edit[] = [];
        for(const specifier of specifiers) {
            const name = this.context.node(specifier).children[0] ?? -1;
            if(name >= 0) { edits.push(new Edit(this.context.start(specifier), this.context.start(name), '')); }
        }
        const insertion = this.context.start(index) + 6;
        edits.push(new Edit(insertion, insertion, ' type'));
        return edits;
    }
    suggestions(index: number): Suggestion[] { return []; }
    visit(index: number, parent: number): void {
        if(this.specifiers(index).length !== 0) {
            // Exact individual fixes are exposed above pending the shared multi-edit Finding API.
            this.context.report(index, '@typescript-eslint/no-import-type-side-effects', 'useTopLevelQualifier', message, '', '', '');
        }
    }
}
export function create(context: RuleContext): Rule { return new Rule(context); }
