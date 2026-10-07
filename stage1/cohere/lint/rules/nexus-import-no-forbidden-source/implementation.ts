import type { RuleContext } from '../../context.ts';
import { entry } from './messages.ts';
export function specifier(context: RuleContext, index: number): number {
    const node = context.node(index);
    if(node.kind === 'ImportDeclaration') {
        for(const child of node.children) { if(['StringLiteral', 'NoSubstitutionTemplateLiteral'].includes(context.node(child).kind)) { return child; } }
    }
    if(node.kind === 'CallExpression') {
        const callee = node.children[0] ?? -1;
        if(callee < 0 || !(context.node(callee).kind === 'ImportKeyword' || (context.node(callee).kind === 'Identifier' && context.node(callee).text === 'require'))) { return -1; }
        context.scanner.pos = context.node(callee).end;
        while(context.scanner.scan() !== 'OpenParenToken') { if(context.scanner.start >= node.end) { return -1; } }
        const opening = context.scanner.pos;
        const argumentsIn: number[] = [];
        for(const child of node.children.slice(1)) { if(context.start(child) >= opening) { argumentsIn.push(child); } }
        const first = argumentsIn[0] ?? -1;
        if(context.node(callee).kind === 'Identifier' && argumentsIn.length !== 1) { return -1; }
        if(first >= 0 && ['StringLiteral', 'NoSubstitutionTemplateLiteral'].includes(context.node(first).kind)) { return first; }
    }
    return -1;
}
export class Rule {
    readonly context: RuleContext;
    readonly filename: string;
    constructor(context: RuleContext, filename: string) { this.context = context; this.filename = filename.split('\\').join('/'); }
    visit(index: number): void {
        const target = specifier(this.context, index);
        if(target < 0) { return; }
        const match = entry(this.context.node(target).text);
        if(match === undefined || (match.owner !== '' && this.filename.includes(match.owner))) { return; }
        this.context.report(target, 'nexus/import-no-forbidden-source', match.id, match.message, 'fix', "'" + match.replacement + "'", '');
    }
}
