import type { RuleContext } from './context.ts';
import { argumentsOf, eventKey, handlerName, emit, first, foreign, last, name, parent } from './shared.ts';

function comparator(context: RuleContext, index: number): boolean {
    for(let cursor = parent(context, index); cursor >= 0; cursor = parent(context, cursor)) {
        if(!['ArrowFunction', 'FunctionExpression'].includes(context.node(cursor).kind)) {
            continue;
        }
        const call = parent(context, cursor);
        if(call < 0 || context.node(call).kind !== 'CallExpression' || argumentsOf(context, call)[0] !== cursor) {
            return false;
        }
        const callee = first(context, call);
        return (
            callee >= 0 &&
            context.node(callee).kind === 'PropertyAccessExpression' &&
            context.node(last(context, callee)).text === 'sort'
        );
    }
    return false;
}
function eventContext(context: RuleContext, index: number): string {
    const declaration = parent(context, index);
    const clause = parent(context, declaration);
    if(
        declaration >= 0 &&
        context.node(declaration).kind === 'VariableDeclaration' &&
        first(context, declaration) === index &&
        clause >= 0 &&
        context.node(clause).kind === 'CatchClause'
    ) {
        return 'error';
    }
    for(let cursor = parent(context, index); cursor >= 0; cursor = parent(context, cursor)) {
        const owner = parent(context, cursor);
        if(owner < 0) {
            break;
        }
        const node = context.node(owner);
        if(
            node.kind === 'PropertyAssignment' &&
            last(context, owner) === cursor &&
            eventKey(context.node(first(context, owner)).text)
        ) {
            return 'event';
        }
        if(node.kind === 'VariableDeclaration' && last(context, owner) === cursor) {
            const binding = name(context, owner);
            if(binding >= 0 && context.node(binding).kind === 'Identifier' && handlerName(context.node(binding).text)) {
                return 'event';
            }
        }
        if(context.node(cursor).kind === 'FunctionDeclaration') {
            const binding = name(context, cursor);
            if(binding >= 0 && handlerName(context.node(binding).text)) {
                return 'event';
            }
        }
    }
    return 'unclear';
}
export function noAmbiguousIdentifier(context: RuleContext, index: number): void {
    const text = context.node(index).text;
    if(text !== '_' && !(text.length === 1 && text >= 'a' && text <= 'z')) {
        return;
    }
    if(foreign(context, index) || ['x', 'y', 'z'].includes(text)) {
        return;
    }
    const rule = 'nexus/consistency-no-ambiguous-identifier';
    if(text === '_') {
        emit(context, index, rule, 'noUnderscore', []);
        return;
    }
    if(text === 'e') {
        const inferred = eventContext(context, index);
        emit(context, index, rule, 'noAmbiguousE', [
            'context',
            inferred === 'error'
                ? ' (appears to be an error)'
                : inferred === 'event'
                  ? ' (appears to be an event)'
                  : ' (context unclear)',
            'suggestedName',
            inferred === 'error' ? 'error' : 'event',
        ]);
        return;
    }
    if((text === 'a' || text === 'b') && comparator(context, index)) {
        return;
    }
    emit(context, index, rule, 'noSingleLetter', ['name', text]);
}

export function visit(context: RuleContext, index: number): void {
    if(context.node(index).kind === 'Identifier' && context.enabled('nexus/consistency-no-ambiguous-identifier')) {
        noAmbiguousIdentifier(context, index);
    }
}
