import type { RuleContext } from '../../context.ts';
import type { ParseNode } from '../../../../typescript/parser/nodes.ts';
import { SuggestionEdit } from '../../suggestions.a';
import { messageUnicodeBomExpected, messageUnicodeBomUnexpected } from './messages.ts';

export class Rule {
    readonly context: RuleContext;
    constructor(context: RuleContext) {
        this.context = context;
    }
    // The finding is the empty range at the file's start, whatever the file node holds, and its repair
    // inserts or deletes the one mark there.
    visit(node: ParseNode, index: number): void {
        const hasMark = this.context.source.startsWith('﻿');
        const require = this.context.settings.read('require', 'never');
        if(require === 'always' && !hasMark) {
            this.context.reportRange(0, 0, 'unicode-bom', 'expected', messageUnicodeBomExpected, [
                new SuggestionEdit(0, 0, '﻿'),
            ]);
        }
        if(require === 'never' && hasMark) {
            this.context.reportRange(0, 0, 'unicode-bom', 'unexpected', messageUnicodeBomUnexpected, [
                new SuggestionEdit(0, 1, ''),
            ]);
        }
    }
}

export function create(context: RuleContext): Rule {
    return new Rule(context);
}
