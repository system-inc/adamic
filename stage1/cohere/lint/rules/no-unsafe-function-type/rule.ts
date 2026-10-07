import type { RuleContext } from '../../context.ts';
import { message } from './messages.ts';

export class Rule {
    readonly context: RuleContext;
    shadowed = false;
    constructor(context: RuleContext) {
        this.context = context;
    }
    prepare(root: number): void {
        this.shadow(root);
    }
    shadow(index: number): void {
        const node = this.context.node(index);
        if(['InterfaceDeclaration', 'TypeAliasDeclaration', 'ClassDeclaration', 'EnumDeclaration'].includes(node.kind)) {
            for(const child of node.children) {
                const name = this.context.node(child);
                if(name.kind === 'Identifier') {
                    if(name.text === 'Function') {
                        this.shadowed = true;
                    }
                    break;
                }
            }
        }
        for(const child of node.children) {
            this.shadow(child);
        }
    }
    visit(index: number): void {
        if(this.shadowed) {
            return;
        }
        const node = this.context.node(index);
        if(node.kind !== 'TypeReference') {
            return;
        }
        const first = node.children[0];
        if(first === undefined) {
            return;
        }
        const name = this.context.node(first);
        if(name.kind === 'Identifier' && name.text === 'Function') {
            this.context.report(first, '@typescript-eslint/no-unsafe-function-type', 'bannedFunctionType', message, '', '', '');
        }
    }
}

export function create(context: RuleContext): Rule {
    return new Rule(context);
}
