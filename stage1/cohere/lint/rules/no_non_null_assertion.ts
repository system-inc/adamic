import { panic } from 'adamic';
import type { RuleContext } from './context.ts';
import { Finding } from '../finding.ts';
import { RepairEdit } from '../repair_edit.ts';
import { policyMessage } from '../volume_messages.ts';
import { first, last, parent, suggest } from './shared.ts';

function assignment(context: RuleContext, index: number): boolean {
    const owner = parent(context, index);
    if(owner < 0) {
        return false;
    }
    const node = context.node(owner);
    switch(node.kind) {
        case 'DeleteExpression':
            return first(context, owner) === index;
        case 'PrefixUnaryExpression':
        case 'PostfixUnaryExpression':
            return ['PlusPlusToken', 'MinusMinusToken'].includes(node.operator);
        case 'ArrayLiteralExpression':
        case 'SpreadElement':
            return true;
        case 'PropertyAssignment': {
            const object = parent(context, owner);
            return (
                last(context, owner) === index &&
                object >= 0 &&
                context.node(object).kind === 'ObjectLiteralExpression' &&
                assignment(context, object)
            );
        }
        case 'ParenthesizedExpression':
        case 'NonNullExpression':
        case 'AsExpression':
        case 'TypeAssertionExpression':
        case 'SatisfiesExpression':
            return assignment(context, owner);
        case 'BinaryExpression':
            return (
                first(context, owner) === index &&
                [
                    'EqualsToken',
                    'PlusEqualsToken',
                    'MinusEqualsToken',
                    'AsteriskEqualsToken',
                    'AsteriskAsteriskEqualsToken',
                    'SlashEqualsToken',
                    'PercentEqualsToken',
                    'LessThanLessThanEqualsToken',
                    'GreaterThanGreaterThanEqualsToken',
                    'GreaterThanGreaterThanGreaterThanEqualsToken',
                    'AmpersandEqualsToken',
                    'BarEqualsToken',
                    'CaretEqualsToken',
                    'BarBarEqualsToken',
                    'AmpersandAmpersandEqualsToken',
                    'QuestionQuestionEqualsToken',
                ].includes(context.node(node.children[1] ?? panic('operator')).kind)
            );
        default:
            return false;
    }
}
export function noNonNullAssertion(context: RuleContext, index: number): void {
    const rule = '@typescript-eslint/no-non-null-assertion';
    const end = context.node(index).end;
    const edits: RepairEdit[] = [];
    const owner = parent(context, index);
    if(owner >= 0 && first(context, owner) === index) {
        const node = context.node(owner);
        const member = ['PropertyAccessExpression', 'ElementAccessExpression'].includes(node.kind);
        if((member && !assignment(context, owner)) || node.kind === 'CallExpression') {
            if(context.hasKind(owner, 'QuestionDotToken')) {
                edits.push(new RepairEdit(end - 1, end, ''));
            }
            else if(node.kind !== 'PropertyAccessExpression') {
                edits.push(new RepairEdit(end - 1, end, '?.'));
            }
            else {
                const dot = context.source.indexOf('.', end);
                if(dot >= 0 && dot < node.end) {
                    edits.push(new RepairEdit(end - 1, end, ''));
                    edits.push(new RepairEdit(dot, dot + 1, '?.'));
                }
            }
        }
    }
    const edit = edits[0];
    const message = policyMessage(rule, 'suggestOptionalChain', []);
    const finding = new Finding(
        rule,
        'noNonNull',
        policyMessage(rule, 'noNonNull', []),
        context.start(index),
        end,
        edit === undefined ? '' : 'suggestion',
        edit?.text ?? '',
        edit === undefined ? '' : message,
    );
    if(edit !== undefined) {
        finding.editStart = edit.start;
        finding.editEnd = edit.end;
        suggest(finding, 'suggestOptionalChain', message, edits);
    }
    context.record(finding);
}

export function visit(context: RuleContext, index: number): void {
    if(
        context.node(index).kind === 'NonNullExpression' &&
        context.enabled('@typescript-eslint/no-non-null-assertion')
    ) {
        noNonNullAssertion(context, index);
    }
}
