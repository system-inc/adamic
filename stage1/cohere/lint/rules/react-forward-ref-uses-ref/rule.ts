import { SuggestionEdit } from '../../suggestions.a';
import { Suggestion } from '../../suggestions.a';
import { space, noEdits } from '../../context.ts';
import type { RuleContext } from '../../context.ts';
import {
    messageForwardRefAddRefParameter,
    messageForwardRefRemoveWrapper,
    messageForwardRefUsesRef,
} from './messages.ts';
export function forwardRefUsesRef(context: RuleContext, index: number): void {
    if(context.kind(index) !== 'CallExpression') {
        return;
    }
    const callee = context.child(index, 0);
    let name = '';
    if(context.kind(callee) === 'Identifier') {
        name = context.node(callee).text;
    }
    if(context.kind(callee) === 'PropertyAccessExpression') {
        name = context.node(context.property(callee)).text;
    }
    if(context.kind(callee) === 'ElementAccessExpression') {
        const property = context.property(callee);
        if(['StringLiteral', 'NoSubstitutionTemplateLiteral'].includes(context.kind(property))) {
            name = context.node(property).text;
        }
    }
    if(name !== 'forwardRef') {
        return;
    }
    const arguments_ = context.arguments(index);
    const first = arguments_[0] ?? -1;
    if(!['ArrowFunction', 'FunctionExpression'].includes(context.kind(first))) {
        return;
    }
    const parameters = context.children(first, 'Parameter');
    let counted = 0;
    for(const parameter of parameters) {
        if(context.has(parameter, 'DotDotDotToken')) {
            return;
        }
        const named = context.name(parameter);
        if(context.kind(named) === 'Identifier' && context.node(named).text === 'this') {
            continue;
        }
        counted++;
    }
    if(counted !== 1) {
        return;
    }
    const suggestions: Suggestion[] = [];
    let ancestor = context.parent(index);
    while(context.kind(ancestor) === 'ParenthesizedExpression') {
        ancestor = context.parent(ancestor);
    }
    if(
        context.kind(first) !== 'FunctionExpression' ||
        context.kind(context.name(first)) === 'Identifier' ||
        context.kind(ancestor) !== 'ExpressionStatement'
    ) {
        suggestions.push(
            new Suggestion('forwardRefRemoveWrapper', messageForwardRefRemoveWrapper, [
                context.edit(index, context.raw(first)),
            ]),
        );
    }
    let start = context.node(parameters[0] ?? -1).pos;
    while(space(context.source.slice(start, start + 1))) {
        start++;
    }
    let open = start;
    while(open > context.node(first).pos && space(context.source.slice(open - 1, open))) {
        open--;
    }
    let end = context.node(parameters[parameters.length - 1] ?? -1).end;
    if(open > context.node(first).pos && context.source.slice(open - 1, open) === '(') {
        start = open - 1;
        while(end < context.source.length && context.source.slice(end, end + 1) !== ')') {
            end++;
        }
        if(end < context.source.length) {
            end++;
        }
    }
    let text = context.source.slice(start, end);
    if(text.endsWith(')')) {
        text = text.slice(0, -1);
    }
    while(space(text.slice(-1))) {
        text = text.slice(0, -1);
    }
    if(text.endsWith(',')) {
        text = text.slice(0, -1);
    }
    if(!text.startsWith('(')) {
        text = `(${text}`;
    }
    suggestions.push(
        new Suggestion('forwardRefAddRefParameter', messageForwardRefAddRefParameter, [
            new SuggestionEdit(start, end, `${text}, ref)`),
        ]),
    );
    context.reportNode(
        first,
        'react/forward-ref-uses-ref',
        'forwardRefUsesRef',
        messageForwardRefUsesRef,
        noEdits,
        suggestions,
    );
}

export class Rule {
    readonly context: RuleContext;
    constructor(context: RuleContext) {
        this.context = context;
    }
    visit(index: number): void {
        forwardRefUsesRef(this.context, index);
    }
}

export function create(context: RuleContext): Rule {
    return new Rule(context);
}
