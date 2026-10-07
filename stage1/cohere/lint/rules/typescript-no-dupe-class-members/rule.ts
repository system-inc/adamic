import type { RuleContext } from '../../context.ts';
import type { Suggestion } from './repairs.ts';

export class Rule {
    readonly context: RuleContext;
    constructor(context: RuleContext) { this.context = context; }
    suggestions(index: number): Suggestion[] { return []; }
    visit(index: number, parent: number): void {
        const node = this.context.node(index);
        if(!['ClassDeclaration', 'ClassExpression'].includes(node.kind)) { return; }
        const seen = new Map<string, string>();
        for(const memberIndex of node.children) {
            const member = this.context.node(memberIndex);
            if(!['MethodDeclaration', 'PropertyDeclaration', 'GetAccessor', 'SetAccessor'].includes(member.kind)) { continue; }
            let name = -1;
            let static_ = false;
            let body = false;
            for(const child of member.children) {
                const kind = this.context.node(child).kind;
                if(kind === 'StaticKeyword') { static_ = true; }
                if(kind === 'Block') { body = true; }
                if(name < 0 && ['Identifier', 'PrivateIdentifier', 'StringLiteral', 'NumericLiteral', 'ComputedPropertyName'].includes(kind)) { name = child; }
            }
            if(name < 0 || this.context.node(name).kind === 'ComputedPropertyName') { continue; }
            if(member.kind !== 'PropertyDeclaration' && !body) { continue; }
            const keyNode = this.context.node(name);
            const text = `${keyNode.kind === 'NumericLiteral' ? 'number' : 'string'}:${keyNode.text}`;
            const key = `${static_ ? 'static' : 'instance'}:${keyNode.kind === 'PrivateIdentifier' ? 'private' : 'public'}:${text}`;
            const previous = seen.get(key);
            if(previous !== undefined) {
                const accessors = ['GetAccessor', 'SetAccessor'];
                if(accessors.includes(previous) && accessors.includes(member.kind) && previous !== member.kind) { continue; }
                this.context.report(name, '@typescript-eslint/no-dupe-class-members', 'noDupeClassMembers', `This class already declares a member named ${text}. The later declaration wins silently, so the earlier one is dead code that reads as live, and nothing in the language or at runtime tells the two apart.`, '', '', '');
            }
            else { seen.set(key, member.kind); }
        }
    }
}
export function create(context: RuleContext): Rule { return new Rule(context); }
