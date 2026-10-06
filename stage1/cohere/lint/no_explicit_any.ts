import type { Batch5Context } from './batch5_context.ts';
import { typescriptFile } from './batch5_helpers.ts';
import { ruleMessage } from './batch5_messages.ts';
import { RepairEdit } from './repair_edit.ts';
export function visit(context: Batch5Context, index: number): void {
    const rule = '@typescript-eslint/no-explicit-any';
    if(context.node(index).kind !== 'AnyKeyword' || !context.enabled(rule) || !typescriptFile(context.parser.path)) {
        return;
    }
    if(context.settings.read('ignorerestargs', 'false') === 'true') {
        for(let owner = context.parent(index); owner >= 0; owner = context.parent(owner)) {
            if(context.node(owner).kind === 'Parameter' && context.hasKind(owner, 'DotDotDotToken')) {
                return;
            }
        }
    }
    const edits: RepairEdit[] = [];
    if(context.settings.read('fixtounknown', 'false') === 'true') {
        edits.push(new RepairEdit(context.start(index), context.node(index).end, 'unknown'));
    }
    context.report(index, rule, 'unexpectedAny', ruleMessage('unexpectedAny'), edits, []);
}
