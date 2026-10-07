import { panic } from 'adamic';
import type { RuleContext } from '../../context.ts';
import { Finding } from '../../finding.ts';
import { description } from './messages.ts';

export class Rule {
    readonly context: RuleContext;
    constructor(context: RuleContext) { this.context = context; }
    visit(index: number, parent: number): void {
        if(parent >= 0 && this.context.node(parent).kind === 'ModuleDeclaration') { return; }
        const node = this.context.node(index);
        let afterModifiers = node.pos;
        let reportStart = node.pos;
        let name = -1;
        for(const child of node.children) {
            const item = this.context.node(child);
            if(item.kind.endsWith('Keyword')) {
                afterModifiers = Math.max(afterModifiers, item.end);
                if(item.kind === 'ExportKeyword') { reportStart = Math.max(reportStart, item.end); }
            }
            else { name = child; break; }
        }
        if(name < 0 || this.context.node(name).kind !== 'Identifier') { return; }
        const scanner = this.context.scanner;
        scanner.pos = afterModifiers; scanner.scan();
        const keywordStart = scanner.start;
        const keywordEnd = scanner.pos;
        if(this.context.source.slice(keywordStart, keywordEnd) !== 'module') { return; }
        scanner.pos = reportStart; scanner.scan();
        const finding = new Finding('@typescript-eslint/prefer-namespace-keyword', 'preferNamespaceKeyword', description,
            scanner.start, node.end, 'fix', 'namespace', '');
        finding.editStart = keywordStart; finding.editEnd = keywordEnd;
        this.context.findings.push(finding);
    }
}
export function create(context: RuleContext): Rule { return new Rule(context); }
