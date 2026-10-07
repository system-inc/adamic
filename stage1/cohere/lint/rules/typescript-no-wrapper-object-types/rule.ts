import type { RuleContext } from '../../context.ts';
import type { ParseNode } from '../../../../typescript/parser/nodes.ts';
import { messageBannedWrapperObjectType } from './messages.ts';

export class Rule {
    readonly context: RuleContext;
    // declaredTypes holds the file's own type names, which shadow a wrapper of the same name.
    readonly declaredTypes = new Set<string>();
    constructor(context: RuleContext) {
        this.context = context;
    }
    prepare(index: number): void {
        const kind = this.context.node(index).kind;
        if(
            kind === 'InterfaceDeclaration' ||
            kind === 'TypeAliasDeclaration' ||
            kind === 'ClassDeclaration' ||
            kind === 'EnumDeclaration'
        ) {
            const name = this.context.name(index);
            if(name >= 0 && this.context.node(name).kind === 'Identifier') {
                this.declaredTypes.add(this.context.node(name).text);
            }
        }
        for(const child of this.context.node(index).children) {
            this.prepare(child);
        }
    }
    // A wrapper is fixed to its primitive, except in an implements clause, where only an object type fits.
    visit(node: ParseNode, index: number): void {
        const name = node.children[0] ?? -1;
        if(name < 0 || this.context.node(name).kind !== 'Identifier') {
            return;
        }
        const text = this.context.node(name).text;
        if(
            !['BigInt', 'Boolean', 'Number', 'Object', 'String', 'Symbol'].includes(text) ||
            this.declaredTypes.has(text)
        ) {
            return;
        }
        let parent = this.context.parent(index);
        let implementsHeritage = false;
        for(let depth = 0; depth < 2 && parent >= 0; depth++) {
            if(
                this.context.node(parent).kind === 'HeritageClause' &&
                this.context.node(parent).operator === 'ImplementsKeyword'
            ) {
                implementsHeritage = true;
            }
            parent = this.context.parent(parent);
        }
        this.context.report(
            name,
            '@typescript-eslint/no-wrapper-object-types',
            'bannedWrapperObjectType',
            messageBannedWrapperObjectType,
            implementsHeritage ? '' : 'fix',
            implementsHeritage ? '' : text.toLowerCase(),
            '',
        );
    }
}

export function create(context: RuleContext): Rule {
    return new Rule(context);
}
