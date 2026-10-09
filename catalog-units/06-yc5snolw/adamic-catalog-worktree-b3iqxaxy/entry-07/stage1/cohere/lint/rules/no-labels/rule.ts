import { panic } from 'adamic';
import type { RuleContext } from '../../context.ts';
import type { ParseNode } from '../../../../typescript/parser/nodes.ts';
import { messageUnexpectedLabel, messageUnexpectedLabelInBreak, messageUnexpectedLabelInContinue } from './messages.ts';

// Upstream judges the whole file from its SourceFile listener, so every finding exists before the walk
// reaches any other node. Walking from the file here keeps that order: a labeled `continue` reports
// no-labels ahead of no-continue at the same position, as Go does.
export class Rule {
    readonly context: RuleContext;
    constructor(context: RuleContext) {
        this.context = context;
    }
    visit(node: ParseNode, index: number): void {
        this.walk(index);
    }
    walk(index: number): void {
        const node = this.context.node(index);
        if(node.kind === 'LabeledStatement' && !this.allowedLabel(index)) {
            this.context.report(index, 'no-labels', 'unexpectedLabel', messageUnexpectedLabel, '', '', '');
        }
        if((node.kind === 'BreakStatement' || node.kind === 'ContinueStatement') && node.children.length > 0) {
            const name = this.context.node(node.children[0] ?? panic('label')).text;
            let allowed = false;
            let cursor = this.context.parent(index);
            while(cursor >= 0) {
                if(
                    this.context.node(cursor).kind === 'LabeledStatement' &&
                    this.context.node(this.context.node(cursor).children[0] ?? panic('label name')).text === name
                ) {
                    allowed = this.allowedLabel(cursor);
                    break;
                }
                cursor = this.context.parent(cursor);
            }
            if(!allowed) {
                this.context.report(
                    index,
                    'no-labels',
                    node.kind === 'BreakStatement' ? 'unexpectedLabelInBreak' : 'unexpectedLabelInContinue',
                    node.kind === 'BreakStatement' ? messageUnexpectedLabelInBreak : messageUnexpectedLabelInContinue,
                    '',
                    '',
                    '',
                );
            }
        }
        for(const child of node.children) {
            this.walk(child);
        }
    }
    // allowedLabel is whether the options permit this label: allowLoop for a loop's, allowSwitch for a switch's.
    allowedLabel(index: number): boolean {
        const body = this.context.node(this.context.node(index).children[1] ?? panic('labeled body')).kind;
        return ['ForStatement', 'ForInStatement', 'ForOfStatement', 'WhileStatement', 'DoStatement'].includes(body)
            ? this.context.settings.read('allowloop', 'false') === 'true'
            : body === 'SwitchStatement' && this.context.settings.read('allowswitch', 'false') === 'true';
    }
}

export function create(context: RuleContext): Rule {
    return new Rule(context);
}
