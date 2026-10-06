// Port of pinned Go cohere's no-script-url.
import type { RuleContext } from './rule_context.ts';

const messageScriptUrl =
    'A `javascript:` URL runs whatever follows the colon as code, so this string is an eval that happens to be spelled as a link. Anything that can reach the value that builds it can execute in the page, and a content security policy that blocks inline script blocks this too, so it tends to fail in production rather than in review. A real handler does the same work without turning a string into a program.';

export function visit(ctx: RuleContext, index: number): void {
    if(!ctx.enabled('no-script-url')) return;
    const node = ctx.node(index);
    if(node.kind !== 'StringLiteral' && node.kind !== 'NoSubstitutionTemplateLiteral') return;
    const parent = ctx.parent(index);
    if(
        node.kind === 'NoSubstitutionTemplateLiteral' &&
        parent >= 0 &&
        ctx.node(parent).kind === 'TaggedTemplateExpression'
    )
        return;
    if(node.text.slice(0, 11).toLowerCase() === ['javascript', ':'].join(''))
        ctx.report(index, 'no-script-url', 'unexpectedScriptURL', messageScriptUrl);
}
