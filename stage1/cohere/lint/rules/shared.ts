import { panic } from 'adamic';
import type { RuleContext } from './context.ts';
import { Finding } from '../finding.ts';
import type { RepairEdit } from '../repair_edit.ts';
import { RepairSuggestion } from '../repair_suggestion.ts';
import { space } from '../comments.ts';
import { foldPoint } from '../unicode.ts';
import { policyMessage } from '../volume_messages.ts';

export function trim(value: string): string {
    let start = 0;
    let end = value.length;
    while(start < end && space(value[start] ?? '')) {
        start++;
    }
    while(end > start && space(value[end - 1] ?? '')) {
        end--;
    }
    return value.slice(start, end);
}
export function ancestry(context: RuleContext, index: number, owner: number): void {
    context.setParent(index, owner);
    for(const child of context.node(index).children) {
        ancestry(context, child, index);
    }
}
export function parent(context: RuleContext, index: number): number {
    return context.parents[index] ?? -1;
}
export function first(context: RuleContext, index: number): number {
    return context.node(index).children[0] ?? -1;
}
export function last(context: RuleContext, index: number): number {
    const children = context.node(index).children;
    return children[children.length - 1] ?? -1;
}
export function name(context: RuleContext, index: number): number {
    for(const child of context.node(index).children) {
        if(
            [
                'Identifier',
                'StringLiteral',
                'NumericLiteral',
                'PrivateIdentifier',
                'ComputedPropertyName',
                'ObjectBindingPattern',
                'ArrayBindingPattern',
            ].includes(context.node(child).kind)
        ) {
            return child;
        }
    }
    return -1;
}
export function emit(context: RuleContext, index: number, rule: string, id: string, fields: string[]): Finding {
    const finding = new Finding(
        rule,
        id,
        policyMessage(rule, id, fields),
        context.start(index),
        context.node(index).end,
        '',
        '',
        '',
    );
    context.record(finding);
    return finding;
}
export function suggest(finding: Finding, id: string, message: string, edits: RepairEdit[]): void {
    finding.addSuggestion(new RepairSuggestion(id, message, edits));
}
export function betweenToken(context: RuleContext, start: number, end: number, kind: string): boolean {
    let cursor = start;
    while(cursor < end) {
        const token = context.scanAt(cursor);
        if(token.start >= end) {
            return false;
        }
        if(context.scanner.kind === kind) {
            return true;
        }
        if(token.end <= cursor) {
            return false;
        }
        cursor = token.end;
    }
    return false;
}

export function foreign(context: RuleContext, index: number): boolean {
    const owner = parent(context, index);
    if(owner < 0) {
        return false;
    }
    const node = context.node(owner);
    switch(node.kind) {
        case 'PropertyAccessExpression':
            return last(context, owner) === index && context.node(first(context, owner)).kind !== 'ThisKeyword';
        case 'PropertyAssignment':
        case 'PropertySignature':
            return first(context, owner) === index;
        case 'BindingElement':
            return (
                first(context, owner) === index &&
                node.children.length > 1 &&
                betweenToken(
                    context,
                    context.node(index).end,
                    context.start(node.children[1] ?? panic('binding child')),
                    'ColonToken',
                )
            );
        case 'ImportSpecifier':
        case 'ExportSpecifier':
        case 'NamespaceImport':
        case 'ImportClause':
            return true;
        case 'QualifiedName':
            return last(context, owner) === index;
        case 'TypeReference':
            return first(context, owner) === index;
        default:
            return false;
    }
}
export function argumentsOf(context: RuleContext, index: number): number[] {
    const node = context.node(index);
    return node.list <= 0 ? [] : node.children.slice(node.children.length - node.list);
}

export function handlerName(value: string): boolean {
    let lower = '';
    for(let cursor = 0; cursor < value.length; cursor++) {
        const point = value.codePointAt(cursor) ?? 0;
        const folded = foldPoint(point);
        lower += folded >= 65 && folded <= 90 ? String.fromCodePoint(folded + 32) : String.fromCodePoint(folded);
        if(point > 65535) {
            cursor++;
        }
    }
    if(lower.includes('handle') || lower.includes('event')) {
        return true;
    }
    for(let index = 0; index + 2 < lower.length; index++) {
        const following = lower[index + 2] ?? '';
        if(lower.slice(index, index + 2) === 'on' && following >= 'a' && following <= 'z') {
            return true;
        }
    }
    return false;
}
export function eventKey(value: string): boolean {
    const following = value[2] ?? '';
    return value.startsWith('on') && following >= 'A' && following <= 'Z';
}
