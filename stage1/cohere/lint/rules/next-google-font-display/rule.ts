import { panic } from 'adamic';
import type { RuleContext } from '../../context.ts';
import { diagnostic } from './decision.ts';
import { missing, discouraged } from './messages.ts';

export class Rule {
    readonly context: RuleContext;
    constructor(context: RuleContext) { this.context = context; }
    visit(index: number): void {
        const node = this.context.node(index);
        const tag = node.children[0] ?? panic('JSX opening without tag');
        if(this.context.node(tag).kind !== 'Identifier' || this.context.node(tag).text !== 'link') { return; }
        for(const child of node.children) {
            if(this.context.node(child).kind !== 'JsxAttributes') { continue; }
            for(const attribute of this.context.node(child).children) {
                const property = this.context.node(attribute);
                if(property.kind !== 'JsxAttribute') { continue; }
                const name = property.children[0] ?? panic('JSX attribute without name');
                const identifier = this.context.node(name);
                if(identifier.kind !== 'Identifier' || identifier.text.toLowerCase() !== 'href') { continue; }
                const value = property.children[1] ?? -1;
                if(value < 0 || this.context.node(value).kind !== 'StringLiteral') { return; }
                const id = diagnostic(this.context.node(value).text);
                if(id !== '') { this.context.report(index, '@next/next/google-font-display', id, id === 'googleFontDisplayMissing' ? missing : discouraged, '', '', ''); }
                return;
            }
        }
    }
}
export function create(context: RuleContext): Rule { return new Rule(context); }
