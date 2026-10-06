// Port of pinned Go cohere's @typescript-eslint/prefer-namespace-keyword.
import type { RuleContext } from './rule_context.ts';

const messagePreferNamespaceKeyword =
    'This declares a namespace with the `module` keyword, which TypeScript has deprecated and plans to make a parse error. It means the same thing as `namespace` and reads as an ECMAScript module, which it is not. Write `namespace` instead.';

export function visit(ctx: RuleContext, index: number): void {
    if(!ctx.enabled('@typescript-eslint/prefer-namespace-keyword')) return;
    const node = ctx.node(index);
    if(node.kind !== 'ModuleDeclaration') return;
    const parent = ctx.parent(index);
    if(parent >= 0 && ctx.node(parent).kind === 'ModuleDeclaration') return;
    let keywordPosition = node.pos;
    let reportPosition = node.pos;
    for(const child of node.children) {
        if(ctx.node(child).kind.endsWith('Keyword') && ctx.node(child).kind !== 'GlobalKeyword') {
            keywordPosition = Math.max(keywordPosition, ctx.node(child).end);
            if(ctx.node(child).kind === 'ExportKeyword') reportPosition = Math.max(reportPosition, ctx.node(child).end);
        }
    }
    if(ctx.scanAt(keywordPosition) !== 'ModuleKeyword') return;
    const start = ctx.scanStart();
    const end = ctx.scanEnd();
    if(ctx.scanner.scan() !== 'Identifier') return;
    const finding = ctx.report(
        index,
        '@typescript-eslint/prefer-namespace-keyword',
        'preferNamespaceKeyword',
        messagePreferNamespaceKeyword,
        'fix',
        'namespace',
    );
    finding.editStart = start;
    finding.editEnd = end;
    ctx.scanAt(reportPosition);
    ctx.replaceRange(finding, ctx.scanner.start, node.end);
}
