import type { RuleContext } from './context.ts';
import { Finding } from '../finding.ts';
import { policyMessage } from '../volume_messages.ts';
import { argumentsOf, first, last, name, parent, trim } from './shared.ts';

function claimed(context: RuleContext, call: number): string {
    if(call < 0 || context.node(call).kind !== 'CallExpression') {
        return '';
    }
    const callee = first(context, call);
    if(callee < 0 || context.node(callee).kind !== 'PropertyAccessExpression') {
        return '';
    }
    const property = context.node(last(context, callee)).text;
    const receiver = context.node(first(context, callee));
    if(
        receiver.kind === 'Identifier' &&
        receiver.text === 'React' &&
        (property === 'forwardRef' || property.startsWith('use'))
    ) {
        return property;
    }
    return property === 'addEventListener' ? property : '';
}
function inherited(context: RuleContext, index: number, onlyThis: boolean): boolean {
    const node = context.node(index);
    if(node.kind === 'ThisKeyword') {
        return true;
    }
    if(
        !onlyThis &&
        (node.kind === 'SuperKeyword' || (node.kind === 'MetaProperty' && node.operator === 'NewKeyword'))
    ) {
        return true;
    }
    if(!onlyThis && node.kind === 'Identifier' && node.text === 'arguments') {
        const owner = parent(context, index);
        if(owner < 0) {
            return true;
        }
        const kind = context.node(owner).kind;
        if(kind === 'PropertyAccessExpression' && last(context, owner) === index) {
            return false;
        }
        if(
            [
                'PropertyAssignment',
                'PropertyDeclaration',
                'PropertySignature',
                'MethodDeclaration',
                'MethodSignature',
            ].includes(kind) &&
            name(context, owner) === index
        ) {
            return false;
        }
        return true;
    }
    if(
        ['FunctionDeclaration', 'FunctionExpression', 'MethodDeclaration'].includes(node.kind) ||
        (onlyThis
            ? ['ClassDeclaration', 'ClassExpression'].includes(node.kind)
            : ['Constructor', 'GetAccessor', 'SetAccessor'].includes(node.kind))
    ) {
        return false;
    }
    for(const child of node.children) {
        if(inherited(context, child, onlyThis)) {
            return true;
        }
    }
    return false;
}
function report(context: RuleContext, index: number, id: string, fields: string[]): void {
    const node = context.node(index);
    let safe = true;
    for(const child of node.children) {
        if(inherited(context, child, false)) {
            safe = false;
        }
    }
    let signatureStart = context.start(index);
    let async = '';
    let arrowStart = -1;
    let genericEnd = -1;
    for(const child of node.children) {
        const part = context.node(child);
        if(part.kind === 'AsyncKeyword') {
            async = 'async ';
            signatureStart = context.scanAt(part.end).start;
        }
        if(part.kind === 'TypeParameter') {
            genericEnd = part.end;
        }
        if(part.kind === 'EqualsGreaterThanToken') {
            arrowStart = context.start(child);
        }
    }
    let generics = '';
    let parametersStart = signatureStart;
    if(genericEnd >= 0) {
        let token = context.scanAt(genericEnd);
        if(context.scanner.kind === 'CommaToken') {
            token = context.scanAt(token.end);
        }
        const angle = token.start;
        if(angle >= arrowStart || context.source[angle] !== '>') {
            safe = false;
        }
        generics = trim(context.source.slice(signatureStart, angle + 1));
        parametersStart = angle + 1;
    }
    if(arrowStart < parametersStart) {
        safe = false;
    }
    let signature = trim(context.source.slice(parametersStart, arrowStart));
    if(!signature.startsWith('(')) {
        signature = `(${signature})`;
    }
    let body = last(context, index);
    const block = context.node(body).kind === 'Block';
    if(!block) {
        body = context.unwrap(body);
    }
    const bodyText = block ? context.text(body) : `{ return ${context.text(body)}; }`;
    let replacement = `${async}function${generics}${signature} ${bodyText}`;
    for(let cursor = index; parent(context, cursor) >= 0;) {
        cursor = parent(context, cursor);
        if(context.start(cursor) !== context.start(index)) {
            break;
        }
        if(context.node(cursor).kind === 'ExpressionStatement') {
            replacement = `(${replacement})`;
            break;
        }
    }
    const rule = 'nexus/consistency-no-multiline-arrow-function';
    context.record(
        new Finding(
            rule,
            id,
            policyMessage(rule, id, fields),
            context.start(index),
            node.end,
            safe ? 'fix' : '',
            safe ? replacement : '',
            '',
        ),
    );
}
export function noMultilineArrowFunction(context: RuleContext, index: number): void {
    if(context.node(index).kind === 'CallExpression') {
        const kind = claimed(context, index);
        if(kind === '') {
            return;
        }
        const argument = argumentsOf(context, index)[kind === 'addEventListener' ? 1 : 0] ?? -1;
        if(argument >= 0 && context.node(argument).kind === 'ArrowFunction') {
            report(
                context,
                argument,
                kind === 'addEventListener' ? 'addEventListenerArrow' : 'reactHookArrow',
                kind === 'addEventListener' ? [] : ['hookName', kind],
            );
        }
        return;
    }
    const body = last(context, index);
    if(body < 0 || context.node(body).kind !== 'Block' || inherited(context, body, true)) {
        return;
    }
    const text = context.text(index);
    if(!text.includes('\r') && !text.includes('\n') && !text.includes('\u2028') && !text.includes('\u2029')) {
        return;
    }
    if(claimed(context, parent(context, index)) !== '') {
        return;
    }
    report(context, index, 'multilineArrow', []);
}

export function visit(context: RuleContext, index: number): void {
    if(
        ['CallExpression', 'ArrowFunction'].includes(context.node(index).kind) &&
        context.enabled('nexus/consistency-no-multiline-arrow-function')
    ) {
        noMultilineArrowFunction(context, index);
    }
}
