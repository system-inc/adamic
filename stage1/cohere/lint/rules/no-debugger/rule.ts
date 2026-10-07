import type { RuleContext } from '../../context.ts';
import { debuggerMessage } from './messages.ts';

export class Rule {
    readonly context: RuleContext;
    constructor(context: RuleContext) {
        this.context = context;
    }
    visit(index: number, parent: number): void {
        const node = this.context.node(index);
        const parentKind = parent < 0 ? '' : this.context.node(parent).kind;
        if(node.kind === 'DebuggerStatement') {
            const removable = ['Block', 'SourceFile', 'ModuleBlock', 'CaseClause', 'DefaultClause'].includes(
                parentKind,
            );
            this.context.report(
                index,
                'no-debugger',
                'unexpectedDebugger',
                debuggerMessage,
                removable ? 'fix' : '',
                '',
                '',
            );
        }
    }
}

export function create(context: RuleContext): Rule {
    return new Rule(context);
}
